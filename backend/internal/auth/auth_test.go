package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/aray/cerebray/backend/internal/auth"
	mw "github.com/aray/cerebray/backend/internal/middleware"
)

// fakeUser is a minimal user struct returned by the fake GetUser func.
type fakeUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// testHarness bundles everything needed to run auth integration tests.
type testHarness struct {
	server   *httptest.Server
	client   *http.Client
	mini     *miniredis.Miniredis
	sessions *auth.SessionStore
}

func newTestHarness(t *testing.T) *testHarness {
	t.Helper()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("starting miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	sessions := auth.NewSessionStore(rdb)

	// Wire up GetUserIDFromContext so HandleMe can read the user ID.
	auth.GetUserIDFromContext = mw.GetUserID

	handler := auth.NewHandler(auth.HandlerConfig{
		Sessions:     sessions,
		UpsertUser:   fakeUpsertUser,
		GetUser:      fakeGetUser,
		BaseURL:      "", // empty so redirects stay relative to the test server
		SecureCookie: false,
		LocalMode:    true,
		LocalEmail:   "test@cerebray.local",
		LocalName:    "Test User",
	})

	r := chi.NewRouter()

	// Auth routes (public)
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", handler.HandleLogin)
		r.Post("/logout", handler.HandleLogout)
	})

	// Protected routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.RequireAuth(sessions))
		r.Get("/auth/me", handler.HandleMe)
	})

	server := httptest.NewServer(r)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("creating cookie jar: %v", err)
	}

	// Client that does NOT follow redirects so we can inspect the 307 response.
	noRedirectClient := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t.Cleanup(func() {
		server.Close()
		rdb.Close()
		mini.Close()
	})

	return &testHarness{
		server:   server,
		client:   noRedirectClient,
		mini:     mini,
		sessions: sessions,
	}
}

// newFollowRedirectsClient returns a client that shares the same cookie jar
// but follows redirects normally.
func (h *testHarness) newFollowRedirectsClient() *http.Client {
	return &http.Client{Jar: h.client.Jar}
}

// fakeUpsertUser always returns user ID 42.
func fakeUpsertUser(_ context.Context, _, _, _, _ string) (int64, error) {
	return 42, nil
}

// fakeGetUser returns a fakeUser for ID 42, error for anything else.
func fakeGetUser(_ context.Context, id int64) (interface{}, error) {
	if id == 42 {
		return fakeUser{ID: 42, Email: "test@cerebray.local", Name: "Test User"}, nil
	}
	return nil, fmt.Errorf("user %d not found", id)
}

func TestLocalLogin_CreatesSessionAndRedirects(t *testing.T) {
	h := newTestHarness(t)

	resp, err := h.client.Get(h.server.URL + "/auth/login")
	if err != nil {
		t.Fatalf("GET /auth/login: %v", err)
	}
	defer resp.Body.Close()

	// Should redirect to /dashboard
	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307 redirect, got %d", resp.StatusCode)
	}

	loc := resp.Header.Get("Location")
	if loc != "/dashboard" {
		t.Errorf("redirect location = %q, want %q", loc, "/dashboard")
	}

	// Should have set the session cookie
	var found bool
	for _, c := range resp.Cookies() {
		if c.Name == auth.SessionCookieName {
			found = true
			if c.Value == "" {
				t.Error("session cookie value is empty")
			}
			break
		}
	}
	if !found {
		t.Error("session cookie not set after login")
	}
}

func TestHandleMe_WithValidSession(t *testing.T) {
	h := newTestHarness(t)

	// Login first to establish a session
	resp, err := h.client.Get(h.server.URL + "/auth/login")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	resp.Body.Close()

	// Now call /api/v1/auth/me - cookie jar carries the session
	resp, err = h.client.Get(h.server.URL + "/api/v1/auth/me")
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var user fakeUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if user.ID != 42 {
		t.Errorf("user ID = %d, want 42", user.ID)
	}
	if user.Email != "test@cerebray.local" {
		t.Errorf("user email = %q, want %q", user.Email, "test@cerebray.local")
	}
	if user.Name != "Test User" {
		t.Errorf("user name = %q, want %q", user.Name, "Test User")
	}
}

func TestHandleMe_WithoutSession(t *testing.T) {
	h := newTestHarness(t)

	// Call /api/v1/auth/me without logging in
	resp, err := h.client.Get(h.server.URL + "/api/v1/auth/me")
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogout_ClearsSession(t *testing.T) {
	h := newTestHarness(t)

	// Login
	resp, err := h.client.Get(h.server.URL + "/auth/login")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	resp.Body.Close()

	// Verify session works
	resp, err = h.client.Get(h.server.URL + "/api/v1/auth/me")
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me before logout: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 before logout, got %d", resp.StatusCode)
	}

	// Logout
	resp, err = h.client.Post(h.server.URL+"/auth/logout", "", nil)
	if err != nil {
		t.Fatalf("POST /auth/logout: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("logout status = %d, want 204", resp.StatusCode)
	}

	// Session should be gone - /api/v1/auth/me should return 401
	resp, err = h.client.Get(h.server.URL + "/api/v1/auth/me")
	if err != nil {
		t.Fatalf("GET /api/v1/auth/me after logout: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 after logout, got %d", resp.StatusCode)
	}
}
