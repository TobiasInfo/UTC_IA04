package obstacles

import (
	"UTC_IA04/pkg/models"
	"sync"
)

// Obstacle represents an immovable object in the environment
type Obstacle struct {
	uid        int
	Position   models.Position
	POIType    models.POIType
	Capacity   int
	CurrentUse int
	mu         sync.RWMutex
}

// NewObstacle creates a new instance of an Obstacle
func NewObstacle(uid int, position models.Position, poiType models.POIType, capacity int) Obstacle {
	return Obstacle{
		uid:        uid,
		Position:   position,
		POIType:    poiType,
		Capacity:   capacity,
		CurrentUse: 0,
		mu:         sync.RWMutex{},
	}
}

func (o *Obstacle) GetPOIType() models.POIType {
	return o.POIType
}