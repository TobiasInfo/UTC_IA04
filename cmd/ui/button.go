package ui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Button struct {
	X, Y, Width, Height float64
	Text                string
	OnClick             func()
	lastClicked         time.Time
}

func (b *Button) Draw(screen *ebiten.Image) {
	buttonImage := ebiten.NewImage(int(b.Width), int(b.Height))
	buttonImage.Fill(color.RGBA{0, 128, 255, 255}) // Blue background
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(b.X, b.Y)
	screen.DrawImage(buttonImage, opts)

	textX := int(b.X + b.Width/4)
	textY := int(b.Y + b.Height/3)
	ebitenutil.DebugPrintAt(screen, b.Text, textX, textY)
}

func (b *Button) Update(mx, my float64, pressed bool) {
	if time.Since(b.lastClicked) < 500*time.Millisecond {
		return
	}

	if pressed && mx >= b.X && mx <= b.X+b.Width && my >= b.Y && my <= b.Y+b.Height {
		b.lastClicked = time.Now()
		if b.OnClick != nil {
			b.OnClick()
		}
	}
}
