package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/middleware"
)

type TagHandlers struct {
	queries sqlc.Querier
}

func NewTagHandlers(q sqlc.Querier) *TagHandlers {
	return &TagHandlers{queries: q}
}

func (h *TagHandlers) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	tags, err := h.queries.ListTagsByUser(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list tags")
		return
	}
	JSON(w, http.StatusOK, tags)
}

type addTagRequest struct {
	Name string `json:"name"`
}

func (h *TagHandlers) AddToNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	var req addTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		Error(w, http.StatusBadRequest, "tag name is required")
		return
	}

	// Confirm the note is the caller's before creating anything. AddNoteTag
	// carries the same guard in SQL, but checking here lets a foreign note
	// return 404 instead of looking like a silent no-op.
	if _, err := h.queries.GetNoteByID(r.Context(), sqlc.GetNoteByIDParams{
		ID:     noteID,
		UserID: userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			Error(w, http.StatusNotFound, "note not found")
			return
		}
		Error(w, http.StatusInternalServerError, "failed to load note")
		return
	}

	tag, err := h.queries.CreateTag(r.Context(), sqlc.CreateTagParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create tag")
		return
	}

	// ON CONFLICT DO NOTHING returns no rows when the tag is already attached,
	// which is a successful no-op rather than a failure.
	if _, err := h.queries.AddNoteTag(r.Context(), sqlc.AddNoteTagParams{
		NoteID: noteID,
		TagID:  tag.ID,
		UserID: userID,
	}); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		Error(w, http.StatusInternalServerError, "failed to add tag to note")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandlers) RemoveFromNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}
	tagID, err := URLParamInt64(r, "tagId")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid tag ID")
		return
	}

	if err := h.queries.RemoveNoteTag(r.Context(), sqlc.RemoveNoteTagParams{
		NoteID: noteID,
		TagID:  tagID,
		UserID: userID,
	}); err != nil {
		Error(w, http.StatusInternalServerError, "failed to remove tag")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
