package persons

import (
	"UTC_IA04/pkg/models"
)

type ProfileType int

const (
	Adventurous ProfileType = iota
	Cautious
	Social
	Independent
)

type PersonProfile struct {
	Type                   ProfileType
	BaseMovementSpeed      float64
	CrowdFollowingTendency float64
	POIInterestRates       map[models.POIType]float64
	StaminaLevel           float64 // 0.0 to 1.0
	PersonalSpace          float64 // Minimum desired distance from others
	RestThreshold          float64 // When stamina drops below this, person seeks rest
	MalaiseResistance      float64 // Base resistance to malaise
}

func NewPersonProfile(profileType ProfileType) PersonProfile {
	profile := PersonProfile{
		Type:             profileType,
		StaminaLevel:     1.0,
		POIInterestRates: make(map[models.POIType]float64),
	}

	switch profileType {
	case Adventurous:
		profile.BaseMovementSpeed = 1.2
		profile.CrowdFollowingTendency = 0.3
		profile.PersonalSpace = 1.0
		profile.RestThreshold = 0.3
		profile.MalaiseResistance = 0.8
		profile.POIInterestRates[models.MainStage] = 0.9
		profile.POIInterestRates[models.SecondaryStage] = 0.8
		profile.POIInterestRates[models.FoodStand] = 0.6
		profile.POIInterestRates[models.DrinkStand] = 0.7

	case Cautious:
		profile.BaseMovementSpeed = 0.8
		profile.CrowdFollowingTendency = 0.5
		profile.PersonalSpace = 2.0
		profile.RestThreshold = 0.6
		profile.MalaiseResistance = 0.6
		profile.POIInterestRates[models.RestArea] = 0.9
		profile.POIInterestRates[models.Toilet] = 0.8
		profile.POIInterestRates[models.DrinkStand] = 0.7
		profile.POIInterestRates[models.FoodStand] = 0.6

	case Social:
		profile.BaseMovementSpeed = 1.0
		profile.CrowdFollowingTendency = 0.8
		profile.PersonalSpace = 0.8
		profile.RestThreshold = 0.4
		profile.MalaiseResistance = 0.7
		profile.POIInterestRates[models.FoodStand] = 0.9
		profile.POIInterestRates[models.DrinkStand] = 0.9
		profile.POIInterestRates[models.MainStage] = 0.7
		profile.POIInterestRates[models.SecondaryStage] = 0.6

	case Independent:
		profile.BaseMovementSpeed = 1.1
		profile.CrowdFollowingTendency = 0.2
		profile.PersonalSpace = 1.5
		profile.RestThreshold = 0.5
		profile.MalaiseResistance = 0.9
		profile.POIInterestRates[models.RestArea] = 0.6
		profile.POIInterestRates[models.FoodStand] = 0.6
		profile.POIInterestRates[models.MainStage] = 0.6
		profile.POIInterestRates[models.DrinkStand] = 0.6
		profile.POIInterestRates[models.SecondaryStage] = 0.6
	}

	// Set default values for common POIs if not already set
	for _, poiType := range []models.POIType{
		models.Toilet,
		models.DrinkStand,
		models.RestArea,
	} {
		if _, exists := profile.POIInterestRates[poiType]; !exists {
			profile.POIInterestRates[poiType] = 0.5
		}
	}

	return profile
}
