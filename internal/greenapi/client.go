package greenapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var methods = map[string]string{
	"getSettings":      http.MethodGet,
	"getStateInstance": http.MethodGet,
	"sendMessage":      http.MethodPost,
	"sendFileByUrl":    http.MethodPost,
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) (*Client, error) {
	normalizedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if normalizedBaseURL == "" {
		return nil, errors.New("green api base url is required")
	}

	parsed, err := url.Parse(normalizedBaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid green api base url %q", normalizedBaseURL)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported green api base url scheme %q", parsed.Scheme)
	}

	return &Client{
		baseURL: normalizedBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) Call(ctx context.Context, idInstance, apiTokenInstance, method string, payload map[string]any) (json.RawMessage, error) {
	httpMethod, ok := methods[method]
	if !ok {
		return nil, fmt.Errorf("unsupported method %q", method)
	}

	endpoint := fmt.Sprintf(
		"%s/waInstance%s/%s/%s",
		c.baseURL,
		url.PathEscape(idInstance),
		method,
		url.PathEscape(apiTokenInstance),
	)

	var body io.Reader = http.NoBody
	if httpMethod == http.MethodPost {
		if payload == nil {
			payload = map[string]any{}
		}

		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	if httpMethod == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		msg := strings.TrimSpace(string(rawBody))
		if msg == "" {
			msg = "empty response"
		}
		return nil, fmt.Errorf("green-api status %d: %s", resp.StatusCode, msg)
	}

	if len(rawBody) == 0 {
		return json.RawMessage("null"), nil
	}

	if !json.Valid(rawBody) {
		wrapped, err := json.Marshal(map[string]any{
			"raw": string(rawBody),
		})
		if err != nil {
			return nil, fmt.Errorf("wrap non-json response: %w", err)
		}
		return wrapped, nil
	}

	return rawBody, nil
}
