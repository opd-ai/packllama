package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// FocusManager maintains an ordered list of focusable widgets and routes
// keyboard focus between them. Tab advances focus; Shift+Tab reverses it.
// Enter and Space activate the focused widget when it implements Activatable.
type FocusManager struct {
	widgets []Widget
	current int // index of the currently focused widget, or -1 if none
}

// NewFocusManager creates a FocusManager with no registered widgets.
func NewFocusManager() *FocusManager {
	return &FocusManager{current: -1}
}

// Register adds w to the focus cycle if it is focusable.
// Widgets are focused in registration order.
func (f *FocusManager) Register(w Widget) {
	if !w.Focusable() {
		return
	}
	f.widgets = append(f.widgets, w)
}

// Current returns the widget that currently holds focus, or nil if none.
func (f *FocusManager) Current() Widget {
	if f.current < 0 || f.current >= len(f.widgets) {
		return nil
	}
	return f.widgets[f.current]
}

// SetFocus moves focus to the widget at index i. Out-of-range values clear focus.
func (f *FocusManager) SetFocus(i int) {
	f.clearCurrent()
	if i < 0 || i >= len(f.widgets) {
		f.current = -1
		return
	}
	f.current = i
	f.widgets[i].SetFocused(true)
}

// Update handles Tab, Shift+Tab, Enter, and Space to manage focus and activation.
func (f *FocusManager) Update() error {
	if len(f.widgets) == 0 {
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			f.moveFocus(-1)
		} else {
			f.moveFocus(1)
		}
	}
	if f.current >= 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			f.activateCurrent()
		}
	}
	return nil
}

// moveFocus shifts focus by delta (-1 or +1), wrapping around the list.
func (f *FocusManager) moveFocus(delta int) {
	n := len(f.widgets)
	if n == 0 {
		return
	}
	next := (f.current + delta + n) % n
	f.SetFocus(next)
}

// activateCurrent calls Activate on the focused widget if it implements Activatable.
func (f *FocusManager) activateCurrent() {
	w := f.Current()
	if w == nil {
		return
	}
	if a, ok := w.(Activatable); ok {
		a.Activate()
	}
}

// clearCurrent removes focus from the currently focused widget.
func (f *FocusManager) clearCurrent() {
	if f.current >= 0 && f.current < len(f.widgets) {
		f.widgets[f.current].SetFocused(false)
	}
}
