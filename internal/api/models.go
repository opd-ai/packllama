package api

import (
	"net/http"
	"time"

	"github.com/opd-ai/packllama/internal/service"
)

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
				ID:      m.ID,
				Object:  "model",
				Created: modelCreatedAt(m.Created),
				OwnedBy: m.OwnedBy,
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
					ID:      m.ID,
					Object:  "model",
					Created: modelCreatedAt(m.Created),
					OwnedBy: m.OwnedBy,
				})
				return
			}
		}
		writeError(w, http.StatusNotFound, "model '"+id+"' not found")
	}
}

// modelCreatedAt returns t when non-zero, otherwise the current Unix timestamp.
func modelCreatedAt(t int64) int64 {
	if t != 0 {
		return t
	}
	return time.Now().Unix()
}
