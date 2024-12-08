package models

import (
	"math"
)

type Vector struct {
	X, Y float64
}

//func (v *Vector) rotate(angle float64) (float64, float64) {
//	rad := angle * math.Pi / 180
//	cos := math.Cos(rad)
//	sin := math.Sin(rad)
//	newX := v.X*cos - v.Y*sin
//	newY := v.X*sin + v.Y*cos
//	return newX, newY
//}
//
//func (v *Vector) RotateInt(angle float64) Position {
//	newX, newY := v.rotate(angle)
//	return Position{X: float64(int(newX)), Y: float64(int(newY))}
//}
//
//func (v *Vector) RotateFloat(angle float64) Position {
//	newX, newY := v.rotate(angle)
//	return Position{X: math.Round(newX*10) / 10, Y: math.Round(newY*10) / 10}
//}

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

	// 			for u := 0; u < 10; u++ {
	//				for j := 0; j < 10; j++ {
	//					position := models.Position{X: d.Position.X + positionMaster.X + (float64(u) / 10), Y: d.Position.Y + positionMaster.Y + (float64(j) / 10)}

	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			for u := 0; u < 10; u++ {
				for j := 0; j < 10; j++ {
					tempX := math.Round(float64(x)+(float64(u)/10)*100) / 100
					tempY := math.Round(float64(y)+(float64(j)/10)*100) / 100
					if tempX*tempX+tempY*tempY <= float64(radius*radius) {
						floatValues = append(floatValues, Position{X: tempX, Y: tempY})
					}
				}
			}
			//if x*x+y*y <= radius*radius {
			//	intValues = append(intValues, Position{X: float64(x), Y: float64(y)})
			//}
		}
	}

	//for x := -float64(radius); x <= float64(radius); x += 0.1 {
	//	for y := -float64(radius); y <= float64(radius); y += 0.1 {
	//		if x*x+y*y <= float64(radius*radius) {
	//			floatValues = append(floatValues, Position{X: x, Y: y})
	//		}
	//	}
	//

	return floatValues, intValues
}
