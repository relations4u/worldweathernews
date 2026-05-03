// Package handler enthält die HTTP-Handler des Backends.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/relations4u/worldweathernews/apps/backend/internal/storage"
	"github.com/relations4u/worldweathernews/apps/backend/internal/version"
)

// Deps bündelt alles, was die Handler brauchen.
type Deps struct {
	StartedAt time.Time
	DB        *pgxpool.Pool
	Redis     *redis.Client
}

type healthResponse struct {
	Status  string       `json:"status"`
	Version version.Info `json:"version"`
	Uptime  string       `json:"uptime"`
}

// Health ist der Liveness-Endpoint. Antwortet 200 wenn der Prozess läuft.
func Health(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status:  "ok",
			Version: version.Get(),
			Uptime:  time.Since(deps.StartedAt).Round(time.Second).String(),
		})
	}
}

type readyComponent struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type readyResponse struct {
	Status     string           `json:"status"`
	Components []readyComponent `json:"components"`
}

// Ready ist der Readiness-Endpoint. Pingt DB und Redis und gibt 503 zurück,
// wenn auch nur eines fehlt.
func Ready(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		components := make([]readyComponent, 0, 2)
		overall := "ok"

		if err := storage.PostgresHealth(ctx, deps.DB); err != nil {
			overall = "degraded"
			components = append(components, readyComponent{
				Name: "postgres", Status: "down", Error: err.Error(),
			})
		} else {
			components = append(components, readyComponent{Name: "postgres", Status: "ok"})
		}

		if err := storage.RedisHealth(ctx, deps.Redis); err != nil {
			overall = "degraded"
			components = append(components, readyComponent{
				Name: "redis", Status: "down", Error: err.Error(),
			})
		} else {
			components = append(components, readyComponent{Name: "redis", Status: "ok"})
		}

		status := http.StatusOK
		if overall != "ok" {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, readyResponse{Status: overall, Components: components})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
