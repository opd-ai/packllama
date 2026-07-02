package api

import "encoding/json"

// Message represents a single chat message in OpenAI format.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest is the request body for POST /v1/chat/completions.
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature *float64  `json:"temperature,omitempty"`
	TopP        *float64  `json:"top_p,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
	Stop        []string  `json:"stop,omitempty"`
	Stream      bool      `json:"stream"`
}

// ChatCompletionChoice is one candidate response in a non-streaming reply.
type ChatCompletionChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// UsageInfo reports token counts.
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse is the non-streaming response for POST /v1/chat/completions.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   UsageInfo              `json:"usage"`
}

// ChatCompletionChunkDelta carries the incremental content in a streaming chunk.
type ChatCompletionChunkDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ChatCompletionChunkChoice is one choice in a streaming chunk.
type ChatCompletionChunkChoice struct {
	Index        int                      `json:"index"`
	Delta        ChatCompletionChunkDelta `json:"delta"`
	FinishReason *string                  `json:"finish_reason"`
}

// ChatCompletionChunk is a single Server-Sent Events frame in a streaming response.
type ChatCompletionChunk struct {
	ID      string                      `json:"id"`
	Object  string                      `json:"object"`
	Created int64                       `json:"created"`
	Model   string                      `json:"model"`
	Choices []ChatCompletionChunkChoice `json:"choices"`
}

// CompletionRequest is the request body for POST /v1/completions.
type CompletionRequest struct {
	Model     string   `json:"model"`
	Prompt    string   `json:"prompt"`
	Suffix    string   `json:"suffix,omitempty"`
	MaxTokens *int     `json:"max_tokens,omitempty"`
	Stop      []string `json:"stop,omitempty"`
	Stream    bool     `json:"stream"`
}

// CompletionChoice is one candidate in a non-streaming completion response.
type CompletionChoice struct {
	Index        int    `json:"index"`
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

// CompletionResponse is the non-streaming response for POST /v1/completions.
type CompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   UsageInfo          `json:"usage"`
}

// ModelObject represents a model entry in the OpenAI models list.
type ModelObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`

	// Extended metadata — zero/empty when not yet populated by the inference backend.
	ContextLength  int64  `json:"context_length,omitempty"`
	ParameterCount int64  `json:"parameter_count,omitempty"`
	Quantization   string `json:"quantization,omitempty"`
}

// EmbeddingRequest is the request body for POST /v1/embeddings.
type EmbeddingRequest struct {
	Model string `json:"model"`
	// Input may be a single string or an array of strings.
	// After decoding it is always stored as a slice.
	Input []string `json:"-"`
}

// UnmarshalJSON handles both string and []string values for the input field.
func (r *EmbeddingRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		Model string          `json:"model"`
		Input json.RawMessage `json:"input"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Model = raw.Model
	if len(raw.Input) == 0 {
		return nil
	}
	// Try array first, then single string.
	var arr []string
	if err := json.Unmarshal(raw.Input, &arr); err == nil {
		r.Input = arr
		return nil
	}
	var s string
	if err := json.Unmarshal(raw.Input, &s); err != nil {
		return err
	}
	r.Input = []string{s}
	return nil
}

// EmbeddingObject is one entry in the embeddings response data array.
type EmbeddingObject struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

// EmbeddingResponse is the response for POST /v1/embeddings.
type EmbeddingResponse struct {
	Object string            `json:"object"`
	Data   []EmbeddingObject `json:"data"`
	Model  string            `json:"model"`
	Usage  UsageInfo         `json:"usage"`
}

// ModelListResponse is the response for GET /v1/models.
type ModelListResponse struct {
	Object string        `json:"object"`
	Data   []ModelObject `json:"data"`
}
