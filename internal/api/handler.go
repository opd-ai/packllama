package api

import (
	"log/slog"
	"net/http"

	"github.com/opd-ai/packllama/internal/service"
)

// NewHandler builds the HTTP handler for the packllama API. When svc is nil,
// all inference endpoints return 503 Service Unavailable.
func NewHandler(logger *slog.Logger, allowedOrigins []string, svc service.InferenceService) http.Handler {
	health := service.NewHealthService()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, health.Status())
	})

	if svc != nil {
		mux.HandleFunc("POST /v1/chat/completions", handleChatCompletions(svc))
		mux.HandleFunc("POST /v1/completions", handleCompletions(svc))
		mux.HandleFunc("GET /v1/models", handleListModels(svc))
		mux.HandleFunc("GET /v1/models/{model_id}", handleGetModel(svc))
	} else {
		unavailable := func(w http.ResponseWriter, r *http.Request) {
			writeError(w, http.StatusServiceUnavailable, "inference service not configured")
		}
		mux.HandleFunc("POST /v1/chat/completions", unavailable)
		mux.HandleFunc("POST /v1/completions", unavailable)
		mux.HandleFunc("GET /v1/models", unavailable)
		mux.HandleFunc("GET /v1/models/{model_id}", unavailable)
	}

	return chain(mux,
		withCORS(allowedOrigins),
		withRequestID,
		withLogging(logger),
		withRecovery(logger),
	)
}
