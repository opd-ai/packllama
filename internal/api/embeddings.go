package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/opd-ai/packllama/internal/service"
)

// handleEmbeddings handles POST /v1/embeddings.
func handleEmbeddings(svc service.InferenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		if err := validateEmbeddingRequest(req); err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}

		svcReq := service.EmbeddingRequest{
			InferenceRequest: service.InferenceRequest{Model: req.Model},
			Input:            req.Input,
			Dimensions:       req.Dimensions,
		}
		vectors, err := svc.Embed(r.Context(), svcReq)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		data := make([]EmbeddingObject, len(vectors))
		for i, v := range vectors {
			data[i] = EmbeddingObject{
				Object:    "embedding",
				Index:     v.Index,
				Embedding: v.Embedding,
			}
		}
		writeJSON(w, http.StatusOK, EmbeddingResponse{
			Object: "list",
			Data:   data,
			Model:  req.Model,
		})
	}
}

func validateEmbeddingRequest(req EmbeddingRequest) error {
	if req.Model == "" {
		return fmt.Errorf("field 'model' is required")
	}
	if len(req.Input) == 0 {
		return fmt.Errorf("field 'input' must not be empty")
	}
	return nil
}
