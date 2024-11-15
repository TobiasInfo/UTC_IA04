package simulation

import (
	"UTC_IA04/pkg/models"
)

// Obstacle represents an immovable object in the environment
type Obstacle struct {
	Position models.Position
}

// NewObstacle creates a new instance of an Obstacle
func NewObstacle(position models.Position) *Obstacle {
	return &Obstacle{
		Position: position,
	}
}

// IsBlocking checks if the obstacle blocks a given position
func (o *Obstacle) IsBlocking(target models.Position) bool {
	return o.Position.X == target.X && o.Position.Y == target.Y
}
