package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/relations4u/worldweathernews/apps/backend/internal/api"
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

// apiTestServer baut den Strict-Server gegen den APIHandler, montiert auf einen
// Chi-Router mit RequestID-Middleware. Genug für End-to-End-Test eines Endpoints
// ohne den vollen Backend-Boot. Pool=nil triggert den dev-Fallback-Pfad im
// Handler (leere Listen, 404 für Detail).
func apiTestServer() http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	apiHandler := handler.NewAPIHandler(nil)
	strict := api.NewStrictHandler(apiHandler, nil)
	api.HandlerFromMux(strict, r)
	return r
}

func TestPing_ReturnsPongWithTraceID(t *testing.T) {
	srv := apiTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}

	var body api.PingResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Message != "pong" {
		t.Errorf("message: got %q want pong", body.Message)
	}
	if body.TraceId == "" {
		t.Errorf("trace id missing — RequestID middleware not applied?")
	}
}

func TestListLocations_NilPoolReturnsEmptyListWithAttribution(t *testing.T) {
	srv := apiTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/locations", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want 200", rec.Code)
	}

	var body struct {
		Results     []api.Location `json:"results"`
		Attribution string         `json:"attribution"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Results == nil {
		t.Errorf("results field must be present (even if empty)")
	}
	if body.Attribution == "" {
		t.Errorf("attribution missing")
	}
}

func TestGetLocationDetail_NilPoolReturns404(t *testing.T) {
	srv := apiTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/locations/potsdam", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rec.Code)
	}
}
