package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"qgis-ai-assistant/internal/llm"
	"qgis-ai-assistant/internal/models"
)

type RegenerateHandler struct {
	llmClient *llm.Client
}

func NewRegenerateHandler(llmClient *llm.Client) *RegenerateHandler {
	return &RegenerateHandler{
		llmClient: llmClient,
	}
}

func (h *RegenerateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RegenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("=== REGENERATE REQUEST ===")
	log.Printf("Original prompt: %s", req.OriginalPrompt)
	log.Printf("Attempt: %d", req.Attempt)
	log.Printf("Error: %s", req.ErrorMessage)

	if req.Attempt > 3 {
		resp := models.GenerateResponse{
			Error: "Maximum retry attempts exceeded",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	code, explanation, usedLayers, warnings, err := h.llmClient.RegenerateCode(
		req.OriginalPrompt,
		req.FailedCode,
		req.ErrorMessage,
		req.Context,
		req.Attempt,
	)

	if err != nil {
		log.Printf("Error regenerating code: %v", err)
		resp := models.GenerateResponse{
			Error: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := models.GenerateResponse{
		Code:        code,
		Explanation: explanation,
		UsedLayers:  usedLayers,
		Warnings:    warnings,
	}

	log.Printf("Code regenerated successfully (attempt %d)", req.Attempt)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
