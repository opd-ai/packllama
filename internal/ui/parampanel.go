package ui

import (
	"fmt"
	"image"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// paramPresetNames lists the preset names used by ParamPanel in display order.
var paramPresetNames = [3]string{"creative", "balanced", "precise"}

// ParamPanel is a composite widget for viewing and editing InferenceParams.
// It provides model selection, basic sliders, preset buttons, an advanced-
// parameters toggle, real-time validation, and save/load support.
type ParamPanel struct {
	baseWidget
	params       InferenceParams
	models       []string
	showAdvanced bool
	errorText    string
	paramsPath   string
	onChange     func(p InferenceParams)

	// sub-widgets
	modelDrop    *Dropdown
	tempSlider   *Slider
	topPSlider   *Slider
	advToggle    *Checkbox
	maxTokInput  *TextInput
	ctxInput     *TextInput
	repeatSlider *Slider
	topKInput    *TextInput
	presetBtns   [3]*Button // creative, balanced, precise
}

// NewParamPanel creates a ParamPanel initialised with params and optional model list.
func NewParamPanel(params InferenceParams, models []string, theme Theme) *ParamPanel {
	pp := &ParamPanel{
		baseWidget: baseWidget{theme: theme},
		params:     params,
		models:     models,
	}
	pp.buildWidgets(theme)
	return pp
}

// buildWidgets initialises all child sub-widgets.
func (pp *ParamPanel) buildWidgets(t Theme) {
	pp.modelDrop = NewDropdown(pp.models, t)
	pp.tempSlider = NewSlider(0, 2, pp.params.Temperature, t)
	pp.topPSlider = NewSlider(0, 1, pp.params.TopP, t)
	pp.advToggle = NewCheckbox("Advanced", t)
	pp.maxTokInput = NewTextInput("max tokens", t)
	pp.ctxInput = NewTextInput("context length", t)
	pp.repeatSlider = NewSlider(0.5, 2, pp.params.RepeatPen, t)
	pp.topKInput = NewTextInput("top_k", t)
	pp.maxTokInput.SetValue(strconv.Itoa(pp.params.MaxTokens))
	pp.ctxInput.SetValue(strconv.Itoa(pp.params.ContextLen))
	pp.topKInput.SetValue(strconv.Itoa(pp.params.TopK))
	for i, name := range paramPresetNames {
		pp.presetBtns[i] = NewButton(name, t, func() { pp.ApplyPreset(name) })
	}
	pp.wireCallbacks()
}

// wireCallbacks connects sub-widget callbacks to update pp.params.
func (pp *ParamPanel) wireCallbacks() {
	pp.tempSlider.OnChange(func(v float64) { pp.setAndValidate(func(p *InferenceParams) { p.Temperature = v }) })
	pp.topPSlider.OnChange(func(v float64) { pp.setAndValidate(func(p *InferenceParams) { p.TopP = v }) })
	pp.repeatSlider.OnChange(func(v float64) { pp.setAndValidate(func(p *InferenceParams) { p.RepeatPen = v }) })
	pp.modelDrop.OnChange(func(_ int, v string) { pp.setAndValidate(func(p *InferenceParams) { p.Model = v }) })
	pp.advToggle.OnChange(func(checked bool) { pp.showAdvanced = checked })
}

// Focusable reports that the param panel itself is not focusable (children are).
func (pp *ParamPanel) Focusable() bool { return false }

// SetModels replaces the model list in the dropdown.
func (pp *ParamPanel) SetModels(models []string) {
	pp.models = models
	pp.modelDrop.SetOptions(models)
}

// Params returns the current InferenceParams (validated).
func (pp *ParamPanel) Params() InferenceParams { return pp.params }

// ApplyPreset applies a named preset and refreshes sub-widget states.
func (pp *ParamPanel) ApplyPreset(name string) {
	pp.params = ParamPreset(name)
	pp.params.Model = pp.modelDrop.Value()
	pp.syncWidgets()
	pp.validate()
}

// SetParamsPath configures the path used by Save/Load.
func (pp *ParamPanel) SetParamsPath(path string) { pp.paramsPath = path }

// Save persists the current parameters using the configured path.
func (pp *ParamPanel) Save() error { return pp.params.Save(pp.paramsPath) }

// Load reads parameters from the configured path and refreshes widgets.
func (pp *ParamPanel) Load() error {
	p, err := LoadParams(pp.paramsPath)
	if err != nil {
		return err
	}
	pp.params = p
	pp.syncWidgets()
	pp.validate()
	return nil
}

// OnChange registers a callback called after every validated parameter change.
func (pp *ParamPanel) OnChange(fn func(p InferenceParams)) { pp.onChange = fn }

// Update propagates to all sub-widgets and reads text fields.
func (pp *ParamPanel) Update() error {
	pp.modelDrop.Update()  //nolint:errcheck
	pp.tempSlider.Update() //nolint:errcheck
	pp.topPSlider.Update() //nolint:errcheck
	pp.advToggle.Update()  //nolint:errcheck
	for _, btn := range pp.presetBtns {
		btn.Update() //nolint:errcheck
	}
	if pp.showAdvanced {
		pp.maxTokInput.Update()  //nolint:errcheck
		pp.ctxInput.Update()     //nolint:errcheck
		pp.repeatSlider.Update() //nolint:errcheck
		pp.topKInput.Update()    //nolint:errcheck
	}
	pp.syncFromTextInputs()
	return nil
}

// Draw renders the panel background, labels, sub-widgets, and error text.
func (pp *ParamPanel) Draw(screen *ebiten.Image) {
	fillRect(screen, pp.bounds, pp.theme.Background)
	pp.drawSection(screen)
	if pp.errorText != "" {
		drawText(screen, pp.errorText,
			pp.bounds.Min.X+pp.theme.Padding,
			pp.bounds.Max.Y-CharHeight-pp.theme.Padding,
			pp.theme.Error)
	}
}

// drawSection renders visible rows of controls.
func (pp *ParamPanel) drawSection(screen *ebiten.Image) {
	pp.layoutWidgets()
	pp.drawRow(screen, "model", pp.modelDrop)
	pp.drawRow(screen, "temp", pp.tempSlider)
	pp.drawRow(screen, "top_p", pp.topPSlider)
	pp.drawPresetButtons(screen)
	pp.advToggle.Draw(screen)
	if pp.showAdvanced {
		pp.drawRow(screen, "max tok", pp.maxTokInput)
		pp.drawRow(screen, "context", pp.ctxInput)
		pp.drawRow(screen, "repeat", pp.repeatSlider)
		pp.drawRow(screen, "top_k", pp.topKInput)
	}
}

// drawRow renders a label + widget pair on one row.
func (pp *ParamPanel) drawRow(screen *ebiten.Image, label string, w Widget) {
	b := w.Bounds()
	lx := pp.bounds.Min.X + pp.theme.Padding
	ly := b.Min.Y + (b.Dy()-CharHeight)/2
	drawText(screen, label+":", lx, ly, pp.theme.TextMuted)
	w.Draw(screen)
}

// drawPresetButtons draws three preset action buttons.
func (pp *ParamPanel) drawPresetButtons(screen *ebiten.Image) {
	for _, btn := range pp.presetBtns {
		btn.Draw(screen)
	}
}

// layoutWidgets distributes control bounds in the panel.
func (pp *ParamPanel) layoutWidgets() {
	lx := pp.bounds.Min.X + pp.theme.Padding + 7*CharWidth + pp.theme.Padding
	rw := pp.bounds.Max.X - lx - pp.theme.Padding
	rh := pp.rowH()
	pp.setRowBounds(pp.modelDrop, lx, 0, rw, rh)
	pp.setRowBounds(pp.tempSlider, lx, 1, rw, rh)
	pp.setRowBounds(pp.topPSlider, lx, 2, rw, rh)
	// position preset buttons on row 3
	x := lx
	y := pp.bounds.Min.Y + pp.theme.Padding + 3*rh
	for i, name := range paramPresetNames {
		w := len(name)*CharWidth + pp.theme.Padding*2
		pp.presetBtns[i].SetBounds(image.Rect(x, y, x+w, y+rh-pp.theme.Margin))
		x += w + pp.theme.Margin
	}
	pp.advToggle.SetBounds(image.Rect(
		pp.bounds.Min.X+pp.theme.Padding, pp.bounds.Min.Y+pp.theme.Padding+4*rh,
		pp.bounds.Max.X-pp.theme.Padding, pp.bounds.Min.Y+pp.theme.Padding+5*rh-pp.theme.Margin))
	if pp.showAdvanced {
		pp.setRowBounds(pp.maxTokInput, lx, 5, rw, rh)
		pp.setRowBounds(pp.ctxInput, lx, 6, rw, rh)
		pp.setRowBounds(pp.repeatSlider, lx, 7, rw, rh)
		pp.setRowBounds(pp.topKInput, lx, 8, rw, rh)
	}
}

// setRowBounds positions widget w at the given row index.
func (pp *ParamPanel) setRowBounds(w Widget, lx, row, rw, rh int) {
	y := pp.bounds.Min.Y + pp.theme.Padding + row*rh
	w.SetBounds(image.Rect(lx, y, lx+rw, y+rh-pp.theme.Margin))
}

// rowH returns the pixel height of one parameter row.
func (pp *ParamPanel) rowH() int { return CharHeight + pp.theme.Padding*2 }

// syncWidgets updates sub-widget values from pp.params.
func (pp *ParamPanel) syncWidgets() {
	pp.tempSlider.SetValue(pp.params.Temperature)
	pp.topPSlider.SetValue(pp.params.TopP)
	pp.repeatSlider.SetValue(pp.params.RepeatPen)
	pp.maxTokInput.SetValue(strconv.Itoa(pp.params.MaxTokens))
	pp.ctxInput.SetValue(strconv.Itoa(pp.params.ContextLen))
	pp.topKInput.SetValue(strconv.Itoa(pp.params.TopK))
}

// syncFromTextInputs reads text fields and updates pp.params integer fields.
func (pp *ParamPanel) syncFromTextInputs() {
	if v, err := strconv.Atoi(pp.maxTokInput.Value()); err == nil {
		pp.setAndValidate(func(p *InferenceParams) { p.MaxTokens = v })
	}
	if v, err := strconv.Atoi(pp.ctxInput.Value()); err == nil {
		pp.setAndValidate(func(p *InferenceParams) { p.ContextLen = v })
	}
	if v, err := strconv.Atoi(pp.topKInput.Value()); err == nil {
		pp.setAndValidate(func(p *InferenceParams) { p.TopK = v })
	}
}

// setAndValidate applies a mutation and re-validates, emitting an error string.
func (pp *ParamPanel) setAndValidate(mutate func(p *InferenceParams)) {
	mutate(&pp.params)
	pp.validate()
}

// validate runs Validate and stores the error text (empty on success).
func (pp *ParamPanel) validate() {
	if err := pp.params.Validate(); err != nil {
		pp.errorText = fmt.Sprintf("⚠ %s", err.Error())
		return
	}
	pp.errorText = ""
	if pp.onChange != nil {
		pp.onChange(pp.params)
	}
}
