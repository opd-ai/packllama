package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

// uiFace is the default bitmap font used for all UI text rendering.
var uiFace = textv2.NewGoXFace(basicfont.Face7x13)

// CharWidth and CharHeight are the pixel dimensions of one uiFace character cell.
const (
	CharWidth  = 7
	CharHeight = 13
)

// fillRect draws a solid-color rectangle onto screen.
func fillRect(screen *ebiten.Image, r image.Rectangle, clr color.Color) {
	ebitenutil.DrawRect(screen,
		float64(r.Min.X), float64(r.Min.Y),
		float64(r.Dx()), float64(r.Dy()), clr)
}

// drawBorder draws a hollow rectangle outline of the given pixel width.
func drawBorder(screen *ebiten.Image, r image.Rectangle, width int, clr color.Color) {
	if width <= 0 {
		return
	}
	x, y := float64(r.Min.X), float64(r.Min.Y)
	fw, fh, w := float64(r.Dx()), float64(r.Dy()), float64(width)
	ebitenutil.DrawRect(screen, x, y, fw, w, clr)
	ebitenutil.DrawRect(screen, x, y+fh-w, fw, w, clr)
	ebitenutil.DrawRect(screen, x, y, w, fh, clr)
	ebitenutil.DrawRect(screen, x+fw-w, y, w, fh, clr)
}

// drawText renders str with its top-left corner at (x, y) using the default font.
func drawText(screen *ebiten.Image, str string, x, y int, clr color.Color) {
	ascent := uiFace.Metrics().HAscent
	op := &textv2.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y)+ascent)
	op.ColorScale.ScaleWithColor(clr)
	textv2.Draw(screen, str, uiFace, op)
}

// inBounds reports whether pixel position (mx, my) lies inside rectangle r.
func inBounds(r image.Rectangle, mx, my int) bool {
	return mx >= r.Min.X && mx < r.Max.X && my >= r.Min.Y && my < r.Max.Y
}
