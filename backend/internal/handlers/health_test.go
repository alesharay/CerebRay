package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	r := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	HandleHealth(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("HandleHealth() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	decodeJSON(t, w.Body, &resp)

	if resp["status"] != "ok" {
		t.Errorf("HandleHealth() status field = %q, want %q", resp["status"], "ok")
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("HandleHealth() Content-Type = %q, want %q", ct, "application/json")
	}
}
