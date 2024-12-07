package assets

import "UTC_IA04/pkg/models"

// POIIcon returns an SVG string for the given POI type
func POIIcon(poiType models.POIType) string {
	switch poiType {
	case 0: // MedicalTent
		return "img/premiersecours.png"
	case 1: // ChargingStation
		return "img/premiersecours.png"
	case 2: // Toilet
		return "img/premiersecours.png"
	case 3: // DrinkStand
		return "img/premiersecours.png"
	case 4: // FoodStand
		return "img/premiersecours.png"
	case 5: // MainStage
		return "img/premiersecours.png"
	case 6: // SecondaryStage
		return "img/premiersecours.png"
	case 7: // RestArea
		return "img/premiersecours.png"
	default:
		return ""
	}
}
