package simulation

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"math/rand"
)

type Simulation struct {
	Map        *Map
	DroneRange int
}

// Initialize the simulation with the given number of drones, crowd members, and obstacles
func NewSimulation(numDrones, numCrowdMembers, numObstacles int) *Simulation {
	s := &Simulation{
		Map:        GetMap(30, 20),
		DroneRange: 10,
	}
	s.Initialize(numDrones, numCrowdMembers, numObstacles)
	return s
}

func (s *Simulation) Initialize(nDrones int, nCrowd int, nObstacles int) {

	s.createDrones(nDrones)

	for i := 0; i < nCrowd; i++ {
		member := NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20)
		s.Map.AddCrowdMember(member)
	}

	for i := 0; i < nObstacles; i++ {
		obstacle := NewObstacle(models.Position{X: 10, Y: 10})
		s.Map.AddObstacle(obstacle)
	}
}

func (s *Simulation) createDrones(n int) {
	// Example detection function
	detectionFunc := func() []float64 {
		pourcentageArray := make([]float64, s.DroneRange+1)
		for i := 0; i < s.DroneRange+1; i++ {
			pourcentageArray[i] = 1.0 - float64(i)/float64(s.DroneRange+1)
		}
		return pourcentageArray
	}

	droneSeeFunction := func(d *SurveillanceDrone) []CrowdMember {
		// Get the current cell of the drone
		currentCell := d.Position
		rangeDrone := s.DroneRange

		Vector := models.Vector{X: currentCell.X, Y: currentCell.Y}
		_, valuesInt := Vector.GenerateCircleValues(rangeDrone)

		droneInformations := make([]CrowdMember, 0)

		probs := d.DetectionPrecisionFunc()
		fmt.Println(probs)

		for i := range s.Map.Cells {
			if len(s.Map.Cells[i].Drones) > 0 {
				fmt.Printf("%.2f, %.2f : %d Citizens\n", s.Map.Cells[i].Position.X, s.Map.Cells[i].Position.Y, len(s.Map.Cells[i].CrowdMembers))
			}
			if len(s.Map.Cells[i].CrowdMembers) > 0 {
				fmt.Printf("%.2f, %.2f : %d Citizens\n", s.Map.Cells[i].Position.X, s.Map.Cells[i].Position.Y, len(s.Map.Cells[i].CrowdMembers))
			}
		}

		// Get The crowd members around the drone in the range of the drone
		for i := 0; i < len(valuesInt); i++ {
			position := models.Position{X: valuesInt[i].X, Y: valuesInt[i].Y}
			// Ensure the position is within the map boundaries
			if cell, exists := s.Map.Cells[position]; exists {
				if !s.Map.IsBlocked(position) {
					fmt.Printf("1- Accessed : %.2f, %.2f \n", valuesInt[i].X, valuesInt[i].Y)
					fmt.Println("Size of the crowd members", len(cell.CrowdMembers))
					for _, member := range cell.CrowdMembers {
						distance := currentCell.CalculateDistance(member.Position)
						fmt.Println("Distance between drone and member", distance)
						fmt.Printf("2- Crowd member : %.2f, %.2f \n", member.Position.X, member.Position.Y)
						if rand.Float64() < probs[int(distance)] {
							fmt.Printf("Drone %d sees crowd member %d at distance %.2f\n", d.ID, member.ID, distance)
							fmt.Printf("Drone %d sees crowd member %d at distance %.2f\n", d.ID, member.ID, distance)
							droneInformations = append(droneInformations, *member)
						}
					}
				}
			}
		}
		return droneInformations
	}

	for i := 0; i < n; i++ {
		drone := NewSurveillanceDrone(i, models.Position{X: 0, Y: 0}, 100.0, detectionFunc, droneSeeFunction)
		s.Map.AddDrone(drone)
	}
}

func (s *Simulation) StartSimulation(numberIteration int) {
	fmt.Println("Simulation started")
	// Logic to initialize and run the simulation

	s.Initialize(5, 40, 3)

	for tick := 0; tick < numberIteration; tick++ {
		// Update drones
		for _, cell := range s.Map.Cells {
			for _, drone := range cell.Drones {
				drone.Myturn()
			}
		}

		// Update crowd members
		for _, cell := range s.Map.Cells {
			for _, member := range cell.CrowdMembers {
				member.Myturn()
			}
		}
	}

	for _, cell := range s.Map.Cells {
		for _, drone := range cell.Drones {
			fmt.Printf("Position du drone %d : X = %f , Y = %f\n", drone.ID, drone.Position.X, drone.Position.Y)
		}
	}

	for _, cell := range s.Map.Cells {
		for _, member := range cell.CrowdMembers {
			fmt.Printf("Position du membre %d : X = %f , Y = %f\n", member.ID, member.Position.X, member.Position.Y)
		}
	}
}

// Update the simulation state
func (s *Simulation) Update() {
	// Update drones
	for _, cell := range s.Map.Cells {
		for _, drone := range cell.Drones {
			drone.Myturn()
		}
	}

	// Update crowd members
	for _, cell := range s.Map.Cells {
		for _, member := range cell.CrowdMembers {
			member.Myturn()
		}
	}
}

func (s *Simulation) UpdateCrowdSize(newSize int) {
	currentSize := 0
	for _, cell := range s.Map.Cells {
		currentSize += len(cell.CrowdMembers)
	}

	if newSize > currentSize {
		for i := currentSize; i < newSize; i++ {
			member := NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20)
			s.Map.AddCrowdMember(member)
		}
	} else if newSize < currentSize {
		for _, cell := range s.Map.Cells {
			for len(cell.CrowdMembers) > 0 && currentSize > newSize {
				cell.CrowdMembers = cell.CrowdMembers[:len(cell.CrowdMembers)-1]
				currentSize--
			}
		}
	}
}

func (s *Simulation) UpdateDroneSize(newSize int) {
	currentSize := 0
	for _, cell := range s.Map.Cells {
		currentSize += len(cell.Drones)
	}

	if newSize > currentSize {
		for i := currentSize; i < newSize; i++ {
			s.createDrones(1)
		}
	} else if newSize < currentSize {
		for _, cell := range s.Map.Cells {
			for len(cell.Drones) > 0 && currentSize > newSize {
				cell.Drones = cell.Drones[:len(cell.Drones)-1]
				currentSize--
			}
		}
	}
}

func (s *Simulation) CountCrowdMembersInDistress() int {
	count := 0
	for _, cell := range s.Map.Cells {
		for _, member := range cell.CrowdMembers {
			if member.InDistress {
				count++
			}
		}
	}
	return count
}
