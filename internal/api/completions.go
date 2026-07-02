package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/opd-ai/packllama/internal/service"
)

// handleCompletions handles POST /v1/completions.
func handleCompletions(svc service.InferenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		if err := validateCompletionRequest(req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		svcReq := completionRequestToService(req)
		if req.Stream {
			streamCompletionResponse(w, r, svc, svcReq, req.Model)
			return
		}
		respondCompletionFull(w, r, svc, svcReq, req.Model)
	}
}

func validateCompletionRequest(req CompletionRequest) error {
	if req.Model == "" {
		return fmt.Errorf("field 'model' is required")
	}
	if req.Prompt == "" {
		return fmt.Errorf("field 'prompt' is required")
	}
	return nil
}

func completionRequestToService(req CompletionRequest) service.TextRequest {
	return service.TextRequest{
		InferenceRequest: service.InferenceRequest{
			Model:     req.Model,
			MaxTokens: req.MaxTokens,
			Stop:      req.Stop,
		},
		Prompt: req.Prompt,
		Suffix: req.Suffix,
	}
}

func respondCompletionFull(w http.ResponseWriter, r *http.Request, svc service.InferenceService, req service.TextRequest, model string) {
	chunks, err := svc.Complete(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var text string
	finishReason := "stop"
	for chunk := range chunks {
		if chunk.Err != nil {
			writeError(w, http.StatusInternalServerError, chunk.Err.Error())
			return
		}
		text += chunk.Text
		if chunk.FinishReason != "" {
			finishReason = chunk.FinishReason
		}
	}

	resp := CompletionResponse{
		ID:      newRequestID(),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []CompletionChoice{{
			Index:        0,
			Text:         text,
			FinishReason: finishReason,
		}},
	}
	writeJSON(w, http.StatusOK, resp)
}

func streamCompletionResponse(w http.ResponseWriter, r *http.Request, svc service.InferenceService, req service.TextRequest, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	chunks, err := svc.Complete(r.Context(), req)
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
		event := map[string]any{
			"id":      id,
			"object":  "text_completion",
			"created": now,
			"model":   model,
			"choices": []map[string]any{{
				"index":         0,
				"text":          chunk.Text,
				"finish_reason": nilIfEmpty(chunk.FinishReason),
			}},
		}
		writeSSEEvent(w, flusher, event)
	}
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
