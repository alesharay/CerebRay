package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/middleware"
)

// NoteHandlers contains handlers for note operations.
type NoteHandlers struct {
	queries *sqlc.Queries
}

func NewNoteHandlers(q *sqlc.Queries) *NoteHandlers {
	return &NoteHandlers{queries: q}
}

func (h *NoteHandlers) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	limit := QueryInt(r, "limit", 50)
	offset := QueryInt(r, "offset", 0)
	status := r.URL.Query().Get("status")
	tag := r.URL.Query().Get("tag")

	if tag != "" {
		notes, err := h.queries.ListNotesByTag(r.Context(), sqlc.ListNotesByTagParams{
			UserID: userID,
			Name:   tag,
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to list notes")
			return
		}
		JSON(w, http.StatusOK, notes)
		return
	}

	if status != "" {
		notes, err := h.queries.ListNotesByStatus(r.Context(), sqlc.ListNotesByStatusParams{
			UserID: userID,
			Status: sqlc.NoteStatus(status),
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			Error(w, http.StatusInternalServerError, "failed to list notes")
			return
		}
		JSON(w, http.StatusOK, notes)
		return
	}

	notes, err := h.queries.ListNotesByUser(r.Context(), sqlc.ListNotesByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list notes")
		return
	}
	JSON(w, http.StatusOK, notes)
}

func (h *NoteHandlers) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	note, err := h.queries.GetNoteByID(r.Context(), sqlc.GetNoteByIDParams{
		ID:     noteID,
		UserID: userID,
	})
	if err != nil {
		Error(w, http.StatusNotFound, "note not found")
		return
	}
	JSON(w, http.StatusOK, note)
}

type createNoteRequest struct {
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	LaymansTerms string `json:"laymans_terms"`
	Analogy      string `json:"analogy"`
	CoreIdea     string `json:"core_idea"`
	Body         string `json:"body"`
	Components   string `json:"components"`
	WhyItMatters string `json:"why_it_matters"`
	Examples     string `json:"examples"`
	Templates    string `json:"templates"`
	Additional   string `json:"additional"`
	NoteType     string `json:"note_type"`
	Status       string `json:"status"`
	TLP          string `json:"tlp"`
	SourceChatID *int64 `json:"source_chat_id"`
}

