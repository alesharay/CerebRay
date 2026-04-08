package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/middleware"
)

type ConversationHandlers struct {
	queries *sqlc.Queries
}

func NewConversationHandlers(q *sqlc.Queries) *ConversationHandlers {
	return &ConversationHandlers{queries: q}
}

func (h *ConversationHandlers) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	limit := QueryInt(r, "limit", 50)
	offset := QueryInt(r, "offset", 0)

	convos, err := h.queries.ListConversations(r.Context(), sqlc.ListConversationsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list conversations")
		return
	}
	JSON(w, http.StatusOK, convos)
}

func (h *ConversationHandlers) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	convoID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid conversation ID")
		return
	}

	convo, err := h.queries.GetConversation(r.Context(), sqlc.GetConversationParams{
		ID:     convoID,
		UserID: userID,
	})
	if err != nil {
		Error(w, http.StatusNotFound, "conversation not found")
		return
	}

	messages, err := h.queries.ListMessages(r.Context(), convoID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to list messages")
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"conversation": convo,
		"messages":     messages,
	})
}

type createConversationRequest struct {
	Title string `json:"title"`
	Topic string `json:"topic"`
}

func (h *ConversationHandlers) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req createConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		req.Title = "New Conversation"
	}

	convo, err := h.queries.CreateConversation(r.Context(), sqlc.CreateConversationParams{
		UserID: userID,
		Title:  req.Title,
		Topic:  req.Topic,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to create conversation")
		return
	}
	JSON(w, http.StatusCreated, convo)
}

func (h *ConversationHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	convoID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid conversation ID")
		return
	}

	if err := h.queries.DeleteConversation(r.Context(), sqlc.DeleteConversationParams{
		ID:     convoID,
		UserID: userID,
	}); err != nil {
		Error(w, http.StatusInternalServerError, "failed to delete conversation")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
