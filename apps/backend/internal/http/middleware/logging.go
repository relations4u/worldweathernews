// Package middleware enthält HTTP-Middleware-Komponenten des Backends.
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/relations4u/worldweathernews/apps/backend/internal/observability"
)

// RequestLogger loggt jede Request einmal nach Abschluss.
// Hängt den Logger zusätzlich an den Context (mit request_id-Attribut),
// damit Handler ihn via observability.FromContext nutzen können.
func RequestLogger(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			reqID := chimw.GetReqID(r.Context())
			log := base.With(slog.String("request_id", reqID))

			ctx := observability.WithLogger(r.Context(), log)
			next.ServeHTTP(ww, r.WithContext(ctx))

			log.LogAttrs(ctx, slog.LevelInfo, "http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}
