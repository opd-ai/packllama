package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opd-ai/packllama/internal/service"
)

// --- invalid JSON body ---

func TestChatCompletions_InvalidJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader([]byte("{bad json"))))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestCompletions_InvalidJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/completions", bytes.NewReader([]byte("{bad json"))))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestEmbeddings_InvalidJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader([]byte("{bad json"))))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

// --- nil service returns 503 ---

func TestNilService_ChatCompletions(t *testing.T) {
	recorder := httptest.NewRecorder()
	body, _ := json.Marshal(ChatCompletionRequest{Model: "m", Messages: []Message{{Role: "user", Content: "hi"}}})
	newTestHandler(nil).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

func TestNilService_Completions(t *testing.T) {
	recorder := httptest.NewRecorder()
	body, _ := json.Marshal(CompletionRequest{Model: "m", Prompt: "p"})
	newTestHandler(nil).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

func TestNilService_Models(t *testing.T) {
	recorder := httptest.NewRecorder()
	newTestHandler(nil).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodGet, "/v1/models", nil))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

func TestNilService_Embeddings(t *testing.T) {
	recorder := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]any{"model": "m", "input": "hi"})
	newTestHandler(nil).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(body)))

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
}

// --- service errors ---

func TestEmbeddings_ServiceError(t *testing.T) {
	svc := &stubInference{embedErr: errors.New("embed backend down")}
	body, _ := json.Marshal(map[string]any{"model": "m", "input": "hello"})

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/embeddings", bytes.NewReader(body)))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

func TestListModels_ServiceError(t *testing.T) {
	svc := &stubInference{listErr: errors.New("registry unavailable")}

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodGet, "/v1/models", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}

// --- model metadata propagation ---

func TestListModels_IncludesMetadata(t *testing.T) {
	svc := &stubInference{
		models: []service.ModelInfo{
			{
				ID:             "llama-7b-q4",
				OwnedBy:        "local",
				Created:        1_000_000,
				ContextLength:  4096,
				ParameterCount: 7_000_000_000,
				Quantization:   "Q4_K_M",
			},
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
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 model, got %d", len(resp.Data))
	}
	m := resp.Data[0]
	if m.ContextLength != 4096 {
		t.Fatalf("expected context_length=4096, got %d", m.ContextLength)
	}
	if m.ParameterCount != 7_000_000_000 {
		t.Fatalf("expected parameter_count=7000000000, got %d", m.ParameterCount)
	}
	if m.Quantization != "Q4_K_M" {
		t.Fatalf("expected quantization=Q4_K_M, got %q", m.Quantization)
	}
}

func TestGetModel_IncludesMetadata(t *testing.T) {
	svc := &stubInference{
		models: []service.ModelInfo{{
			ID:            "llama-7b-q4",
			OwnedBy:       "local",
			ContextLength: 4096,
			Quantization:  "Q4_K_M",
		}},
	}

	recorder := httptest.NewRecorder()
	newTestHandler(svc).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodGet, "/v1/models/llama-7b-q4", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var m ModelObject
	if err := json.NewDecoder(recorder.Body).Decode(&m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if m.ContextLength != 4096 {
		t.Fatalf("expected context_length=4096, got %d", m.ContextLength)
	}
	if m.Quantization != "Q4_K_M" {
		t.Fatalf("expected quantization=Q4_K_M, got %q", m.Quantization)
	}
}

// --- chat validation edge cases ---

func TestChatCompletions_EmptyRoleInMessage(t *testing.T) {
	body, _ := json.Marshal(ChatCompletionRequest{
		Model:    "m",
		Messages: []Message{{Role: "", Content: "hi"}},
	})
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", recorder.Code)
	}
}

func TestChatCompletions_EmptyContentInMessage(t *testing.T) {
	body, _ := json.Marshal(ChatCompletionRequest{
		Model:    "m",
		Messages: []Message{{Role: "user", Content: ""}},
	})
	recorder := httptest.NewRecorder()
	newTestHandler(&stubInference{}).ServeHTTP(recorder,
		httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body)))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", recorder.Code)
	}
}
