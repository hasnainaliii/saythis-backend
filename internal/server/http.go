package server

import (
	"encoding/json"
	"log"
	"net/http"
	"saythis-backend/internal/config"
	AuthHandler "saythis-backend/internal/src/auth/handler"
	AuthRepository "saythis-backend/internal/src/auth/repository"
	AuthUseCase "saythis-backend/internal/src/auth/usecase"
	"saythis-backend/internal/src/user/repository"
	"saythis-backend/internal/src/user/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config, logger *log.Logger) http.Handler {
	logger.Println("[BOOT] Initializing NewRouter and dependencies...")
	mux := http.NewServeMux()

	// -----------------------------
	// Database Dependencies
	// -----------------------------

	logger.Println("[BOOT] Setting up User layer...")
	userRepo := repository.NewPostgresUserRepository(db, logger)
	userUseCase := usecase.NewUserUseCase(userRepo, logger)

	logger.Println("[BOOT] Setting up Auth layer...")
	authRepo := AuthRepository.NewPostgresAuthRepository(db, logger)
	authUseCase := AuthUseCase.NewRegisterAuthUseCase(authRepo, logger)

	logger.Println("[BOOT] Setting up Orchestrator...")
	registerOrchestrator := AuthUseCase.NewRegisterOrchestrator(
		db,
		userUseCase,
		authUseCase,
		userRepo,
		authRepo,
		logger,
	)

	logger.Println("[BOOT] Setting up Auth Handlers...")
	registerHandler := AuthHandler.NewRegisterHandler(registerOrchestrator, logger)

	// -----------------------------
	// Routes---USER&AUTH Routes
	// -----------------------------

	logger.Println("[BOOT] Registering Routes...")
	mux.Handle("POST /auth/register", registerHandler)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		logger.Println("[HEALTH] GET /health check")
		if err := db.Ping(r.Context()); err != nil {
			logger.Printf("[HEALTH] [ERROR] Database ping failed: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	logger.Println("[BOOT] Router initialization complete ✅")
	return mux
}
