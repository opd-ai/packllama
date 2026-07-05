package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// MultiLineInput is an editable multi-line text field.
// Pressing Enter submits (calls onSubmit if set); Shift+Enter inserts a newline.
type MultiLineInput struct {
	baseWidget
	lines       [][]rune
	cursorRow   int
	cursorCol   int
	blinkTick   int
	placeholder string
	onSubmit    func(text string)
}

// NewMultiLineInput creates an empty input field with an optional placeholder.
func NewMultiLineInput(placeholder string, theme Theme) *MultiLineInput {
	return &MultiLineInput{
		baseWidget:  baseWidget{theme: theme},
		lines:       [][]rune{{}},
		placeholder: placeholder,
	}
}

// Focusable reports that the multi-line input accepts keyboard focus.
func (m *MultiLineInput) Focusable() bool { return true }

// OnSubmit registers a callback invoked with the current text on Enter.
func (m *MultiLineInput) OnSubmit(fn func(text string)) { m.onSubmit = fn }

// Value returns the current text with newlines between lines.
func (m *MultiLineInput) Value() string {
	result := make([]rune, 0, 128)
	for i, line := range m.lines {
		result = append(result, line...)
		if i < len(m.lines)-1 {
			result = append(result, '\n')
		}
	}
	return string(result)
}

// Clear resets the input to an empty state.
func (m *MultiLineInput) Clear() {
	m.lines = [][]rune{{}}
	m.cursorRow, m.cursorCol = 0, 0
}

// Update handles character input and special keys.
func (m *MultiLineInput) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if inBounds(m.bounds, mx, my) {
			m.grabFocus()
		} else if m.focusCallback == nil {
			m.focused = false
		}
	}
	if !m.focused {
		return nil
	}
	m.blinkTick++
	m.handleEnter()
	m.handleChars()
	m.handleEditKeys()
	return nil
}

// handleEnter processes Enter (submit) and Shift+Enter (newline).
func (m *MultiLineInput) handleEnter() {
	if !inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return
	}
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		m.insertNewline()
		return
	}
	if m.onSubmit != nil {
		m.onSubmit(m.Value())
	}
	m.Clear()
}

// insertNewline splits the current line at the cursor position.
func (m *MultiLineInput) insertNewline() {
	cur := m.lines[m.cursorRow]
	before, after := append([]rune{}, cur[:m.cursorCol]...), append([]rune{}, cur[m.cursorCol:]...)
	m.lines = append(m.lines[:m.cursorRow], append([][]rune{before, after}, m.lines[m.cursorRow+1:]...)...)
	m.cursorRow++
	m.cursorCol = 0
}

// handleChars appends typed runes at the cursor position.
func (m *MultiLineInput) handleChars() {
	for _, r := range ebiten.AppendInputChars(nil) {
		line := m.lines[m.cursorRow]
		line = append(line[:m.cursorCol], append([]rune{r}, line[m.cursorCol:]...)...)
		m.lines[m.cursorRow] = line
		m.cursorCol++
	}
}

// handleEditKeys processes Backspace, Delete, and arrow navigation.
func (m *MultiLineInput) handleEditKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		m.backspace()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && m.cursorCol > 0 {
		m.cursorCol--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && m.cursorCol < len(m.lines[m.cursorRow]) {
		m.cursorCol++
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && m.cursorRow > 0 {
		m.cursorRow--
		m.clampCol()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && m.cursorRow < len(m.lines)-1 {
		m.cursorRow++
		m.clampCol()
	}
}

// backspace deletes the character before the cursor, merging lines when needed.
func (m *MultiLineInput) backspace() {
	if m.cursorCol > 0 {
		line := m.lines[m.cursorRow]
		m.lines[m.cursorRow] = append(line[:m.cursorCol-1], line[m.cursorCol:]...)
		m.cursorCol--
		return
	}
	if m.cursorRow > 0 {
		prev := m.lines[m.cursorRow-1]
		m.cursorCol = len(prev)
		merged := append(append([]rune{}, prev...), m.lines[m.cursorRow]...)
		m.lines = append(m.lines[:m.cursorRow-1], append([][]rune{merged}, m.lines[m.cursorRow+1:]...)...)
		m.cursorRow--
	}
}

// clampCol ensures cursorCol is within the current line's length.
func (m *MultiLineInput) clampCol() {
	if m.cursorCol > len(m.lines[m.cursorRow]) {
		m.cursorCol = len(m.lines[m.cursorRow])
	}
}

// Draw renders the multi-line input with border, text, and cursor.
func (m *MultiLineInput) Draw(screen *ebiten.Image) {
	fillRect(screen, m.bounds, m.theme.Surface)
	borderClr := m.theme.Border
	if m.focused {
		borderClr = m.theme.Primary
	}
	drawBorder(screen, m.bounds, m.theme.BorderWidth, borderClr)
	if m.isEmpty() && !m.focused {
		drawText(screen, m.placeholder,
			m.bounds.Min.X+m.theme.Padding,
			m.bounds.Min.Y+m.theme.Padding,
			m.theme.TextMuted)
		return
	}
	m.drawLines(screen)
}

// drawLines renders each text line and the blinking cursor.
func (m *MultiLineInput) drawLines(screen *ebiten.Image) {
	x := m.bounds.Min.X + m.theme.Padding
	y := m.bounds.Min.Y + m.theme.Padding
	for row, line := range m.lines {
		drawText(screen, string(line), x, y, m.theme.Text)
		if m.focused && row == m.cursorRow && (m.blinkTick/30)%2 == 0 {
			cx := x + m.cursorCol*CharWidth
			drawText(screen, "|", cx, y, m.theme.Primary)
		}
		y += CharHeight + 2
		if y > m.bounds.Max.Y-m.theme.Padding {
			break
		}
	}
}

// isEmpty reports whether all lines are empty.
func (m *MultiLineInput) isEmpty() bool {
	for _, line := range m.lines {
		if len(line) > 0 {
			return false
		}
	}
	return true
}
