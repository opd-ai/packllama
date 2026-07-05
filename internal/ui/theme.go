package ui

import "image/color"

// Theme defines visual styling used by all UI widgets.
type Theme struct {
	// Background is the window/panel fill color.
	Background color.RGBA
	// Surface is the widget background color.
	Surface color.RGBA
	// Primary is the accent and highlight color.
	Primary color.RGBA
	// Text is the default text color.
	Text color.RGBA
	// TextMuted is used for placeholder and secondary text.
	TextMuted color.RGBA
	// Border is the widget border color.
	Border color.RGBA
	// Hover is the background when the pointer is over a widget.
	Hover color.RGBA
	// Active is the background when a widget is pressed.
	Active color.RGBA
	// Error is used for validation-error indicators.
	Error color.RGBA
	// Padding is the inner spacing in pixels.
	Padding int
	// Margin is the outer spacing in pixels.
	Margin int
	// BorderWidth is the border thickness in pixels.
	BorderWidth int
}

// DefaultDark returns a dark Catppuccin Mocha-inspired theme.
func DefaultDark() Theme {
	return Theme{
		Background:  color.RGBA{R: 0x1e, G: 0x1e, B: 0x2e, A: 0xff},
		Surface:     color.RGBA{R: 0x31, G: 0x32, B: 0x44, A: 0xff},
		Primary:     color.RGBA{R: 0x89, G: 0xb4, B: 0xfa, A: 0xff},
		Text:        color.RGBA{R: 0xcd, G: 0xd6, B: 0xf4, A: 0xff},
		TextMuted:   color.RGBA{R: 0x6c, G: 0x70, B: 0x86, A: 0xff},
		Border:      color.RGBA{R: 0x45, G: 0x47, B: 0x5a, A: 0xff},
		Hover:       color.RGBA{R: 0x45, G: 0x47, B: 0x5a, A: 0xff},
		Active:      color.RGBA{R: 0x58, G: 0x5b, B: 0x70, A: 0xff},
		Error:       color.RGBA{R: 0xf3, G: 0x8b, B: 0xa8, A: 0xff},
		Padding:     8,
		Margin:      4,
		BorderWidth: 1,
	}
}

// DefaultLight returns a light Catppuccin Latte-inspired theme.
func DefaultLight() Theme {
	return Theme{
		Background:  color.RGBA{R: 0xef, G: 0xf1, B: 0xf5, A: 0xff},
		Surface:     color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		Primary:     color.RGBA{R: 0x17, G: 0x93, B: 0xd1, A: 0xff},
		Text:        color.RGBA{R: 0x4c, G: 0x4f, B: 0x69, A: 0xff},
		TextMuted:   color.RGBA{R: 0x8c, G: 0x8f, B: 0xa8, A: 0xff},
		Border:      color.RGBA{R: 0xcc, G: 0xd0, B: 0xda, A: 0xff},
		Hover:       color.RGBA{R: 0xe6, G: 0xe9, B: 0xef, A: 0xff},
		Active:      color.RGBA{R: 0xdc, G: 0xe0, B: 0xe8, A: 0xff},
		Error:       color.RGBA{R: 0xd2, G: 0x0f, B: 0x39, A: 0xff},
		Padding:     8,
		Margin:      4,
		BorderWidth: 1,
	}
}
