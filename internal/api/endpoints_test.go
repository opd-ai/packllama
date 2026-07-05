package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/opd-ai/packllama/internal/service"
)

// stubInference is a test double for service.InferenceService.
type stubInference struct {
	chatChunks []service.ChatChunk
	textChunks []service.CompletionChunk
	models     []service.ModelInfo
	embeddings []service.EmbeddingVector
	chatErr    error
	textErr    error
	listErr    error
	embedErr   error
	loadErr    error
	unloadErr  error
	loadModel  service.ModelInfo
	lastLoad   service.ModelLoadRequest
	lastUnload string
}

func (s *stubInference) ChatComplete(_ context.Context, _ service.ChatRequest) (<-chan service.ChatChunk, error) {
	if s.chatErr != nil {
		return nil, s.chatErr
	}
	ch := make(chan service.ChatChunk, len(s.chatChunks))
	for _, c := range s.chatChunks {
		ch <- c
	}
	close(ch)
	return ch, nil
}

func (s *stubInference) Complete(_ context.Context, _ service.TextRequest) (<-chan service.CompletionChunk, error) {
	if s.textErr != nil {
		return nil, s.textErr
	}
	ch := make(chan service.CompletionChunk, len(s.textChunks))
	for _, c := range s.textChunks {
		ch <- c
	}
	close(ch)
	return ch, nil
}

func (s *stubInference) Embed(_ context.Context, _ service.EmbeddingRequest) ([]service.EmbeddingVector, error) {
	return s.embeddings, s.embedErr
}

func (s *stubInference) ListModels(_ context.Context) ([]service.ModelInfo, error) {
	return s.models, s.listErr
}

func (s *stubInference) LoadModel(_ context.Context, req service.ModelLoadRequest) (service.ModelInfo, error) {
	s.lastLoad = req
	return s.loadModel, s.loadErr
}

func (s *stubInference) UnloadModel(_ context.Context, id string) error {
	s.lastUnload = id
	return s.unloadErr
}

func newTestHandler(svc service.InferenceService) http.Handler {
	return NewHandler(testLogger(io.Discard), nil, svc)
}

// flushableRecorder wraps httptest.ResponseRecorder to implement http.Flusher,
// which is required for SSE streaming endpoints.
type flushableRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flushableRecorder) Flush() { f.ResponseRecorder.Flush() }

func newFlushableRecorder() *flushableRecorder {
	return &flushableRecorder{httptest.NewRecorder()}
}

// --- /v1/chat/completions ---

func TestChatCompletions_FullResponse(t *testing.T) {
	svc := &stubInference{
		chatChunks: []service.ChatChunk{
			{Content: "Hello"},
			{Content: " world", FinishReason: "stop"},
		},
	}
	body, _ := json.Marshal(ChatCompletionRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	newTestHandler(svc).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body)
	}
	var resp ChatCompletionResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}
	if got := resp.Choices[0].Message.Content; got != "Hello world" {
		t.Fatalf("expected 'Hello world', got %q", got)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Fatalf("expected finish_reason stop, got %q", resp.Choices[0].FinishReason)
	}
	if resp.Object != "chat.completion" {
		t.Fatalf("unexpected object %q", resp.Object)
	}
}

func TestChatCompletions_StreamResponse(t *testing.T) {
	svc := &stubInference{
		chatChunks: []service.ChatChunk{
			{Content: "chunk1"},
			{Content: "chunk2", FinishReason: "stop"},
		},
	}
	body, _ := json.Marshal(ChatCompletionRequest{
		Model:    "test-model",
		Messages: []Message{{Role: "user", Content: "Hi"}},
		Stream:   true,
	})

	recorder := newFlushableRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	newTestHandler(svc).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	ct := recorder.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected text/event-stream, got %q", ct)
	}

	lines := sseLines(t, recorder.Body.String())
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 data lines, got %d", len(lines))
	}
	if lines[len(lines)-1] != "[DONE]" {
		t.Fatalf("expected last line [DONE], got %q", lines[len(lines)-1])
	}

	var chunk ChatCompletionChunk
	if err := json.Unmarshal([]byte(lines[0]), &chunk); err != nil {
		t.Fatalf("unmarshal first chunk: %v", err)
	}
	if chunk.Object != "chat.completion.chunk" {
		t.Fatalf("unexpected object %q", chunk.Object)
	}
}

func TestChatCompletions_MissingModel(t *testing.T) {
	body, _ := json.Marshal(ChatCompletionRequest{
		Messages: []Message{{Role: "user", Content: "Hi"}},
	})
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", recorder.Code)
	}
}

func TestChatCompletions_MissingMessages(t *testing.T) {
	body, _ := json.Marshal(ChatCompletionRequest{Model: "test"})
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", recorder.Code)
	}
}

func TestChatCompletions_BadJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader("{bad}")))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestChatCompletions_NoService(t *testing.T) {
	body, _ := json.Marshal(ChatCompletionRequest{
		Model:    "x",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	recorder := httptest.NewRecorder()
	NewHandler(testLogger(io.Discard), nil, nil).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

// --- /v1/completions ---

func TestCompletions_FullResponse(t *testing.T) {
	svc := &stubInference{
		textChunks: []service.CompletionChunk{
			{Text: "answer", FinishReason: "stop"},
		},
	}
	body, _ := json.Marshal(CompletionRequest{Model: "m", Prompt: "Q:"})

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body)
	}
	var resp CompletionResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Choices[0].Text != "answer" {
		t.Fatalf("expected 'answer', got %q", resp.Choices[0].Text)
	}
}

func TestCompletions_MissingPrompt(t *testing.T) {
	body, _ := json.Marshal(CompletionRequest{Model: "m"})
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", recorder.Code)
	}
}

