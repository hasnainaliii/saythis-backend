package server

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"saythis-backend/internal/config"
	"saythis-backend/internal/middleware"
	"saythis-backend/internal/user/handler"
	"saythis-backend/internal/user/repository"
	"saythis-backend/internal/user/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) http.Handler {

	// ************************
	// ── Dependency wiring ───
	// ************************

	userRepo := repository.NewPostgresUserRepo(db)
	userUseCase := usecase.NewUserUseCase(userRepo)
	createHandler := handler.NewCreateUserHandler(userUseCase)

	// ***************
	// ── Routes ───
	// ***************

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/users", createHandler)

	// ── Middleware ────────────────────────────────────────────────────────────
	// Rate limiter: 10 requests/second sustained, bursts up to 20.
	// Applied globally here; for tighter per-route limits (e.g. registration),
	// wrap individual handlers instead.
	rateLimiter := middleware.NewRateLimiter(rate.Every(time.Second), 20)

	corsMiddleware := middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
	})

	// Chain execution order (left → right on an incoming request):
	//   RequestID → CORS → RateLimit → mux
	//
	// RequestID is outermost so every log line emitted by any middleware or
	// handler can include the request ID from the context.
	return middleware.Chain(mux,
		middleware.RequestID,
		corsMiddleware,
		rateLimiter.Limit,
	)
}
