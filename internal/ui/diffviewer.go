package ui

import (
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// DiffKind classifies a line in a diff result.
type DiffKind int

const (
	// DiffEqual means the line is unchanged.
	DiffEqual DiffKind = iota
	// DiffAdded means the line was inserted in the new version.
	DiffAdded
	// DiffRemoved means the line was deleted from the old version.
	DiffRemoved
)

// DiffLine is one line in a computed diff.
type DiffLine struct {
	Text string
	Kind DiffKind
}

// ComputeDiff returns a list of DiffLines comparing oldText to newText.
// It uses an LCS-based algorithm to produce an edit sequence.
func ComputeDiff(oldText, newText string) []DiffLine {
	a := strings.Split(oldText, "\n")
	b := strings.Split(newText, "\n")
	lcs := computeLCS(a, b)
	return buildDiff(a, b, lcs)
}

// computeLCS returns the longest-common-subsequence of lines a and b.
func computeLCS(a, b []string) []string {
	n, m := len(a), len(b)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] > dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var result []string
	i, j := 0, 0
	for i < n && j < m {
		if a[i] == b[j] {
			result = append(result, a[i])
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			i++
		} else {
			j++
		}
	}
	return result
}

// buildDiff constructs the DiffLine slice from old, new, and LCS slices.
func buildDiff(a, b, lcs []string) []DiffLine {
	var out []DiffLine
	ai, bi, li := 0, 0, 0
	for li < len(lcs) {
		for ai < len(a) && a[ai] != lcs[li] {
			out = append(out, DiffLine{Text: a[ai], Kind: DiffRemoved})
			ai++
		}
		for bi < len(b) && b[bi] != lcs[li] {
			out = append(out, DiffLine{Text: b[bi], Kind: DiffAdded})
			bi++
		}
		out = append(out, DiffLine{Text: lcs[li], Kind: DiffEqual})
		ai++
		bi++
		li++
	}
	for ; ai < len(a); ai++ {
		out = append(out, DiffLine{Text: a[ai], Kind: DiffRemoved})
	}
	for ; bi < len(b); bi++ {
		out = append(out, DiffLine{Text: b[bi], Kind: DiffAdded})
	}
	return out
}

// DiffViewer displays a computed diff with added/removed line highlighting.
type DiffViewer struct {
	baseWidget
	lines        []DiffLine
	scrollOffset int
}

// NewDiffViewer creates a DiffViewer with the given theme.
func NewDiffViewer(theme Theme) *DiffViewer {
	return &DiffViewer{baseWidget: baseWidget{theme: theme}}
}

// Focusable reports that the diff viewer accepts keyboard focus.
func (dv *DiffViewer) Focusable() bool { return true }

// SetDiff replaces the displayed diff lines.
func (dv *DiffViewer) SetDiff(lines []DiffLine) {
	dv.lines = lines
	dv.scrollOffset = 0
}

// SetTexts computes and displays a diff between oldText and newText.
func (dv *DiffViewer) SetTexts(oldText, newText string) {
	dv.SetDiff(ComputeDiff(oldText, newText))
}

// Update handles scroll-wheel input.
func (dv *DiffViewer) Update() error {
	mx, my := ebiten.CursorPosition()
	if !inBounds(dv.bounds, mx, my) {
		return nil
	}
	_, dy := ebiten.Wheel()
	if dy > 0 && dv.scrollOffset > 0 {
		dv.scrollOffset--
	} else if dy < 0 {
		dv.scrollOffset++
		dv.clampScroll()
	}
	return nil
}

// Draw renders the diff with color-coded lines.
func (dv *DiffViewer) Draw(screen *ebiten.Image) {
	fillRect(screen, dv.bounds, dv.theme.Surface)
	drawBorder(screen, dv.bounds, dv.theme.BorderWidth, dv.theme.Border)
	visible := dv.visibleCount()
	x := dv.bounds.Min.X + dv.theme.Padding
	y := dv.bounds.Min.Y + dv.theme.Padding
	for i := dv.scrollOffset; i < dv.scrollOffset+visible && i < len(dv.lines); i++ {
		dl := dv.lines[i]
		bg, prefix, clr := dv.lineStyle(dl.Kind)
		if bg != nil {
			row := image.Rect(dv.bounds.Min.X, y, dv.bounds.Max.X, y+CharHeight+2)
			fillRect(screen, row, bg)
		}
		drawText(screen, prefix+dl.Text, x, y, clr)
		y += CharHeight + 2
	}
}

// lineStyle returns the background color, prefix character, and text color
// for a line based on its diff kind.
func (dv *DiffViewer) lineStyle(kind DiffKind) (bg color.Color, prefix string, clr color.Color) {
	switch kind {
	case DiffAdded:
		return color.RGBA{R: 0x1e, G: 0x3a, B: 0x1e, A: 0xff}, "+ ", color.RGBA{R: 0xa6, G: 0xe3, B: 0xa1, A: 0xff}
	case DiffRemoved:
		return color.RGBA{R: 0x3a, G: 0x1e, B: 0x1e, A: 0xff}, "- ", dv.theme.Error
	default:
		return nil, "  ", dv.theme.Text
	}
}

// visibleCount returns the number of lines that fit in the bounds.
func (dv *DiffViewer) visibleCount() int {
	inner := dv.bounds.Dy() - 2*dv.theme.Padding
	if inner <= 0 {
		return 0
	}
	return inner / (CharHeight + 2)
}

// clampScroll keeps scrollOffset within valid range.
func (dv *DiffViewer) clampScroll() {
	visible := dv.visibleCount()
	max := len(dv.lines) - visible
	if max < 0 {
		max = 0
	}
	if dv.scrollOffset > max {
		dv.scrollOffset = max
	}
}
