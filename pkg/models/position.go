package models

import "math"

type Position struct {
	X, Y float64
}

// Round rounds the position to the nearest integer
// e.g., (1.1, 2.2) -> (1, 2)
// Deprecated: Don't use this, will be removed soon
func (p Position) Round() Position {
	return Position{
		X: math.Round(p.X*10) / 10,
		Y: math.Round(p.Y*10) / 10,
	}
}

func (p *Position) CalculateDistance(other Position) float64 {
	return math.Sqrt(math.Pow(p.X-other.X, 2) + math.Pow(p.Y-other.Y, 2))
}

func (p *Position) CalculateManhattanDistance(other Position) float64 {
	return math.Abs(p.X-other.X) + math.Abs(p.Y-other.Y)
}