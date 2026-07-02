package api

import (
	"log/slog"
	"net/http"

	"github.com/opd-ai/packllama/internal/service"
)

func NewHandler(logger *slog.Logger, allowedOrigins []string) http.Handler {
	health := service.NewHealthService()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, health.Status())
	})

	return chain(mux,
		withCORS(allowedOrigins),
		withRequestID,
		withLogging(logger),
		withRecovery(logger),
	)
}
