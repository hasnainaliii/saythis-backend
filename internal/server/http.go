package server

import (
	"net/http"
	"saythis-backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool, cfg *config.Config) http.Handler {

	mux := http.NewServeMux()

	return mux
}
