package oauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestIsAllowedOrigin(t *testing.T) {
	allowed := []string{"worldweathernews.com", "research.worldweathernews.com", "localhost"}
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"bare host match", "worldweathernews.com", true},
		{"subdomain match", "blog.worldweathernews.com", true},
		{"deeply nested subdomain", "a.b.research.worldweathernews.com", true},
		{"https URL match", "https://research.worldweathernews.com", true},
		{"localhost", "localhost", true},
		{"empty", "", false},
		{"different host", "evil.example.com", false},
		{"suffix-only match must not pass", "evilworldweathernews.com", false},
		{"invalid URL", "://broken", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsAllowedOrigin(tc.in, allowed); got != tc.want {
				t.Fatalf("IsAllowedOrigin(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseAllowedDomains(t *testing.T) {
	got := ParseAllowedDomains(" worldweathernews.com, ,research.worldweathernews.com ")
	if len(got) != 2 || got[0] != "worldweathernews.com" || got[1] != "research.worldweathernews.com" {
		t.Fatalf("unexpected: %#v", got)
	}
}

func TestAuth_RedirectsAndSetsCSRFCookie(t *testing.T) {
	h := NewHandler(Config{
		ClientID:       "id",
		ClientSecret:   "secret",
		AllowedDomains: []string{"worldweathernews.com"},
		PublicBaseURL:  "https://cms-auth.worldweathernews.com",
	})

	req := httptest.NewRequest(http.MethodGet, "/auth?provider=github&site_id=https://worldweathernews.com&scope=repo,user", nil)
	rec := httptest.NewRecorder()
	h.Auth(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	loc := rec.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	if u.Host != "github.com" {
		t.Fatalf("location host = %q, want github.com", u.Host)
	}
	if u.Query().Get("client_id") != "id" {
		t.Fatalf("client_id missing")
	}
	if u.Query().Get("redirect_uri") != "https://cms-auth.worldweathernews.com/callback" {
		t.Fatalf("redirect_uri = %q", u.Query().Get("redirect_uri"))
	}
	state := u.Query().Get("state")
	if state == "" {
		t.Fatalf("state missing")
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != csrfCookieName || cookies[0].Value != state {
		t.Fatalf("csrf cookie missing or mismatched: %#v", cookies)
	}
}

func TestAuth_RejectsUnknownProvider(t *testing.T) {
	h := NewHandler(Config{AllowedDomains: []string{"worldweathernews.com"}})
	req := httptest.NewRequest(http.MethodGet, "/auth?provider=gitlab&site_id=worldweathernews.com", nil)
	rec := httptest.NewRecorder()
	h.Auth(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestAuth_RejectsDisallowedOrigin(t *testing.T) {
	h := NewHandler(Config{ClientID: "id", AllowedDomains: []string{"worldweathernews.com"}})
	req := httptest.NewRequest(http.MethodGet, "/auth?provider=github&site_id=evil.example.com", nil)
	rec := httptest.NewRecorder()
	h.Auth(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCallback_Success(t *testing.T) {
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != githubTokenURL {
			t.Fatalf("unexpected URL %s", r.URL.String())
		}
		body, _ := io.ReadAll(r.Body)
		var in map[string]string
		_ = json.Unmarshal(body, &in)
		if in["code"] != "abc" || in["client_id"] != "id" || in["client_secret"] != "secret" {
			t.Fatalf("payload mismatch: %#v", in)
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"access_token":"gho_xxx"}`)),
			Header:     make(http.Header),
		}, nil
	})
	h := NewHandler(Config{
		ClientID:     "id",
		ClientSecret: "secret",
		HTTPClient:   &http.Client{Transport: rt},
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/callback?code=abc&state=xyz", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "xyz"})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "authorization:github:success:") {
		t.Fatalf("body does not contain success payload: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "gho_xxx") {
		t.Fatalf("token missing from response")
	}
}

func TestCallback_StateMismatch(t *testing.T) {
	h := NewHandler(Config{})
	req := httptest.NewRequest(http.MethodGet, "/callback?code=abc&state=xyz", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "different"})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestCallback_TokenExchangeFailure(t *testing.T) {
	rt := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 401,
			Body:       io.NopCloser(strings.NewReader(`{"error":"bad_verification_code","error_description":"nope"}`)),
			Header:     make(http.Header),
		}, nil
	})
	h := NewHandler(Config{HTTPClient: &http.Client{Transport: rt}})
	req := httptest.NewRequest(http.MethodGet, "/callback?code=abc&state=xyz", nil)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "xyz"})
	rec := httptest.NewRecorder()
	h.Callback(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "bad_verification_code") {
		t.Fatalf("error not propagated: %s", rec.Body.String())
	}
}
