// Package config provides configuration loading and validation for packllama.
// Configuration can be supplied via a JSON file, environment variables, or
// command-line flags; flags take precedence over environment variables which
// take precedence over the file.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// ModelParams holds inference parameter overrides for a specific model.
// Any field left at its zero value is ignored; the request or global default is used.
type ModelParams struct {
	// Temperature overrides the sampling temperature (range 0.0–2.0). Nil = no override.
	Temperature *float64 `json:"temperature,omitempty"`
	// TopP overrides nucleus sampling probability (range 0.0–1.0). Nil = no override.
	TopP *float64 `json:"top_p,omitempty"`
	// MaxTokens overrides the maximum number of generated tokens. Nil = no override.
	MaxTokens *int `json:"max_tokens,omitempty"`
	// Stop overrides the stop sequences. Nil = no override.
	Stop []string `json:"stop,omitempty"`
}

// Config holds the complete runtime configuration for packllama.
type Config struct {
	// Server settings.
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// CORS.
	AllowedOrigins []string `json:"allowed_origins"`

	// Logging.
	LogLevel    string `json:"log_level"`
	LogFormat   string `json:"log_format"`    // "text" or "json"
	LogRequests bool   `json:"log_requests"`  // log request body (verbose)
	LogResponses bool  `json:"log_responses"` // log response body (verbose)

	// Inference.
	ModelsDir    string `json:"models_dir"`
	DefaultModel string `json:"default_model"`
	// PreloadModels lists model IDs to load into the inference backend at startup.
	PreloadModels []string `json:"preload_models,omitempty"`
	// ModelOverrides maps model IDs to parameter overrides that supersede global
	// defaults when that model is used for inference.
	ModelOverrides map[string]ModelParams `json:"model_overrides,omitempty"`

	// Behaviour.
	DisableUI bool `json:"disable_ui"`
}

// Default returns a Config with sensible defaults applied.
func Default() Config {
	return Config{
		Host:            "127.0.0.1",
		Port:            8080,
		ShutdownTimeout: 5 * time.Second,
		LogLevel:        "info",
		LogFormat:       "text",
		DisableUI:       false,
	}
}

// Validate returns an error if any field holds an invalid value.
func (c Config) Validate() error {
	if c.Host == "" {
		return errors.New("host must not be empty")
	}
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("port %d is out of range [0, 65535]", c.Port)
	}
	if c.ShutdownTimeout < 0 {
		return errors.New("shutdown_timeout must not be negative")
	}
	switch strings.ToLower(c.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log_level %q must be one of: debug, info, warn, error", c.LogLevel)
	}
	switch strings.ToLower(c.LogFormat) {
	case "text", "json":
	default:
		return fmt.Errorf("log_format %q must be one of: text, json", c.LogFormat)
	}
	return nil
}

// LoadFile reads a JSON configuration file into c, leaving fields that are not
// present in the file at their existing values.
func (c *Config) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(c); err != nil {
		return fmt.Errorf("decode config file: %w", err)
	}
	return nil
}

// ApplyEnv overwrites fields in c with values found in the environment.
// All environment variables are prefixed with PACKLLAMA_.
func (c *Config) ApplyEnv() {
	if v := os.Getenv("PACKLLAMA_HOST"); v != "" {
		c.Host = v
	}
	if v := os.Getenv("PACKLLAMA_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Port = n
		}
	}
	if v := os.Getenv("PACKLLAMA_SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.ShutdownTimeout = d
		}
	}
	if v := os.Getenv("PACKLLAMA_ALLOWED_ORIGINS"); v != "" {
		c.AllowedOrigins = splitComma(v)
	}
	if v := os.Getenv("PACKLLAMA_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	if v := os.Getenv("PACKLLAMA_LOG_FORMAT"); v != "" {
		c.LogFormat = v
	}
	if v := os.Getenv("PACKLLAMA_LOG_REQUESTS"); v != "" {
		c.LogRequests = isTruthy(v)
	}
	if v := os.Getenv("PACKLLAMA_LOG_RESPONSES"); v != "" {
		c.LogResponses = isTruthy(v)
	}
	if v := os.Getenv("PACKLLAMA_MODELS_DIR"); v != "" {
		c.ModelsDir = v
	}
	if v := os.Getenv("PACKLLAMA_DEFAULT_MODEL"); v != "" {
		c.DefaultModel = v
	}
	if v := os.Getenv("PACKLLAMA_PRELOAD_MODELS"); v != "" {
		c.PreloadModels = splitComma(v)
	}
	if v := os.Getenv("PACKLLAMA_DISABLE_UI"); v != "" {
		c.DisableUI = isTruthy(v)
	}
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func isTruthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
