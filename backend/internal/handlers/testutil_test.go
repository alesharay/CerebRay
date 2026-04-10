package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/ai"
	"github.com/aray/cerebray/backend/internal/metrics"
	"github.com/aray/cerebray/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
)

// fakeQuerier implements sqlc.Querier with injectable function fields.
// Any method without a function set returns zero values and nil error.
type fakeQuerier struct {
	AddNoteTagFn                   func(ctx context.Context, arg sqlc.AddNoteTagParams) error
	CountNotesByStatusFn           func(ctx context.Context, userID int64) ([]sqlc.CountNotesByStatusRow, error)
	CreateConnectionFn             func(ctx context.Context, arg sqlc.CreateConnectionParams) (sqlc.Connection, error)
	CreateConversationFn           func(ctx context.Context, arg sqlc.CreateConversationParams) (sqlc.Conversation, error)
	CreateGlossaryTermFn           func(ctx context.Context, arg sqlc.CreateGlossaryTermParams) (sqlc.GlossaryTerm, error)
	CreateMessageFn                func(ctx context.Context, arg sqlc.CreateMessageParams) (sqlc.Message, error)
	CreateNoteFn                   func(ctx context.Context, arg sqlc.CreateNoteParams) (sqlc.Note, error)
	CreateNoteEventFn              func(ctx context.Context, arg sqlc.CreateNoteEventParams) (sqlc.NoteEvent, error)
	CreateTagFn                    func(ctx context.Context, arg sqlc.CreateTagParams) (sqlc.Tag, error)
	CreateUserFn                   func(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error)
	DeleteConnectionFn             func(ctx context.Context, id int64) error
	DeleteConversationFn           func(ctx context.Context, arg sqlc.DeleteConversationParams) error
	DeleteGlossaryTermFn           func(ctx context.Context, arg sqlc.DeleteGlossaryTermParams) error
	DeleteNoteFn                   func(ctx context.Context, arg sqlc.DeleteNoteParams) error
	DeleteTagFn                    func(ctx context.Context, arg sqlc.DeleteTagParams) error
	GetConnectionDensityFn         func(ctx context.Context, userID int64) (sqlc.GetConnectionDensityRow, error)
	GetConversationFn              func(ctx context.Context, arg sqlc.GetConversationParams) (sqlc.Conversation, error)
	GetConversationConversionRateFn func(ctx context.Context, userID int64) (sqlc.GetConversationConversionRateRow, error)
	GetFleetingNotesWithAgeFn      func(ctx context.Context, userID int64) ([]sqlc.GetFleetingNotesWithAgeRow, error)
	GetGlossaryCountFn             func(ctx context.Context, userID int64) (int32, error)
	GetGraphDataFn                 func(ctx context.Context, userID int64) ([]sqlc.GetGraphDataRow, error)
	GetLifecycleMetricsFn          func(ctx context.Context, arg sqlc.GetLifecycleMetricsParams) ([]sqlc.GetLifecycleMetricsRow, error)
	GetLifecycleTrendFn            func(ctx context.Context, arg sqlc.GetLifecycleTrendParams) ([]sqlc.GetLifecycleTrendRow, error)
	GetMonthlyUsageFn              func(ctx context.Context, userID *int64) (sqlc.GetMonthlyUsageRow, error)
	GetNoteByIDFn                  func(ctx context.Context, arg sqlc.GetNoteByIDParams) (sqlc.Note, error)
	GetNoteBySourceChatFn          func(ctx context.Context, arg sqlc.GetNoteBySourceChatParams) (sqlc.Note, error)
	GetNoteCurrentStatusFn         func(ctx context.Context, arg sqlc.GetNoteCurrentStatusParams) (sqlc.NoteStatus, error)
	GetNoteTypeDistributionFn      func(ctx context.Context, userID int64) ([]sqlc.GetNoteTypeDistributionRow, error)
	GetSleepingNoteStatsFn         func(ctx context.Context, userID int64) (sqlc.GetSleepingNoteStatsRow, error)
	GetStaleNotesFn                func(ctx context.Context, arg sqlc.GetStaleNotesParams) ([]sqlc.GetStaleNotesRow, error)
	GetTagsForNoteFn               func(ctx context.Context, noteID int64) ([]sqlc.Tag, error)
	GetUserByEmailFn               func(ctx context.Context, email string) (sqlc.User, error)
	GetUserByIDFn                  func(ctx context.Context, id int64) (sqlc.User, error)
	GetUserByOIDCSubjectFn         func(ctx context.Context, oidcSubject string) (sqlc.User, error)
	ListConnectionsForNoteFn       func(ctx context.Context, sourceID int64) ([]sqlc.ListConnectionsForNoteRow, error)
	ListConversationsFn            func(ctx context.Context, arg sqlc.ListConversationsParams) ([]sqlc.Conversation, error)
	ListGlossaryTermsFn            func(ctx context.Context, userID int64) ([]sqlc.ListGlossaryTermsRow, error)
	ListMessagesFn                 func(ctx context.Context, conversationID int64) ([]sqlc.Message, error)
	ListNotesByStatusFn            func(ctx context.Context, arg sqlc.ListNotesByStatusParams) ([]sqlc.Note, error)
	ListNotesByTagFn               func(ctx context.Context, arg sqlc.ListNotesByTagParams) ([]sqlc.Note, error)
	ListNotesByUserFn              func(ctx context.Context, arg sqlc.ListNotesByUserParams) ([]sqlc.ListNotesByUserRow, error)
	ListTagsByUserFn               func(ctx context.Context, userID int64) ([]sqlc.ListTagsByUserRow, error)
	LogAIUsageFn                   func(ctx context.Context, arg sqlc.LogAIUsageParams) (sqlc.AiUsageLog, error)
	RecentNotesFn                  func(ctx context.Context, arg sqlc.RecentNotesParams) ([]sqlc.Note, error)
	RemoveNoteTagFn                func(ctx context.Context, arg sqlc.RemoveNoteTagParams) error
	SearchNotesFn                  func(ctx context.Context, arg sqlc.SearchNotesParams) ([]sqlc.SearchNotesRow, error)
	UpdateConversationTitleFn      func(ctx context.Context, arg sqlc.UpdateConversationTitleParams) (sqlc.Conversation, error)
	UpdateGlossaryTermFn           func(ctx context.Context, arg sqlc.UpdateGlossaryTermParams) (sqlc.GlossaryTerm, error)
	UpdateNoteFn                   func(ctx context.Context, arg sqlc.UpdateNoteParams) (sqlc.Note, error)
	UpdateNoteSourceChatFn         func(ctx context.Context, arg sqlc.UpdateNoteSourceChatParams) (sqlc.Note, error)
	UpdateNoteStatusFn             func(ctx context.Context, arg sqlc.UpdateNoteStatusParams) (sqlc.Note, error)
	UpdateUserFn                   func(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error)
	UpdateUserPreferencesFn        func(ctx context.Context, arg sqlc.UpdateUserPreferencesParams) (sqlc.User, error)
}

