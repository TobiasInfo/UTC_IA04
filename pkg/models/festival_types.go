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

type ZoneConfig struct {
	Type    ZoneType
	StartX  int
	StartY  int
	EndX    int
	EndY    int
	MinPOIs map[POIType]int 
}

type FestivalConfig struct {
	MapWidth     int
	MapHeight    int
	Zones        []ZoneConfig
	POILocations []POILocation
}

type POILocation struct {
	Type     POIType
	Position Position
	Capacity int
}
