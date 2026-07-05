package main

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opd-ai/packllama/internal/config"
)

func TestBuildService_WarnsOnUnknownDefaultModel(t *testing.T) {
	cfg := config.Default()
	cfg.ModelsDir = t.TempDir()
	cfg.DefaultModel = "missing-model"

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	svc := buildService(cfg, logger)
	if svc == nil {
		t.Fatal("expected registry service")
	}
	if !strings.Contains(logs.String(), "default model not found during discovery") {
		t.Fatalf("expected warning about missing default model, got %q", logs.String())
	}
}

func TestBuildService_AutoDownloadsModels(t *testing.T) {
	cfg := config.Default()
	cfg.ModelsDir = t.TempDir()
	cfg.ModelDownloads = []string{"owner/repo/model.gguf"}

	var got []string
	orig := downloadHuggingFaceModel
	downloadHuggingFaceModel = func(_ context.Context, modelsDir, ref string) (string, error) {
		got = append(got, ref)
		path := filepath.Join(modelsDir, "model.gguf")
		return path, os.WriteFile(path, []byte("x"), 0o600)
	}
	t.Cleanup(func() { downloadHuggingFaceModel = orig })

	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	if svc := buildService(cfg, logger); svc == nil {
		t.Fatal("expected registry service")
	}
	if len(got) != 1 || got[0] != "owner/repo/model.gguf" {
		t.Fatalf("unexpected downloaded refs: %v", got)
	}
}
