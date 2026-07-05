package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// SysInfoPanel is a display widget that shows system resource gauges and
// inference metadata. Values are set externally via Update.
type SysInfoPanel struct {
	baseWidget
	info SysInfo
}

// NewSysInfoPanel creates a SysInfoPanel with the given theme.
func NewSysInfoPanel(theme Theme) *SysInfoPanel {
	return &SysInfoPanel{baseWidget: baseWidget{theme: theme}}
}

// Focusable reports that the system info panel does not accept keyboard focus.
func (p *SysInfoPanel) Focusable() bool { return false }

// SetInfo replaces the displayed SysInfo snapshot.
func (p *SysInfoPanel) SetInfo(info SysInfo) { p.info = info }

// Update is a no-op; data is pushed via SetInfo.
func (p *SysInfoPanel) Update() error { return nil }

// Draw renders all system info rows.
func (p *SysInfoPanel) Draw(screen *ebiten.Image) {
	fillRect(screen, p.bounds, p.theme.Surface)
	drawBorder(screen, p.bounds, p.theme.BorderWidth, p.theme.Border)
	p.drawRows(screen)
}

// drawRows renders each metric row inside the panel.
func (p *SysInfoPanel) drawRows(screen *ebiten.Image) {
	rh := CharHeight + p.theme.Padding*2
	rows := []struct {
		label string
		fn    func(*ebiten.Image, image.Rectangle)
	}{
		{"CPU", p.drawCPURow},
		{"RAM", p.drawMemRow},
		{"GPU", p.drawGPURow},
		{"Speed", p.drawSpeedRow},
		{"Model", p.drawModelRow},
		{"Queue", p.drawQueueRow},
		{"Health", p.drawHealthRow},
	}
	for i, row := range rows {
		y := p.bounds.Min.Y + p.theme.Padding + i*rh
		r := image.Rect(p.bounds.Min.X, y, p.bounds.Max.X, y+rh)
		drawText(screen, row.label+":", r.Min.X+p.theme.Padding, r.Min.Y+(rh-CharHeight)/2, p.theme.TextMuted)
		row.fn(screen, r)
	}
}

// drawCPURow renders the CPU percentage gauge.
func (p *SysInfoPanel) drawCPURow(screen *ebiten.Image, r image.Rectangle) {
	p.drawGauge(screen, r, p.info.CPUPercent/100, gaugeColor(p.info.CPUPercent/100, p.theme))
	label := fmt.Sprintf("%.1f%%", p.info.CPUPercent)
	p.drawGaugeLabel(screen, r, label)
}

// drawMemRow renders the memory percentage gauge.
func (p *SysInfoPanel) drawMemRow(screen *ebiten.Image, r image.Rectangle) {
	pct := p.info.MemPercent() / 100
	p.drawGauge(screen, r, pct, gaugeColor(pct, p.theme))
	label := fmt.Sprintf("%.1f%%  %s / %s",
		p.info.MemPercent(),
		formatBytes(p.info.MemUsedBytes),
		formatBytes(p.info.MemTotalBytes))
	p.drawGaugeLabel(screen, r, label)
}

// drawGPURow renders the GPU info string.
func (p *SysInfoPanel) drawGPURow(screen *ebiten.Image, r image.Rectangle) {
	label := p.info.GPULabel
	if label == "" {
		label = "n/a"
	}
	drawText(screen, label, r.Min.X+p.theme.Padding+7*CharWidth, r.Min.Y+(r.Dy()-CharHeight)/2, p.theme.Text)
}

// drawSpeedRow renders the inference tokens/sec indicator.
func (p *SysInfoPanel) drawSpeedRow(screen *ebiten.Image, r image.Rectangle) {
	label := fmt.Sprintf("%.1f tok/s", p.info.TokensPerSec)
	drawText(screen, label, r.Min.X+p.theme.Padding+7*CharWidth, r.Min.Y+(r.Dy()-CharHeight)/2, p.theme.Text)
}

// drawModelRow renders the model load status.
func (p *SysInfoPanel) drawModelRow(screen *ebiten.Image, r image.Rectangle) {
	status := p.info.ModelStatus
	if status == "" {
		status = "none"
	}
	drawText(screen, status, r.Min.X+p.theme.Padding+7*CharWidth, r.Min.Y+(r.Dy()-CharHeight)/2, p.theme.Text)
}

// drawQueueRow renders the pending request queue length.
func (p *SysInfoPanel) drawQueueRow(screen *ebiten.Image, r image.Rectangle) {
	label := fmt.Sprintf("%d pending", p.info.QueueLen)
	drawText(screen, label, r.Min.X+p.theme.Padding+7*CharWidth, r.Min.Y+(r.Dy()-CharHeight)/2, p.theme.Text)
}

// drawHealthRow renders the overall health indicator.
func (p *SysInfoPanel) drawHealthRow(screen *ebiten.Image, r image.Rectangle) {
	clr := p.theme.Primary
	label := "● OK"
	if !p.info.HealthOK() {
		clr = p.theme.Error
		label = "● Warning"
	}
	drawText(screen, label, r.Min.X+p.theme.Padding+7*CharWidth, r.Min.Y+(r.Dy()-CharHeight)/2, clr)
}

// drawGauge draws a filled progress bar in the value area of row r.
func (p *SysInfoPanel) drawGauge(screen *ebiten.Image, r image.Rectangle, fraction float64, clr color.Color) {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	lx := r.Min.X + p.theme.Padding + 7*CharWidth + p.theme.Padding
	gy := r.Min.Y + r.Dy()/2 - 4
	gw := r.Max.X - lx - p.theme.Padding - 8*CharWidth
	track := image.Rect(lx, gy, lx+gw, gy+8)
	fillRect(screen, track, p.theme.Background)
	fill := image.Rect(lx, gy, lx+int(float64(gw)*fraction), gy+8)
	fillRect(screen, fill, clr)
	drawBorder(screen, track, p.theme.BorderWidth, p.theme.Border)
}

// drawGaugeLabel draws the label text to the right of the gauge.
func (p *SysInfoPanel) drawGaugeLabel(screen *ebiten.Image, r image.Rectangle, label string) {
	lx := r.Max.X - p.theme.Padding - len(label)*CharWidth
	drawText(screen, label, lx, r.Min.Y+(r.Dy()-CharHeight)/2, p.theme.Text)
}

// gaugeColor returns green/yellow/red based on load fraction.
func gaugeColor(fraction float64, t Theme) color.Color {
	switch {
	case fraction > 0.85:
		return t.Error
	case fraction > 0.60:
		return color.RGBA{R: 0xf9, G: 0xe2, B: 0xaf, A: 0xff} // yellow
	default:
		return t.Primary
	}
}

// formatBytes converts a byte count to a human-readable string.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}
