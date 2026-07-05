package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Button is a clickable widget that invokes a callback on activation.
type Button struct {
	baseWidget
	label   string
	onClick func()
	hovered bool
	pressed bool
}

// NewButton creates a Button with the given label, theme, and click handler.
// onClick may be nil.
func NewButton(label string, theme Theme, onClick func()) *Button {
	return &Button{
		baseWidget: baseWidget{theme: theme},
		label:      label,
		onClick:    onClick,
	}
}

// Focusable reports that buttons accept keyboard focus.
func (b *Button) Focusable() bool { return true }

// Activate triggers the button's click handler, satisfying Activatable.
func (b *Button) Activate() {
	if b.onClick != nil {
		b.onClick()
	}
}

// Update processes mouse hover and click events.
func (b *Button) Update() error {
	mx, my := ebiten.CursorPosition()
	b.hovered = inBounds(b.bounds, mx, my)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && b.hovered {
		b.pressed = true
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && b.pressed {
		b.pressed = false
		if b.hovered && b.onClick != nil {
			b.onClick()
		}
	}
	return nil
}

// Draw renders the button background, border, and label.
func (b *Button) Draw(screen *ebiten.Image) {
	bg := b.theme.Surface
	switch {
	case b.pressed:
		bg = b.theme.Active
	case b.hovered:
		bg = b.theme.Hover
	}
	fillRect(screen, b.bounds, bg)
	borderClr := b.theme.Border
	if b.focused {
		borderClr = b.theme.Primary
	}
	drawBorder(screen, b.bounds, b.theme.BorderWidth, borderClr)
	b.drawLabel(screen)
}

// drawLabel centers the button label inside b.bounds.
func (b *Button) drawLabel(screen *ebiten.Image) {
	tw := len(b.label) * CharWidth
	x := b.bounds.Min.X + (b.bounds.Dx()-tw)/2
	y := b.bounds.Min.Y + (b.bounds.Dy()-CharHeight)/2
	inner := image.Rect(
		b.bounds.Min.X+b.theme.Padding, b.bounds.Min.Y,
		b.bounds.Max.X-b.theme.Padding, b.bounds.Max.Y,
	)
	if x < inner.Min.X {
		x = inner.Min.X
	}
	drawText(screen, b.label, x, y, b.theme.Text)
}
