package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/opd-ai/packllama/internal/service"
)

// handleChatCompletions handles POST /v1/chat/completions.
func handleChatCompletions(svc service.InferenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		if err := validateChatRequest(req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		svcReq := chatRequestToService(req)
		if req.Stream {
			streamChatResponse(w, r, svc, svcReq, req.Model)
			return
		}
		respondChatFull(w, r, svc, svcReq, req.Model)
	}
}

func validateChatRequest(req ChatCompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("field 'model' is required")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("field 'messages' must not be empty")
	}
	for i, m := range req.Messages {
		if m.Role == "" {
			return fmt.Errorf("messages[%d].role is required", i)
		}
		if m.Content == "" {
			return fmt.Errorf("messages[%d].content is required", i)
		}
	}
	return nil
}

func chatRequestToService(req ChatCompletionRequest) service.ChatRequest {
	msgs := make([]service.ChatMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = service.ChatMessage{Role: m.Role, Content: m.Content}
	}
	return service.ChatRequest{
		InferenceRequest: service.InferenceRequest{
			Model:       req.Model,
			Temperature: req.Temperature,
			TopP:        req.TopP,
			MaxTokens:   req.MaxTokens,
			Stop:        req.Stop,
		},
		Messages: msgs,
	}
}

func respondChatFull(w http.ResponseWriter, r *http.Request, svc service.InferenceService, req service.ChatRequest, model string) {
	chunks, err := svc.ChatComplete(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var content string
	finishReason := "stop"
	for chunk := range chunks {
		if chunk.Err != nil {
			writeError(w, http.StatusInternalServerError, chunk.Err.Error())
			return
		}
		content += chunk.Content
		if chunk.FinishReason != "" {
			finishReason = chunk.FinishReason
		}
	}

	resp := ChatCompletionResponse{
		ID:      newRequestID(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []ChatCompletionChoice{{
			Index:        0,
			Message:      Message{Role: "assistant", Content: content},
			FinishReason: finishReason,
		}},
	}
	writeJSON(w, http.StatusOK, resp)
}

func streamChatResponse(w http.ResponseWriter, r *http.Request, svc service.InferenceService, req service.ChatRequest, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	chunks, err := svc.ChatComplete(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	id := newRequestID()
	now := time.Now().Unix()

	for chunk := range chunks {
		if chunk.Err != nil {
			writeSSEError(w, flusher, chunk.Err)
			return
		}

		var finishReason *string
		if chunk.FinishReason != "" {
			finishReason = &chunk.FinishReason
		}

		event := ChatCompletionChunk{
			ID:      id,
			Object:  "chat.completion.chunk",
			Created: now,
			Model:   model,
			Choices: []ChatCompletionChunkChoice{{
				Index:        0,
				Delta:        ChatCompletionChunkDelta{Content: chunk.Content},
				FinishReason: finishReason,
			}},
		}
		writeSSEEvent(w, flusher, event)
	}
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}
