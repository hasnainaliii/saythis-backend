package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

// RequestIDKey is the context key used to store and retrieve the request ID.
const RequestIDKey contextKey = "request_id"

// RequestID is a middleware that ensures every request carries a unique ID.
//
// If the incoming request already has an X-Request-ID header, that value is
// reused (useful when a gateway or client generates the ID). Otherwise a new
// UUID v4 is generated. The ID is:
//   - stored in the request context (retrieve with GetRequestID)
//   - echoed back in the X-Request-ID response header
//
// This makes it trivial to correlate a client error report with a server log line.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID stored in the context.
// Returns an empty string if no ID is present.
func GetRequestID(ctx context.Context) string {
	id, _ := ctx.Value(RequestIDKey).(string)
	return id
}
