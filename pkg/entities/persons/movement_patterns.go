// movement_patterns.go
// Replace entire file contents in pkg/entities/persons/movement_patterns.go

package persons

import (
	"UTC_IA04/pkg/models"
	"math"
	"math/rand"
	"time"
)

type MovementPattern int

const (
	MainEventFocused MovementPattern = iota
	SocialButterfly
	EarlyExiter
	LateArrival
	FoodEnthusiast
)

type ZonePreference struct {
	EntranceZoneWeight float64
	MainZoneWeight     float64
	ExitZoneWeight     float64
	POIPreferences     map[models.POIType]float64
	ExitTime           time.Duration
	LastPOIVisit       map[models.POIType]time.Time
}

func GetZonePreference(pattern MovementPattern) ZonePreference {
	pref := ZonePreference{
		POIPreferences: make(map[models.POIType]float64),
		LastPOIVisit:   make(map[models.POIType]time.Time),
	}

	// Base exit times
	baseExitTime := 3 * time.Hour

	switch pattern {
	case MainEventFocused:
		pref.EntranceZoneWeight = 0.1
		pref.MainZoneWeight = 0.8
		pref.ExitZoneWeight = 0.1
		pref.ExitTime = baseExitTime + time.Duration(rand.Intn(60))*time.Minute
		pref.POIPreferences[models.MainStage] = 0.9
		pref.POIPreferences[models.SecondaryStage] = 0.7
		pref.POIPreferences[models.FoodStand] = 0.4
		pref.POIPreferences[models.DrinkStand] = 0.4

	case SocialButterfly:
		pref.EntranceZoneWeight = 0.3
		pref.MainZoneWeight = 0.4
		pref.ExitZoneWeight = 0.3
		pref.ExitTime = baseExitTime + time.Duration(rand.Intn(120))*time.Minute
		pref.POIPreferences[models.MainStage] = 0.7
		pref.POIPreferences[models.SecondaryStage] = 0.7
		pref.POIPreferences[models.FoodStand] = 0.7
		pref.POIPreferences[models.DrinkStand] = 0.7
		pref.POIPreferences[models.RestArea] = 0.7

	case EarlyExiter:
		pref.EntranceZoneWeight = 0.2
		pref.MainZoneWeight = 0.3
		pref.ExitZoneWeight = 0.5
		pref.ExitTime = baseExitTime - time.Duration(rand.Intn(90))*time.Minute
		pref.POIPreferences[models.MainStage] = 0.3
		pref.POIPreferences[models.SecondaryStage] = 0.3
		pref.POIPreferences[models.FoodStand] = 0.4

	case LateArrival:
		pref.EntranceZoneWeight = 0.6
		pref.MainZoneWeight = 0.3
		pref.ExitZoneWeight = 0.1
		pref.ExitTime = baseExitTime + time.Duration(rand.Intn(180))*time.Minute
		pref.POIPreferences[models.MainStage] = 0.5
		pref.POIPreferences[models.SecondaryStage] = 0.5
		pref.POIPreferences[models.FoodStand] = 0.5

	case FoodEnthusiast:
		pref.EntranceZoneWeight = 0.2
		pref.MainZoneWeight = 0.6
		pref.ExitZoneWeight = 0.2
		pref.ExitTime = baseExitTime
		pref.POIPreferences[models.FoodStand] = 0.9
		pref.POIPreferences[models.DrinkStand] = 0.9
		pref.POIPreferences[models.MainStage] = 0.3
	}

	// Set defaults for any unset POIs
	for _, poiType := range []models.POIType{
		models.Toilet,
		models.RestArea,
		models.DrinkStand,
	} {
		if _, exists := pref.POIPreferences[poiType]; !exists {
			pref.POIPreferences[poiType] = 0.5
		}
		// Initialize last visit times to simulation start
		pref.LastPOIVisit[poiType] = time.Now()
	}

	return pref
}

func (z *ZonePreference) ShouldMoveToZone(currentZone string, entryTime time.Time) bool {
	timeSinceEntry := time.Since(entryTime)

	switch currentZone {
	case "entrance":
		// Always encourage moving from entrance after some time
		if timeSinceEntry > 15*time.Minute {
			return true
		}
		return rand.Float64() > z.EntranceZoneWeight

	case "main":
		// Consider moving to exit if approaching exit time
		if timeSinceEntry > z.ExitTime-30*time.Minute {
			return true
		}
		return rand.Float64() > z.MainZoneWeight

	case "exit":
		// Strong tendency to stay in exit zone once there
		return rand.Float64() > 0.9

	default:
		return false
	}
}

func (z *ZonePreference) GetPOIPreference(poiType models.POIType) float64 {
	if pref, exists := z.POIPreferences[poiType]; exists {
		return pref
	}
	return 0.5 // default preference
}

func (z *ZonePreference) ShouldVisitPOI(poiType models.POIType) bool {
	baseProbability := z.GetPOIPreference(poiType)
	lastVisit, exists := z.LastPOIVisit[poiType]

	if !exists {
		z.LastPOIVisit[poiType] = time.Now()
		return rand.Float64() < baseProbability
	}

	// Increase probability based on time since last visit
	timeSinceVisit := time.Since(lastVisit)
	timeMultiplier := math.Min(timeSinceVisit.Hours()/2.0, 1.0)
	adjustedProbability := baseProbability * (1 + timeMultiplier)

	return rand.Float64() < adjustedProbability
}

func (z *ZonePreference) UpdatePOIVisit(poiType models.POIType) {
	z.LastPOIVisit[poiType] = time.Now()
}

func (z *ZonePreference) GetNextZone(currentZone string, entryTime time.Time) string {
	if z.ShouldMoveToZone(currentZone, entryTime) {
		switch currentZone {
		case "entrance":
			return "main"
		case "main":
			if time.Since(entryTime) > z.ExitTime-30*time.Minute {
				return "exit"
			}
		}
	}
	return currentZone
}
