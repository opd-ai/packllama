package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// WindowConfig holds configuration for the Ebitengine window and game loop.
type WindowConfig struct {
	// Title is the window title bar text.
	Title string
	// Width is the logical screen width in pixels.
	Width int
	// Height is the logical screen height in pixels.
	Height int
}

// DefaultWindowConfig returns a 1280×720 window configuration.
func DefaultWindowConfig() WindowConfig {
	return WindowConfig{
		Title:  "packllama",
		Width:  1280,
		Height: 720,
	}
}

// Game implements ebiten.Game and manages the root layout container plus
// keyboard focus for all registered widgets.
type Game struct {
	cfg   WindowConfig
	theme Theme
	root  *Container
	focus *FocusManager
}

// NewGame creates a Game with the given window config and theme.
// Use g.Root() to add child widgets before calling Run.
func NewGame(cfg WindowConfig, theme Theme) *Game {
	root := NewContainer(Vertical, theme)
	root.SetBounds(image.Rect(0, 0, cfg.Width, cfg.Height))
	return &Game{
		cfg:   cfg,
		theme: theme,
		root:  root,
		focus: NewFocusManager(),
	}
}

// Root returns the top-level container for adding widgets.
func (g *Game) Root() *Container { return g.root }

// Focus returns the FocusManager for registering focusable widgets.
func (g *Game) Focus() *FocusManager { return g.focus }

// Update advances the focus manager and root container each game tick.
func (g *Game) Update() error {
	if err := g.focus.Update(); err != nil {
		return err
	}
	return g.root.Update()
}

// Draw clears the background and renders the root container.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(g.theme.Background)
	g.root.Draw(screen)
}

// Layout returns the logical screen dimensions, ignoring the OS window size.
func (g *Game) Layout(_, _ int) (int, int) {
	return g.cfg.Width, g.cfg.Height
}

// Run initializes the window and starts the Ebitengine game loop.
// It blocks until the window is closed or an error occurs.
func (g *Game) Run() error {
	ebiten.SetWindowSize(g.cfg.Width, g.cfg.Height)
	ebiten.SetWindowTitle(g.cfg.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(g)
}
