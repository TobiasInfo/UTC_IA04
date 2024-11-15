package simulation

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"math/rand"
)

// SurveillanceDrone represents a drone in the simulation
type SurveillanceDrone struct {
	ID                      int
	Position                models.Position
	Battery                 float64
	DetectionPrecisionFunc  func(x int) []int
	ReportedZonesByCentrale []models.Position
}

// NewSurveillanceDrone creates a new instance of SurveillanceDrone
func NewSurveillanceDrone(id int, position models.Position, battery float64, detectionFunc func(x int) []int) *SurveillanceDrone {
	return &SurveillanceDrone{
		ID:                      id,
		Position:                position,
		Battery:                 battery,
		DetectionPrecisionFunc:  detectionFunc,
		ReportedZonesByCentrale: []models.Position{},
	}
}

// Move updates the drone's position to the destination
func (d *SurveillanceDrone) Move(destination models.Position) {
	if d.Battery <= 0 {
		fmt.Printf("Drone %d cannot move. Battery is empty.\n", d.ID)
		return
	}

	fmt.Printf("Drone %d moving from %v to %v\n", d.ID, d.Position, destination)
	d.Position = destination
	d.Battery -= 1.0 // Simulate battery usage
}

// DetectIncident simulates the detection of incidents in the drone's vicinity
func (d *SurveillanceDrone) DetectIncident(radius int) map[models.Position][]int {
	detectedIncidents := make(map[models.Position][]int)

	// Simulate random detections based on the detection precision function
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			position := models.Position{
				X: d.Position.X + float64(x),
				Y: d.Position.Y + float64(y),
			}
			detectedIncidents[position] = []int{
				rand.Intn(5), // Random number of people in distress
				rand.Intn(20), // Random number of people in the zone
			}
		}
	}
	fmt.Printf("Drone %d detected incidents: %v\n", d.ID, detectedIncidents)
	return detectedIncidents
}

// Communicate sends data to a centrale or another protocol
func (d *SurveillanceDrone) Communicate(protocol string) {
	switch protocol {
	case "Centrale":
		fmt.Printf("Drone %d communicating with Centrale.\n", d.ID)
		// Add logic for communication
	default:
		fmt.Printf("Drone %d: Unknown protocol '%s'\n", d.ID, protocol)
	}
}
