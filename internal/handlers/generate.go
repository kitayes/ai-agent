package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"qgis-ai-assistant/internal/llm"
	"qgis-ai-assistant/internal/models"
	"qgis-ai-assistant/internal/validator"
)

type GenerateHandler struct {
	llmClient *llm.Client
	validator *validator.Validator
}

func NewGenerateHandler(llmClient *llm.Client) *GenerateHandler {
	return &GenerateHandler{
		llmClient: llmClient,
		validator: validator.NewValidator(),
	}
}

func (h *GenerateHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("=== NEW REQUEST ===")
	log.Printf("Prompt: %s", req.Prompt)
	if req.Context != nil {
		log.Printf("Project: %s", req.Context.Project.Name)
		log.Printf("Layers: %d", len(req.Context.Layers))
	} else {
		log.Printf("Context: Not provided (legacy mode)")
	}

	code, explanation, usedLayers, warnings, err := h.llmClient.GenerateCodeWithContext(req.Prompt, req.Context)
	if err != nil {
		log.Printf("Error generating code: %v", err)
		resp := models.GenerateResponse{
			Error: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// SECURITY: Validate generated code
	validationResult := h.validator.ValidateCode(code)

	if !validationResult.IsValid {
		log.Printf("⚠️ Generated code failed validation!")
		log.Printf("Validation errors: %v", validationResult.Errors)

		// Add validation errors to warnings
		for _, err := range validationResult.Errors {
			warnings = append(warnings, "🔒 SECURITY: "+err)
		}

		// If code is dangerous (score < 50), reject it
		if validationResult.Score < 50 {
			resp := models.GenerateResponse{
				Error:    "Сгенерированный код не прошел проверку безопасности. Попробуйте переформулировать запрос.",
				Warnings: warnings,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	// Add validation warnings
	for _, warning := range validationResult.Warnings {
		warnings = append(warnings, warning)
	}

	log.Printf("Code validation: Score=%d", validationResult.Score)

	resp := models.GenerateResponse{
		Code:        code,
		Explanation: explanation,
		UsedLayers:  usedLayers,
		Warnings:    warnings,
	}

	log.Printf("Code generated successfully")
	log.Printf("Used layers: %v", usedLayers)
	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
