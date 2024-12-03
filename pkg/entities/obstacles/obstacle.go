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

// IsBlocking checks if the obstacle blocks a given position
func (o *Obstacle) IsBlocking(target models.Position) bool {
	return o.Position.X == target.X && o.Position.Y == target.Y
}

func (o *Obstacle) GetUid() int {
	return o.uid
}

// IsFull checks if the POI is at capacity
func (o *Obstacle) IsFull() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.CurrentUse >= o.Capacity
}

// IncrementUsage increases the current usage count
func (o *Obstacle) IncrementUsage() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.CurrentUse < o.Capacity {
		o.CurrentUse++
		return true
	}
	return false
}

// DecrementUsage decreases the current usage count
func (o *Obstacle) DecrementUsage() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.CurrentUse > 0 {
		o.CurrentUse--
	}
}

func (o *Obstacle) GetPOIType() models.POIType {
	return o.POIType
}

func (o *Obstacle) GetCapacity() int {
	return o.Capacity
}
