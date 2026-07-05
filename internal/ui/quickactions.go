package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// QuickAction is a named action with a callback.
type QuickAction struct {
	// Label is the button text.
	Label string
	// Run is the callback invoked when the action is activated.
	Run func(selectedText string)
}

// QuickActionBar renders a horizontal row of action buttons for common
// code-assistant operations (Explain, Generate Tests, Refactor, etc.).
type QuickActionBar struct {
	baseWidget
	actions      []QuickAction
	buttons      []*Button
	selectedText string
}

// NewQuickActionBar creates a bar with the given theme and actions.
// Pass nil actions to use the default set.
func NewQuickActionBar(actions []QuickAction, theme Theme) *QuickActionBar {
	if actions == nil {
		actions = defaultQuickActions()
	}
	qab := &QuickActionBar{
		baseWidget: baseWidget{theme: theme},
		actions:    actions,
	}
	qab.buildButtons(theme)
	return qab
}

// defaultQuickActions returns the built-in set of quick actions.
func defaultQuickActions() []QuickAction {
	return []QuickAction{
		{Label: "Explain"},
		{Label: "Gen Tests"},
		{Label: "Refactor"},
		{Label: "Find Bugs"},
		{Label: "Document"},
	}
}

// buildButtons creates one Button per action.
func (qab *QuickActionBar) buildButtons(t Theme) {
	qab.buttons = make([]*Button, len(qab.actions))
	for i, a := range qab.actions {
		idx := i
		act := a
		qab.buttons[idx] = NewButton(act.Label, t, func() {
			if act.Run != nil {
				act.Run(qab.selectedText)
			}
		})
	}
}

// Focusable reports that the action bar does not claim focus itself.
func (qab *QuickActionBar) Focusable() bool { return false }

// SetSelectedText sets the text that will be passed to action callbacks.
func (qab *QuickActionBar) SetSelectedText(text string) { qab.selectedText = text }

// SetAction replaces the callback for the action at index i.
func (qab *QuickActionBar) SetAction(i int, fn func(text string)) {
	if i >= 0 && i < len(qab.actions) {
		qab.actions[i].Run = fn
		qab.buildButtons(qab.theme)
		qab.layoutButtons()
	}
}

// SetBounds positions the bar and distributes buttons evenly.
func (qab *QuickActionBar) SetBounds(r image.Rectangle) {
	qab.bounds = r
	qab.layoutButtons()
}

// layoutButtons distributes button bounds across the bar width.
func (qab *QuickActionBar) layoutButtons() {
	if len(qab.buttons) == 0 {
		return
	}
	n := len(qab.buttons)
	totalMargin := qab.theme.Margin * (n - 1)
	bw := (qab.bounds.Dx() - totalMargin) / n
	x := qab.bounds.Min.X
	for _, btn := range qab.buttons {
		btn.SetBounds(image.Rect(x, qab.bounds.Min.Y, x+bw, qab.bounds.Max.Y))
		x += bw + qab.theme.Margin
	}
}

// Update propagates to all buttons.
func (qab *QuickActionBar) Update() error {
	for _, btn := range qab.buttons {
		if err := btn.Update(); err != nil {
			return err
		}
	}
	return nil
}

// Draw renders all buttons.
func (qab *QuickActionBar) Draw(screen *ebiten.Image) {
	fillRect(screen, qab.bounds, qab.theme.Background)
	for _, btn := range qab.buttons {
		btn.Draw(screen)
	}
}
