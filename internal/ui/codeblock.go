package ui

import (
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// coloredSpan is a fragment of text that should be drawn in a specific color.
type coloredSpan struct {
	text string
	clr  color.Color
}

// drawSpans renders a slice of colored spans left-to-right starting at (x, y).
func drawSpans(screen *ebiten.Image, spans []coloredSpan, x, y int) {
	for _, s := range spans {
		drawText(screen, s.text, x, y, s.clr)
		x += len(s.text) * CharWidth
	}
}

// CodeBlock is a widget that displays a snippet of code with:
//   - language label in the header
//   - syntax-highlighted text using coloredSpans
//   - a "Copy" button that calls onCopy with the raw code text
type CodeBlock struct {
	baseWidget
	code    string
	lang    string
	onCopy  func(code string)
	copyBtn *Button
}

// NewCodeBlock creates a CodeBlock for the given language and code text.
func NewCodeBlock(lang, code string, theme Theme) *CodeBlock {
	cb := &CodeBlock{
		baseWidget: baseWidget{theme: theme},
		code:       code,
		lang:       lang,
	}
	cb.copyBtn = NewButton("Copy", theme, func() {
		if cb.onCopy != nil {
			cb.onCopy(cb.code)
		}
	})
	return cb
}

// Focusable reports that code blocks accept keyboard focus.
func (cb *CodeBlock) Focusable() bool { return true }

// OnCopy registers a callback invoked when the Copy button is activated.
func (cb *CodeBlock) OnCopy(fn func(code string)) { cb.onCopy = fn }

// SetBounds sets the widget bounds and positions the copy button.
func (cb *CodeBlock) SetBounds(r image.Rectangle) {
	cb.bounds = r
	bw := 6*CharWidth + cb.theme.Padding*2
	bh := CharHeight + cb.theme.Padding
	cb.copyBtn.SetBounds(image.Rect(
		r.Max.X-bw-cb.theme.Padding, r.Min.Y+cb.theme.Padding/2,
		r.Max.X-cb.theme.Padding, r.Min.Y+cb.theme.Padding/2+bh,
	))
}

// Update propagates to the copy button.
func (cb *CodeBlock) Update() error {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		cb.focused = inBounds(cb.bounds, mx, my)
	}
	return cb.copyBtn.Update()
}

// Draw renders the header bar, highlighted code, and copy button.
func (cb *CodeBlock) Draw(screen *ebiten.Image) {
	fillRect(screen, cb.bounds, cb.theme.Background)
	drawBorder(screen, cb.bounds, cb.theme.BorderWidth, cb.theme.Border)
	cb.drawHeader(screen)
	cb.drawCode(screen)
	cb.copyBtn.Draw(screen)
}

// drawHeader renders the language label bar at the top of the block.
func (cb *CodeBlock) drawHeader(screen *ebiten.Image) {
	hdr := image.Rect(cb.bounds.Min.X, cb.bounds.Min.Y,
		cb.bounds.Max.X, cb.bounds.Min.Y+CharHeight+cb.theme.Padding)
	fillRect(screen, hdr, cb.theme.Surface)
	lang := cb.lang
	if lang == "" {
		lang = "code"
	}
	drawText(screen, lang, hdr.Min.X+cb.theme.Padding, hdr.Min.Y+(hdr.Dy()-CharHeight)/2, cb.theme.TextMuted)
}

// drawCode renders syntax-highlighted lines below the header.
func (cb *CodeBlock) drawCode(screen *ebiten.Image) {
	x := cb.bounds.Min.X + cb.theme.Padding
	y := cb.bounds.Min.Y + CharHeight + cb.theme.Padding*2
	for _, line := range splitLines(cb.code) {
		spans := HighlightLine(line, cb.lang)
		drawSpans(screen, spans, x, y)
		y += CharHeight + 2
		if y > cb.bounds.Max.Y-cb.theme.Padding {
			break
		}
	}
}

// splitLines splits s on newlines into individual lines.
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
