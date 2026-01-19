package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"qgis-ai-assistant/internal/handlers"
	"qgis-ai-assistant/internal/llm"
)

type Server struct {
	httpServer *http.Server
	llmClient  *llm.Client
}

func New(port string, llmClient *llm.Client) *Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/echo", corsMiddleware(handlers.EchoHandler))

	generateHandler := handlers.NewGenerateHandler(llmClient)
	mux.HandleFunc("/api/generate", corsMiddleware(generateHandler.Handle))

	regenerateHandler := handlers.NewRegenerateHandler(llmClient)
	mux.HandleFunc("/api/regenerate", corsMiddleware(regenerateHandler.Handle))

	mux.HandleFunc("/api/validate", corsMiddleware(handlers.ValidateHandler))

	analyzeHandler := handlers.NewAnalyzeHandler(llmClient)
	mux.HandleFunc("/api/analyze-screenshot", corsMiddleware(analyzeHandler.Handle))

	// Data fetching endpoints
	dataSearchHandler := handlers.NewDataSearchHandler(llmClient)
	mux.HandleFunc("/api/data/search", corsMiddleware(dataSearchHandler.Handle))

	dataFetchHandler := handlers.NewDataFetchHandler("./downloads")
	mux.HandleFunc("/api/data/fetch", corsMiddleware(dataFetchHandler.Handle))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		llmClient: llmClient,
	}
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
