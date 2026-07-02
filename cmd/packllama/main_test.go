package main

import (
	"bytes"
	"log/slog"
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
