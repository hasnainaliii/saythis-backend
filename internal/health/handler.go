package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// DBPinger is satisfied by *pgxpool.Pool and makes the handler unit-testable.
type DBPinger interface {
	Ping(ctx context.Context) error
}

type status string

const (
	statusHealthy   status = "healthy"
	statusUnhealthy status = "unhealthy"
)

type checkResult struct {
	Status    status `json:"status"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}

type response struct {
	Status      status                 `json:"status"`
	Uptime      string                 `json:"uptime"`
	Timestamp   time.Time              `json:"timestamp"`
	Environment string                 `json:"environment"`
	Checks      map[string]checkResult `json:"checks"`
}

// Handler reports the health of the service and its dependencies.
// It is intentionally registered outside the rate-limiter middleware chain.
type Handler struct {
	db        DBPinger
	env       string
	startTime time.Time
}

func NewHandler(db DBPinger, env string, startTime time.Time) *Handler {
	return &Handler{db: db, env: env, startTime: startTime}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// ── Database check ────────────────────────────────────────────────────────
	dbResult := checkResult{Status: statusHealthy}
	t := time.Now()
	if err := h.db.Ping(ctx); err != nil {
		dbResult.Status = statusUnhealthy
		dbResult.Error = err.Error()
	} else {
		dbResult.LatencyMs = time.Since(t).Milliseconds()
	}

	// ── Overall status ────────────────────────────────────────────────────────
	overall := statusHealthy
	httpCode := http.StatusOK
	if dbResult.Status == statusUnhealthy {
		overall = statusUnhealthy
		httpCode = http.StatusServiceUnavailable
	}

	resp := response{
		Status:      overall,
		Uptime:      time.Since(h.startTime).Round(time.Second).String(),
		Timestamp:   time.Now().UTC(),
		Environment: h.env,
		Checks: map[string]checkResult{
			"database": dbResult,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(httpCode)
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}
