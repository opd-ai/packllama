package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Slider is a horizontal range-input widget. Values are clamped to [Min, Max].
type Slider struct {
	baseWidget
	Min      float64
	Max      float64
	value    float64
	dragging bool
	onChange func(value float64)
}

// NewSlider creates a Slider with the given range and initial value.
func NewSlider(min, max, initial float64, theme Theme) *Slider {
	s := &Slider{
		baseWidget: baseWidget{theme: theme},
		Min:        min,
		Max:        max,
	}
	s.SetValue(initial)
	return s
}

// Focusable reports that sliders accept keyboard focus.
func (s *Slider) Focusable() bool { return true }

// Value returns the current slider value.
func (s *Slider) Value() float64 { return s.value }

// SetValue clamps v to [Min, Max] and stores it.
func (s *Slider) SetValue(v float64) {
	switch {
	case v < s.Min:
		s.value = s.Min
	case v > s.Max:
		s.value = s.Max
	default:
		s.value = v
	}
}

// OnChange registers a callback invoked whenever the value changes.
func (s *Slider) OnChange(fn func(value float64)) { s.onChange = fn }

// Update handles drag interactions to adjust the slider value.
func (s *Slider) Update() error {
	mx, my := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && inBounds(s.bounds, mx, my) {
		s.dragging = true
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		s.dragging = false
	}
	if s.dragging {
		s.updateFromMouse(mx, my)
	}
	if s.focused {
		s.handleKeys()
	}
	return nil
}

// updateFromMouse sets the value proportionally to the mouse x position.
func (s *Slider) updateFromMouse(mx, _ int) {
	track := s.trackBounds()
	t := float64(mx-track.Min.X) / float64(track.Dx())
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	newVal := s.Min + t*(s.Max-s.Min)
	s.SetValue(newVal)
	if s.onChange != nil {
		s.onChange(s.value)
	}
}

// handleKeys adjusts the value with arrow keys when focused.
func (s *Slider) handleKeys() {
	step := (s.Max - s.Min) / 20
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		s.SetValue(s.value + step)
		if s.onChange != nil {
			s.onChange(s.value)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		s.SetValue(s.value - step)
		if s.onChange != nil {
			s.onChange(s.value)
		}
	}
}

// Draw renders the slider track, fill, thumb, and current value label.
func (s *Slider) Draw(screen *ebiten.Image) {
	fillRect(screen, s.bounds, s.theme.Surface)
	track := s.trackBounds()
	fillRect(screen, track, s.theme.Border)
	fill := s.fillBounds(track)
	fillRect(screen, fill, s.theme.Primary)
	thumb := s.thumbBounds(track)
	fillRect(screen, thumb, s.theme.Text)
	label := fmt.Sprintf("%.2f", s.value)
	x := s.bounds.Min.X + s.bounds.Dx() - len(label)*CharWidth - s.theme.Padding
	y := s.bounds.Min.Y + (s.bounds.Dy()-CharHeight)/2
	drawText(screen, label, x, y, s.theme.TextMuted)
}

// trackBounds returns the horizontal track rectangle.
func (s *Slider) trackBounds() image.Rectangle {
	cx := s.bounds.Min.X + s.theme.Padding
	cy := s.bounds.Min.Y + s.bounds.Dy()/2 - 2
	return image.Rect(cx, cy, s.bounds.Max.X-s.theme.Padding*4-6*CharWidth, cy+4)
}

// fillBounds returns the filled portion of the track up to the current value.
func (s *Slider) fillBounds(track image.Rectangle) image.Rectangle {
	if s.Max == s.Min {
		return track
	}
	t := (s.value - s.Min) / (s.Max - s.Min)
	w := int(float64(track.Dx()) * t)
	return image.Rect(track.Min.X, track.Min.Y, track.Min.X+w, track.Max.Y)
}

// thumbBounds returns the rectangle for the draggable thumb.
func (s *Slider) thumbBounds(track image.Rectangle) image.Rectangle {
	const thumbW, thumbH = 10, 16
	if s.Max == s.Min {
		return image.Rect(track.Min.X, track.Min.Y-thumbH/2, track.Min.X+thumbW, track.Min.Y+thumbH/2)
	}
	t := (s.value - s.Min) / (s.Max - s.Min)
	cx := track.Min.X + int(float64(track.Dx())*t)
	cy := s.bounds.Min.Y + s.bounds.Dy()/2
	return image.Rect(cx-thumbW/2, cy-thumbH/2, cx+thumbW/2, cy+thumbH/2)
}
