package server

import (
	"net/http"

	"saythis-backend/internal/config"
	"saythis-backend/internal/middleware"
	authHandler "saythis-backend/internal/src/auth/handler"
	authRepository "saythis-backend/internal/src/auth/repository"
	authService "saythis-backend/internal/src/auth/service"
	authUseCase "saythis-backend/internal/src/auth/usecase"
	userHandler "saythis-backend/internal/src/user/handler"
	userRepository "saythis-backend/internal/src/user/repository"
	userUseCase "saythis-backend/internal/src/user/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	tokenService := authService.NewTokenService(cfg)

	userRepo := userRepository.NewPostgresUserRepository(db)
	authRepo := authRepository.NewPostgresAuthRepository(db)

	userUC := userUseCase.NewUserUseCase(userRepo)
	registerAuthUC := authUseCase.NewRegisterAuthUseCase(authRepo)
	loginUC := authUseCase.NewLoginAuthUseCase(authRepo, tokenService)

	registerOrchestrator := authUseCase.NewRegisterOrchestrator(
		db,
		userUC,
		registerAuthUC,
		userRepo,
		authRepo,
	)

	registerHandler := authHandler.NewRegisterHandler(registerOrchestrator)
	loginHandler := authHandler.NewLoginHandler(loginUC)
	refreshHandler := authHandler.NewRefreshHandler(tokenService)
	meHandler := userHandler.NewMeHandler()
	deleteMeHandler := userHandler.NewDeleteMeHandler(userUC)

	authMW := middleware.AuthMiddleware(tokenService)

	mux := http.NewServeMux()

	// Public Routes
	mux.Handle("POST /api/v1/auth/register", registerHandler)
	mux.Handle("POST /api/v1/auth/login", loginHandler)
	mux.Handle("POST /api/v1/auth/refresh", refreshHandler)

	// Protected Routes
	mux.Handle("GET /api/v1/user/me", authMW(meHandler))
	mux.Handle("DELETE /api/v1/users/me", authMW(deleteMeHandler))

	zap.S().Info("✅ Router initialized")

	return applyMiddleware(mux)
}

func applyMiddleware(handler http.Handler) http.Handler {
	return middleware.RecoveryMiddleware(
		middleware.SecurityMiddleware(handler),
	)
}