func (h *NoteHandlers) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		Error(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.NoteType == "" {
		req.NoteType = "concept"
	}
	if req.Status == "" {
		req.Status = "fleeting"
	}
	if req.TLP == "" {
		req.TLP = "clear"
	}

	note, err := h.queries.CreateNote(r.Context(), sqlc.CreateNoteParams{
		UserID:       userID,
		Title:        req.Title,
		Summary:      req.Summary,
		LaymansTerms: req.LaymansTerms,
		Analogy:      req.Analogy,
		CoreIdea:     req.CoreIdea,
		Body:         req.Body,
		Components:   req.Components,
		WhyItMatters: req.WhyItMatters,
		Examples:     req.Examples,
		Templates:    req.Templates,
		Additional:   req.Additional,
		NoteType:     sqlc.NoteType(req.NoteType),
		Status:       sqlc.NoteStatus(req.Status),
		Tlp:          sqlc.NoteTlp(req.TLP),
		SourceChatID: req.SourceChatID,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create note")
		return
	}

	// Log the creation event
	h.queries.CreateNoteEvent(r.Context(), sqlc.CreateNoteEventParams{
		NoteID:     note.ID,
		UserID:     userID,
		FromStatus: sqlc.NullNoteStatus{Valid: false},
		ToStatus:   note.Status,
	})

	JSON(w, http.StatusCreated, note)
}

func (h *NoteHandlers) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	var req createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	note, err := h.queries.UpdateNote(r.Context(), sqlc.UpdateNoteParams{
		ID:           noteID,
		UserID:       userID,
		Title:        req.Title,
		Summary:      req.Summary,
		LaymansTerms: req.LaymansTerms,
		Analogy:      req.Analogy,
		CoreIdea:     req.CoreIdea,
		Body:         req.Body,
		Components:   req.Components,
		WhyItMatters: req.WhyItMatters,
		Examples:     req.Examples,
		Templates:    req.Templates,
		Additional:   req.Additional,
		NoteType:     sqlc.NoteType(req.NoteType),
		Status:       sqlc.NoteStatus(req.Status),
		Tlp:          sqlc.NoteTlp(req.TLP),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to update note")
		return
	}
	JSON(w, http.StatusOK, note)
}

func (h *NoteHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	if err := h.queries.DeleteNote(r.Context(), sqlc.DeleteNoteParams{
		ID:     noteID,
		UserID: userID,
	}); err != nil {
		Error(w, http.StatusInternalServerError, "failed to delete note")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NoteHandlers) transitionStatus(w http.ResponseWriter, r *http.Request, toStatus sqlc.NoteStatus, errMsg string) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	// Capture current status before the transition
	fromStatus, err := h.queries.GetNoteCurrentStatus(r.Context(), sqlc.GetNoteCurrentStatusParams{
		ID:     noteID,
		UserID: userID,
	})
	if err != nil {
		Error(w, http.StatusNotFound, "note not found")
		return
	}

	note, err := h.queries.UpdateNoteStatus(r.Context(), sqlc.UpdateNoteStatusParams{
		ID:     noteID,
		UserID: userID,
		Status: toStatus,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, errMsg)
		return
	}

	// Log the transition event
	h.queries.CreateNoteEvent(r.Context(), sqlc.CreateNoteEventParams{
		NoteID:     noteID,
		UserID:     userID,
		FromStatus: sqlc.NullNoteStatus{NoteStatus: fromStatus, Valid: true},
		ToStatus:   toStatus,
	})

	JSON(w, http.StatusOK, note)
}

func (h *NoteHandlers) Promote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	noteID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note ID")
		return
	}

	// Fetch the note to get the title and current status
	note, err := h.queries.GetNoteByID(r.Context(), sqlc.GetNoteByIDParams{
		ID:     noteID,
		UserID: userID,
	})
	if err != nil {
		Error(w, http.StatusNotFound, "note not found")
		return
	}

	fromStatus := note.Status

	// Create a conversation linked to this note
	convo, err := h.queries.CreateConversation(r.Context(), sqlc.CreateConversationParams{
		UserID: userID,
		Title:  note.Title,
		Topic:  note.Title,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}

	// Link the conversation to the note
	_, err = h.queries.UpdateNoteSourceChat(r.Context(), sqlc.UpdateNoteSourceChatParams{
		ID:           noteID,
		UserID:       userID,
		SourceChatID: &convo.ID,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to link conversation")
		return
	}

	// Transition status to active
	updated, err := h.queries.UpdateNoteStatus(r.Context(), sqlc.UpdateNoteStatusParams{
		ID:     noteID,
		UserID: userID,
		Status: sqlc.NoteStatusActive,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to promote note")
		return
	}

	// Log the transition event
	h.queries.CreateNoteEvent(r.Context(), sqlc.CreateNoteEventParams{
		NoteID:     noteID,
		UserID:     userID,
		FromStatus: sqlc.NullNoteStatus{NoteStatus: fromStatus, Valid: true},
		ToStatus:   sqlc.NoteStatusActive,
	})

	JSON(w, http.StatusOK, map[string]interface{}{
		"note":            updated,
		"conversation_id": convo.ID,
	})
}

func (h *NoteHandlers) Sleep(w http.ResponseWriter, r *http.Request) {
	h.transitionStatus(w, r, sqlc.NoteStatusSleeping, "failed to move note to echoes")
}

func (h *NoteHandlers) Archive(w http.ResponseWriter, r *http.Request) {
	h.transitionStatus(w, r, sqlc.NoteStatusArchived, "failed to archive note")
}

func (h *NoteHandlers) Search(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	query := r.URL.Query().Get("q")
	if query == "" {
		Error(w, http.StatusBadRequest, "q parameter required")
		return
	}
	limit := QueryInt(r, "limit", 50)
	offset := QueryInt(r, "offset", 0)

	notes, err := h.queries.SearchNotes(r.Context(), sqlc.SearchNotesParams{
		UserID:         userID,
		PlaintoTsquery: query,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "search failed")
		return
	}
	JSON(w, http.StatusOK, notes)
}
