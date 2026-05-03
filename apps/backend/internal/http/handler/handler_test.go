package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/relations4u/worldweathernews/apps/backend/internal/http/handler"
)

func TestHealth_ReturnsOKWithVersion(t *testing.T) {
	deps := handler.Deps{StartedAt: time.Now()}

	r := chi.NewRouter()
	r.Get("/health", handler.Health(deps))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}

	var body struct {
		Status  string `json:"status"`
		Version struct {
			Version string `json:"version"`
		} `json:"version"`
		Uptime string `json:"uptime"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("status: got %q want ok", body.Status)
	}
	if body.Version.Version == "" {
		t.Errorf("version field missing")
	}
	if body.Uptime == "" {
		t.Errorf("uptime field missing")
	}
}

func TestPing_ReturnsPongWithTraceID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Get("/ping", handler.Ping())

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}

	var body struct {
		Message string `json:"message"`
		TraceID string `json:"traceId"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Message != "pong" {
		t.Errorf("message: got %q want pong", body.Message)
	}
	if body.TraceID == "" {
		t.Errorf("trace id missing — RequestID middleware not applied?")
	}
}
