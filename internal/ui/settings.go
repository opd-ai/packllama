package ui

import (
	"encoding/json"
	"fmt"
	"os"
)

// UserPreferences holds all persisted UI settings.
type UserPreferences struct {
	// ThemeName is "dark" or "light".
	ThemeName string `json:"theme"`
	// FontScale is a multiplier for future scalable-font support (default 1.0).
	FontScale float64 `json:"font_scale"`
	// DefaultModel is the model ID selected by default at startup.
	DefaultModel string `json:"default_model"`
	// AutoSave controls whether conversations are auto-saved on every message.
	AutoSave bool `json:"auto_save"`
	// ConversationsDir is the directory where conversations are persisted.
	ConversationsDir string `json:"conversations_dir"`
}

// DefaultPreferences returns sensible default user preferences.
func DefaultPreferences() UserPreferences {
	return UserPreferences{
		ThemeName: "dark",
		FontScale: 1.0,
		AutoSave:  true,
	}
}

// Theme returns the Theme corresponding to the ThemeName field.
func (p UserPreferences) Theme() Theme {
	if p.ThemeName == "light" {
		return DefaultLight()
	}
	return DefaultDark()
}

// Save writes the preferences as JSON to path.
func (p UserPreferences) Save(path string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal preferences: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write preferences: %w", err)
	}
	return nil
}

// LoadPreferences reads preferences from path.
// Returns DefaultPreferences on any error.
func LoadPreferences(path string) (UserPreferences, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultPreferences(), fmt.Errorf("read preferences: %w", err)
	}
	var p UserPreferences
	if err := json.Unmarshal(data, &p); err != nil {
		return DefaultPreferences(), fmt.Errorf("unmarshal preferences: %w", err)
	}
	if p.FontScale <= 0 {
		p.FontScale = 1.0
	}
	return p, nil
}
