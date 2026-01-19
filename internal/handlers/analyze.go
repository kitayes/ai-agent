package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"qgis-ai-assistant/internal/llm"
	"qgis-ai-assistant/internal/models"
)

type AnalyzeScreenshotRequest struct {
	ImageBase64 string          `json:"imageBase64,omitempty"`
	ImagePath   string          `json:"imagePath,omitempty"`
	Prompt      string          `json:"prompt"`
	Context     *models.Context `json:"context,omitempty"`
}

type AnalyzeScreenshotResponse struct {
	Analysis         string   `json:"analysis"`
	SuggestedActions []string `json:"suggestedActions"`
	GeneratedCode    string   `json:"generatedCode,omitempty"`
	Explanation      string   `json:"explanation"`
	Warnings         []string `json:"warnings,omitempty"`
	Error            string   `json:"error,omitempty"`
}

type AnalyzeHandler struct {
	llmClient *llm.Client
}

func NewAnalyzeHandler(llmClient *llm.Client) *AnalyzeHandler {
	return &AnalyzeHandler{
		llmClient: llmClient,
	}
}

func (h *AnalyzeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeScreenshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("=== SCREENSHOT ANALYSIS REQUEST ===")
	log.Printf("Prompt: %s", req.Prompt)
	log.Printf("Has image: %v", req.ImageBase64 != "" || req.ImagePath != "")

	// Get image bytes
	var imageBytes []byte
	var err error

	if req.ImageBase64 != "" {
		imageBytes, err = base64.StdEncoding.DecodeString(req.ImageBase64)
		if err != nil {
			log.Printf("Error decoding base64 image: %v", err)
			resp := AnalyzeScreenshotResponse{
				Error: "Invalid base64 image data",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
	} else if req.ImagePath != "" {
		file, err := os.Open(req.ImagePath)
		if err != nil {
			log.Printf("Error opening image file: %v", err)
			resp := AnalyzeScreenshotResponse{
				Error: fmt.Sprintf("Could not open image file: %v", err),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}
		defer file.Close()

		imageBytes, err = io.ReadAll(file)
		if err != nil {
			log.Printf("Error reading image file: %v", err)
			resp := AnalyzeScreenshotResponse{
				Error: "Could not read image file",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(resp)
			return
		}
	} else {
		resp := AnalyzeScreenshotResponse{
			Error: "No image provided (imageBase64 or imagePath required)",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	log.Printf("Image size: %d bytes", len(imageBytes))

	// Analyze with Gemini Vision
	analysis, suggestedActions, code, explanation, warnings, err := h.llmClient.AnalyzeMapScreenshot(
		imageBytes,
		req.Prompt,
		req.Context,
	)

	if err != nil {
		log.Printf("Error analyzing screenshot: %v", err)
		resp := AnalyzeScreenshotResponse{
			Error: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := AnalyzeScreenshotResponse{
		Analysis:         analysis,
		SuggestedActions: suggestedActions,
		GeneratedCode:    code,
		Explanation:      explanation,
		Warnings:         warnings,
	}

	log.Printf("Screenshot analysis completed successfully")
	log.Printf("Suggested actions: %d", len(suggestedActions))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
