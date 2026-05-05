// Package main ist der Einstiegspunkt des wwn-backend-Binaries.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/relations4u/worldweathernews/apps/backend/internal/config"
	httpapp "github.com/relations4u/worldweathernews/apps/backend/internal/http"
	"github.com/relations4u/worldweathernews/apps/backend/internal/http/handler"
	"github.com/relations4u/worldweathernews/apps/backend/internal/observability"
	"github.com/relations4u/worldweathernews/apps/backend/internal/storage"
	"github.com/relations4u/worldweathernews/apps/backend/internal/version"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck())
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Getenv("WWN_CONFIG_FILE"))
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log := observability.NewLogger(cfg.Logging.Format, cfg.Logging.Level)
	log = log.With(slog.String("service", "wwn-backend"), slog.Group("version",
		slog.String("v", version.Version),
		slog.String("commit", version.Commit),
	))
	slog.SetDefault(log)

	log.Info("starting", slog.String("environment", cfg.Environment), slog.Int("port", cfg.HTTP.Port))

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tracingShutdown, err := observability.InitTracing(rootCtx, observability.TracingConfig{
		Enabled:     cfg.Observability.Tracing.Enabled,
		Endpoint:    cfg.Observability.Tracing.Endpoint,
		ServiceName: "wwn-backend",
		Environment: cfg.Environment,
		Version:     version.Version,
	})
	if err != nil {
		// Tracing-Probleme dürfen den Start nicht killen — Tempo könnte
		// schlicht nicht laufen. Loggen, weiter.
		log.Warn("tracing init failed, continuing without traces", slog.Any("err", err))
		tracingShutdown = func(context.Context) error { return nil }
	} else if cfg.Observability.Tracing.Enabled {
		log.Info("tracing enabled", slog.String("endpoint", cfg.Observability.Tracing.Endpoint))
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracingShutdown(shutdownCtx); err != nil {
			log.Warn("tracing shutdown failed", slog.Any("err", err))
		}
	}()

	pool, err := storage.NewPool(rootCtx, cfg.Database)
	if err != nil {
		if cfg.Environment == "dev" {
			log.Warn("postgres unavailable, continuing in dev", slog.Any("err", err))
		} else {
			return fmt.Errorf("postgres: %w", err)
		}
	}
	defer func() {
		if pool != nil {
			pool.Close()
		}
	}()

	rds, err := storage.NewRedis(rootCtx, cfg.Redis)
	if err != nil {
		if cfg.Environment == "dev" {
			log.Warn("redis unavailable, continuing in dev", slog.Any("err", err))
		} else {
			return fmt.Errorf("redis: %w", err)
		}
	}
	defer func() {
		if rds != nil {
			_ = rds.Close()
		}
	}()

	metrics := observability.NewMetrics()
	deps := handler.Deps{StartedAt: time.Now(), DB: pool, Redis: rds}
	router := httpapp.NewRouter(cfg, deps, log, metrics)

	srv := &stdhttp.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Info("http listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-rootCtx.Done():
		log.Info("shutdown signal received")
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info("stopped")
	return nil
}

// runHealthcheck wird vom Docker-HEALTHCHECK aufgerufen. Pingt den eigenen
// /health-Endpoint und exited mit 0 oder 1.
// Hält keine externen Abhängigkeiten zu liveness — pingt nur den Prozess.
func runHealthcheck() int {
	port := os.Getenv("WWN_HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	if _, err := strconv.Atoi(port); err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: invalid WWN_HTTP_PORT %q\n", port)
		return 1
	}

	client := &stdhttp.Client{Timeout: 2 * time.Second}
	url := "http://127.0.0.1:" + port + "/health"
	// gosec G107/G704: URL ist lokal (loopback) und aus validiertem ENV,
	// kein Caller-Context-Risiko im Healthcheck-CLI.
	req, err := stdhttp.NewRequest(stdhttp.MethodGet, url, nil) //nolint:noctx,gosec // siehe oben
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %v\n", err)
		return 1
	}
	resp, err := client.Do(req) //nolint:gosec // loopback-only, statisch
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %v\n", err)
		return 1
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != stdhttp.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck: status %d\n", resp.StatusCode)
		return 1
	}
	return 0
}
