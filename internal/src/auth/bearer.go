package auth

import (
	"context"
	"net/http"
	"strings"

	"saythis-backend/internal/helper"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

// BearerAuth validates the JWT in the Authorization header.
// On success it injects the parsed Claims into the request context.
// Attach this to any route that requires a logged-in user.
//
// Expected header format:  Authorization: Bearer <access_token>
func BearerAuth(cfg JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				helper.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				helper.Error(w, http.StatusUnauthorized, "authorization header must be: Bearer <token>")
				return
			}

			claims, err := ValidateAccessToken(cfg, parts[1])
			if err != nil {
				helper.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext retrieves the validated JWT claims from the request context.
// Returns nil, false if the request was not authenticated (route not protected).
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}
