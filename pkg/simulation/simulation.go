package simulation

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"math/rand"
)

type Simulation struct {
	Map        *Map
	DroneRange int
	MoveChan   chan models.MovementRequest

	//Debug vars
	debug     bool
	hardDebug bool
}

// Initialize the simulation with the given number of drones, crowd members, and obstacles
func NewSimulation(numDrones, numCrowdMembers, numObstacles int) *Simulation {
	s := &Simulation{
		Map:        GetMap(30, 20),
		DroneRange: 2,
		MoveChan:   make(chan models.MovementRequest),
		debug:      false,
		hardDebug:  false,
	}
	s.Initialize(numDrones, numCrowdMembers, numObstacles)
	go s.handleMovementRequests()
	return s
}

func (s *Simulation) handleMovementRequests() {
	for req := range s.MoveChan {
		if !s.Map.IsBlocked(req.NewPosition) {
			if req.MemberType != "person" && req.MemberType != "drone" {
				req.ResponseChan <- models.MovementResponse{Authorized: false}
				continue
			}

			var entity interface{}

			if req.MemberType == "drone" {
				for _, drone := range s.Map.Drones {
					if drone.ID == req.MemberID {
						entity = drone
						break
					}
				}
			}

			if req.MemberType == "person" {
				for _, person := range s.Map.Persons {
					if person.ID == req.MemberID {
						entity = person
						break
					}
				}
			}

			s.Map.MoveEntity(entity, req.NewPosition)
			req.ResponseChan <- models.MovementResponse{Authorized: true}
		} else {
			req.ResponseChan <- models.MovementResponse{Authorized: false}
		}
	}
}

func (s *Simulation) Initialize(nDrones int, nCrowd int, nObstacles int) {

	s.createDrones(nDrones)

	for i := 0; i < nCrowd; i++ {
		member := NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20, s.Map.Width, s.Map.Height, s.MoveChan)
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
			pourcentageArray[i] = 1.0 - float64(i)/float64(s.DroneRange)
		}
		return pourcentageArray
	}

	droneSeeFunction := func(d *Drone) []Person {
		// Get the current cell of the drone
		currentCell := d.Position
		rangeDrone := s.DroneRange

		Vector := models.Vector{X: currentCell.X, Y: currentCell.Y}
		_, valuesInt := Vector.GenerateCircleValues(rangeDrone)

		droneInformations := make([]Person, 0)

		probs := d.DetectionPrecisionFunc()

		if s.hardDebug {
			fmt.Println("Detection probabilities & Int values")
			fmt.Println(valuesInt)
			fmt.Println(probs)
		}

		// Get The crowd members around the drone in the range of the drone
		for i := 0; i < len(valuesInt); i++ {
			position := valuesInt[i]
			// Ensure the position is within the map boundaries
			if cell, exists := s.Map.Cells[position]; exists {
				distance := currentCell.CalculateDistance(position)
				if s.debug {
					fmt.Printf("Distance : %.2f\n", distance)
				}
				if !s.Map.IsBlocked(position) {
					if s.hardDebug {
						fmt.Printf("1- Accessed : %.2f, %.2f \n", valuesInt[i].X, valuesInt[i].Y)
						fmt.Println("Size of the crowd members", len(cell.Persons))
					}
					for _, member := range cell.Persons {
						if member.Position.X != position.X || member.Position.Y != position.Y {
							fmt.Printf("ATTENTION -- ACCES MEMBRE ET CELLULE NON MEMBRE --- MEMBRE : %.2f, %.2f --- CELLEULE %.2f, %.2f \n", member.Position.X, member.Position.Y, position.X, position.Y)
						}
						if rand.Float64() < probs[int(distance)] {
							fmt.Printf("Drone %d (%.2f, %.2f) sees crowd member %d (%.2f, %.2f) at distance %.2f\n", d.ID, d.Position.X, d.Position.Y, member.ID, member.Position.X, member.Position.Y, distance)
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
			for _, member := range cell.Persons {
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
		for _, member := range cell.Persons {
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
		for _, member := range cell.Persons {
			member.Myturn()
		}
	}
}

func (s *Simulation) UpdateCrowdSize(newSize int) {
	currentSize := 0
	for _, cell := range s.Map.Cells {
		currentSize += len(cell.Persons)
	}

	if newSize > currentSize {
		for i := currentSize; i < newSize; i++ {
			member := NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20, s.Map.Width, s.Map.Height, s.MoveChan)
			s.Map.AddCrowdMember(member)
		}
	} else if newSize < currentSize {
		for _, cell := range s.Map.Cells {
			for len(cell.Persons) > 0 && currentSize > newSize {
				cell.Persons = cell.Persons[:len(cell.Persons)-1]
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
		for _, member := range cell.Persons {
			if member.InDistress {
				count++
			}
		}
	}
	return count
}
