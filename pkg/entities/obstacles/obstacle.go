package obstacles

import (
	"UTC_IA04/pkg/models"
)

// Obstacle represents an immovable object in the environment
type Obstacle struct {
	uid      int
	Position models.Position
}

// NewObstacle creates a new instance of an Obstacle
func NewObstacle(uid int, position models.Position) Obstacle {
	return Obstacle{
		uid:      uid,
		Position: position,
	}
}

// IsBlocking checks if the obstacles blocks a given position
func (o *Obstacle) IsBlocking(target models.Position) bool {
	return o.Position.X == target.X && o.Position.Y == target.Y
}

func (o *Obstacle) GetUid() int {
	return o.uid
}
