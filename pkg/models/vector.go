package models

import "math"

type Vector struct {
	X, Y float64
}

func (v *Vector) rotate(angle float64) (float64, float64) {
	rad := angle * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	newX := v.X*cos - v.Y*sin
	newY := v.X*sin + v.Y*cos
	return newX, newY
}

func (v *Vector) RotateInt(angle float64) Position {
	newX, newY := v.rotate(angle)
	return Position{X: float64(int(newX)), Y: float64(int(newY))}
}

func (v *Vector) RotateFloat(angle float64) Position {
	newX, newY := v.rotate(angle)
	return Position{X: math.Round(newX*10) / 10, Y: math.Round(newY*10) / 10}
}

// GenerateCircleValues generates the values of a circle with the given radius
func (v *Vector) GenerateCircleValues(radius int) ([]Position, []Position) {
	var floatValues []Position
	var intValues []Position

	intValues = append(intValues, Position{X: 0, Y: 0})
	floatValues = append(intValues, Position{X: 0, Y: 0})

	remarkableAngles := []float64{0, 30, 45, 60, 90, 120, 135, 150, 180, 210, 225, 240, 270, 300, 315, 330}
	for _, angle := range remarkableAngles {
		v.X, v.Y = float64(radius), 0
		floatValues = append(floatValues, v.RotateFloat(angle))
		intValues = append(intValues, v.RotateInt(angle))
	}

	return floatValues, intValues
}
