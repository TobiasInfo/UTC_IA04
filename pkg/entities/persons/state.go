package persons

import (
	"UTC_IA04/pkg/models"
)

type PersonState int

const (
	Exploring PersonState = iota
	SeekingPOI
	Resting
	InQueue
	Leaving
	InDistress
)

type StateData struct {
	CurrentState PersonState
	TargetPOI    *models.POIType
	TimeInState  int
	LastRestTime int
	CurrentPath  []models.Position
}

func NewStateData() StateData {
	return StateData{
		CurrentState: Exploring,
		TimeInState:  0,
		LastRestTime: 0,
		CurrentPath:  make([]models.Position, 0),
	}
}

func (s *StateData) UpdateState(person *Person) {
	s.TimeInState++

	switch s.CurrentState {
	case Exploring:
		// Check if person needs rest
		if person.Profile.StaminaLevel <= person.Profile.RestThreshold {
			s.CurrentState = SeekingPOI
			s.TargetPOI = poiTypePtr(models.RestArea)
		}

	case SeekingPOI:
		// If reached POI, transition to InQueue
		if person.HasReachedPOI() {
			s.CurrentState = InQueue
			s.TimeInState = 0
		}

	case Resting:
		// After sufficient rest, return to exploring
		if s.TimeInState > 50 && person.Profile.StaminaLevel > 0.8 {
			s.CurrentState = Exploring
			s.TimeInState = 0
			s.LastRestTime = 0
		}

	case InQueue:
		// After using POI, might need rest or can explore
		if person.Profile.StaminaLevel < person.Profile.RestThreshold {
			s.CurrentState = Resting
		} else {
			s.CurrentState = Exploring
		}
		s.TimeInState = 0
	}

	// Override with distress state if needed
	if person.InDistress {
		s.CurrentState = InDistress
	}
}

func poiTypePtr(poi models.POIType) *models.POIType {
	return &poi
}
