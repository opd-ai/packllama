package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// StatusBar displays inference metadata: model name, tokens/sec, and token counts.
type StatusBar struct {
	baseWidget
	model            string
	tokensPerSec     float64
	promptTokens     int
	completionTokens int
}

// NewStatusBar creates a StatusBar with the given theme.
func NewStatusBar(theme Theme) *StatusBar {
	return &StatusBar{baseWidget: baseWidget{theme: theme}}
}

// Focusable reports that the status bar does not accept keyboard focus.
func (s *StatusBar) Focusable() bool { return false }

// SetModel updates the displayed model name.
func (s *StatusBar) SetModel(name string) { s.model = name }

// SetTokensPerSec updates the displayed generation speed.
func (s *StatusBar) SetTokensPerSec(tps float64) { s.tokensPerSec = tps }

// SetTokenCounts updates the displayed prompt and completion token counts.
func (s *StatusBar) SetTokenCounts(prompt, completion int) {
	s.promptTokens = prompt
	s.completionTokens = completion
}

// Update is a no-op for the status bar (display only).
func (s *StatusBar) Update() error { return nil }

// Draw renders the status bar background and text metrics.
func (s *StatusBar) Draw(screen *ebiten.Image) {
	fillRect(screen, s.bounds, s.theme.Surface)
	drawBorder(screen, s.bounds, s.theme.BorderWidth, s.theme.Border)
	x := s.bounds.Min.X + s.theme.Padding
	y := s.bounds.Min.Y + (s.bounds.Dy()-CharHeight)/2
	drawText(screen, s.statusText(), x, y, s.theme.TextMuted)
}

// statusText composes the one-line status string.
func (s *StatusBar) statusText() string {
	model := s.model
	if model == "" {
		model = "(no model)"
	}
	return fmt.Sprintf("model: %s  |  %.1f tok/s  |  prompt: %d  completion: %d  total: %d",
		model, s.tokensPerSec,
		s.promptTokens, s.completionTokens,
		s.promptTokens+s.completionTokens)
}
