package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	c := Default()
	if c.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %q", c.Host)
	}
	if c.Port != 8080 {
		t.Errorf("expected port 8080, got %d", c.Port)
	}
	if c.ShutdownTimeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", c.ShutdownTimeout)
	}
	if c.LogLevel != "info" {
		t.Errorf("expected log level info, got %q", c.LogLevel)
	}
}

func TestValidate_Valid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Errorf("default config should be valid: %v", err)
	}
}

func TestValidate_EmptyHost(t *testing.T) {
	c := Default()
	c.Host = ""
	if err := c.Validate(); err == nil {
		t.Error("expected error for empty host")
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	c := Default()
	c.Port = 99999
	if err := c.Validate(); err == nil {
		t.Error("expected error for out-of-range port")
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	c := Default()
	c.LogLevel = "verbose"
	if err := c.Validate(); err == nil {
		t.Error("expected error for unknown log level")
	}
}

func TestValidate_InvalidLogFormat(t *testing.T) {
	c := Default()
	c.LogFormat = "yaml"
	if err := c.Validate(); err == nil {
		t.Error("expected error for unknown log format")
	}
}

func TestLoadFile(t *testing.T) {
	data := map[string]any{
		"host": "0.0.0.0",
		"port": 9090,
	}
	path := writeJSON(t, data)

	c := Default()
	if err := c.LoadFile(path); err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if c.Host != "0.0.0.0" {
		t.Errorf("expected 0.0.0.0, got %q", c.Host)
	}
	if c.Port != 9090 {
		t.Errorf("expected 9090, got %d", c.Port)
	}
	// Fields absent from file keep their default.
	if c.LogLevel != "info" {
		t.Errorf("expected log level info (default), got %q", c.LogLevel)
	}
}

func TestLoadFile_Missing(t *testing.T) {
	c := Default()
	if err := c.LoadFile("/nonexistent/path/config.json"); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFile_BadJSON(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "cfg*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	f.WriteString("{bad json}")
	f.Close()

	c := Default()
	if err := c.LoadFile(f.Name()); err == nil {
		t.Error("expected error for bad JSON")
	}
}

func TestApplyEnv(t *testing.T) {
	t.Setenv("PACKLLAMA_HOST", "0.0.0.0")
	t.Setenv("PACKLLAMA_PORT", "9000")
	t.Setenv("PACKLLAMA_LOG_LEVEL", "debug")
	t.Setenv("PACKLLAMA_LOG_FORMAT", "json")
	t.Setenv("PACKLLAMA_DISABLE_UI", "true")
	t.Setenv("PACKLLAMA_ALLOWED_ORIGINS", "http://a.example, http://b.example")
	t.Setenv("PACKLLAMA_SHUTDOWN_TIMEOUT", "10s")

	c := Default()
	c.ApplyEnv()

	if c.Host != "0.0.0.0" {
		t.Errorf("expected 0.0.0.0, got %q", c.Host)
	}
	if c.Port != 9000 {
		t.Errorf("expected 9000, got %d", c.Port)
	}
	if c.LogLevel != "debug" {
		t.Errorf("expected debug, got %q", c.LogLevel)
	}
	if c.LogFormat != "json" {
		t.Errorf("expected json, got %q", c.LogFormat)
	}
	if !c.DisableUI {
		t.Error("expected DisableUI=true")
	}
	if len(c.AllowedOrigins) != 2 {
		t.Errorf("expected 2 origins, got %v", c.AllowedOrigins)
	}
	if c.ShutdownTimeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", c.ShutdownTimeout)
	}
}

func TestApplyEnv_InvalidPort(t *testing.T) {
	t.Setenv("PACKLLAMA_PORT", "notanumber")
	c := Default()
	c.ApplyEnv()
	if c.Port != 8080 {
		t.Errorf("invalid port env should be ignored; got %d", c.Port)
	}
}

func TestIsTruthy(t *testing.T) {
	truthy := []string{"1", "true", "TRUE", "yes", "YES", "on", "ON"}
	for _, v := range truthy {
		if !isTruthy(v) {
			t.Errorf("%q should be truthy", v)
		}
	}
	falsy := []string{"0", "false", "no", "off", ""}
	for _, v := range falsy {
		if isTruthy(v) {
			t.Errorf("%q should not be truthy", v)
		}
	}
}

func writeJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestApplyEnv_NewFields(t *testing.T) {
	t.Setenv("PACKLLAMA_LOG_REQUESTS", "true")
	t.Setenv("PACKLLAMA_LOG_RESPONSES", "1")
	t.Setenv("PACKLLAMA_PRELOAD_MODELS", "llama-7b, codellama-13b")

	c := Default()
	c.ApplyEnv()

	if !c.LogRequests {
		t.Error("expected LogRequests=true")
	}
	if !c.LogResponses {
		t.Error("expected LogResponses=true")
	}
	if len(c.PreloadModels) != 2 {
		t.Errorf("expected 2 preload models, got %v", c.PreloadModels)
	}
}

func TestModelOverrides_RoundTrip(t *testing.T) {
	temp := 0.5
	tokens := 512
	cfg := Config{
		ModelOverrides: map[string]ModelParams{
			"llama-7b": {
				Temperature: &temp,
				MaxTokens:   &tokens,
				Stop:        []string{"</s>"},
			},
		},
	}
	path := writeJSON(t, cfg)
	loaded := Default()
	if err := loaded.LoadFile(path); err != nil {
		t.Fatalf("load: %v", err)
	}
	p, ok := loaded.ModelOverrides["llama-7b"]
	if !ok {
		t.Fatal("expected llama-7b override to be present")
	}
	if p.Temperature == nil || *p.Temperature != 0.5 {
		t.Errorf("unexpected Temperature: %v", p.Temperature)
	}
	if p.MaxTokens == nil || *p.MaxTokens != 512 {
		t.Errorf("unexpected MaxTokens: %v", p.MaxTokens)
	}
	if len(p.Stop) != 1 || p.Stop[0] != "</s>" {
		t.Errorf("unexpected Stop: %v", p.Stop)
	}
}
