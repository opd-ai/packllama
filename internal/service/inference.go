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

	// ContextLength is the maximum context window in tokens. Zero when unknown.
	ContextLength int64
	// ParameterCount is the number of model parameters. Zero when unknown.
	ParameterCount int64
	// Quantization describes the weight quantization scheme (e.g. "Q4_K_M"). Empty when unknown.
	Quantization string
}

// EmbeddingRequest carries parameters for a single or batch embedding request.
type EmbeddingRequest struct {
	InferenceRequest
	Input      []string // one or more texts to embed
	Dimensions *int     // optional output dimension reduction; nil = backend default
}

// EmbeddingVector holds one embedding result.
type EmbeddingVector struct {
	Index     int
	Embedding []float32
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

	// Embed returns embedding vectors for each input in req.Input.
	Embed(ctx context.Context, req EmbeddingRequest) ([]EmbeddingVector, error)

	// ListModels returns the models currently available for inference.
	ListModels(ctx context.Context) ([]ModelInfo, error)
}
