// Package oauth implements the GitHub OAuth handshake used by Sveltia CMS.
//
// Sveltia (and its predecessor Decap CMS) open a popup at <base>/auth, expect
// a redirect to GitHub, and consume the token via postMessage from <base>/callback.
// Logic is a direct port of the previous Cloudflare-Worker implementation; the
// flow stays minimal — only `provider=github`, allowed origins are validated
// explicitly, CSRF state is bound via a Secure HttpOnly cookie.
package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	// #nosec G101 — public GitHub endpoint, not a credential.
	githubTokenURL   = "https://github.com/login/oauth/access_token"
	csrfCookieName   = "csrf-token"
	csrfCookieMaxAge = 600
)

// Config carries the runtime configuration of the OAuth proxy.
type Config struct {
	ClientID       string
	ClientSecret   string
	AllowedDomains []string
	PublicBaseURL  string
	HTTPClient     *http.Client
	Logger         *slog.Logger
}

// Handler implements the Sveltia-compatible /auth + /callback handshake.
type Handler struct {
	cfg Config
}

// NewHandler returns a Handler with HTTPClient and Logger defaulted when nil.
func NewHandler(cfg Config) *Handler {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = http.DefaultClient
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	return &Handler{cfg: cfg}
}

// Auth starts the GitHub OAuth handshake by redirecting to GitHub with a
// freshly minted CSRF state bound to a Secure HttpOnly cookie.
func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	if provider := r.URL.Query().Get("provider"); provider != "github" {
		http.Error(w, "Unsupported provider", http.StatusBadRequest)
		return
	}
	siteID := r.URL.Query().Get("site_id")
	if !IsAllowedOrigin(siteID, h.cfg.AllowedDomains) {
		h.cfg.Logger.Warn("rejected origin", "site_id", siteID)
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = "repo,user"
	}
	state, err := randomState()
	if err != nil {
		h.cfg.Logger.Error("randomState", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	redirectURI := strings.TrimRight(h.cfg.PublicBaseURL, "/") + "/callback"
	ghURL, _ := url.Parse(githubAuthorizeURL)
	q := ghURL.Query()
	q.Set("client_id", h.cfg.ClientID)
	q.Set("scope", scope)
	q.Set("state", state)
	q.Set("redirect_uri", redirectURI)
	ghURL.RawQuery = q.Encode()

	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   csrfCookieMaxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, ghURL.String(), http.StatusFound)
}

// Callback completes the OAuth handshake: it validates the CSRF state,
// exchanges the GitHub code for an access token, and delivers the result via
// postMessage to the opener.
func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	cookie, _ := r.Cookie(csrfCookieName)
	if code == "" || state == "" || cookie == nil || cookie.Value != state {
		h.cfg.Logger.Warn("invalid state")
		writeResponseHTML(w, http.StatusBadRequest, "error", map[string]string{
			"error": "invalid state",
		})
		return
	}

	body, _ := json.Marshal(map[string]string{
		"client_id":     h.cfg.ClientID,
		"client_secret": h.cfg.ClientSecret,
		"code":          code,
	})
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, githubTokenURL, strings.NewReader(string(body)))
	if err != nil {
		h.cfg.Logger.Error("build token request", "err", err)
		writeResponseHTML(w, http.StatusInternalServerError, "error", map[string]string{"error": "internal"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := h.cfg.HTTPClient.Do(req)
	if err != nil {
		h.cfg.Logger.Error("token exchange", "err", err)
		writeResponseHTML(w, http.StatusBadGateway, "error", map[string]string{"error": "token exchange failed"})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(resp.Body)
	var tok struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	_ = json.Unmarshal(raw, &tok)

	clearCookie(w)

	if resp.StatusCode/100 != 2 || tok.AccessToken == "" {
		h.cfg.Logger.Warn("token exchange rejected", "status", resp.StatusCode, "err", tok.Error)
		writeResponseHTML(w, http.StatusUnauthorized, "error", map[string]string{
			"error":       firstNonEmpty(tok.Error, "token exchange failed"),
			"description": tok.ErrorDescription,
		})
		return
	}

	writeResponseHTML(w, http.StatusOK, "success", map[string]string{
		"provider": "github",
		"token":    tok.AccessToken,
	})
}

// Index serves a plain banner so that operators hitting the bare host can
// verify the service is up without speaking the full OAuth handshake.
func (h *Handler) Index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "wwn-cms-auth: GitHub OAuth proxy for Sveltia CMS\n")
}

// Healthz answers Docker HEALTHCHECK probes.
func (h *Handler) Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "ok")
}

// IsAllowedOrigin accepts a site_id query value (bare host or full URL) and
// returns true if its hostname matches one of the allowed bare hostnames, or
// is a subdomain thereof.
func IsAllowedOrigin(raw string, allowed []string) bool {
	if raw == "" {
		return false
	}
	withScheme := raw
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		withScheme = "https://" + raw
	}
	u, err := url.Parse(withScheme)
	if err != nil || u.Hostname() == "" {
		return false
	}
	host := u.Hostname()
	for _, d := range allowed {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// ParseAllowedDomains parses a comma-separated list of bare hostnames, trimming
// whitespace and dropping empty entries.
func ParseAllowedDomains(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func writeResponseHTML(w http.ResponseWriter, status int, kind string, payload map[string]string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	data, _ := json.Marshal(payload)
	_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Authorizing…</title></head>
<body>
<script>
(function() {
  var data = %s;
  function receive(event) {
    if (event.data === 'authorizing:github') {
      window.removeEventListener('message', receive, false);
      window.opener.postMessage(
        'authorization:github:%s:' + JSON.stringify(data),
        event.origin
      );
    }
  }
  window.addEventListener('message', receive, false);
  window.opener.postMessage('authorizing:github', '*');
})();
</script>
</body>
</html>`, string(data), kind)
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}