func TestCompletions_StreamResponse(t *testing.T) {
	svc := &stubInference{
		textChunks: []service.CompletionChunk{
			{Text: "tok1"},
			{Text: "tok2", FinishReason: "stop"},
		},
	}
	body, _ := json.Marshal(CompletionRequest{Model: "m", Prompt: "p", Stream: true})

	recorder := newFlushableRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	lines := sseLines(t, recorder.Body.String())
	if lines[len(lines)-1] != "[DONE]" {
		t.Fatalf("expected [DONE], got %q", lines[len(lines)-1])
	}
}

// --- /v1/models ---

func TestListModels(t *testing.T) {
	svc := &stubInference{
		models: []service.ModelInfo{
			{ID: "model-a", OwnedBy: "local", Created: 1_000_000},
			{ID: "model-b", OwnedBy: "local", Created: 2_000_000},
		},
	}

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodGet, "/v1/models", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var resp ModelListResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Object != "list" {
		t.Fatalf("expected object=list, got %q", resp.Object)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 models, got %d", len(resp.Data))
	}
}

func TestGetModel_Found(t *testing.T) {
	svc := &stubInference{
		models: []service.ModelInfo{{ID: "model-a", OwnedBy: "local", Created: 1_000_000}},
	}

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodGet, "/v1/models/model-a", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var m ModelObject
	if err := json.NewDecoder(recorder.Body).Decode(&m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if m.ID != "model-a" {
		t.Fatalf("expected model-a, got %q", m.ID)
	}
}

func TestGetModel_NotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodGet, "/v1/models/nonexistent", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestLoadModel(t *testing.T) {
	svc := &stubInference{
		loadModel: service.ModelInfo{ID: "new-model", OwnedBy: "local", Created: 1_000_000},
	}
	body := `{"path":"/models/new-model.gguf"}`

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/models", strings.NewReader(body)))

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if svc.lastLoad.Path != "/models/new-model.gguf" {
		t.Fatalf("expected load path to be passed through, got %q", svc.lastLoad.Path)
	}
}

func TestLoadModel_ValidationError(t *testing.T) {
	svc := &stubInference{loadErr: service.ErrModelPathRequired}
	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/models", strings.NewReader(`{"path":""}`)))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", recorder.Code)
	}
}

func TestLoadModel_Conflict(t *testing.T) {
	svc := &stubInference{loadErr: service.ErrModelAlreadyExists}
	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/models", strings.NewReader(`{"path":"/models/existing.gguf"}`)))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}
}

func TestUnloadModel(t *testing.T) {
	svc := &stubInference{}
	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodDelete, "/v1/models/model-a", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if svc.lastUnload != "model-a" {
		t.Fatalf("expected model-a unload, got %q", svc.lastUnload)
	}
}

func TestUnloadModel_NotFound(t *testing.T) {
	svc := &stubInference{unloadErr: service.ErrModelNotFound}
	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodDelete, "/v1/models/missing", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

// sseLines extracts the JSON payload from each "data: ..." SSE line.
func sseLines(t *testing.T, body string) []string {
	t.Helper()
	var out []string
	scanner := bufio.NewScanner(strings.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			out = append(out, strings.TrimPrefix(line, "data: "))
		}
	}
	return out
}

// --- /v1/embeddings ---

func TestEmbeddings_SingleInput(t *testing.T) {
	svc := &stubInference{
		embeddings: []service.EmbeddingVector{
			{Index: 0, Embedding: []float32{0.1, 0.2, 0.3}},
		},
	}
	body, _ := json.Marshal(map[string]any{
		"model": "embed-model",
		"input": "hello world",
	})

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(body)))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body)
	}
	var resp EmbeddingResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Object != "list" {
		t.Fatalf("expected object=list, got %q", resp.Object)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 embedding, got %d", len(resp.Data))
	}
	if resp.Data[0].Object != "embedding" {
		t.Fatalf("expected object=embedding, got %q", resp.Data[0].Object)
	}
	if len(resp.Data[0].Embedding) != 3 {
		t.Fatalf("expected 3 dimensions, got %d", len(resp.Data[0].Embedding))
	}
}

func TestEmbeddings_BatchInput(t *testing.T) {
	svc := &stubInference{
		embeddings: []service.EmbeddingVector{
			{Index: 0, Embedding: []float32{0.1}},
			{Index: 1, Embedding: []float32{0.2}},
		},
	}
	body, _ := json.Marshal(map[string]any{
		"model": "embed-model",
		"input": []string{"text one", "text two"},
	})

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(body)))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var resp EmbeddingResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(resp.Data))
	}
}

func TestEmbeddings_MissingModel(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"input": "hi"})
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(body)))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", recorder.Code)
	}
}

func TestEmbeddings_MissingInput(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"model": "m"})
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(body)))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", recorder.Code)
	}
}
