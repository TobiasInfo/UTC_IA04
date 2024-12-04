package models

import "math"

type Position struct {
	X, Y float64
}

func (p Position) Round() Position {
	return Position{
		X: math.Round(p.X*10) / 10,
		Y: math.Round(p.Y*10) / 10,
	}
}

func (p *Position) CalculateDistance(other Position) float64 {
	return math.Sqrt(math.Pow(p.X-other.X, 2) + math.Pow(p.Y-other.Y, 2))
}
