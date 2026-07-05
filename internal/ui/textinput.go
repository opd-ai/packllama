package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// TextInput is a single-line editable text field.
type TextInput struct {
	baseWidget
	value       []rune
	cursor      int
	placeholder string
	blinkTick   int
}

// NewTextInput creates an empty TextInput with an optional placeholder.
func NewTextInput(placeholder string, theme Theme) *TextInput {
	return &TextInput{
		baseWidget:  baseWidget{theme: theme},
		placeholder: placeholder,
	}
}

// Focusable reports that text inputs accept keyboard focus.
func (t *TextInput) Focusable() bool { return true }

// Value returns the current text content.
func (t *TextInput) Value() string { return string(t.value) }

// SetValue replaces the text content and moves the cursor to the end.
func (t *TextInput) SetValue(s string) {
	t.value = []rune(s)
	t.cursor = len(t.value)
}

// Update handles character input, backspace, and cursor movement.
func (t *TextInput) Update() error {
	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		t.focused = inBounds(t.bounds, mx, my)
	}
	if !t.focused {
		return nil
	}
	t.blinkTick++
	t.handleChars()
	t.handleSpecialKeys()
	return nil
}

// handleChars appends printable input characters at the cursor position.
func (t *TextInput) handleChars() {
	for _, r := range ebiten.AppendInputChars(nil) {
		t.value = append(t.value[:t.cursor], append([]rune{r}, t.value[t.cursor:]...)...)
		t.cursor++
	}
}

// handleSpecialKeys processes Backspace, Delete, and arrow navigation.
func (t *TextInput) handleSpecialKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && t.cursor > 0 {
		t.value = append(t.value[:t.cursor-1], t.value[t.cursor:]...)
		t.cursor--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) && t.cursor < len(t.value) {
		t.value = append(t.value[:t.cursor], t.value[t.cursor+1:]...)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && t.cursor > 0 {
		t.cursor--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && t.cursor < len(t.value) {
		t.cursor++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		t.cursor = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		t.cursor = len(t.value)
	}
}

// Draw renders the text field border, content, placeholder, and cursor.
func (t *TextInput) Draw(screen *ebiten.Image) {
	fillRect(screen, t.bounds, t.theme.Surface)
	borderClr := t.theme.Border
	if t.focused {
		borderClr = t.theme.Primary
	}
	drawBorder(screen, t.bounds, t.theme.BorderWidth, borderClr)
	x := t.bounds.Min.X + t.theme.Padding
	y := t.bounds.Min.Y + (t.bounds.Dy()-CharHeight)/2
	if len(t.value) == 0 && !t.focused {
		drawText(screen, t.placeholder, x, y, t.theme.TextMuted)
		return
	}
	drawText(screen, string(t.value), x, y, t.theme.Text)
	if t.focused && (t.blinkTick/30)%2 == 0 {
		cx := x + t.cursor*CharWidth
		drawText(screen, "|", cx, y, t.theme.Primary)
	}
}
