package ui

import (
	"encoding/json"
	"fmt"
	"os"
)

// InferenceParams holds all adjustable inference parameters.
type InferenceParams struct {
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	MaxTokens   int     `json:"max_tokens"`
	ContextLen  int     `json:"context_length"`
	RepeatPen   float64 `json:"repeat_penalty"`
	TopK        int     `json:"top_k"`
	Seed        int64   `json:"seed"`
}

// DefaultParams returns a balanced set of inference defaults.
func DefaultParams() InferenceParams {
	return InferenceParams{
		Temperature: 0.7,
		TopP:        0.9,
		MaxTokens:   2048,
		ContextLen:  4096,
		RepeatPen:   1.1,
		TopK:        40,
		Seed:        -1,
	}
}

// ParamPreset returns a named preset (creative, balanced, or precise).
// Unrecognised names return DefaultParams.
func ParamPreset(name string) InferenceParams {
	switch name {
	case "creative":
		p := DefaultParams()
		p.Temperature = 1.2
		p.TopP = 0.95
		p.TopK = 60
		return p
	case "precise":
		p := DefaultParams()
		p.Temperature = 0.2
		p.TopP = 0.7
		p.TopK = 20
		return p
	default: // balanced
		return DefaultParams()
	}
}

// Validate checks that all parameter values are within acceptable ranges.
// It returns a non-nil error describing the first violation found.
func (p InferenceParams) Validate() error {
	if p.Temperature < 0 || p.Temperature > 2 {
		return fmt.Errorf("temperature must be in [0, 2], got %.2f", p.Temperature)
	}
	if p.TopP < 0 || p.TopP > 1 {
		return fmt.Errorf("top_p must be in [0, 1], got %.2f", p.TopP)
	}
	if p.MaxTokens < 1 {
		return fmt.Errorf("max_tokens must be >= 1, got %d", p.MaxTokens)
	}
	if p.ContextLen < 1 {
		return fmt.Errorf("context_length must be >= 1, got %d", p.ContextLen)
	}
	if p.RepeatPen < 0 {
		return fmt.Errorf("repeat_penalty must be >= 0, got %.2f", p.RepeatPen)
	}
	return nil
}

// Save writes the parameters as JSON to path.
func (p InferenceParams) Save(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write params: %w", err)
	}
	return nil
}

// LoadParams reads InferenceParams from path.
func LoadParams(path string) (InferenceParams, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultParams(), fmt.Errorf("read params: %w", err)
	}
	var p InferenceParams
	if err := json.Unmarshal(data, &p); err != nil {
		return DefaultParams(), fmt.Errorf("unmarshal params: %w", err)
	}
	return p, nil
}
