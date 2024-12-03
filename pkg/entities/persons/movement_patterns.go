package persons

import (
	"UTC_IA04/pkg/models"
	"math/rand"
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
}

func GetZonePreference(pattern MovementPattern) ZonePreference {
	pref := ZonePreference{
		POIPreferences: make(map[models.POIType]float64),
	}

	switch pattern {
	case MainEventFocused:
		pref.EntranceZoneWeight = 0.1
		pref.MainZoneWeight = 0.8
		pref.ExitZoneWeight = 0.1
		// POI preferences
		pref.POIPreferences[models.MainStage] = 0.9
		pref.POIPreferences[models.SecondaryStage] = 0.7
		pref.POIPreferences[models.FoodStand] = 0.4
		pref.POIPreferences[models.DrinkStand] = 0.4

	case SocialButterfly:
		pref.EntranceZoneWeight = 0.3
		pref.MainZoneWeight = 0.4
		pref.ExitZoneWeight = 0.3
		// High interest in all POIs
		pref.POIPreferences[models.MainStage] = 0.7
		pref.POIPreferences[models.SecondaryStage] = 0.7
		pref.POIPreferences[models.FoodStand] = 0.7
		pref.POIPreferences[models.DrinkStand] = 0.7
		pref.POIPreferences[models.RestArea] = 0.7
		pref.POIPreferences[models.Toilet] = 0.7

	case EarlyExiter:
		pref.EntranceZoneWeight = 0.2
		pref.MainZoneWeight = 0.3
		pref.ExitZoneWeight = 0.5
		// Lower POI interest
		pref.POIPreferences[models.MainStage] = 0.3
		pref.POIPreferences[models.SecondaryStage] = 0.3
		pref.POIPreferences[models.FoodStand] = 0.4
		pref.POIPreferences[models.DrinkStand] = 0.4

	case LateArrival:
		pref.EntranceZoneWeight = 0.6
		pref.MainZoneWeight = 0.3
		pref.ExitZoneWeight = 0.1
		// Normal POI interest
		pref.POIPreferences[models.MainStage] = 0.5
		pref.POIPreferences[models.SecondaryStage] = 0.5
		pref.POIPreferences[models.FoodStand] = 0.5
		pref.POIPreferences[models.DrinkStand] = 0.5

	case FoodEnthusiast:
		pref.EntranceZoneWeight = 0.2
		pref.MainZoneWeight = 0.6
		pref.ExitZoneWeight = 0.2
		// High food/drink interest
		pref.POIPreferences[models.FoodStand] = 0.9
		pref.POIPreferences[models.DrinkStand] = 0.9
		pref.POIPreferences[models.MainStage] = 0.3
		pref.POIPreferences[models.SecondaryStage] = 0.3
	}

	// Set defaults for any unset POIs
	for _, poiType := range []models.POIType{
		models.Toilet,
		models.RestArea,
	} {
		if _, exists := pref.POIPreferences[poiType]; !exists {
			pref.POIPreferences[poiType] = 0.5
		}
	}

	return pref
}

// Helper to determine if we should move between zones based on preference
func (z *ZonePreference) ShouldMoveToZone(currentZone string) bool {
	switch currentZone {
	case "entrance":
		return rand.Float64() > z.EntranceZoneWeight
	case "main":
		return rand.Float64() > z.MainZoneWeight
	case "exit":
		return rand.Float64() > z.ExitZoneWeight
	default:
		return false
	}
}

// Helper to get POI preference
func (z *ZonePreference) GetPOIPreference(poiType models.POIType) float64 {
	if pref, exists := z.POIPreferences[poiType]; exists {
		return pref
	}
	return 0.5 // default preference
}
