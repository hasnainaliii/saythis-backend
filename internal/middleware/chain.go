package middleware

import "net/http"

// Chain wraps handler h with the given middlewares in declaration order.
// The first middleware in the list is the outermost (executes first on the
// way in, last on the way out), matching the intuitive reading order.
//
// Example:
//
//	Chain(mux, RequestID, corsMiddleware, rateLimiter.Limit)
//
// Execution order on an incoming request: RequestID → CORS → RateLimit → mux
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Apply in reverse so that middlewares[0] ends up as the outermost wrapper.
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
