package handlers

import (
	"encoding/json"
	"net/http"
)

type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Response string `json:"response"`
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EchoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resp := EchoResponse{
		Response: req.Message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
