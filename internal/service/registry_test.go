package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/packllama/internal/modelstore"
	"github.com/opd-ai/packllama/internal/service"
)

func TestRegistryService_ListModels_Empty(t *testing.T) {
	reg := modelstore.New()
	svc := service.NewRegistryService(reg)

	models, err := svc.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("expected no models, got %d", len(models))
	}
}

func TestRegistryService_ListModels_WithEntries(t *testing.T) {
	dir := t.TempDir()
	// Create fake .gguf files.
	for _, name := range []string{"model-a.gguf", "model-b.gguf"} {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
		f.Close()
	}

	reg := modelstore.New()
	if err := reg.Scan(dir, false); err != nil {
		t.Fatalf("scan: %v", err)
	}

	svc := service.NewRegistryService(reg)
	models, err := svc.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(models))
	}

	// Each model should have OwnedBy set.
	for _, m := range models {
		if m.OwnedBy == "" {
			t.Fatalf("expected OwnedBy to be set for model %q", m.ID)
		}
	}
}

func TestRegistryService_ListModels_DerivesIDFromFilename(t *testing.T) {
	dir := t.TempDir()
	f, err := os.Create(filepath.Join(dir, "llama-7b-q4.gguf"))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	f.Close()

	reg := modelstore.New()
	if err := reg.Scan(dir, false); err != nil {
		t.Fatalf("scan: %v", err)
	}

	svc := service.NewRegistryService(reg)
	models, err := svc.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if models[0].ID != "llama-7b-q4" {
		t.Fatalf("expected ID=llama-7b-q4, got %q", models[0].ID)
	}
}

func TestRegistryService_InferenceMethods_Unavailable(t *testing.T) {
	reg := modelstore.New()
	svc := service.NewRegistryService(reg)
	ctx := context.Background()

	if _, err := svc.ChatComplete(ctx, service.ChatRequest{}); !errors.Is(err, service.ErrInferenceNotAvailable) {
		t.Fatalf("expected ErrInferenceNotAvailable from ChatComplete, got %v", err)
	}
	if _, err := svc.Complete(ctx, service.TextRequest{}); !errors.Is(err, service.ErrInferenceNotAvailable) {
		t.Fatalf("expected ErrInferenceNotAvailable from Complete, got %v", err)
	}
	if _, err := svc.Embed(ctx, service.EmbeddingRequest{}); !errors.Is(err, service.ErrInferenceNotAvailable) {
		t.Fatalf("expected ErrInferenceNotAvailable from Embed, got %v", err)
	}
}

func TestRegistryService_NonExistentDir(t *testing.T) {
	reg := modelstore.New()
	// Scanning a non-existent dir should be a no-op (not an error).
	if err := reg.Scan("/nonexistent/path/to/models", false); err != nil {
		t.Fatalf("expected no error for non-existent dir, got %v", err)
	}
	svc := service.NewRegistryService(reg)
	models, err := svc.ListModels(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("expected no models, got %d", len(models))
	}
}

func TestRegistryService_LoadModel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "loaded.gguf")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	f.Close()

	svc := service.NewRegistryService(modelstore.New())
	model, err := svc.LoadModel(context.Background(), service.ModelLoadRequest{Path: path})
	if err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	if model.ID != "loaded" {
		t.Fatalf("expected loaded, got %q", model.ID)
	}
}

func TestRegistryService_LoadModel_Errors(t *testing.T) {
	svc := service.NewRegistryService(modelstore.New())
	if _, err := svc.LoadModel(context.Background(), service.ModelLoadRequest{}); !errors.Is(err, service.ErrModelPathRequired) {
		t.Fatalf("expected ErrModelPathRequired, got %v", err)
	}
	if _, err := svc.LoadModel(context.Background(), service.ModelLoadRequest{Path: "/missing.gguf"}); !errors.Is(err, service.ErrInvalidModelPath) {
		t.Fatalf("expected ErrInvalidModelPath, got %v", err)
	}
}

func TestRegistryService_UnloadModel(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "loaded.gguf")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	f.Close()

	svc := service.NewRegistryService(modelstore.New())
	if _, err := svc.LoadModel(context.Background(), service.ModelLoadRequest{Path: path}); err != nil {
		t.Fatalf("LoadModel: %v", err)
	}
	if err := svc.UnloadModel(context.Background(), "loaded"); err != nil {
		t.Fatalf("UnloadModel: %v", err)
	}
	if err := svc.UnloadModel(context.Background(), "loaded"); !errors.Is(err, service.ErrModelNotFound) {
		t.Fatalf("expected ErrModelNotFound, got %v", err)
	}
}
