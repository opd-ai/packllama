package ui

import (
	"fmt"
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ChatView is a scrollable widget that displays a conversation history.
// It supports streaming updates and simple message editing/deletion.
type ChatView struct {
	baseWidget
	messages     []displayMessage
	scrollOffset int // first visible line index
	dirty        bool
	cachedLines  []displayLine
	selected     int // selected message index for edit/delete (-1 = none)
}

// displayMessage tracks one chat message in the view.
type displayMessage struct {
	role    string
	content string
}

// NewChatView creates an empty ChatView with the given theme.
func NewChatView(theme Theme) *ChatView {
	return &ChatView{
		baseWidget: baseWidget{theme: theme},
		selected:   -1,
		dirty:      true,
	}
}

// Focusable reports that the chat view can receive keyboard focus for scrolling.
func (cv *ChatView) Focusable() bool { return true }

// AppendMessage adds a new message and scrolls to the bottom.
func (cv *ChatView) AppendMessage(role, content string) {
	cv.messages = append(cv.messages, displayMessage{role: role, content: content})
	cv.dirty = true
	cv.scrollToBottom()
}

// UpdateStreamChunk appends delta text to the last message.
// If no messages exist, it creates an assistant message.
func (cv *ChatView) UpdateStreamChunk(delta string) {
	if len(cv.messages) == 0 {
		cv.messages = append(cv.messages, displayMessage{role: string(RoleAssistant)})
	}
	last := &cv.messages[len(cv.messages)-1]
	last.content += delta
	cv.dirty = true
	cv.scrollToBottom()
}

// DeleteMessage removes the message at index i.
func (cv *ChatView) DeleteMessage(i int) {
	if i < 0 || i >= len(cv.messages) {
		return
	}
	cv.messages = append(cv.messages[:i], cv.messages[i+1:]...)
	cv.dirty = true
	if cv.selected >= len(cv.messages) {
		cv.selected = len(cv.messages) - 1
	}
}

// EditMessage replaces the content of message at index i.
func (cv *ChatView) EditMessage(i int, content string) {
	if i >= 0 && i < len(cv.messages) {
		cv.messages[i].content = content
		cv.dirty = true
	}
}

// Update handles scroll-wheel and keyboard navigation.
func (cv *ChatView) Update() error {
	mx, my := ebiten.CursorPosition()
	if inBounds(cv.bounds, mx, my) {
		cv.handleScroll()
	}
	if cv.focused {
		cv.handleKeys()
	}
	return nil
}

// handleScroll processes mouse wheel events.
func (cv *ChatView) handleScroll() {
	_, dy := ebiten.Wheel()
	if dy > 0 && cv.scrollOffset > 0 {
		cv.scrollOffset--
	} else if dy < 0 {
		cv.scrollOffset++
		cv.clampScroll()
	}
}

// handleKeys handles keyboard navigation over messages.
func (cv *ChatView) handleKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		if cv.selected < len(cv.messages)-1 {
			cv.selected++
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && cv.selected > 0 {
		cv.selected--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) && cv.selected >= 0 {
		cv.DeleteMessage(cv.selected)
	}
}

// Draw renders visible message lines inside the bounds.
func (cv *ChatView) Draw(screen *ebiten.Image) {
	fillRect(screen, cv.bounds, cv.theme.Surface)
	drawBorder(screen, cv.bounds, cv.theme.BorderWidth, cv.theme.Border)

	lines := cv.lines()
	visible := cv.visibleLineCount()
	y := cv.bounds.Min.Y + cv.theme.Padding
	x := cv.bounds.Min.X + cv.theme.Padding

	for i := cv.scrollOffset; i < cv.scrollOffset+visible && i < len(lines); i++ {
		cv.drawLine(screen, lines[i], x, y)
		y += CharHeight + 2
	}
}

// drawLine renders a single display line with appropriate styling.
func (cv *ChatView) drawLine(screen *ebiten.Image, l displayLine, x, y int) {
	lineRect := image.Rect(x, y, cv.bounds.Max.X-cv.theme.Padding, y+CharHeight+2)
	if l.isCode {
		fillRect(screen, lineRect, cv.theme.Background)
	}
	if l.msgIdx >= 0 && l.msgIdx == cv.selected {
		fillRect(screen, lineRect, cv.theme.Hover)
	}
	clr := cv.theme.Text
	if l.isCode {
		clr = cv.theme.Primary
	}
	drawText(screen, l.text, x, y, clr)
}

// lines returns the cached (or freshly computed) display lines.
func (cv *ChatView) lines() []displayLine {
	if cv.dirty {
		cv.cachedLines = cv.computeLines()
		cv.dirty = false
	}
	return cv.cachedLines
}

// computeLines converts all messages into flat display lines.
func (cv *ChatView) computeLines() []displayLine {
	var out []displayLine
	for i, msg := range cv.messages {
		header := fmt.Sprintf("[%s]", msg.role)
		out = append(out, displayLine{text: header, msgIdx: i})
		for _, seg := range parseSegments(msg.content) {
			for _, dl := range segmentLines(seg) {
				dl.msgIdx = i
				out = append(out, dl)
			}
		}
		out = append(out, displayLine{text: "", msgIdx: -1})
	}
	return out
}

// visibleLineCount returns the number of lines fitting in the current bounds.
func (cv *ChatView) visibleLineCount() int {
	inner := cv.bounds.Dy() - 2*cv.theme.Padding
	if inner <= 0 {
		return 0
	}
	return inner / (CharHeight + 2)
}

// scrollToBottom scrolls to display the last lines.
func (cv *ChatView) scrollToBottom() {
	lines := cv.computeLines()
	visible := cv.visibleLineCount()
	off := len(lines) - visible
	if off < 0 {
		off = 0
	}
	cv.scrollOffset = off
}

// clampScroll keeps scrollOffset in valid range.
func (cv *ChatView) clampScroll() {
	lines := cv.lines()
	visible := cv.visibleLineCount()
	max := len(lines) - visible
	if max < 0 {
		max = 0
	}
	if cv.scrollOffset > max {
		cv.scrollOffset = max
	}
	if cv.scrollOffset < 0 {
		cv.scrollOffset = 0
	}
}

// MessageCount returns the number of displayed messages.
func (cv *ChatView) MessageCount() int { return len(cv.messages) }

// SelectedIndex returns the index of the currently selected message, or -1.
func (cv *ChatView) SelectedIndex() int { return cv.selected }

// Clear removes all messages from the view.
func (cv *ChatView) Clear() {
	cv.messages = cv.messages[:0]
	cv.cachedLines = nil
	cv.dirty = true
	cv.scrollOffset = 0
	cv.selected = -1
}

// LoadConversationMessages populates the view from a Conversation.
func (cv *ChatView) LoadConversationMessages(c *Conversation) {
	cv.Clear()
	for _, m := range c.Messages {
		cv.messages = append(cv.messages, displayMessage{
			role:    string(m.Role),
			content: m.Content,
		})
	}
	cv.dirty = true
	cv.scrollToBottom()
}

// wordWrapLine splits a long line into segments fitting within maxChars.
func wordWrapLine(line string, maxChars int) []string {
	if maxChars <= 0 || len(line) <= maxChars {
		return []string{line}
	}
	var out []string
	for len(line) > maxChars {
		cut := strings.LastIndexByte(line[:maxChars], ' ')
		if cut <= 0 {
			cut = maxChars
		}
		out = append(out, line[:cut])
		line = strings.TrimPrefix(line[cut:], " ")
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}
