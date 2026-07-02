package service

import "context"

// ChatChunk is a single piece of streamed chat completion content.
type ChatChunk struct {
	Content      string
	FinishReason string // non-empty on the final chunk
	Err          error
}

// CompletionChunk is a single piece of streamed text completion content.
type CompletionChunk struct {
	Text         string
	FinishReason string // non-empty on the final chunk
	Err          error
}

// InferenceRequest carries parameters common to both chat and text completions.
type InferenceRequest struct {
	Model       string
	Temperature *float64
	TopP        *float64
	MaxTokens   *int
	Stop        []string
}

// ChatRequest wraps a chat-specific inference request.
type ChatRequest struct {
	InferenceRequest
	Messages []ChatMessage
}

// ChatMessage is a role+content pair for chat completions.
type ChatMessage struct {
	Role    string
	Content string
}

// TextRequest wraps a text completion inference request.
type TextRequest struct {
	InferenceRequest
	Prompt string
	Suffix string
}

// ModelInfo holds metadata for a single loaded or available model.
type ModelInfo struct {
	ID      string
	OwnedBy string
	Created int64
}

// InferenceService is the interface the API layer uses for LLM inference.
// Implementations are expected to be thread-safe.
type InferenceService interface {
	// ChatComplete begins a chat completion. The returned channel emits one or
	// more chunks and is closed after the final chunk (FinishReason != "").
	// Callers must drain or abandon the channel; ctx cancellation stops generation.
	ChatComplete(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)

	// Complete begins a text completion. The returned channel behaves the same
	// as ChatComplete.
	Complete(ctx context.Context, req TextRequest) (<-chan CompletionChunk, error)

	// ListModels returns the models currently available for inference.
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
