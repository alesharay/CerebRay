package handlers

import (
	"net/http"
	"time"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/middleware"
)

type DashboardHandlers struct {
	queries *sqlc.Queries
}

func NewDashboardHandlers(q *sqlc.Queries) *DashboardHandlers {
	return &DashboardHandlers{queries: q}
}

func (h *DashboardHandlers) Stats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	counts, err := h.queries.CountNotesByStatus(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	stats := map[string]int64{
		"inbox":  0,
		"echoes": 0,
		"codex":  0,
	}
	for _, c := range counts {
		switch c.Status {
		case sqlc.NoteStatusFleeting:
			stats["inbox"] = c.Count
		case sqlc.NoteStatusSleeping:
			stats["echoes"] = c.Count
		case sqlc.NoteStatusActive, sqlc.NoteStatusLinked:
			stats["codex"] += c.Count
		}
	}

	recent, err := h.queries.RecentNotes(r.Context(), sqlc.RecentNotesParams{
		UserID: userID,
		Limit:  10,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get recent notes")
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"stats":  stats,
		"recent": recent,
	})
}

// readinessFields are the note fields that contribute to the readiness score.
var readinessFields = 10

func readinessScore(n sqlc.GetFleetingNotesWithAgeRow) float64 {
	filled := 0
	if n.Summary != "" {
		filled++
	}
	if n.CoreIdea != "" {
		filled++
	}
	if n.Body != "" {
		filled++
	}
	if n.LaymansTerms != "" {
		filled++
	}
	if n.Analogy != "" {
		filled++
	}
	if n.Components != "" {
		filled++
	}
	if n.WhyItMatters != "" {
		filled++
	}
	if n.Examples != "" {
		filled++
	}
	if n.Templates != "" {
		filled++
	}
	if n.Additional != "" {
		filled++
	}
	return float64(filled) / float64(readinessFields)
}

type inboxItem struct {
	ID             int64    `json:"id"`
	Title          string   `json:"title"`
	NoteType       string   `json:"note_type"`
	AgeSeconds     int64    `json:"age_seconds"`
	ReadinessScore float64  `json:"readiness_score"`
	CreatedAt      string   `json:"created_at"`
}

type lifecycleEntry struct {
	Action          string `json:"action"`
	Count           int64  `json:"count"`
	AvgDwellSeconds int64  `json:"avg_dwell_seconds"`
}

type strengthScore struct {
	Overall           float64 `json:"overall"`
	ConnectionDensity float64 `json:"connection_density"`
	OrphanCount       int32   `json:"orphan_count"`
	ActiveNotes       int32   `json:"active_notes"`
	TotalConnections  int32   `json:"total_connections"`
	GlossaryCount     int32   `json:"glossary_count"`
	GlossaryCoverage  float64 `json:"glossary_coverage"`
	SleepingBacklog   int32   `json:"sleeping_backlog"`
	SleepingAvgAge    int64   `json:"sleeping_avg_age_seconds"`
	TypeDistribution  []typeCount `json:"type_distribution"`
}

type typeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

type conversionRate struct {
	TotalConversations     int32   `json:"total_conversations"`
	ConversationsWithNotes int32   `json:"conversations_with_notes"`
	Rate                   float64 `json:"rate"`
}

type staleNote struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	NoteType  string `json:"note_type"`
	UpdatedAt string `json:"updated_at"`
	StaleDays int    `json:"stale_days"`
}

