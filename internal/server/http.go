package server

import (
	"net/http"
	"saythis-backend/internal/config"
	"saythis-backend/internal/middleware"
	AuthHandler "saythis-backend/internal/src/auth/handler"
	AuthRepository "saythis-backend/internal/src/auth/repository"
	AuthUseCase "saythis-backend/internal/src/auth/usecase"
	"saythis-backend/internal/src/user/repository"
	"saythis-backend/internal/src/user/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// -----------------------------
	// Database Dependencies
	// -----------------------------

	userRepo := repository.NewPostgresUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo)

	authRepo := AuthRepository.NewPostgresAuthRepository(db)
	authUseCase := AuthUseCase.NewRegisterAuthUseCase(authRepo)

	registerOrchestrator := AuthUseCase.NewRegisterOrchestrator(
		db,
		userUseCase,
		authUseCase,
		userRepo,
		authRepo,
	)

	registerHandler := AuthHandler.NewRegisterHandler(registerOrchestrator)

	// -----------------------------
	// Routes---USER&AUTH Routes
	// -----------------------------

	mux.Handle("POST /auth/register", registerHandler)

	zap.S().Info("✅ Router initialized successfully")

	// -----------------------------
	// Apply Middleware Chain
	// -----------------------------
	// Order: Recovery -> Security -> Router

	handler := middleware.RecoveryMiddleware(
		middleware.SecurityMiddleware(mux),
	)

	return handler
}
