package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Dropdown is a select widget that shows one option at a time and expands
// to display all choices on click.
// Note: expanded state draws on top of adjacent widgets; position accordingly.
type Dropdown struct {
	baseWidget
	options      []string
	selected     int
	expanded     bool
	onChangeFunc func(index int, value string)
}

// NewDropdown creates a Dropdown pre-populated with options.
// The first option is selected by default.
func NewDropdown(options []string, theme Theme) *Dropdown {
	return &Dropdown{
		baseWidget: baseWidget{theme: theme},
		options:    options,
	}
}

// Focusable reports that dropdowns accept keyboard focus.
func (d *Dropdown) Focusable() bool { return true }

// Value returns the currently selected option string.
func (d *Dropdown) Value() string {
	if len(d.options) == 0 {
		return ""
	}
	return d.options[d.selected]
}

// SelectedIndex returns the index of the currently selected option.
func (d *Dropdown) SelectedIndex() int { return d.selected }

// SetOptions replaces the option list and resets the selection to index 0.
func (d *Dropdown) SetOptions(opts []string) {
	d.options = opts
	d.selected = 0
	d.expanded = false
}

// OnChange registers a callback invoked whenever the selection changes.
func (d *Dropdown) OnChange(fn func(index int, value string)) {
	d.onChangeFunc = fn
}

// Update processes mouse clicks to toggle expansion and select options.
func (d *Dropdown) Update() error {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return nil
	}
	mx, my := ebiten.CursorPosition()
	if !d.expanded {
		if inBounds(d.bounds, mx, my) {
			d.expanded = true
		}
		return nil
	}
	d.handleExpandedClick(mx, my)
	return nil
}

// handleExpandedClick selects an option or collapses the dropdown.
func (d *Dropdown) handleExpandedClick(mx, my int) {
	for i := range d.options {
		r := d.optionBounds(i)
		if inBounds(r, mx, my) {
			d.selected = i
			d.expanded = false
			if d.onChangeFunc != nil {
				d.onChangeFunc(i, d.options[i])
			}
			return
		}
	}
	d.expanded = false
}

// Draw renders the collapsed header and, if expanded, the options list.
func (d *Dropdown) Draw(screen *ebiten.Image) {
	d.drawHeader(screen)
	if d.expanded {
		d.drawOptions(screen)
	}
}

// drawHeader renders the selected-value row with a down-arrow indicator.
func (d *Dropdown) drawHeader(screen *ebiten.Image) {
	fillRect(screen, d.bounds, d.theme.Surface)
	borderClr := d.theme.Border
	if d.focused {
		borderClr = d.theme.Primary
	}
	drawBorder(screen, d.bounds, d.theme.BorderWidth, borderClr)
	x := d.bounds.Min.X + d.theme.Padding
	y := d.bounds.Min.Y + (d.bounds.Dy()-CharHeight)/2
	label := fmt.Sprintf("%s ▾", d.Value())
	drawText(screen, label, x, y, d.theme.Text)
}

// drawOptions renders the expanded option list below the header.
func (d *Dropdown) drawOptions(screen *ebiten.Image) {
	for i, opt := range d.options {
		r := d.optionBounds(i)
		bg := d.theme.Surface
		if i == d.selected {
			bg = d.theme.Active
		}
		fillRect(screen, r, bg)
		drawBorder(screen, r, d.theme.BorderWidth, d.theme.Border)
		x := r.Min.X + d.theme.Padding
		y := r.Min.Y + (r.Dy()-CharHeight)/2
		drawText(screen, opt, x, y, d.theme.Text)
	}
}

// optionBounds returns the rectangle for the i-th option row.
func (d *Dropdown) optionBounds(i int) image.Rectangle {
	h := d.bounds.Dy()
	return image.Rect(
		d.bounds.Min.X,
		d.bounds.Max.Y+i*h,
		d.bounds.Max.X,
		d.bounds.Max.Y+(i+1)*h,
	)
}
