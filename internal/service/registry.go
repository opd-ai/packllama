package service

import (
	"context"
	"errors"

	"github.com/opd-ai/packllama/internal/modelstore"
)

// ErrInferenceNotAvailable is returned by inference methods when no backend is
// configured. Callers that only need model listing can ignore this error path.
var ErrInferenceNotAvailable = errors.New("inference backend not available")

// RegistryService is an InferenceService backed by a modelstore.Registry.
// It serves ListModels from the registry and returns ErrInferenceNotAvailable
// for ChatComplete, Complete, and Embed until an inference backend is wired in.
type RegistryService struct {
	registry *modelstore.Registry
}

// NewRegistryService returns a RegistryService that reads model information
// from the provided Registry.
func NewRegistryService(registry *modelstore.Registry) *RegistryService {
	return &RegistryService{registry: registry}
}

// ListModels returns all models currently registered in the registry.
func (s *RegistryService) ListModels(_ context.Context) ([]ModelInfo, error) {
	entries := s.registry.List()
	models := make([]ModelInfo, len(entries))
	for i, e := range entries {
		models[i] = ModelInfo{
			ID:             e.ID,
			OwnedBy:        e.OwnedBy,
			Created:        e.ModTime.Unix(),
			ContextLength:  e.ContextLength,
			ParameterCount: e.ParameterCount,
			Quantization:   e.Quantization,
		}
	}
	return models, nil
}

// ChatComplete returns ErrInferenceNotAvailable; inference requires a backend.
func (s *RegistryService) ChatComplete(_ context.Context, _ ChatRequest) (<-chan ChatChunk, error) {
	return nil, ErrInferenceNotAvailable
}

// Complete returns ErrInferenceNotAvailable; inference requires a backend.
func (s *RegistryService) Complete(_ context.Context, _ TextRequest) (<-chan CompletionChunk, error) {
	return nil, ErrInferenceNotAvailable
}

// Embed returns ErrInferenceNotAvailable; inference requires a backend.
func (s *RegistryService) Embed(_ context.Context, _ EmbeddingRequest) ([]EmbeddingVector, error) {
	return nil, ErrInferenceNotAvailable
}
