package assets

import (
	"UTC_IA04/pkg/models"
)

// POIIcon returns an SVG string for the given POI type
func POIIcon(poiType models.POIType) string {
	switch poiType {
	case 0: // MedicalTent
		return "img/tente-secours.png"
	case 1: // ChargingStation
		return "img/station chargement.png"
	case 2: // Toilet
		return "img/toilets-real.png"
	case 3: // DrinkStand
		return "img/drink-stand-real.png"
	case 4: // FoodStand
		return "img/food-real.png"
	case 5: // MainStage
		return "img/mainstage-real.png"
	case 6: // SecondaryStage
		return "img/stage-second-real.png"
	case 7: // RestArea
		return "img/aire-repos-real.png"
	default: //POI pas défini
		return "img/bugs.png"
	}
}

func PoiScale(poiType models.POIType) float64 {
	switch poiType {
	case 0: // MedicalTent
		return 0.14
	case 1: // ChargingStation
		return 0.10
	case 2: // Toilet
		return 0.10
	case 3: // DrinkStand
		return 0.10
	case 4: // FoodStand
		return 0.10
	case 5: // MainStage
		return 0.25
	case 6: // SecondaryStage
		return 0.15
	case 7: // RestArea
		return 0.15
	default: //POI pas défini
		return 0.10
	}
}
