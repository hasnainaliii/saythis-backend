package helper

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// errorEnvelope is the standard error body returned to API consumers.
// Every error response looks like: {"error": "some message"}
type errorEnvelope struct {
	Error string `json:"error"`
}

// JSON serialises data as JSON, sets Content-Type, and writes the status code.
// The encode error is logged but not returned — headers are already sent.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// Error writes a structured {"error": "..."} JSON response.
// Use this everywhere instead of http.Error so clients always receive JSON.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, errorEnvelope{Error: message})
}

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
