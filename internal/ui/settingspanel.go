package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	appName    = "packllama"
	appVersion = "dev"
	appLicense = "MIT"
)

// SettingsPanel is a composite widget for viewing and editing UserPreferences.
// It exposes theme selection, font scale, default model, auto-save toggle,
// a clear-history action, and an about section.
type SettingsPanel struct {
	baseWidget
	prefs       UserPreferences
	prefsPath   string
	onChange    func(p UserPreferences)
	onClearHist func()

	// sub-widgets
	themeDrop    *Dropdown
	fontSlider   *Slider
	modelInput   *TextInput
	autoSaveBox  *Checkbox
	clearBtn     *Button
}

// NewSettingsPanel creates a panel pre-loaded with prefs and the given theme.
func NewSettingsPanel(prefs UserPreferences, theme Theme) *SettingsPanel {
	sp := &SettingsPanel{
		baseWidget: baseWidget{theme: theme},
		prefs:      prefs,
	}
	sp.buildWidgets(theme)
	return sp
}

// buildWidgets constructs and wires all sub-widgets.
func (sp *SettingsPanel) buildWidgets(t Theme) {
	sp.themeDrop = NewDropdown([]string{"dark", "light"}, t)
	if sp.prefs.ThemeName == "light" {
		sp.themeDrop.SetOptions([]string{"dark", "light"})
	}
	sp.fontSlider = NewSlider(0.5, 3.0, sp.prefs.FontScale, t)
	sp.modelInput = NewTextInput("default model", t)
	sp.modelInput.SetValue(sp.prefs.DefaultModel)
	sp.autoSaveBox = NewCheckbox("Auto-save conversations", t)
	sp.autoSaveBox.SetChecked(sp.prefs.AutoSave)
	sp.clearBtn = NewButton("Clear History", t, func() {
		if sp.onClearHist != nil {
			sp.onClearHist()
		}
	})
	sp.wireCallbacks()
}

// wireCallbacks connects sub-widget changes to preference updates.
func (sp *SettingsPanel) wireCallbacks() {
	sp.themeDrop.OnChange(func(_ int, v string) {
		sp.prefs.ThemeName = v
		sp.emit()
	})
	sp.fontSlider.OnChange(func(v float64) {
		sp.prefs.FontScale = v
		sp.emit()
	})
	sp.autoSaveBox.OnChange(func(checked bool) {
		sp.prefs.AutoSave = checked
		sp.emit()
	})
}

// Focusable reports that the settings panel itself is not directly focusable.
func (sp *SettingsPanel) Focusable() bool { return false }

// Prefs returns the current preferences.
func (sp *SettingsPanel) Prefs() UserPreferences { return sp.prefs }

// SetPrefsPath configures the path used by Save.
func (sp *SettingsPanel) SetPrefsPath(path string) { sp.prefsPath = path }

// Save persists the current preferences using the configured path.
func (sp *SettingsPanel) Save() error { return sp.prefs.Save(sp.prefsPath) }

// OnChange registers a callback called when any preference changes.
func (sp *SettingsPanel) OnChange(fn func(p UserPreferences)) { sp.onChange = fn }

// OnClearHistory registers a callback for the clear-history button.
func (sp *SettingsPanel) OnClearHistory(fn func()) { sp.onClearHist = fn }

// Update propagates to all interactive sub-widgets and syncs text fields.
func (sp *SettingsPanel) Update() error {
	sp.themeDrop.Update()   //nolint:errcheck
	sp.fontSlider.Update()  //nolint:errcheck
	sp.modelInput.Update()  //nolint:errcheck
	sp.autoSaveBox.Update() //nolint:errcheck
	sp.clearBtn.Update()    //nolint:errcheck
	if m := sp.modelInput.Value(); m != sp.prefs.DefaultModel {
		sp.prefs.DefaultModel = m
		sp.emit()
	}
	return nil
}

// Draw renders the settings panel with all rows and an about section.
func (sp *SettingsPanel) Draw(screen *ebiten.Image) {
	fillRect(screen, sp.bounds, sp.theme.Background)
	sp.layoutWidgets()
	sp.drawRow(screen, "Theme", sp.themeDrop)
	sp.drawRow(screen, "Font scale", sp.fontSlider)
	sp.drawRow(screen, "Default model", sp.modelInput)
	sp.autoSaveBox.Draw(screen)
	sp.clearBtn.Draw(screen)
	sp.drawAbout(screen)
}

// drawRow renders a label+widget pair on one row.
func (sp *SettingsPanel) drawRow(screen *ebiten.Image, label string, w Widget) {
	b := w.Bounds()
	lx := sp.bounds.Min.X + sp.theme.Padding
	ly := b.Min.Y + (b.Dy()-CharHeight)/2
	drawText(screen, label+":", lx, ly, sp.theme.TextMuted)
	w.Draw(screen)
}

// drawAbout renders the about/info section at the bottom of the panel.
func (sp *SettingsPanel) drawAbout(screen *ebiten.Image) {
	y := sp.bounds.Max.Y - 3*(CharHeight+sp.theme.Padding)
	x := sp.bounds.Min.X + sp.theme.Padding
	drawText(screen, fmt.Sprintf("%s  %s", appName, appVersion), x, y, sp.theme.Primary)
	drawText(screen, "License: "+appLicense, x, y+CharHeight+sp.theme.Padding, sp.theme.TextMuted)
	drawText(screen, "https://github.com/opd-ai/packllama", x, y+2*(CharHeight+sp.theme.Padding), sp.theme.TextMuted)
}

// layoutWidgets distributes sub-widget bounds inside the panel.
func (sp *SettingsPanel) layoutWidgets() {
	lx := sp.bounds.Min.X + sp.theme.Padding + 14*CharWidth + sp.theme.Padding
	rw := sp.bounds.Max.X - lx - sp.theme.Padding
	rh := sp.rowH()
	sp.setRowBounds(sp.themeDrop, lx, 0, rw, rh)
	sp.setRowBounds(sp.fontSlider, lx, 1, rw, rh)
	sp.setRowBounds(sp.modelInput, lx, 2, rw, rh)
	sp.autoSaveBox.SetBounds(image.Rect(
		sp.bounds.Min.X+sp.theme.Padding, sp.bounds.Min.Y+sp.theme.Padding+3*rh,
		sp.bounds.Max.X-sp.theme.Padding, sp.bounds.Min.Y+sp.theme.Padding+4*rh-sp.theme.Margin))
	sp.clearBtn.SetBounds(image.Rect(
		sp.bounds.Min.X+sp.theme.Padding, sp.bounds.Min.Y+sp.theme.Padding+4*rh+sp.theme.Margin,
		sp.bounds.Min.X+sp.theme.Padding+14*CharWidth, sp.bounds.Min.Y+sp.theme.Padding+5*rh))
}

// setRowBounds positions widget w at the given row index.
func (sp *SettingsPanel) setRowBounds(w Widget, lx, row, rw, rh int) {
	y := sp.bounds.Min.Y + sp.theme.Padding + row*rh
	w.SetBounds(image.Rect(lx, y, lx+rw, y+rh-sp.theme.Margin))
}

// rowH returns the pixel height of one settings row.
func (sp *SettingsPanel) rowH() int { return CharHeight + sp.theme.Padding*2 }

// emit calls the onChange callback if set.
func (sp *SettingsPanel) emit() {
	if sp.onChange != nil {
		sp.onChange(sp.prefs)
	}
}
