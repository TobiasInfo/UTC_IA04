package interfaces

import (
	"UTC_IA04/pkg/models"
	"sync"
)

type DroneMemory struct {
	DronePatrolPath   []models.Position
	DroneActualTarget models.Position
	ReturningToStart  bool
	Persons           struct {
		PersonsToSave sync.Map
	}
}
