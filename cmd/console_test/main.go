package main

import (
	"UTC_IA04/pkg/models"
	"UTC_IA04/pkg/simulation"
	"fmt"
)

func main() {
	// Get the singleton instance of the map
	mapInstance := simulation.GetMap(10, 10)

	// Add a drone to the map
	drone := &simulation.SurveillanceDrone{Position: models.Position{X: 5, Y: 5}}
	mapInstance.AddDrone(drone)

	// Add a crowd member to the map
	crowdMember := &simulation.CrowdMember{ID: 1, Position: models.Position{X: 5, Y: 5}}
	mapInstance.AddCrowdMember(crowdMember)

	// Move a drone
	newPosition := models.Position{X: 6, Y: 6}
	mapInstance.MoveEntity(drone, newPosition)

	// Move a crowd member
	newPosition2 := models.Position{X: 6, Y: 6}
	mapInstance.MoveEntity(crowdMember, newPosition2)

	// Check if the map position is blocked
	if mapInstance.IsBlocked(models.Position{X: 6, Y: 6}) {
		fmt.Println("Position (6,6) is blocked")
	} else {
		fmt.Println("Position (6,6) is free")
	}
	drone2 := &simulation.SurveillanceDrone{Position: models.Position{X: 6, Y: 6}}
	mapInstance.AddDrone(drone2)
	// Display how many drones are in the cell
	fmt.Printf("Number of drones in cell (6,6): %d\n", len(mapInstance.GetDrones(models.Position{X: 6, Y: 6})))
}
