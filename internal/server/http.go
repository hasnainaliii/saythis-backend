package server

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"saythis-backend/internal/config"
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
	// Routes
	// *******************

	mux := http.NewServeMux()

	// Public auth routes
	mux.Handle("POST /api/v1/auth/register", registerHandler)
	mux.Handle("POST /api/v1/auth/login", loginHandler)
	mux.Handle("POST /api/v1/auth/refresh", refreshHandler)
	mux.Handle("POST /api/v1/auth/verify-email", verifyEmailHandler)
	mux.Handle("POST /api/v1/auth/forgot-password", forgotPasswordHandler)
	mux.Handle("POST /api/v1/auth/reset-password", resetPasswordHandler)

	// Protected user routes
	mux.Handle("GET /api/v1/users/me", bearerAuth(getProfileHandler))
	mux.Handle("PATCH /api/v1/users/me", bearerAuth(updateProfileHandler))
	mux.Handle("PATCH /api/v1/users/me/avatar", bearerAuth(updateAvatarHandler))
	mux.Handle("DELETE /api/v1/users/me", bearerAuth(deleteAccountHandler))

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
