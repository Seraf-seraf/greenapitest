package main

import (
	"log"

	"greenapi-form/internal/config"
	"greenapi-form/internal/httpserver"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatal(err)
	}

	server, err := httpserver.New(cfg.Server.ListenAddress(), cfg.GreenAPI.BaseURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server started on http://localhost%s", server.Addr())
	log.Printf("GREEN-API endpoint: %s", cfg.GreenAPI.BaseURL)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
