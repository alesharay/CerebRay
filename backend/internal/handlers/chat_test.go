package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aray/cerebray/backend/db/sqlc"
)

func TestChatHandlers_GetUsage(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(q *fakeQuerier)
		wantStatus int
		wantTokens int32
	}{
		{
			name: "returns usage data",
			setup: func(q *fakeQuerier) {
				q.GetMonthlyUsageFn = func(_ context.Context, userID *int64) (sqlc.GetMonthlyUsageRow, error) {
					if userID == nil || *userID != 1 {
						t.Errorf("expected userID pointer to 1, got %v", userID)
					}
					return sqlc.GetMonthlyUsageRow{
						TotalInputTokens:  1500,
						TotalOutputTokens: 3000,
						TotalRequests:     10,
					}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantTokens: 1500,
		},
		{
			name: "db error returns 500",
			setup: func(q *fakeQuerier) {
				q.GetMonthlyUsageFn = func(_ context.Context, _ *int64) (sqlc.GetMonthlyUsageRow, error) {
					return sqlc.GetMonthlyUsageRow{}, fmt.Errorf("query failed")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "zero usage",
			setup: func(q *fakeQuerier) {
				q.GetMonthlyUsageFn = func(_ context.Context, _ *int64) (sqlc.GetMonthlyUsageRow, error) {
					return sqlc.GetMonthlyUsageRow{}, nil
				}
			},
			wantStatus: http.StatusOK,
			wantTokens: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &fakeQuerier{}
			tt.setup(q)
			m := newTestMetrics()
			h := NewChatHandlers(q, &mockProvider{}, m)

			r := reqWithUser("GET", "/api/v1/settings/usage", nil, 1)
			w := httptest.NewRecorder()

			h.GetUsage(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("GetUsage() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var usage sqlc.GetMonthlyUsageRow
				decodeJSON(t, w.Body, &usage)
				if usage.TotalInputTokens != tt.wantTokens {
					t.Errorf("GetUsage() input tokens = %d, want %d", usage.TotalInputTokens, tt.wantTokens)
				}
			}
		})
	}
}

func TestChatHandlers_SendMessage_NoProvider(t *testing.T) {
	q := &fakeQuerier{}
	m := newTestMetrics()
	h := NewChatHandlers(q, nil, m)

	r := reqWithUserAndChiCtx("POST", "/api/v1/conversations/1/messages", nil, 1, map[string]string{"id": "1"})
	w := httptest.NewRecorder()

	h.SendMessage(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("SendMessage() without provider status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}

	var resp map[string]string
	decodeJSON(t, w.Body, &resp)
	if resp["error"] != "AI not configured" {
		t.Errorf("expected error 'AI not configured', got %q", resp["error"])
	}
}
