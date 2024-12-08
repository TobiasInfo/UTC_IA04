package models

import (
	"math"
)

type Vector struct {
	X, Y float64
}

// GenerateCircleValues generates the values of a circle with the given radius
// func (v *Vector) GenerateCircleValues(radius int) ([]Position, []Position) {
func (v *Vector) GenerateCircleValues(radius int) ([]Position, []Position) {
	var floatValues []Position
	var intValues []Position

	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			if x*x+y*y <= radius*radius {
				intValues = append(intValues, Position{X: float64(x), Y: float64(y)})
			}
		}
	}

	for x := -float64(radius); x <= float64(radius); x += 0.1 {
		for y := -float64(radius); y <= float64(radius); y += 0.1 {
			if x*x+y*y <= float64(radius*radius) {
				tempX := math.Round(x*100) / 100
				tempY := math.Round(y*100) / 100
				floatValues = append(floatValues, Position{X: tempX, Y: tempY})
			}
		}
	}

	floatValues = append(floatValues, Position{X: 0, Y: float64(radius)})
	floatValues = append(floatValues, Position{X: float64(radius), Y: 0})

	return floatValues, intValues
}
