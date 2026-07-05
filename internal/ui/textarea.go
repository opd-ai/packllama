package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// TextArea is a read-only, scrollable multi-line text display widget.
type TextArea struct {
	baseWidget
	lines        []string
	scrollOffset int // first visible line index
}

// NewTextArea creates an empty TextArea with the given theme.
func NewTextArea(theme Theme) *TextArea {
	return &TextArea{baseWidget: baseWidget{theme: theme}}
}

// Focusable reports that text areas do not accept keyboard focus.
func (t *TextArea) Focusable() bool { return false }

// SetText replaces the displayed content with the given string, splitting on newlines.
func (t *TextArea) SetText(s string) {
	t.lines = strings.Split(s, "\n")
	t.clampScroll()
}

// AppendLine appends a single line of text and scrolls to the bottom.
func (t *TextArea) AppendLine(line string) {
	t.lines = append(t.lines, line)
	t.scrollToBottom()
}

// ScrollToBottom scrolls to show the last line.
func (t *TextArea) ScrollToBottom() { t.scrollToBottom() }

// Update handles mouse-wheel scrolling when the cursor is over the widget.
func (t *TextArea) Update() error {
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

// Draw renders the visible text lines inside the text area's bounds.
func (t *TextArea) Draw(screen *ebiten.Image) {
	fillRect(screen, t.bounds, t.theme.Surface)
	drawBorder(screen, t.bounds, t.theme.BorderWidth, t.theme.Border)

	visible := t.visibleLineCount()
	y := t.bounds.Min.Y + t.theme.Padding
	x := t.bounds.Min.X + t.theme.Padding
	for i := 0; i < visible; i++ {
		idx := t.scrollOffset + i
		if idx >= len(t.lines) {
			break
		}
		drawText(screen, t.lines[idx], x, y, t.theme.Text)
		y += CharHeight + 2
	}
}

// visibleLineCount returns the number of lines that fit in the current bounds.
func (t *TextArea) visibleLineCount() int {
	inner := t.bounds.Dy() - 2*t.theme.Padding
	if inner <= 0 {
		return 0
	}
	return inner / (CharHeight + 2)
}

// scrollToBottom scrolls so the last line is visible.
func (t *TextArea) scrollToBottom() {
	visible := t.visibleLineCount()
	if len(t.lines) > visible {
		t.scrollOffset = len(t.lines) - visible
	}
}

// clampScroll ensures scrollOffset is within valid range.
func (t *TextArea) clampScroll() {
	visible := t.visibleLineCount()
	maxOff := len(t.lines) - visible
	if maxOff < 0 {
		maxOff = 0
	}
	if t.scrollOffset > maxOff {
		t.scrollOffset = maxOff
	}
	if t.scrollOffset < 0 {
		t.scrollOffset = 0
	}
}
