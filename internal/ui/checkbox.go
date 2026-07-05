package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Checkbox is a toggleable boolean widget with an optional text label.
type Checkbox struct {
	baseWidget
	checked  bool
	label    string
	onChange func(checked bool)
}

// NewCheckbox creates an unchecked Checkbox with the given label and theme.
func NewCheckbox(label string, theme Theme) *Checkbox {
	return &Checkbox{
		baseWidget: baseWidget{theme: theme},
		label:      label,
	}
}

// Focusable reports that checkboxes accept keyboard focus.
func (c *Checkbox) Focusable() bool { return true }

// Checked returns the current checked state.
func (c *Checkbox) Checked() bool { return c.checked }

// SetChecked sets the checked state without triggering the onChange callback.
func (c *Checkbox) SetChecked(checked bool) { c.checked = checked }

// OnChange registers a callback invoked whenever the checked state changes.
func (c *Checkbox) OnChange(fn func(checked bool)) { c.onChange = fn }

// Activate toggles the checkbox, satisfying Activatable.
func (c *Checkbox) Activate() {
	c.checked = !c.checked
	if c.onChange != nil {
		c.onChange(c.checked)
	}
}

// Update handles mouse clicks and keyboard activation on the checkbox.
func (c *Checkbox) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		if inBounds(c.bounds, mx, my) {
			c.Activate()
		}
	}
	return nil
}

// Draw renders the checkbox box, optional checkmark, and label.
func (c *Checkbox) Draw(screen *ebiten.Image) {
	box := c.boxBounds()
	fillRect(screen, box, c.theme.Surface)
	borderClr := c.theme.Border
	if c.focused {
		borderClr = c.theme.Primary
	}
	drawBorder(screen, box, c.theme.BorderWidth, borderClr)
	if c.checked {
		inner := image.Rect(box.Min.X+3, box.Min.Y+3, box.Max.X-3, box.Max.Y-3)
		fillRect(screen, inner, c.theme.Primary)
	}
	if c.label != "" {
		lx := box.Max.X + c.theme.Padding
		ly := c.bounds.Min.Y + (c.bounds.Dy()-CharHeight)/2
		drawText(screen, c.label, lx, ly, c.theme.Text)
	}
}

// boxBounds returns the square checkbox area at the left of the bounds.
func (c *Checkbox) boxBounds() image.Rectangle {
	size := c.bounds.Dy() - 2*c.theme.Padding
	if size < 8 {
		size = 8
	}
	x := c.bounds.Min.X + c.theme.Padding
	y := c.bounds.Min.Y + (c.bounds.Dy()-size)/2
	return image.Rect(x, y, x+size, y+size)
}
