package server

import (
	"encoding/json"
	"log"
	"net/http"
	"saythis-backend/internal/config"
	"saythis-backend/internal/src/user/handler"
	"saythis-backend/internal/src/user/repository"
	"saythis-backend/internal/src/user/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()

	// -----------------------------
	// User routes
	// -----------------------------
	userRepo := repository.NewPostgresUserRepository(db)    // Postgres implementation
	userUseCase := usecase.NewUserUseCase(userRepo, logger) // Inject repo into usecase
	registerHandler := handler.NewRegisterUserHandler(userUseCase, logger)

	mux.Handle("POST /users/register", registerHandler)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	return mux
}
