package config

import (
	"errors"
	"fmt"
	"net"
	neturl "net/url"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath = "configs/config.yml"

	EnvConfigPath      = "APP_CONFIG_PATH"
	EnvServerHost      = "APP_SERVER_HOST"
	EnvServerPort      = "APP_SERVER_PORT"
	EnvGreenAPIBaseURL = "GREEN_API_BASE_URL"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	GreenAPI GreenAPIConfig `yaml:"green_api"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type GreenAPIConfig struct {
	BaseURL string `yaml:"base_url"`
}

func Load(path string) (Config, error) {
	var cfg Config

	resolvedPath := strings.TrimSpace(path)
	if resolvedPath == "" {
		resolvedPath = strings.TrimSpace(os.Getenv(EnvConfigPath))
	}

	if resolvedPath == "" {
		resolvedPath = DefaultConfigPath
	}

	raw, err := os.ReadFile(resolvedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("config file %q not found", resolvedPath)
		}
		return Config{}, fmt.Errorf("read config %q: %w", resolvedPath, err)
	}

	if len(strings.TrimSpace(string(raw))) == 0 {
		return Config{}, fmt.Errorf("config file %q is empty", resolvedPath)
	}

	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", resolvedPath, err)
	}

	applyEnvOverrides(&cfg)
	if err := validate(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if value := strings.TrimSpace(os.Getenv(EnvServerHost)); value != "" {
		cfg.Server.Host = value
	}

	if value := strings.TrimSpace(os.Getenv(EnvServerPort)); value != "" {
		port, err := strconv.Atoi(value)
		if err != nil {
			cfg.Server.Port = -1
		} else {
			cfg.Server.Port = port
		}
	}

	if value := strings.TrimSpace(os.Getenv(EnvGreenAPIBaseURL)); value != "" {
		cfg.GreenAPI.BaseURL = value
	}
}

func validate(cfg *Config) error {
	cfg.Server.Host = strings.TrimSpace(cfg.Server.Host)
	if cfg.Server.Host == "" {
		return errors.New("config: server.host is required")
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return errors.New("config: server.port must be in range 1..65535")
	}

	cfg.GreenAPI.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.GreenAPI.BaseURL), "/")
	if cfg.GreenAPI.BaseURL == "" {
		return errors.New("config: green_api.base_url is required")
	}

	parsed, err := neturl.Parse(cfg.GreenAPI.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("config: invalid green_api.base_url %q", cfg.GreenAPI.BaseURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("config: unsupported green_api.base_url scheme %q", parsed.Scheme)
	}

	return nil
}

func (s ServerConfig) ListenAddress() string {
	host := strings.TrimSpace(s.Host)
	if host == "" || host == "0.0.0.0" {
		return fmt.Sprintf(":%d", s.Port)
	}

	return net.JoinHostPort(host, strconv.Itoa(s.Port))
}
