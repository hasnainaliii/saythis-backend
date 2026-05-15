package server

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"saythis-backend/internal/config"
	"saythis-backend/internal/health"
	"saythis-backend/internal/middleware"
	"saythis-backend/internal/src/auth"
	authhandler "saythis-backend/internal/src/auth/handler"
	authrepo "saythis-backend/internal/src/auth/repository"
	authusecase "saythis-backend/internal/src/auth/usecase"
	userhandler "saythis-backend/internal/src/user/handler"
	userrepo "saythis-backend/internal/src/user/repository"
	userusecase "saythis-backend/internal/src/user/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	startTime := time.Now()

	// *******************
	// Shared infra
	// *******************

	jwtCfg := auth.NewJWTConfig(cfg)
	bearerAuth := auth.BearerAuth(jwtCfg)

	// *******************
	// Repositories
	// *******************

	userRepo := userrepo.NewPostgresUserRepo(db)
	authRepo := authrepo.NewPostgresAuthRepo(db)

	// *******************
	// Auth
	// *******************

	emailSender := auth.NewResendClient(cfg.ResendAPIKey, "auth@hasn.me")
	authUseCase := authusecase.NewAuthUseCase(authRepo, userRepo, jwtCfg, emailSender, cfg.FrontendURL)

	registerHandler := authhandler.NewRegisterHandler(authUseCase)
	loginHandler := authhandler.NewLoginHandler(authUseCase)
	refreshHandler := authhandler.NewRefreshHandler(authUseCase)
	verifyEmailHandler := authhandler.NewVerifyEmailHandler(authUseCase)
	forgotPasswordHandler := authhandler.NewForgotPasswordHandler(authUseCase)
	resetPasswordHandler := authhandler.NewResetPasswordHandler(authUseCase)

	// *******************
	// User (protected)
	// *******************

	cloudinaryUploader := userusecase.MustNewCloudinaryUploader(cfg.CloudinaryURL)
	userUseCase := userusecase.NewUserUseCase(userRepo, authRepo, cloudinaryUploader)
	getProfileHandler := userhandler.NewGetProfileHandler(userUseCase)
	deleteAccountHandler := userhandler.NewDeleteAccountHandler(userUseCase)
	updateProfileHandler := userhandler.NewUpdateProfileHandler(userUseCase)
	updateAvatarHandler := userhandler.NewUpdateAvatarHandler(userUseCase)

	// *******************
	// API routes (rate-limited)
	// *******************

	apiMux := http.NewServeMux()

	// Public auth routes
	apiMux.Handle("POST /api/v1/auth/register", registerHandler)
	apiMux.Handle("POST /api/v1/auth/login", loginHandler)
	apiMux.Handle("POST /api/v1/auth/refresh", refreshHandler)
	apiMux.Handle("POST /api/v1/auth/verify-email", verifyEmailHandler)
	apiMux.Handle("POST /api/v1/auth/forgot-password", forgotPasswordHandler)
	apiMux.Handle("POST /api/v1/auth/reset-password", resetPasswordHandler)

	// Protected user routes
	apiMux.Handle("GET /api/v1/users/me", bearerAuth(getProfileHandler))
	apiMux.Handle("PATCH /api/v1/users/me", bearerAuth(updateProfileHandler))
	apiMux.Handle("PATCH /api/v1/users/me/avatar", bearerAuth(updateAvatarHandler))
	apiMux.Handle("DELETE /api/v1/users/me", bearerAuth(deleteAccountHandler))

	// *******************
	// Middleware
	// *******************

	rateLimiter := middleware.NewRateLimiter(rate.Every(time.Second), 20)

	corsMiddleware := middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Request-ID"},
	})

	// *******************
	// Top-level mux
	// Health is registered here — outside the rate-limiter — so uptime monitors
	// (which poll every 30 s) never consume API rate-limit budget.
	// *******************

	mux := http.NewServeMux()
	mux.Handle("GET /health", health.NewHandler(db, cfg.AppEnv, startTime))
	mux.Handle("/", middleware.Chain(apiMux,
		middleware.RequestID,
		corsMiddleware,
		rateLimiter.Limit,
	))

	return mux
}
