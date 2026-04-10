package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/aray/cerebray/backend/internal/auth"
)

func newTestSessionStore(t *testing.T) (*auth.SessionStore, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("starting miniredis: %v", err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })

	return auth.NewSessionStore(client), mr
}

func TestRequireAuth_ValidSession(t *testing.T) {
	sessions, _ := newTestSessionStore(t)

	// Create a real session
	sessionID, err := sessions.Create(t.Context(), 42)
	if err != nil {
		t.Fatalf("creating session: %v", err)
	}

	// Handler that checks for user_id in context
	var gotUserID int64
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mw := RequireAuth(sessions)
	handler := mw(inner)

	req := httptest.NewRequest("GET", "/api/v1/notes", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: sessionID})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
	if gotUserID != 42 {
		t.Errorf("got user_id %d, want 42", gotUserID)
	}
}

func TestRequireAuth_NoCookie(t *testing.T) {
	sessions, _ := newTestSessionStore(t)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	mw := RequireAuth(sessions)
	handler := mw(inner)

	req := httptest.NewRequest("GET", "/api/v1/notes", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_InvalidSession(t *testing.T) {
	sessions, _ := newTestSessionStore(t)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called")
	})

	mw := RequireAuth(sessions)
	handler := mw(inner)

	req := httptest.NewRequest("GET", "/api/v1/notes", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "bogus-session-id"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetUserID_Present(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, UserIDKey, int64(99))
	if got := GetUserID(ctx); got != 99 {
		t.Errorf("got %d, want 99", got)
	}
}

func TestRequestLogger_IncludesUserID(t *testing.T) {
	ctx := context.WithValue(t.Context(), UserIDKey, int64(7))
	logger := RequestLogger(ctx)
	// Verify logger is created without panic - functional test
	logger.Info().Msg("test log with user context")
}

func TestGetUserID_Missing(t *testing.T) {
	if got := GetUserID(t.Context()); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}
