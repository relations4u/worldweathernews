// wwn-cms-auth is the self-hosted OAuth proxy for Sveltia CMS.
//
// Endpoints:
//
//	GET /auth      → 302 to github.com/login/oauth/authorize
//	GET /callback  → exchanges the code, posts the token back to the opener
//	GET /healthz   → 200 ok (used by docker HEALTHCHECK)
//	GET /          → plain banner
//
// Configuration via environment variables:
//
//	WWN_CMS_AUTH_LISTEN          (default :8090)
//	WWN_CMS_AUTH_PUBLIC_BASE_URL (e.g. https://cms-auth.worldweathernews.com)
//	WWN_CMS_AUTH_CLIENT_ID       (GitHub OAuth App)
//	WWN_CMS_AUTH_CLIENT_SECRET   (GitHub OAuth App)
//	WWN_CMS_AUTH_ALLOWED_DOMAINS (CSV, e.g. worldweathernews.com,research.worldweathernews.com)
//	WWN_CMS_AUTH_LOG_FORMAT      (json|text, default json)
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/relations4u/worldweathernews/apps/cms-auth/internal/oauth"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}

	logger := newLogger()
	slog.SetDefault(logger)

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("config", "err", err)
		os.Exit(1)
	}

	h := oauth.NewHandler(oauth.Config{
		ClientID:       cfg.ClientID,
		ClientSecret:   cfg.ClientSecret,
		AllowedDomains: cfg.AllowedDomains,
		PublicBaseURL:  cfg.PublicBaseURL,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		Logger:         logger,
	})

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Get("/auth", h.Auth)
	r.Get("/callback", h.Callback)
	r.Get("/healthz", h.Healthz)
	r.Get("/", h.Index)

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("listening", "addr", cfg.Listen, "public_base_url", cfg.PublicBaseURL,
			"allowed_domains", cfg.AllowedDomains)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
	}
}

type config struct {
	Listen         string
	PublicBaseURL  string
	ClientID       string
	ClientSecret   string
	AllowedDomains []string
}

func loadConfig() (config, error) {
	c := config{
		Listen:         envDefault("WWN_CMS_AUTH_LISTEN", ":8090"),
		PublicBaseURL:  os.Getenv("WWN_CMS_AUTH_PUBLIC_BASE_URL"),
		ClientID:       os.Getenv("WWN_CMS_AUTH_CLIENT_ID"),
		ClientSecret:   os.Getenv("WWN_CMS_AUTH_CLIENT_SECRET"),
		AllowedDomains: oauth.ParseAllowedDomains(os.Getenv("WWN_CMS_AUTH_ALLOWED_DOMAINS")),
	}
	missing := []string{}
	if c.PublicBaseURL == "" {
		missing = append(missing, "WWN_CMS_AUTH_PUBLIC_BASE_URL")
	}
	if c.ClientID == "" {
		missing = append(missing, "WWN_CMS_AUTH_CLIENT_ID")
	}
	if c.ClientSecret == "" {
		missing = append(missing, "WWN_CMS_AUTH_CLIENT_SECRET")
	}
	if len(c.AllowedDomains) == 0 {
		missing = append(missing, "WWN_CMS_AUTH_ALLOWED_DOMAINS")
	}
	if len(missing) > 0 {
		return c, errors.New("missing required env: " + strings.Join(missing, ", "))
	}
	return c, nil
}

func newLogger() *slog.Logger {
	if os.Getenv("WWN_CMS_AUTH_LOG_FORMAT") == "text" {
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// runHealthcheck is invoked via `wwn-cms-auth healthcheck` from the Docker
// HEALTHCHECK directive. Distroless images do not ship wget/curl, so the
// binary speaks HTTP to itself.
func runHealthcheck() {
	addr := envDefault("WWN_CMS_AUTH_LISTEN", ":8090")
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + addr + "/healthz")
	if err != nil || resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	_ = resp.Body.Close()
}
