// Package http baut den Chi-Router samt Middleware und Routen für das Backend.
package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/relations4u/worldweathernews/apps/backend/internal/api"
	"github.com/relations4u/worldweathernews/apps/backend/internal/config"
	"github.com/relations4u/worldweathernews/apps/backend/internal/http/handler"
	mw "github.com/relations4u/worldweathernews/apps/backend/internal/http/middleware"
	"github.com/relations4u/worldweathernews/apps/backend/internal/observability"
)

// NewRouter baut den Chi-Router mit Standard-Middleware und allen Routen.
func NewRouter(cfg *config.Config, deps handler.Deps, log *slog.Logger, m *observability.Metrics) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mw.RequestLogger(log))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.HTTP.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	r.Use(metricsMiddleware(m))

	// System-Endpoints (bewusst außerhalb OpenAPI).
	r.Get("/health", handler.Health(deps))
	r.Get("/ready", handler.Ready(deps))
	if cfg.MetricsEnabled {
		r.Handle("/metrics", promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{Registry: m.Registry}))
	}

	// OpenAPI-generierte Routen unter /api/v1/.
	apiHandler := handler.NewAPIHandler()
	strictHandler := api.NewStrictHandler(apiHandler, nil)
	api.HandlerFromMux(strictHandler, r)

	return r
}

// metricsMiddleware füllt die HTTP-Metriken pro Request.
// Path nutzt das Chi-RoutePattern (z.B. /api/v1/users/{id}), damit Cardinality
// nicht durch URL-Parameter explodiert.
func metricsMiddleware(m *observability.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			pattern := chi.RouteContext(r.Context()).RoutePattern()
			if pattern == "" {
				pattern = "unknown"
			}
			m.HTTPRequestsTotal.WithLabelValues(r.Method, pattern, strconv.Itoa(ww.Status())).Inc()
			m.HTTPRequestDuration.WithLabelValues(r.Method, pattern).Observe(time.Since(start).Seconds())
		})
	}
}
