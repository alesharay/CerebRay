package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aray/cerebray/backend/db/sqlc"
	"github.com/aray/cerebray/backend/internal/ai"
	"github.com/aray/cerebray/backend/internal/metrics"
	"github.com/aray/cerebray/backend/internal/middleware"
)

// ChatHandlers manages AI-powered chat endpoints.
type ChatHandlers struct {
	queries  sqlc.Querier
	provider ai.Provider
	metrics  *metrics.Metrics
}

// NewChatHandlers creates a new ChatHandlers instance.
func NewChatHandlers(q sqlc.Querier, provider ai.Provider, m *metrics.Metrics) *ChatHandlers {
	return &ChatHandlers{queries: q, provider: provider, metrics: m}
}

type sendMessageRequest struct {
	Content string `json:"content"`
}

// SendMessage handles POST /api/v1/conversations/{id}/messages.
// It stores the user message, builds the prompt with conversation history,
// streams the AI response via SSE, and stores the assistant reply when done.
func (h *ChatHandlers) SendMessage(w http.ResponseWriter, r *http.Request) {
	if h.provider == nil {
		Error(w, http.StatusServiceUnavailable, "AI not configured")
		return
	}

	userID := middleware.GetUserID(r.Context())
	convoID, err := URLParamInt64(r, "id")
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid conversation ID")
		return
	}

	// Verify conversation belongs to user
	_, err = h.queries.GetConversation(r.Context(), sqlc.GetConversationParams{
		ID:     convoID,
		UserID: userID,
	})
	if err != nil {
		Error(w, http.StatusNotFound, "conversation not found")
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Content == "" {
		Error(w, http.StatusBadRequest, "content is required")
		return
	}

	// Store user message (zero tokens, no model for user messages)
	_, err = h.queries.CreateMessage(r.Context(), sqlc.CreateMessageParams{
		ConversationID: convoID,
		Role:           "user",
		Content:        req.Content,
		InputTokens:    0,
		OutputTokens:   0,
		Model:          "",
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to store message")
		return
	}

	// Load full conversation history
	dbMessages, err := h.queries.ListMessages(r.Context(), convoID)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to load messages")
		return
	}

	// Convert to AI message format
	var chatMessages []ai.ChatMessage
	for _, m := range dbMessages {
		chatMessages = append(chatMessages, ai.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// Build system prompt with user's existing tags and notes for context
	tags, _ := h.queries.ListTagsByUser(r.Context(), userID)
	var tagNames []string
	for _, t := range tags {
		tagNames = append(tagNames, t.Name)
	}

	recentNotes, _ := h.queries.RecentNotes(r.Context(), sqlc.RecentNotesParams{
		UserID: userID,
		Limit:  20,
	})
	var noteTitles []string
	for _, n := range recentNotes {
		noteTitles = append(noteTitles, n.Title)
	}

	// Check if this is the first message in a note-linked conversation (expand mode)
	var systemPrompt string
	linkedNote, noteErr := h.queries.GetNoteBySourceChat(r.Context(), sqlc.GetNoteBySourceChatParams{
		SourceChatID: &convoID,
		UserID:       userID,
	})
	if noteErr == nil && len(dbMessages) == 1 {
		// First message in a promoted note's conversation - use expand prompt
		rawThought := linkedNote.Title
		if linkedNote.Body != "" {
			rawThought = linkedNote.Title + "\n\n" + linkedNote.Body
		}
		systemPrompt = ai.BuildExpandPrompt(rawThought, tagNames, noteTitles)
	} else {
		systemPrompt = ai.BuildSystemPrompt(tagNames, noteTitles)
	}

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		Error(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	var fullResponse string
	rlog := middleware.RequestLogger(r.Context())

	// Stream the AI response, sending each chunk as an SSE event
	aiStart := time.Now()
	result, err := h.provider.StreamChat(r.Context(), chatMessages, systemPrompt, func(chunk string) error {
		fullResponse += chunk
		data, _ := json.Marshal(map[string]string{"delta": chunk})
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		return nil
	})
	aiDuration := time.Since(aiStart).Seconds()

	if err != nil {
		rlog.Error().Err(err).Msg("AI stream error")
		h.metrics.AIRequestsTotal.WithLabelValues("unknown", "error").Inc()
		errData, _ := json.Marshal(map[string]string{"error": "AI stream failed"})
		fmt.Fprintf(w, "data: %s\n\n", errData)
		flusher.Flush()
		return
	}

	// Record AI metrics
	h.metrics.AIRequestsTotal.WithLabelValues(result.Model, "success").Inc()
	h.metrics.AITokensTotal.WithLabelValues("input", result.Model).Add(float64(result.InputTokens))
	h.metrics.AITokensTotal.WithLabelValues("output", result.Model).Add(float64(result.OutputTokens))
	h.metrics.AIRequestDuration.WithLabelValues(result.Model).Observe(aiDuration)

	// Store assistant message with token usage
	_, err = h.queries.CreateMessage(r.Context(), sqlc.CreateMessageParams{
		ConversationID: convoID,
		Role:           "assistant",
		Content:        fullResponse,
		InputTokens:    int32(result.InputTokens),
		OutputTokens:   int32(result.OutputTokens),
		Model:          result.Model,
	})
	if err != nil {
		rlog.Error().Err(err).Msg("failed to store assistant message")
	}

	// Log AI usage for budget tracking
	uid := userID
	cid := convoID
	_, err = h.queries.LogAIUsage(r.Context(), sqlc.LogAIUsageParams{
		UserID:         &uid,
		ConversationID: &cid,
		InputTokens:    int32(result.InputTokens),
		OutputTokens:   int32(result.OutputTokens),
		Model:          result.Model,
	})
	if err != nil {
		rlog.Error().Err(err).Msg("failed to log AI usage")
	}

	// Send done event so the client knows the stream is complete
	fmt.Fprintf(w, "data: %s\n\n", `{"done":true}`)
	flusher.Flush()
}

// GetUsage handles GET /api/v1/settings/usage.
// Returns the current month's token usage for the authenticated user.
func (h *ChatHandlers) GetUsage(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid := userID

	usage, err := h.queries.GetMonthlyUsage(r.Context(), &uid)
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to get usage")
		return
	}
	JSON(w, http.StatusOK, usage)
}
