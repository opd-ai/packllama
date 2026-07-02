package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opd-ai/packllama/internal/service"
)

func TestOpenAICompatibility_ChatCompletions(t *testing.T) {
	svc := &stubInference{
		chatChunks: []service.ChatChunk{
			{Content: "Hello"},
			{Content: " world", FinishReason: "stop"},
		},
	}
	body, _ := json.Marshal(map[string]any{
		"model": "llama-3-8b-instruct",
		"messages": []map[string]string{
			{"role": "user", "content": "Say hello"},
		},
	})

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var resp ChatCompletionResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty id")
	}
	if resp.Object != "chat.completion" {
		t.Fatalf("expected object chat.completion, got %q", resp.Object)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Role != "assistant" {
		t.Fatalf("expected assistant role, got %q", resp.Choices[0].Message.Role)
	}
	if resp.Choices[0].Message.Content != "Hello world" {
		t.Fatalf("expected content 'Hello world', got %q", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Fatalf("expected finish_reason stop, got %q", resp.Choices[0].FinishReason)
	}
}

func TestOpenAICompatibility_ChatCompletionsStream(t *testing.T) {
	svc := &stubInference{
		chatChunks: []service.ChatChunk{
			{Content: "chunk1"},
			{Content: "chunk2", FinishReason: "stop"},
		},
	}
	body, _ := json.Marshal(map[string]any{
		"model": "llama-3-8b-instruct",
		"messages": []map[string]string{
			{"role": "user", "content": "Stream"},
		},
		"stream": true,
	})

	recorder := newFlushableRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	lines := sseLines(t, recorder.Body.String())
	if len(lines) < 2 {
		t.Fatalf("expected chunk lines and [DONE], got %d lines", len(lines))
	}
	if lines[len(lines)-1] != "[DONE]" {
		t.Fatalf("expected last line [DONE], got %q", lines[len(lines)-1])
	}

	var firstChunk ChatCompletionChunk
	if err := json.Unmarshal([]byte(lines[0]), &firstChunk); err != nil {
		t.Fatalf("decode first chunk: %v", err)
	}
	if firstChunk.Object != "chat.completion.chunk" {
		t.Fatalf("expected chunk object chat.completion.chunk, got %q", firstChunk.Object)
	}
	if len(firstChunk.Choices) != 1 {
		t.Fatalf("expected 1 chunk choice, got %d", len(firstChunk.Choices))
	}
	if firstChunk.Choices[0].Delta.Content == "" {
		t.Fatal("expected delta content in first chunk")
	}
}

func TestOpenAICompatibility_Completions(t *testing.T) {
	svc := &stubInference{
		textChunks: []service.CompletionChunk{
			{Text: "answer", FinishReason: "stop"},
		},
	}
	body, _ := json.Marshal(map[string]any{
		"model":  "llama-3-8b-instruct",
		"prompt": "Answer briefly:",
	})

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var resp CompletionResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty id")
	}
	if resp.Object != "text_completion" {
		t.Fatalf("expected object text_completion, got %q", resp.Object)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Text != "answer" {
		t.Fatalf("expected text 'answer', got %q", resp.Choices[0].Text)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Fatalf("expected finish_reason stop, got %q", resp.Choices[0].FinishReason)
	}
}
