package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func CorsMiddleware(next http.Handler) http.Handler {
	zap.S().Warn("⚠️  CORS middleware using wildcard (*) - configure properly for production")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