func (f *fakeQuerier) AddNoteTag(ctx context.Context, arg sqlc.AddNoteTagParams) error {
	if f.AddNoteTagFn != nil {
		return f.AddNoteTagFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) CountNotesByStatus(ctx context.Context, userID int64) ([]sqlc.CountNotesByStatusRow, error) {
	if f.CountNotesByStatusFn != nil {
		return f.CountNotesByStatusFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeQuerier) CreateConnection(ctx context.Context, arg sqlc.CreateConnectionParams) (sqlc.Connection, error) {
	if f.CreateConnectionFn != nil {
		return f.CreateConnectionFn(ctx, arg)
	}
	return sqlc.Connection{}, nil
}

func (f *fakeQuerier) CreateConversation(ctx context.Context, arg sqlc.CreateConversationParams) (sqlc.Conversation, error) {
	if f.CreateConversationFn != nil {
		return f.CreateConversationFn(ctx, arg)
	}
	return sqlc.Conversation{}, nil
}

func (f *fakeQuerier) CreateGlossaryTerm(ctx context.Context, arg sqlc.CreateGlossaryTermParams) (sqlc.GlossaryTerm, error) {
	if f.CreateGlossaryTermFn != nil {
		return f.CreateGlossaryTermFn(ctx, arg)
	}
	return sqlc.GlossaryTerm{}, nil
}

func (f *fakeQuerier) CreateMessage(ctx context.Context, arg sqlc.CreateMessageParams) (sqlc.Message, error) {
	if f.CreateMessageFn != nil {
		return f.CreateMessageFn(ctx, arg)
	}
	return sqlc.Message{}, nil
}

func (f *fakeQuerier) CreateNote(ctx context.Context, arg sqlc.CreateNoteParams) (sqlc.Note, error) {
	if f.CreateNoteFn != nil {
		return f.CreateNoteFn(ctx, arg)
	}
	return sqlc.Note{}, nil
}

func (f *fakeQuerier) CreateNoteEvent(ctx context.Context, arg sqlc.CreateNoteEventParams) (sqlc.NoteEvent, error) {
	if f.CreateNoteEventFn != nil {
		return f.CreateNoteEventFn(ctx, arg)
	}
	return sqlc.NoteEvent{}, nil
}

func (f *fakeQuerier) CreateTag(ctx context.Context, arg sqlc.CreateTagParams) (sqlc.Tag, error) {
	if f.CreateTagFn != nil {
		return f.CreateTagFn(ctx, arg)
	}
	return sqlc.Tag{}, nil
}

func (f *fakeQuerier) CreateUser(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	if f.CreateUserFn != nil {
		return f.CreateUserFn(ctx, arg)
	}
	return sqlc.User{}, nil
}

func (f *fakeQuerier) DeleteConnection(ctx context.Context, id int64) error {
	if f.DeleteConnectionFn != nil {
		return f.DeleteConnectionFn(ctx, id)
	}
	return nil
}

func (f *fakeQuerier) DeleteConversation(ctx context.Context, arg sqlc.DeleteConversationParams) error {
	if f.DeleteConversationFn != nil {
		return f.DeleteConversationFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) DeleteGlossaryTerm(ctx context.Context, arg sqlc.DeleteGlossaryTermParams) error {
	if f.DeleteGlossaryTermFn != nil {
		return f.DeleteGlossaryTermFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) DeleteNote(ctx context.Context, arg sqlc.DeleteNoteParams) error {
	if f.DeleteNoteFn != nil {
		return f.DeleteNoteFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) DeleteTag(ctx context.Context, arg sqlc.DeleteTagParams) error {
	if f.DeleteTagFn != nil {
		return f.DeleteTagFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) GetConnectionDensity(ctx context.Context, userID int64) (sqlc.GetConnectionDensityRow, error) {
	if f.GetConnectionDensityFn != nil {
		return f.GetConnectionDensityFn(ctx, userID)
	}
	return sqlc.GetConnectionDensityRow{}, nil
}

func (f *fakeQuerier) GetConversation(ctx context.Context, arg sqlc.GetConversationParams) (sqlc.Conversation, error) {
	if f.GetConversationFn != nil {
		return f.GetConversationFn(ctx, arg)
	}
	return sqlc.Conversation{}, nil
}

func (f *fakeQuerier) GetConversationConversionRate(ctx context.Context, userID int64) (sqlc.GetConversationConversionRateRow, error) {
	if f.GetConversationConversionRateFn != nil {
		return f.GetConversationConversionRateFn(ctx, userID)
	}
	return sqlc.GetConversationConversionRateRow{}, nil
}

func (f *fakeQuerier) GetFleetingNotesWithAge(ctx context.Context, userID int64) ([]sqlc.GetFleetingNotesWithAgeRow, error) {
	if f.GetFleetingNotesWithAgeFn != nil {
		return f.GetFleetingNotesWithAgeFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeQuerier) GetGlossaryCount(ctx context.Context, userID int64) (int32, error) {
	if f.GetGlossaryCountFn != nil {
		return f.GetGlossaryCountFn(ctx, userID)
	}
	return 0, nil
}

func (f *fakeQuerier) GetGraphData(ctx context.Context, userID int64) ([]sqlc.GetGraphDataRow, error) {
	if f.GetGraphDataFn != nil {
		return f.GetGraphDataFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeQuerier) GetLifecycleMetrics(ctx context.Context, arg sqlc.GetLifecycleMetricsParams) ([]sqlc.GetLifecycleMetricsRow, error) {
	if f.GetLifecycleMetricsFn != nil {
		return f.GetLifecycleMetricsFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) GetLifecycleTrend(ctx context.Context, arg sqlc.GetLifecycleTrendParams) ([]sqlc.GetLifecycleTrendRow, error) {
	if f.GetLifecycleTrendFn != nil {
		return f.GetLifecycleTrendFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) GetMonthlyUsage(ctx context.Context, userID *int64) (sqlc.GetMonthlyUsageRow, error) {
	if f.GetMonthlyUsageFn != nil {
		return f.GetMonthlyUsageFn(ctx, userID)
	}
	return sqlc.GetMonthlyUsageRow{}, nil
}

func (f *fakeQuerier) GetNoteByID(ctx context.Context, arg sqlc.GetNoteByIDParams) (sqlc.Note, error) {
	if f.GetNoteByIDFn != nil {
		return f.GetNoteByIDFn(ctx, arg)
	}
	return sqlc.Note{}, nil
}

func (f *fakeQuerier) GetNoteBySourceChat(ctx context.Context, arg sqlc.GetNoteBySourceChatParams) (sqlc.Note, error) {
	if f.GetNoteBySourceChatFn != nil {
		return f.GetNoteBySourceChatFn(ctx, arg)
	}
	return sqlc.Note{}, nil
}

func (f *fakeQuerier) GetNoteCurrentStatus(ctx context.Context, arg sqlc.GetNoteCurrentStatusParams) (sqlc.NoteStatus, error) {
	if f.GetNoteCurrentStatusFn != nil {
		return f.GetNoteCurrentStatusFn(ctx, arg)
	}
	return "", nil
}

func (f *fakeQuerier) GetNoteTypeDistribution(ctx context.Context, userID int64) ([]sqlc.GetNoteTypeDistributionRow, error) {
	if f.GetNoteTypeDistributionFn != nil {
		return f.GetNoteTypeDistributionFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeQuerier) GetSleepingNoteStats(ctx context.Context, userID int64) (sqlc.GetSleepingNoteStatsRow, error) {
	if f.GetSleepingNoteStatsFn != nil {
		return f.GetSleepingNoteStatsFn(ctx, userID)
	}
	return sqlc.GetSleepingNoteStatsRow{}, nil
}

func (f *fakeQuerier) GetStaleNotes(ctx context.Context, arg sqlc.GetStaleNotesParams) ([]sqlc.GetStaleNotesRow, error) {
	if f.GetStaleNotesFn != nil {
		return f.GetStaleNotesFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) GetTagsForNote(ctx context.Context, noteID int64) ([]sqlc.Tag, error) {
	if f.GetTagsForNoteFn != nil {
		return f.GetTagsForNoteFn(ctx, noteID)
	}
	return nil, nil
}

func (f *fakeQuerier) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	if f.GetUserByEmailFn != nil {
		return f.GetUserByEmailFn(ctx, email)
	}
	return sqlc.User{}, nil
}

func (f *fakeQuerier) GetUserByID(ctx context.Context, id int64) (sqlc.User, error) {
	if f.GetUserByIDFn != nil {
		return f.GetUserByIDFn(ctx, id)
	}
	return sqlc.User{}, nil
}

func (f *fakeQuerier) GetUserByOIDCSubject(ctx context.Context, oidcSubject string) (sqlc.User, error) {
	if f.GetUserByOIDCSubjectFn != nil {
		return f.GetUserByOIDCSubjectFn(ctx, oidcSubject)
	}
	return sqlc.User{}, nil
}

func (f *fakeQuerier) ListConnectionsForNote(ctx context.Context, sourceID int64) ([]sqlc.ListConnectionsForNoteRow, error) {
	if f.ListConnectionsForNoteFn != nil {
		return f.ListConnectionsForNoteFn(ctx, sourceID)
	}
	return nil, nil
}

func (f *fakeQuerier) ListConversations(ctx context.Context, arg sqlc.ListConversationsParams) ([]sqlc.Conversation, error) {
	if f.ListConversationsFn != nil {
		return f.ListConversationsFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) ListGlossaryTerms(ctx context.Context, userID int64) ([]sqlc.ListGlossaryTermsRow, error) {
	if f.ListGlossaryTermsFn != nil {
		return f.ListGlossaryTermsFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeQuerier) ListMessages(ctx context.Context, conversationID int64) ([]sqlc.Message, error) {
	if f.ListMessagesFn != nil {
		return f.ListMessagesFn(ctx, conversationID)
	}
	return nil, nil
}

func (f *fakeQuerier) ListNotesByStatus(ctx context.Context, arg sqlc.ListNotesByStatusParams) ([]sqlc.Note, error) {
	if f.ListNotesByStatusFn != nil {
		return f.ListNotesByStatusFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) ListNotesByTag(ctx context.Context, arg sqlc.ListNotesByTagParams) ([]sqlc.Note, error) {
	if f.ListNotesByTagFn != nil {
		return f.ListNotesByTagFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) ListNotesByUser(ctx context.Context, arg sqlc.ListNotesByUserParams) ([]sqlc.ListNotesByUserRow, error) {
	if f.ListNotesByUserFn != nil {
		return f.ListNotesByUserFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) ListTagsByUser(ctx context.Context, userID int64) ([]sqlc.ListTagsByUserRow, error) {
	if f.ListTagsByUserFn != nil {
		return f.ListTagsByUserFn(ctx, userID)
	}
	return nil, nil
}

func (f *fakeQuerier) LogAIUsage(ctx context.Context, arg sqlc.LogAIUsageParams) (sqlc.AiUsageLog, error) {
	if f.LogAIUsageFn != nil {
		return f.LogAIUsageFn(ctx, arg)
	}
	return sqlc.AiUsageLog{}, nil
}

func (f *fakeQuerier) RecentNotes(ctx context.Context, arg sqlc.RecentNotesParams) ([]sqlc.Note, error) {
	if f.RecentNotesFn != nil {
		return f.RecentNotesFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) RemoveNoteTag(ctx context.Context, arg sqlc.RemoveNoteTagParams) error {
	if f.RemoveNoteTagFn != nil {
		return f.RemoveNoteTagFn(ctx, arg)
	}
	return nil
}

func (f *fakeQuerier) SearchNotes(ctx context.Context, arg sqlc.SearchNotesParams) ([]sqlc.SearchNotesRow, error) {
	if f.SearchNotesFn != nil {
		return f.SearchNotesFn(ctx, arg)
	}
	return nil, nil
}

func (f *fakeQuerier) UpdateConversationTitle(ctx context.Context, arg sqlc.UpdateConversationTitleParams) (sqlc.Conversation, error) {
	if f.UpdateConversationTitleFn != nil {
		return f.UpdateConversationTitleFn(ctx, arg)
	}
	return sqlc.Conversation{}, nil
}

func (f *fakeQuerier) UpdateGlossaryTerm(ctx context.Context, arg sqlc.UpdateGlossaryTermParams) (sqlc.GlossaryTerm, error) {
	if f.UpdateGlossaryTermFn != nil {
		return f.UpdateGlossaryTermFn(ctx, arg)
	}
	return sqlc.GlossaryTerm{}, nil
}

func (f *fakeQuerier) UpdateNote(ctx context.Context, arg sqlc.UpdateNoteParams) (sqlc.Note, error) {
	if f.UpdateNoteFn != nil {
		return f.UpdateNoteFn(ctx, arg)
	}
	return sqlc.Note{}, nil
}

func (f *fakeQuerier) UpdateNoteSourceChat(ctx context.Context, arg sqlc.UpdateNoteSourceChatParams) (sqlc.Note, error) {
	if f.UpdateNoteSourceChatFn != nil {
		return f.UpdateNoteSourceChatFn(ctx, arg)
	}
	return sqlc.Note{}, nil
}

func (f *fakeQuerier) UpdateNoteStatus(ctx context.Context, arg sqlc.UpdateNoteStatusParams) (sqlc.Note, error) {
	if f.UpdateNoteStatusFn != nil {
		return f.UpdateNoteStatusFn(ctx, arg)
	}
	return sqlc.Note{}, nil
}

func (f *fakeQuerier) UpdateUser(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	if f.UpdateUserFn != nil {
		return f.UpdateUserFn(ctx, arg)
	}
	return sqlc.User{}, nil
}

func (f *fakeQuerier) UpdateUserPreferences(ctx context.Context, arg sqlc.UpdateUserPreferencesParams) (sqlc.User, error) {
	if f.UpdateUserPreferencesFn != nil {
		return f.UpdateUserPreferencesFn(ctx, arg)
	}
	return sqlc.User{}, nil
}

// mockProvider implements ai.Provider for testing.
type mockProvider struct {
	StreamChatFn func(ctx context.Context, messages []ai.ChatMessage, systemPrompt string, cb ai.StreamCallback) (*ai.ChatResult, error)
}

func (m *mockProvider) StreamChat(ctx context.Context, messages []ai.ChatMessage, systemPrompt string, cb ai.StreamCallback) (*ai.ChatResult, error) {
	if m.StreamChatFn != nil {
		return m.StreamChatFn(ctx, messages, systemPrompt, cb)
	}
	return &ai.ChatResult{}, nil
}

// reqWithUser builds an httptest.Request with the given user ID set in context.
func reqWithUser(method, path string, body io.Reader, userID int64) *http.Request {
	r := httptest.NewRequest(method, path, body)
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

// reqWithUserAndChiCtx builds an httptest.Request with user ID and Chi URL params.
func reqWithUserAndChiCtx(method, path string, body io.Reader, userID int64, params map[string]string) *http.Request {
	r := reqWithUser(method, path, body, userID)
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, rctx)
	return r.WithContext(ctx)
}

// newTestMetrics creates a Metrics instance using an isolated Prometheus registry
// so tests don't conflict with each other or the default registry.
func newTestMetrics() *metrics.Metrics {
	reg := prometheus.NewRegistry()

	m := &metrics.Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_http_requests_total",
		}, []string{"method", "route", "status_code"}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "test_http_request_duration_seconds",
		}, []string{"method", "route"}),
		HTTPRequestsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "test_http_requests_in_flight",
		}),
		AITokensTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_ai_tokens_total",
		}, []string{"type", "model"}),
		AIRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "test_ai_request_duration_seconds",
		}, []string{"model"}),
		AIRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_ai_requests_total",
		}, []string{"model", "status"}),
	}

	reg.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.AITokensTotal,
		m.AIRequestDuration,
		m.AIRequestsTotal,
	)

	return m
}

// sampleNote returns a Note with sensible defaults for testing.
func sampleNote(id, userID int64) sqlc.Note {
	return sqlc.Note{
		ID:       id,
		UserID:   userID,
		Title:    "Test Note",
		Summary:  "A test summary",
		Body:     "Test body content",
		NoteType: sqlc.NoteTypeConcept,
		Status:   sqlc.NoteStatusFleeting,
		Tlp:      sqlc.NoteTlpClear,
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// decodeJSON is a test helper that decodes JSON from an http.Response body.
func decodeJSON(t interface{ Fatal(...any) }, body io.Reader, v any) {
	if err := json.NewDecoder(body).Decode(v); err != nil {
		t.Fatal("failed to decode JSON response:", err)
	}
}
