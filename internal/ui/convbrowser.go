package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ConversationBrowser is a panel that lists conversations and lets the user
// select, create, or delete them.
type ConversationBrowser struct {
	baseWidget
	manager    *ConversationManager
	scrollOffset int
	onSelect   func(i int, c *Conversation)
	onNew      func()
	onDelete   func(i int)
}

// NewConversationBrowser creates a browser backed by the given manager.
func NewConversationBrowser(mgr *ConversationManager, theme Theme) *ConversationBrowser {
	return &ConversationBrowser{
		baseWidget: baseWidget{theme: theme},
		manager:    mgr,
	}
}

// Focusable reports that the browser accepts keyboard focus.
func (cb *ConversationBrowser) Focusable() bool { return true }

// OnSelect registers a callback invoked when a conversation is selected.
func (cb *ConversationBrowser) OnSelect(fn func(i int, c *Conversation)) { cb.onSelect = fn }

// OnNew registers a callback invoked when the "New" action is triggered.
func (cb *ConversationBrowser) OnNew(fn func()) { cb.onNew = fn }

// OnDelete registers a callback invoked when a conversation is deleted.
func (cb *ConversationBrowser) OnDelete(fn func(i int)) { cb.onDelete = fn }

// Update handles mouse click selection and scroll wheel.
func (cb *ConversationBrowser) Update() error {
	mx, my := ebiten.CursorPosition()
	if inBounds(cb.bounds, mx, my) {
		cb.handleScroll()
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && inBounds(cb.bounds, mx, my) {
		cb.handleClick(mx, my)
	}
	if cb.focused {
		cb.handleKeys()
	}
	return nil
}

// handleScroll processes mouse wheel events.
func (cb *ConversationBrowser) handleScroll() {
	_, dy := ebiten.Wheel()
	list := cb.manager.List()
	if dy > 0 && cb.scrollOffset > 0 {
		cb.scrollOffset--
	} else if dy < 0 && cb.scrollOffset < len(list)-cb.visibleCount() {
		cb.scrollOffset++
	}
}

// handleClick selects the conversation row under the cursor.
func (cb *ConversationBrowser) handleClick(mx, my int) {
	row := (my - cb.bounds.Min.Y - cb.theme.Padding) / cb.rowHeight()
	idx := row + cb.scrollOffset
	list := cb.manager.List()
	if idx < 0 || idx >= len(list) {
		return
	}
	cb.manager.SetActive(idx)
	if cb.onSelect != nil {
		cb.onSelect(idx, list[idx])
	}
}

// handleKeys processes Delete (remove selected) and N (new conversation).
func (cb *ConversationBrowser) handleKeys() {
	active := cb.manager.Active()
	list := cb.manager.List()
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) && active != nil {
		for i, c := range list {
			if c == active {
				if cb.onDelete != nil {
					cb.onDelete(i)
				}
				break
			}
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyN) && cb.onNew != nil {
		cb.onNew()
	}
}

// Draw renders the conversation list panel.
func (cb *ConversationBrowser) Draw(screen *ebiten.Image) {
	fillRect(screen, cb.bounds, cb.theme.Surface)
	drawBorder(screen, cb.bounds, cb.theme.BorderWidth, cb.theme.Border)

	list := cb.manager.List()
	visible := cb.visibleCount()
	active := cb.manager.Active()

	for i := 0; i < visible; i++ {
		idx := i + cb.scrollOffset
		if idx >= len(list) {
			break
		}
		cb.drawRow(screen, list[idx], idx, active)
	}
}

// drawRow renders a single conversation entry.
func (cb *ConversationBrowser) drawRow(screen *ebiten.Image, c *Conversation, idx int, active *Conversation) {
	rh := cb.rowHeight()
	y := cb.bounds.Min.Y + cb.theme.Padding + (idx-cb.scrollOffset)*rh
	r := image.Rect(cb.bounds.Min.X+cb.theme.BorderWidth, y,
		cb.bounds.Max.X-cb.theme.BorderWidth, y+rh)
	bg := cb.theme.Surface
	if c == active {
		bg = cb.theme.Active
	}
	fillRect(screen, r, bg)
	title := c.Title
	if title == "" {
		title = c.ID
	}
	drawText(screen, title, r.Min.X+cb.theme.Padding, r.Min.Y+(rh-CharHeight)/2, cb.theme.Text)
}

// visibleCount returns the number of rows that fit in the current bounds.
func (cb *ConversationBrowser) visibleCount() int {
	inner := cb.bounds.Dy() - 2*cb.theme.Padding
	if inner <= 0 {
		return 0
	}
	return inner / cb.rowHeight()
}

// rowHeight returns the pixel height of one conversation row.
func (cb *ConversationBrowser) rowHeight() int {
	return CharHeight + cb.theme.Padding*2
}
