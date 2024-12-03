package models

type ZoneType int
type POIType int

const (
	EntranceZone ZoneType = iota
	MainZone
	ExitZone
)

const (
	MedicalTent POIType = iota
	ChargingStation
	Toilet
	DrinkStand
	FoodStand
	MainStage
	SecondaryStage
	RestArea
)

// ZoneConfig defines a zone in the festival
type ZoneConfig struct {
	Type    ZoneType
	StartX  int
	StartY  int
	EndX    int
	EndY    int
	MinPOIs map[POIType]int // Minimum number of each POI type required in zone
}

// FestivalConfig holds the complete festival layout configuration
type FestivalConfig struct {
	MapWidth     int
	MapHeight    int
	Zones        []ZoneConfig
	POILocations []POILocation
}

// POILocation defines where a POI should be placed
type POILocation struct {
	Type     POIType
	Position Position
	Capacity int // How many people/drones can use this POI at once
}
