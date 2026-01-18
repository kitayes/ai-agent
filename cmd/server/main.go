package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"arcgis-ai-assistant/internal/config"
	"arcgis-ai-assistant/internal/llm"
	"arcgis-ai-assistant/internal/server"
)

func main() {
	log.Println("Starting ArcGIS AI Assistant Server...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	llmClient, err := llm.NewClient(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	srv := server.New(cfg.ServerPort, llmClient)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Printf("Server running on http://localhost:%s", cfg.ServerPort)
	log.Println("Endpoints:")
	log.Println("  - POST /api/echo - Test connectivity")
	log.Println("  - POST /api/generate - Generate ArcPy code")
	log.Println("  - GET /health - Health check")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
