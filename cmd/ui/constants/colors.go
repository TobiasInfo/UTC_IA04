package constants

import "image/color"

// Zone colors with slight transparency for better visibility
var (
	EntranceZoneColor = color.RGBA{135, 206, 235, 180} // Light blue
	MainZoneColor     = color.RGBA{144, 238, 144, 180} // Light green
	ExitZoneColor     = color.RGBA{255, 182, 193, 180} // Light pink
)

// POI colors for different types of points of interest
var (
	MedicalColor      = color.RGBA{255, 0, 0, 255}    // Red
	ChargingColor     = color.RGBA{255, 255, 0, 255}  // Yellow
	ToiletColor       = color.RGBA{128, 128, 128, 255} // Gray
	DrinkStandColor   = color.RGBA{0, 191, 255, 255}  // Deep sky blue
	FoodStandColor    = color.RGBA{255, 165, 0, 255}  // Orange
	MainStageColor    = color.RGBA{148, 0, 211, 255}  // Purple
	SecondaryColor    = color.RGBA{186, 85, 211, 255}  // Medium purple
	RestAreaColor     = color.RGBA{46, 139, 87, 255}   // Sea green
)

// Zone width ratios (total should be 1.0)
const (
	EntranceZoneRatio = 0.2
	MainZoneRatio     = 0.6
	ExitZoneRatio     = 0.2
)