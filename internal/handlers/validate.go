package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"qgis-ai-assistant/internal/validator"
)

type ValidateRequest struct {
	Code string `json:"code"`
}

type ValidateResponse struct {
	validator.ValidationResult
}

func ValidateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	v := validator.NewValidator()
	result := v.ValidateCode(req.Code)

	log.Printf("Code validation: Valid=%v, Score=%d, Errors=%d, Warnings=%d",
		result.IsValid, result.Score, len(result.Errors), len(result.Warnings))

	resp := ValidateResponse{
		ValidationResult: result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
