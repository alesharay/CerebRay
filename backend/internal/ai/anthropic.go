// Package ai provides AI provider implementations for the Zettelkasten chat coach.
//
// The Anthropic provider calls the Messages API directly with streaming enabled.
// Text deltas are sent to a callback as they arrive so the handler can relay them
// over SSE to the browser in real time.
package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

// AnthropicProvider calls the Anthropic Messages API with streaming.
type AnthropicProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewAnthropicProvider creates a provider backed by the Anthropic API.
func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

// anthropicMessage is a single message in the Anthropic request format.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicRequest is the full request body for POST /v1/messages.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
}

// StreamChat sends the conversation to Claude and streams text deltas via cb.
// Returns token usage and model info after the stream completes.
func (p *AnthropicProvider) StreamChat(ctx context.Context, messages []ChatMessage, systemPrompt string, cb StreamCallback) (*ChatResult, error) {
	// Convert to Anthropic message format
	var apiMessages []anthropicMessage
	for _, m := range messages {
		if m.Role == "user" || m.Role == "assistant" {
			apiMessages = append(apiMessages, anthropicMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	body := anthropicRequest{
		Model:     p.model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  apiMessages,
		Stream:    true,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API error %d: %s", resp.StatusCode, respBody)
	}

	result := &ChatResult{
		Model: p.model,
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 512*1024), 512*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || line == "data: [DONE]" {
			continue
		}
		if len(line) < 6 || line[:6] != "data: " {
			continue
		}

		eventJSON := line[6:]
		var event map[string]any
		if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
			continue
		}

		eventType, _ := event["type"].(string)

		switch eventType {
		case "content_block_delta":
			delta, _ := event["delta"].(map[string]any)
			if delta == nil {
				continue
			}
			if delta["type"] == "text_delta" {
				text, _ := delta["text"].(string)
				if text != "" {
					if err := cb(text); err != nil {
						return nil, fmt.Errorf("callback error: %w", err)
					}
				}
			}

		case "message_delta":
			// Extract token usage from the final message_delta event
			usage, _ := event["usage"].(map[string]any)
			if usage != nil {
				if v, ok := usage["output_tokens"].(float64); ok {
					result.OutputTokens = int(v)
				}
			}

		case "message_start":
			// Extract input token count from the initial message event
			msg, _ := event["message"].(map[string]any)
			if msg != nil {
				usage, _ := msg["usage"].(map[string]any)
				if usage != nil {
					if v, ok := usage["input_tokens"].(float64); ok {
						result.InputTokens = int(v)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading stream: %w", err)
	}

	return result, nil
}
