package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// TerminalOutput is a read-only, scrollable terminal-style display widget.
// It strips common ANSI escape sequences before rendering.
type TerminalOutput struct {
	baseWidget
	lines        []string
	scrollOffset int
	maxLines     int
}

// NewTerminalOutput creates a TerminalOutput widget with the given theme.
// maxLines caps history (0 = unlimited).
func NewTerminalOutput(maxLines int, theme Theme) *TerminalOutput {
	return &TerminalOutput{
		baseWidget: baseWidget{theme: theme},
		maxLines:   maxLines,
	}
}

// Focusable reports that the terminal output accepts keyboard focus.
func (t *TerminalOutput) Focusable() bool { return true }

// Write appends text to the terminal, splitting on newlines and stripping ANSI codes.
func (t *TerminalOutput) Write(data string) {
	clean := stripANSI(data)
	parts := strings.Split(clean, "\n")
	if len(t.lines) > 0 && len(parts) > 0 {
		t.lines[len(t.lines)-1] += parts[0]
		parts = parts[1:]
	}
	t.lines = append(t.lines, parts...)
	t.trimHistory()
	t.scrollToBottom()
}

// Clear removes all output from the terminal.
func (t *TerminalOutput) Clear() {
	t.lines = nil
	t.scrollOffset = 0
}

// Update handles mouse-wheel scrolling.
func (t *TerminalOutput) Update() error {
	mx, my := ebiten.CursorPosition()
	if !inBounds(t.bounds, mx, my) {
		return nil
	}
	_, dy := ebiten.Wheel()
	if dy > 0 && t.scrollOffset > 0 {
		t.scrollOffset--
	} else if dy < 0 {
		t.scrollOffset++
		t.clampScroll()
	}
	return nil
}

// Draw renders visible terminal lines.
func (t *TerminalOutput) Draw(screen *ebiten.Image) {
	fillRect(screen, t.bounds, t.theme.Background)
	drawBorder(screen, t.bounds, t.theme.BorderWidth, t.theme.Border)
	visible := t.visibleCount()
	x := t.bounds.Min.X + t.theme.Padding
	y := t.bounds.Min.Y + t.theme.Padding
	for i := t.scrollOffset; i < t.scrollOffset+visible && i < len(t.lines); i++ {
		drawText(screen, t.lines[i], x, y, t.theme.Text)
		y += CharHeight + 2
	}
}

// stripANSI removes ANSI escape sequences (e.g. \x1b[...m) from s.
func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // consume 'm'
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// trimHistory removes the oldest lines when maxLines is exceeded.
func (t *TerminalOutput) trimHistory() {
	if t.maxLines > 0 && len(t.lines) > t.maxLines {
		t.lines = t.lines[len(t.lines)-t.maxLines:]
	}
}

// scrollToBottom scrolls to show the most recent output.
func (t *TerminalOutput) scrollToBottom() {
	visible := t.visibleCount()
	off := len(t.lines) - visible
	if off < 0 {
		off = 0
	}
	t.scrollOffset = off
}

// visibleCount returns the number of lines fitting inside the bounds.
func (t *TerminalOutput) visibleCount() int {
	inner := t.bounds.Dy() - 2*t.theme.Padding
	if inner <= 0 {
		return 0
	}
	return inner / (CharHeight + 2)
}

// clampScroll keeps scrollOffset within valid range.
func (t *TerminalOutput) clampScroll() {
	visible := t.visibleCount()
	max := len(t.lines) - visible
	if max < 0 {
		max = 0
	}
	if t.scrollOffset > max {
		t.scrollOffset = max
	}
}
