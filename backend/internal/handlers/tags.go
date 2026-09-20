package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

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

	tag, err := h.queries.CreateTag(r.Context(), sqlc.CreateTagParams{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create tag")
		return
	}
	if err := h.queries.AddNoteTag(r.Context(), sqlc.AddNoteTagParams{
		NoteID: noteID,
		TagID:  tag.ID,
	}); err != nil {
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
