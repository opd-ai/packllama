package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Direction controls how a Container arranges its children.
type Direction int

const (
	// Vertical stacks children top-to-bottom.
	Vertical Direction = iota
	// Horizontal places children left-to-right.
	Horizontal
)

// Container is a layout widget that arranges child widgets in a single axis.
// Children are sized equally within the available space, separated by Margin.
type Container struct {
	baseWidget
	children  []Widget
	direction Direction
}

// NewContainer creates a Container with the given direction and theme.
func NewContainer(dir Direction, theme Theme) *Container {
	return &Container{
		baseWidget: baseWidget{theme: theme},
		direction:  dir,
	}
}

// Add appends a child widget to the container and re-layouts if bounds are set.
func (c *Container) Add(w Widget) {
	c.children = append(c.children, w)
	if !c.bounds.Empty() {
		c.layout()
	}
}

// Focusable reports that containers are not themselves focusable.
func (c *Container) Focusable() bool { return false }

// SetBounds sets the container's bounds and triggers a re-layout.
func (c *Container) SetBounds(r image.Rectangle) {
	c.bounds = r
	c.layout()
}

// Update propagates the update call to all children.
func (c *Container) Update() error {
	for _, w := range c.children {
		if err := w.Update(); err != nil {
			return err
		}
	}
	return nil
}

// Draw fills the background and renders all children.
func (c *Container) Draw(screen *ebiten.Image) {
	fillRect(screen, c.bounds, c.theme.Background)
	for _, w := range c.children {
		w.Draw(screen)
	}
}

// layout distributes available space equally among children along the axis.
func (c *Container) layout() {
	if len(c.children) == 0 {
		return
	}
	if c.direction == Vertical {
		c.layoutVertical()
	} else {
		c.layoutHorizontal()
	}
}

// layoutVertical distributes the container height among children.
func (c *Container) layoutVertical() {
	n := len(c.children)
	totalMargin := c.theme.Margin * (n - 1)
	cellH := (c.bounds.Dy() - totalMargin) / n
	y := c.bounds.Min.Y
	for _, w := range c.children {
		r := image.Rect(c.bounds.Min.X, y, c.bounds.Max.X, y+cellH)
		w.SetBounds(r)
		y += cellH + c.theme.Margin
	}
}

// layoutHorizontal distributes the container width among children.
func (c *Container) layoutHorizontal() {
	n := len(c.children)
	totalMargin := c.theme.Margin * (n - 1)
	cellW := (c.bounds.Dx() - totalMargin) / n
	x := c.bounds.Min.X
	for _, w := range c.children {
		r := image.Rect(x, c.bounds.Min.Y, x+cellW, c.bounds.Max.Y)
		w.SetBounds(r)
		x += cellW + c.theme.Margin
	}
}
