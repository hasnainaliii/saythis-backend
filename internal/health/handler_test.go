package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"saythis-backend/internal/health"
)

// ── mock ─────────────────────────────────────────────────────────────────────

type mockDB struct{ err error }

func (m *mockDB) Ping(_ context.Context) error { return m.err }

// ── helpers ──────────────────────────────────────────────────────────────────

func newHandler(dbErr error) *health.Handler {
	return health.NewHandler(&mockDB{err: dbErr}, "test", time.Now())
}

func call(h *health.Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return body
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestHealth_Healthy(t *testing.T) {
	rec := call(newHandler(nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}

	body := decode(t, rec)

	if body["status"] != "healthy" {
		t.Errorf("want status=healthy, got %v", body["status"])
	}

	checks := body["checks"].(map[string]any)
	db := checks["database"].(map[string]any)
	if db["status"] != "healthy" {
		t.Errorf("want database.status=healthy, got %v", db["status"])
	}
}

func TestHealth_DBUnhealthy(t *testing.T) {
	rec := call(newHandler(errors.New("connection refused")))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", rec.Code)
	}

	body := decode(t, rec)

	if body["status"] != "unhealthy" {
		t.Errorf("want status=unhealthy, got %v", body["status"])
	}

	checks := body["checks"].(map[string]any)
	db := checks["database"].(map[string]any)
	if db["status"] != "unhealthy" {
		t.Errorf("want database.status=unhealthy, got %v", db["status"])
	}
	if db["error"] == "" {
		t.Error("want non-empty database.error")
	}
}

func TestHealth_ContentType(t *testing.T) {
	rec := call(newHandler(nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("want Content-Type=application/json, got %q", ct)
	}
}

func TestHealth_CacheControl(t *testing.T) {
	rec := call(newHandler(nil))

	cc := rec.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("want Cache-Control=no-store, got %q", cc)
	}
}

func TestHealth_UptimePresent(t *testing.T) {
	h := health.NewHandler(&mockDB{}, "test", time.Now().Add(-5*time.Minute))
	rec := call(h)

	body := decode(t, rec)
	uptime, ok := body["uptime"].(string)
	if !ok || uptime == "" {
		t.Errorf("want non-empty uptime, got %v", body["uptime"])
	}
}

func TestHealth_EnvironmentPresent(t *testing.T) {
	h := health.NewHandler(&mockDB{}, "production", time.Now())
	rec := call(h)

	body := decode(t, rec)
	if body["environment"] != "production" {
		t.Errorf("want environment=production, got %v", body["environment"])
	}
}

func TestHealth_TimestampPresent(t *testing.T) {
	rec := call(newHandler(nil))
	body := decode(t, rec)

	if _, ok := body["timestamp"]; !ok {
		t.Error("want timestamp field in response")
	}
}