type lifecycleTrendPoint struct {
	Week   string `json:"week"`
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

type analyticsResponse struct {
	Inbox          []inboxItem           `json:"inbox"`
	Lifecycle      []lifecycleEntry      `json:"lifecycle"`
	LifecycleTrend []lifecycleTrendPoint  `json:"lifecycle_trend"`
	Strength       strengthScore         `json:"strength"`
	Conversion     conversionRate        `json:"conversion"`
	AIUsage        aiUsageData           `json:"ai_usage"`
	StaleNotes     []staleNote           `json:"stale_notes"`
}

type aiUsageData struct {
	InputTokens  int32 `json:"input_tokens"`
	OutputTokens int32 `json:"output_tokens"`
	Requests     int32 `json:"requests"`
}

func (h *DashboardHandlers) Analytics(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	ctx := r.Context()
	now := time.Now()
	staleDays := int32(QueryInt(r, "stale_days", 14))

	// Inbox items with readiness
	fleeting, err := h.queries.GetFleetingNotesWithAge(ctx, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get inbox")
		return
	}
	inbox := make([]inboxItem, len(fleeting))
	for i, n := range fleeting {
		inbox[i] = inboxItem{
			ID:             n.ID,
			Title:          n.Title,
			NoteType:       string(n.NoteType),
			AgeSeconds:     int64(now.Sub(n.CreatedAt).Seconds()),
			ReadinessScore: readinessScore(n),
			CreatedAt:      n.CreatedAt.Format(time.RFC3339),
		}
	}

	// Lifecycle metrics (last 90 days)
	since := now.AddDate(0, 0, -90)
	lcRows, err := h.queries.GetLifecycleMetrics(ctx, sqlc.GetLifecycleMetricsParams{
		UserID:    userID,
		CreatedAt: since,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get lifecycle metrics")
		return
	}
	lifecycle := make([]lifecycleEntry, len(lcRows))
	for i, row := range lcRows {
		action := "unknown"
		switch row.ToStatus {
		case sqlc.NoteStatusActive:
			action = "promoted"
		case sqlc.NoteStatusSleeping:
			action = "slept"
		case sqlc.NoteStatusArchived:
			action = "archived"
		}
		lifecycle[i] = lifecycleEntry{
			Action:          action,
			Count:           row.TransitionCount,
			AvgDwellSeconds: row.AvgDwellSeconds,
		}
	}

	// Lifecycle trend (weekly counts over last 90 days)
	trendRows, err := h.queries.GetLifecycleTrend(ctx, sqlc.GetLifecycleTrendParams{
		UserID:    userID,
		CreatedAt: since,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get lifecycle trend")
		return
	}
	trend := make([]lifecycleTrendPoint, len(trendRows))
	for i, row := range trendRows {
		action := "unknown"
		switch row.ToStatus {
		case sqlc.NoteStatusActive:
			action = "promoted"
		case sqlc.NoteStatusSleeping:
			action = "slept"
		case sqlc.NoteStatusArchived:
			action = "archived"
		}
		trend[i] = lifecycleTrendPoint{
			Week:   row.Week.Format("2006-01-02"),
			Action: action,
			Count:  row.Count,
		}
	}

	// Connection density and orphans
	connDensity, err := h.queries.GetConnectionDensity(ctx, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get connection density")
		return
	}

	// Note type distribution
	typeDist, err := h.queries.GetNoteTypeDistribution(ctx, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get type distribution")
		return
	}
	types := make([]typeCount, len(typeDist))
	for i, t := range typeDist {
		types[i] = typeCount{Type: string(t.NoteType), Count: t.Count}
	}

	// Glossary coverage
	glossaryCount, err := h.queries.GetGlossaryCount(ctx, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get glossary count")
		return
	}

	// Sleeping note backlog
	sleepStats, err := h.queries.GetSleepingNoteStats(ctx, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get sleeping stats")
		return
	}

	// Compute overall strength score (0-100)
	var density float64
	if connDensity.ActiveNotes > 0 {
		density = float64(connDensity.TotalConnections) / float64(connDensity.ActiveNotes)
	}
	var glossaryCoverage float64
	if connDensity.ActiveNotes > 0 {
		glossaryCoverage = float64(glossaryCount) / float64(connDensity.ActiveNotes)
		if glossaryCoverage > 1 {
			glossaryCoverage = 1
		}
	}
	orphanPenalty := 0.0
	if connDensity.ActiveNotes > 0 {
		orphanPenalty = float64(connDensity.OrphanCount) / float64(connDensity.ActiveNotes)
	}
	typeDiv := float64(len(typeDist)) / 8.0 // 8 possible note types
	if typeDiv > 1 {
		typeDiv = 1
	}

	// Weighted score: density (35%), orphan penalty (25%), glossary (20%), diversity (20%)
	densityCapped := density
	if densityCapped > 3 {
		densityCapped = 3
	}
	overall := (densityCapped/3.0)*35 +
		(1-orphanPenalty)*25 +
		glossaryCoverage*20 +
		typeDiv*20

	var sleepAvgAge int64
	if v, ok := sleepStats.AvgAgeSeconds.(int64); ok {
		sleepAvgAge = v
	}

	strength := strengthScore{
		Overall:           overall,
		ConnectionDensity: density,
		OrphanCount:       connDensity.OrphanCount,
		ActiveNotes:       connDensity.ActiveNotes,
		TotalConnections:  connDensity.TotalConnections,
		GlossaryCount:     glossaryCount,
		GlossaryCoverage:  glossaryCoverage,
		SleepingBacklog:   sleepStats.Count,
		SleepingAvgAge:    sleepAvgAge,
		TypeDistribution:  types,
	}

	// Conversation conversion
	convRate, err := h.queries.GetConversationConversionRate(ctx, userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get conversion rate")
		return
	}
	var rate float64
	if convRate.TotalConversations > 0 {
		rate = float64(convRate.ConversationsWithNotes) / float64(convRate.TotalConversations)
	}

	// AI usage
	usage, err := h.queries.GetMonthlyUsage(ctx, &userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get AI usage")
		return
	}

	// Stale notes
	staleRows, err := h.queries.GetStaleNotes(ctx, sqlc.GetStaleNotesParams{
		UserID:  userID,
		Column2: staleDays,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get stale notes")
		return
	}
	stale := make([]staleNote, len(staleRows))
	for i, s := range staleRows {
		stale[i] = staleNote{
			ID:        s.ID,
			Title:     s.Title,
			NoteType:  string(s.NoteType),
			UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
			StaleDays: int(now.Sub(s.UpdatedAt).Hours() / 24),
		}
	}

	JSON(w, http.StatusOK, analyticsResponse{
		Inbox:          inbox,
		Lifecycle:      lifecycle,
		LifecycleTrend: trend,
		Strength:  strength,
		Conversion: conversionRate{
			TotalConversations:     convRate.TotalConversations,
			ConversationsWithNotes: convRate.ConversationsWithNotes,
			Rate:                   rate,
		},
		AIUsage: aiUsageData{
			InputTokens:  usage.TotalInputTokens,
			OutputTokens: usage.TotalOutputTokens,
			Requests:     usage.TotalRequests,
		},
		StaleNotes: stale,
	})
}
