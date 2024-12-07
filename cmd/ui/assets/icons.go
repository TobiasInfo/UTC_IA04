package assets

import "UTC_IA04/pkg/models"

// POIIcon returns an SVG string for the given POI type
func POIIcon(poiType models.POIType) string {
	switch poiType {
	case 0: // MedicalTent
		return "img/secoursBW.png"
	case 1: // ChargingStation
		return "img/station chargement.png"
	case 2: // Toilet
		return "img/toilettes.png"
	case 3: // DrinkStand
		return "img/bar.png"
	case 4: // FoodStand
		return "img/fast-food.png"
	case 5: // MainStage
		return "img/mainstage.png"
	case 6: // SecondaryStage
		return "img/secondarystage.png"
	case 7: // RestArea
		return "img/aire-de-repos.png"
	default: //POI pas défini
		return "img/bugs.png"
	}
}
