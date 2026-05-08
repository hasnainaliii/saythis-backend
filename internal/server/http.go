package server

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"saythis-backend/internal/auth"
	authhandler "saythis-backend/internal/auth/handler"
	authrepo "saythis-backend/internal/auth/repository"
	authusecase "saythis-backend/internal/auth/usecase"
	"saythis-backend/internal/config"
	"saythis-backend/internal/middleware"
	userrepo "saythis-backend/internal/user/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) http.Handler {

	// *******************
	// User
	// *******************

	userRepo := userrepo.NewPostgresUserRepo(db)

	// *******************
	// Auth
	// *******************

	authRepo := authrepo.NewPostgresAuthRepo(db)
	authUseCase := authusecase.NewAuthUseCase(authRepo, userRepo, auth.NewJWTConfig(cfg))
	registerHandler := authhandler.NewRegisterHandler(authUseCase)
	loginHandler := authhandler.NewLoginHandler(authUseCase)
	refreshHandler := authhandler.NewRefreshHandler(authUseCase)

	// *******************
	// Routes
	// *******************

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/auth/register", registerHandler)
	mux.Handle("POST /api/v1/auth/login", loginHandler)
	mux.Handle("POST /api/v1/auth/refresh", refreshHandler)

	// *******************
	// Middleware
	// *******************

	rateLimiter := middleware.NewRateLimiter(rate.Every(time.Second), 20)

	corsMiddleware := middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
	})

	return middleware.Chain(mux,
		middleware.RequestID,
		corsMiddleware,
		rateLimiter.Limit,
	)
}
