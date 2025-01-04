package ui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Dropdown struct {
	X, Y, Width, Height float64
	Options             []string
	SelectedIndex       int
	IsOpen              bool
	OnSelect            func(int)
	lastClicked         time.Time
}

func (d *Dropdown) Draw(screen *ebiten.Image) {
	boxColor := color.RGBA{200, 200, 200, 255}
	if d.IsOpen {
		boxColor = color.RGBA{220, 220, 220, 255}
	}
	dropdownImage := ebiten.NewImage(int(d.Width), int(d.Height))
	dropdownImage.Fill(boxColor)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(d.X, d.Y)
	screen.DrawImage(dropdownImage, opts)

	if d.SelectedIndex >= 0 && d.SelectedIndex < len(d.Options) {
		textX := int(d.X + 10)
		textY := int(d.Y + d.Height/2 - 7)
		ebitenutil.DebugPrintAt(screen, d.Options[d.SelectedIndex], textX, textY)
	}

	arrowImage := ebiten.NewImage(10, 10)
	arrowImage.Fill(color.RGBA{50, 50, 50, 255})
	arrowOpts := &ebiten.DrawImageOptions{}
	arrowOpts.GeoM.Translate(d.X+d.Width-15, d.Y+d.Height/2-5)
	screen.DrawImage(arrowImage, arrowOpts)

	if d.IsOpen {
		for i, option := range d.Options {
			optionY := d.Y + d.Height*float64(i+1)
			optionImage := ebiten.NewImage(int(d.Width), int(d.Height))
			optionColor := color.RGBA{255, 255, 255, 255}
			if i == d.SelectedIndex {
				optionColor = color.RGBA{0, 128, 255, 255}
			}
			optionImage.Fill(optionColor)
			optionOpts := &ebiten.DrawImageOptions{}
			optionOpts.GeoM.Translate(d.X, optionY)
			screen.DrawImage(optionImage, optionOpts)

			textX := int(d.X + 10)
			textY := int(optionY + d.Height/2 - 7)
			ebitenutil.DebugPrintAt(screen, option, textX, textY)
		}
	}
}

func (d *Dropdown) Update(mx, my float64, pressed bool) {
	if time.Since(d.lastClicked) < 200*time.Millisecond {
		return
	}

	if pressed {
		if mx >= d.X && mx <= d.X+d.Width && my >= d.Y && my <= d.Y+d.Height {
			d.IsOpen = !d.IsOpen
			d.lastClicked = time.Now()
			return
		}

		if d.IsOpen {
			for i := range d.Options {
				optionY := d.Y + d.Height*float64(i+1)
				if mx >= d.X && mx <= d.X+d.Width && my >= optionY && my <= optionY+d.Height {
					d.SelectedIndex = i
					d.IsOpen = false
					d.lastClicked = time.Now()
					if d.OnSelect != nil {
						d.OnSelect(i)
					}
					return
				}
			}
		}
		d.IsOpen = false
	}
}
