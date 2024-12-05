package components

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Tooltip struct {
	Text      []string
	X, Y      float64
	Padding   float64
	BGColor   color.Color
	TextColor color.Color
	Visible   bool
}

func NewTooltip() *Tooltip {
	return &Tooltip{
		Padding:   10,
		BGColor:   color.RGBA{40, 40, 40, 230},
		TextColor: color.White,
	}
}

func (t *Tooltip) Show(x, y float64, text ...string) {
	t.X = x
	t.Y = y
	t.Text = text
	t.Visible = true
}

func (t *Tooltip) Hide() {
	t.Visible = false
}

func (t *Tooltip) Draw(screen *ebiten.Image) {
	if !t.Visible || len(t.Text) == 0 {
		return
	}

	// Calculate dimensions
	maxWidth := 0
	for _, line := range t.Text {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	width := float64(maxWidth)*6 + t.Padding*2
	height := float64(len(t.Text))*15 + t.Padding*2

	// Draw background
	tooltipBG := ebiten.NewImage(int(width), int(height))
	tooltipBG.Fill(t.BGColor)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(t.X, t.Y)
	screen.DrawImage(tooltipBG, op)

	// Draw text
	for i, line := range t.Text {
		ebitenutil.DebugPrintAt(
			screen,
			line,
			int(t.X+t.Padding),
			int(t.Y+t.Padding)+i*15,
		)
	}
}
