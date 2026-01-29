package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				zap.S().Errorw("Panic recovered",
					"error", err,
					"path", r.URL.Path,
					"method", r.Method,
					"stack", string(debug.Stack()),
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"error": map[string]string{
						"code":    "INTERNAL_ERROR",
						"message": "An unexpected error occurred",
					},
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
