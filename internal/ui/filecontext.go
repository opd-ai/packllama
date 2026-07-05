package ui

import (
	"image"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// FileEntry holds information about one file or directory in the context browser.
type FileEntry struct {
	Name  string
	Path  string
	IsDir bool
}

// FileContextBrowser is a panel for browsing and selecting files to attach
// as context to an LLM request.
type FileContextBrowser struct {
	baseWidget
	root         string
	entries      []FileEntry
	selected     []int
	scrollOffset int
	onSelect     func(paths []string)
}

// NewFileContextBrowser creates a browser rooted at dir with the given theme.
func NewFileContextBrowser(dir string, theme Theme) *FileContextBrowser {
	fb := &FileContextBrowser{
		baseWidget: baseWidget{theme: theme},
		root:       dir,
	}
	fb.Refresh()
	return fb
}

// Focusable reports that the file browser accepts keyboard focus.
func (fb *FileContextBrowser) Focusable() bool { return true }

// OnSelect registers a callback called with the currently selected file paths.
func (fb *FileContextBrowser) OnSelect(fn func(paths []string)) { fb.onSelect = fn }

// Refresh re-reads the root directory listing.
func (fb *FileContextBrowser) Refresh() {
	entries, err := os.ReadDir(fb.root)
	if err != nil {
		return
	}
	fb.entries = make([]FileEntry, 0, len(entries))
	for _, e := range entries {
		fb.entries = append(fb.entries, FileEntry{
			Name:  e.Name(),
			Path:  filepath.Join(fb.root, e.Name()),
			IsDir: e.IsDir(),
		})
	}
}

// SelectedPaths returns the paths of all currently selected entries.
func (fb *FileContextBrowser) SelectedPaths() []string {
	paths := make([]string, 0, len(fb.selected))
	for _, idx := range fb.selected {
		if idx < len(fb.entries) {
			paths = append(paths, fb.entries[idx].Path)
		}
	}
	return paths
}

// Update handles mouse clicks and scroll wheel.
func (fb *FileContextBrowser) Update() error {
	mx, my := ebiten.CursorPosition()
	if inBounds(fb.bounds, mx, my) {
		fb.handleScroll()
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && inBounds(fb.bounds, mx, my) {
		fb.handleClick(my)
	}
	return nil
}

// handleScroll processes mouse wheel input.
func (fb *FileContextBrowser) handleScroll() {
	_, dy := ebiten.Wheel()
	if dy > 0 && fb.scrollOffset > 0 {
		fb.scrollOffset--
	} else if dy < 0 && fb.scrollOffset < len(fb.entries)-fb.visibleCount() {
		fb.scrollOffset++
	}
}

// handleClick toggles selection for the row under the cursor.
func (fb *FileContextBrowser) handleClick(my int) {
	rh := fb.rowHeight()
	row := (my - fb.bounds.Min.Y - fb.theme.Padding) / rh
	idx := row + fb.scrollOffset
	if idx < 0 || idx >= len(fb.entries) {
		return
	}
	fb.toggleSelected(idx)
	if fb.onSelect != nil {
		fb.onSelect(fb.SelectedPaths())
	}
}

// toggleSelected adds idx to the selection or removes it if already selected.
func (fb *FileContextBrowser) toggleSelected(idx int) {
	for i, s := range fb.selected {
		if s == idx {
			fb.selected = append(fb.selected[:i], fb.selected[i+1:]...)
			return
		}
	}
	fb.selected = append(fb.selected, idx)
}

// Draw renders the file list with selection highlights.
func (fb *FileContextBrowser) Draw(screen *ebiten.Image) {
	fillRect(screen, fb.bounds, fb.theme.Surface)
	drawBorder(screen, fb.bounds, fb.theme.BorderWidth, fb.theme.Border)
	visible := fb.visibleCount()
	for i := 0; i < visible; i++ {
		idx := i + fb.scrollOffset
		if idx >= len(fb.entries) {
			break
		}
		fb.drawEntry(screen, fb.entries[idx], idx, i)
	}
}

// drawEntry renders one file entry row.
func (fb *FileContextBrowser) drawEntry(screen *ebiten.Image, e FileEntry, idx, row int) {
	rh := fb.rowHeight()
	y := fb.bounds.Min.Y + fb.theme.Padding + row*rh
	r := image.Rect(fb.bounds.Min.X+fb.theme.BorderWidth, y, fb.bounds.Max.X-fb.theme.BorderWidth, y+rh)
	bg := fb.theme.Surface
	if fb.isSelected(idx) {
		bg = fb.theme.Active
	}
	fillRect(screen, r, bg)
	prefix := "  "
	if e.IsDir {
		prefix = "▸ "
	}
	clr := fb.theme.Text
	if e.IsDir {
		clr = fb.theme.Primary
	}
	drawText(screen, prefix+e.Name, r.Min.X+fb.theme.Padding, r.Min.Y+(rh-CharHeight)/2, clr)
}

// isSelected reports whether the entry at idx is selected.
func (fb *FileContextBrowser) isSelected(idx int) bool {
	for _, s := range fb.selected {
		if s == idx {
			return true
		}
	}
	return false
}

// visibleCount returns how many rows fit inside the bounds.
func (fb *FileContextBrowser) visibleCount() int {
	inner := fb.bounds.Dy() - 2*fb.theme.Padding
	if inner <= 0 {
		return 0
	}
	return inner / fb.rowHeight()
}

// rowHeight returns the pixel height of one file entry row.
func (fb *FileContextBrowser) rowHeight() int {
	return CharHeight + fb.theme.Padding*2
}
