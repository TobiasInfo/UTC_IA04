package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func drawRectangle(img *ebiten.Image, x, y, width, height float64, clr color.Color) {
	rect := ebiten.NewImage(int(width), int(height))
	rect.Fill(clr)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x, y)
	img.DrawImage(rect, opts)
}

func drawCircle(img *ebiten.Image, x, y, radius float64, clr color.Color) {
	circle := ebiten.NewImage(int(2*radius), int(2*radius))
	for cy := 0; cy < int(2*radius); cy++ {
		for cx := 0; cx < int(2*radius); cx++ {
			dx := float64(cx) - radius
			dy := float64(cy) - radius
			if dx*dx+dy*dy <= radius*radius {
				circle.Set(cx, cy, clr)
			}
		}
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x-radius, y-radius)
	img.DrawImage(circle, opts)
}

func drawTranslucentCircle(img *ebiten.Image, x, y, radius float64, clr color.Color) {
	circle := ebiten.NewImage(int(2*radius), int(2*radius))
	for cy := 0; cy < int(2*radius); cy++ {
		for cx := 0; cx < int(2*radius); cx++ {
			dx := float64(cx) - radius
			dy := float64(cy) - radius
			if dx*dx+dy*dy <= radius*radius {
				circle.Set(cx, cy, clr)
			}
		}
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x-radius, y-radius)
	img.DrawImage(circle, opts)
}
