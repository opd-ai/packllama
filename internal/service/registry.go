package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/opd-ai/packllama/internal/modelstore"
)

// ErrInferenceNotAvailable is returned by inference methods when no backend is
// configured. Callers that only need model listing can ignore this error path.
var ErrInferenceNotAvailable = errors.New("inference backend not available")

var (
	// ErrModelPathRequired is returned when loading a model without a file path.
	ErrModelPathRequired = errors.New("model path is required")
	// ErrInvalidModelPath is returned when loading a non-GGUF or unreadable file.
	ErrInvalidModelPath = errors.New("invalid model path")
	// ErrModelAlreadyExists is returned when loading a model with a duplicate ID.
	ErrModelAlreadyExists = errors.New("model already exists")
	// ErrModelNotFound is returned when unloading a model that is not registered.
	ErrModelNotFound = errors.New("model not found")
)

// ModelLoadRequest carries input for registering a model through the API.
type ModelLoadRequest struct {
	Path    string
	ID      string
	OwnedBy string
}

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

// LoadModel registers a model file in the registry.
func (s *RegistryService) LoadModel(_ context.Context, req ModelLoadRequest) (ModelInfo, error) {
	if strings.TrimSpace(req.Path) == "" {
		return ModelInfo{}, ErrModelPathRequired
	}
	entry, err := s.registry.AddModelFile(req.Path, req.ID, req.OwnedBy)
	if err != nil {
		switch {
		case errors.Is(err, modelstore.ErrInvalidModelFile):
			return ModelInfo{}, ErrInvalidModelPath
		case errors.Is(err, modelstore.ErrModelAlreadyExists):
			return ModelInfo{}, ErrModelAlreadyExists
		case os.IsNotExist(err):
			return ModelInfo{}, ErrInvalidModelPath
		default:
			return ModelInfo{}, fmt.Errorf("load model: %w", err)
		}
	}
	return ModelInfo{
		ID:             entry.ID,
		OwnedBy:        entry.OwnedBy,
		Created:        entry.ModTime.Unix(),
		ContextLength:  entry.ContextLength,
		ParameterCount: entry.ParameterCount,
		Quantization:   entry.Quantization,
	}, nil
}

// UnloadModel removes a model from the registry.
func (s *RegistryService) UnloadModel(_ context.Context, id string) error {
	if err := s.registry.RemoveModel(id); err != nil {
		if errors.Is(err, modelstore.ErrModelNotFound) {
			return ErrModelNotFound
		}
		return fmt.Errorf("unload model: %w", err)
	}
	return nil
}
