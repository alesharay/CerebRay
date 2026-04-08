package ai

import (
	"context"
)

// StreamCallback is called for each chunk of the AI response.
type StreamCallback func(chunk string) error

// Provider defines the AI chat interface.
type Provider interface {
	// StreamChat sends messages and streams the response via callback.
	StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string, cb StreamCallback) (*ChatResult, error)
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// ChatResult contains metadata about a completed chat interaction.
type ChatResult struct {
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	Model        string `json:"model"`
}
