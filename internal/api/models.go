package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/opd-ai/packllama/internal/service"
)

type modelManagerService interface {
	LoadModel(ctx context.Context, req service.ModelLoadRequest) (service.ModelInfo, error)
	UnloadModel(ctx context.Context, id string) error
}

// handleListModels handles GET /v1/models.
func handleListModels(svc service.InferenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models, err := svc.ListModels(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		objects := make([]ModelObject, len(models))
		for i, m := range models {
			objects[i] = ModelObject{
				ID:             m.ID,
				Object:         "model",
				Created:        modelCreatedAt(m.Created),
				OwnedBy:        m.OwnedBy,
				ContextLength:  m.ContextLength,
				ParameterCount: m.ParameterCount,
				Quantization:   m.Quantization,
			}
		}
		writeJSON(w, http.StatusOK, ModelListResponse{
			Object: "list",
			Data:   objects,
		})
	}
}

// handleGetModel handles GET /v1/models/{model_id}.
func handleGetModel(svc service.InferenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("model_id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "model_id is required")
			return
		}

		models, err := svc.ListModels(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		for _, m := range models {
			if m.ID == id {
				writeJSON(w, http.StatusOK, ModelObject{
					ID:             m.ID,
					Object:         "model",
					Created:        modelCreatedAt(m.Created),
					OwnedBy:        m.OwnedBy,
					ContextLength:  m.ContextLength,
					ParameterCount: m.ParameterCount,
					Quantization:   m.Quantization,
				})
				return
			}
		}
		writeError(w, http.StatusNotFound, "model '"+id+"' not found")
	}
}

// handleLoadModel handles POST /v1/models.
func handleLoadModel(svc service.InferenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		manager, ok := svc.(modelManagerService)
		if !ok {
			writeError(w, http.StatusNotImplemented, "model loading is not supported")
			return
		}

		var req LoadModelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}

		model, err := manager.LoadModel(r.Context(), service.ModelLoadRequest{
			Path:    req.Path,
			ID:      req.ID,
			OwnedBy: req.OwnedBy,
		})
		if err != nil {
			writeModelError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, ModelObject{
			ID:             model.ID,
			Object:         "model",
			Created:        modelCreatedAt(model.Created),
			OwnedBy:        model.OwnedBy,
			ContextLength:  model.ContextLength,
			ParameterCount: model.ParameterCount,
			Quantization:   model.Quantization,
		})
	}
}

// handleUnloadModel handles DELETE /v1/models/{model_id}.
func handleUnloadModel(svc service.InferenceService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		manager, ok := svc.(modelManagerService)
		if !ok {
			writeError(w, http.StatusNotImplemented, "model unloading is not supported")
			return
		}
		id := r.PathValue("model_id")
		if id == "" {
			writeError(w, http.StatusBadRequest, "model_id is required")
			return
		}
		if err := manager.UnloadModel(r.Context(), id); err != nil {
			writeModelError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, DeleteModelResponse{
			ID:      id,
			Object:  "model",
			Deleted: true,
		})
	}
}

func writeModelError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrModelPathRequired):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, service.ErrInvalidModelPath):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrModelAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrModelNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

// modelCreatedAt returns t when non-zero, otherwise the current Unix timestamp.
func modelCreatedAt(t int64) int64 {
	if t != 0 {
		return t
	}
	return time.Now().Unix()
}
