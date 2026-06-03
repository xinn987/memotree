package main

import (
	"log"
	"net/http"

	"memotree/server/api/internal/config"
	"memotree/server/api/internal/httpapi"
)

func main() {
	cfg := config.Load()
	server := &http.Server{
		Addr:    cfg.APIAddr,
		Handler: httpapi.NewRouter(cfg),
	}

	log.Printf("memotree api listening on %s", cfg.APIAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("api server stopped: %v", err)
	}
}
