package ui

import (
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type TextField struct {
	X, Y, Width, Height float64
	Text                string
	IsActive            bool
	OnEnter             func(value int)
}

func (tf *TextField) Draw(screen *ebiten.Image) {
	field := ebiten.NewImage(int(tf.Width), int(tf.Height))
	if tf.IsActive {
		field.Fill(color.RGBA{200, 200, 255, 255}) // Light blue when active
	} else {
		field.Fill(color.RGBA{255, 255, 255, 255}) // White otherwise
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(tf.X, tf.Y)
	screen.DrawImage(field, opts)

	ebitenutil.DebugPrintAt(screen, tf.Text, int(tf.X+5), int(tf.Y+5))
}

func (tf *TextField) Update(mx, my float64, pressed bool, inputChars []rune, enterPressed bool) {
	if pressed && mx >= tf.X && mx <= tf.X+tf.Width && my >= tf.Y && my <= tf.Y+tf.Height {
		tf.IsActive = true
	} else if pressed {
		tf.IsActive = false
	}

	if tf.IsActive {
		for _, char := range inputChars {
			if char >= '0' && char <= '9' {
				tf.Text += string(char)
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyBackspace) && len(tf.Text) > 0 {
			tf.Text = tf.Text[:len(tf.Text)-1]
		}

		if enterPressed {
			value, err := strconv.Atoi(tf.Text)
			if err == nil && tf.OnEnter != nil {
				tf.OnEnter(value)
			}
			tf.IsActive = false
		}
	}
}
