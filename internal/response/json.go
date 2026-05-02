package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse is the standard error envelope returned to API consumers.
// Every error response looks like: {"error": "some message"}
type ErrorResponse struct {
	Error string `json:"error"`
}

// JSON serialises data as JSON, sets the Content-Type header, and writes the status code.
// The error from Encode is logged but cannot be returned — headers are already sent.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// Error writes a structured JSON error response.
// Use this everywhere instead of http.Error so clients always get JSON.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}
