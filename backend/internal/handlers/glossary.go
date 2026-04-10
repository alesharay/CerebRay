package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/middleware"
)

type GlossaryHandlers struct {
	queries sqlc.Querier
}

func NewGlossaryHandlers(q sqlc.Querier) *GlossaryHandlers {
	return &GlossaryHandlers{queries: q}
}

func (h *GlossaryHandlers) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	terms, err := h.queries.ListGlossaryTerms(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list glossary terms")
		return
	}
	JSON(w, http.StatusOK, terms)
}

type glossaryTermRequest struct {
	Term         string `json:"term"`
	Definition   string `json:"definition"`
	SourceNoteID *int64 `json:"source_note_id"`
}

func (h *GlossaryHandlers) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req glossaryTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	term, err := h.queries.CreateGlossaryTerm(r.Context(), sqlc.CreateGlossaryTermParams{
		UserID:       userID,
		Term:         req.Term,
		Definition:   req.Definition,
		SourceNoteID: req.SourceNoteID,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create glossary term")
		return
	}
	JSON(w, http.StatusCreated, term)
}

func (h *GlossaryHandlers) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	termID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid term ID")
		return
	}

	var req glossaryTermRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	term, err := h.queries.UpdateGlossaryTerm(r.Context(), sqlc.UpdateGlossaryTermParams{
		ID:           termID,
		UserID:       userID,
		Term:         req.Term,
		Definition:   req.Definition,
		SourceNoteID: req.SourceNoteID,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to update glossary term")
		return
	}
	JSON(w, http.StatusOK, term)
}

func (h *GlossaryHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	termID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid term ID")
		return
	}

	if err := h.queries.DeleteGlossaryTerm(r.Context(), sqlc.DeleteGlossaryTermParams{
		ID:     termID,
		UserID: userID,
	}); err != nil {
		Error(w, http.StatusInternalServerError, "failed to delete glossary term")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
