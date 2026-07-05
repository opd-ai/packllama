package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Widget is the interface implemented by every UI component.
type Widget interface {
	// Update processes input events and advances the widget state.
	// It must be called once per game tick.
	Update() error
	// Draw renders the widget onto screen.
	Draw(screen *ebiten.Image)
	// Bounds returns the widget's axis-aligned bounding rectangle.
	Bounds() image.Rectangle
	// SetBounds repositions and resizes the widget.
	SetBounds(r image.Rectangle)
	// Focusable reports whether the widget can receive keyboard focus.
	Focusable() bool
	// SetFocused grants or removes keyboard focus from the widget.
	SetFocused(focused bool)
}

// Activatable is an optional interface for widgets that respond to
// keyboard activation (Enter or Space when focused).
type Activatable interface {
	// Activate triggers the widget's primary action.
	Activate()
}

// baseWidget provides shared bounds, focus state, and theme for all widgets.
// It is intended to be embedded, not used directly.
type baseWidget struct {
	bounds        image.Rectangle
	focused       bool
	theme         Theme
	focusCallback func() // set by FocusManager.Register; nil when unmanaged
}

// Bounds returns the widget's bounding rectangle.
func (b *baseWidget) Bounds() image.Rectangle { return b.bounds }

// SetBounds repositions and resizes the widget.
func (b *baseWidget) SetBounds(r image.Rectangle) { b.bounds = r }

// SetFocused grants or removes keyboard focus.
func (b *baseWidget) SetFocused(focused bool) { b.focused = focused }

// setFocusCallback is called by FocusManager.Register so that the widget can
// route click-initiated focus requests through the manager.
func (b *baseWidget) setFocusCallback(fn func()) { b.focusCallback = fn }

// grabFocus requests focus: routes through FocusManager when registered,
// otherwise sets the focused flag directly.
func (b *baseWidget) grabFocus() {
	if b.focusCallback != nil {
		b.focusCallback()
	} else {
		b.focused = true
	}
}
