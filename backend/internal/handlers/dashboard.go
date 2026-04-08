package handlers

import (
	"net/http"

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
