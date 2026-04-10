package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()

	JSON(w, http.StatusCreated, map[string]string{"msg": "created"})

	if w.Code != http.StatusCreated {
		t.Errorf("JSON() status = %d, want %d", w.Code, http.StatusCreated)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("JSON() Content-Type = %q, want %q", ct, "application/json")
	}

	var resp map[string]string
	decodeJSON(t, w.Body, &resp)
	if resp["msg"] != "created" {
		t.Errorf("JSON() body msg = %q, want %q", resp["msg"], "created")
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()

	Error(w, http.StatusBadRequest, "something broke")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Error() status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp map[string]string
	decodeJSON(t, w.Body, &resp)
	if resp["error"] != "something broke" {
		t.Errorf("Error() error field = %q, want %q", resp["error"], "something broke")
	}
}

func TestQueryInt(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		key      string
		fallback int
		want     int
	}{
		{
			name:     "returns parsed value",
			query:    "?limit=25",
			key:      "limit",
			fallback: 50,
			want:     25,
		},
		{
			name:     "returns default when missing",
			query:    "",
			key:      "limit",
			fallback: 50,
			want:     50,
		},
		{
			name:     "returns default for non-numeric",
			query:    "?limit=abc",
			key:      "limit",
			fallback: 50,
			want:     50,
		},
		{
			name:     "returns default for empty value",
			query:    "?limit=",
			key:      "limit",
			fallback: 10,
			want:     10,
		},
		{
			name:     "handles zero",
			query:    "?offset=0",
			key:      "offset",
			fallback: 5,
			want:     0,
		},
		{
			name:     "handles negative",
			query:    "?page=-1",
			key:      "page",
			fallback: 1,
			want:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/test"+tt.query, nil)
			got := QueryInt(r, tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf("QueryInt(%q, %d) = %d, want %d", tt.key, tt.fallback, got, tt.want)
			}
		})
	}
}
