package persons

import (
	"UTC_IA04/pkg/models"
	"math/rand"
)

type PersonState int

const (
	Exploring PersonState = iota
	SeekingPOI
	Resting
	InQueue
	InDistress
)

type StateData struct {
	CurrentState PersonState
	TimeInState  int
	LastRestTime int
	TargetPOI    *models.POIType
	CurrentPath  []models.Position
}

func NewStateData() StateData {
	return StateData{
		CurrentState: Exploring,
		TimeInState:  0,
		LastRestTime: 0,
		TargetPOI:    nil,
		CurrentPath:  make([]models.Position, 0),
	}
}

func (s *StateData) UpdateState(person *Person) {
	s.TimeInState++

	// If person is in distress, override other states
	if person.InDistress {
		s.CurrentState = InDistress
		s.TargetPOI = nil // Clear any POI target
		return
	}

	switch s.CurrentState {
	case Exploring:
		// Check if person needs rest
		if person.Profile.StaminaLevel <= person.Profile.RestThreshold {
			s.CurrentState = SeekingPOI
			s.TargetPOI = poiTypePtr(models.RestArea)
			s.TimeInState = 0
		} else {
			// Chance to seek POI based on interests and pattern
			for poiType, interest := range person.ZonePreference.POIPreferences {
				if interest > 0.7 && rand.Float64() < interest {
					s.CurrentState = SeekingPOI
					s.TargetPOI = poiTypePtr(poiType)
					s.TimeInState = 0
					break
				}
			}
		}

	case SeekingPOI:
		// If reached POI, transition to InQueue
		if person.HasReachedPOI() {
			s.CurrentState = InQueue
			s.TimeInState = 0
		} else if s.TimeInState > 50 { // If taking too long to reach POI
			s.CurrentState = Exploring
			s.TargetPOI = nil
			s.TimeInState = 0
		}

	case Resting:
		// After sufficient rest, return to exploring
		if s.TimeInState > 50 && person.Profile.StaminaLevel > 0.8 {
			s.CurrentState = Exploring
			s.TimeInState = 0
			s.LastRestTime = 0
			s.TargetPOI = nil
		}

	case InQueue:
		// After using POI, return to exploring
		if s.TimeInState > 20 {
			person.Profile.StaminaLevel = 0.8

			s.CurrentState = Exploring
			s.TimeInState = 0
			s.TargetPOI = nil
		}

	case InDistress:

		// Stay in distress until rescued
		// No state change possible until rescue
		return
	}
}

func poiTypePtr(poi models.POIType) *models.POIType {
	return &poi
}
