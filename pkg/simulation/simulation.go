package simulation

import (
	"UTC_IA04/pkg/entities/drones"
	"UTC_IA04/pkg/entities/obstacles"
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"math/rand"
	"sync"
)

type Simulation struct {
	Map            *Map
	DroneRange     int
	MoveChan       chan models.MovementRequest
	DeadChan       chan models.DeadRequest
	Persons        []persons.Person
	Drones         []drones.Drone
	Obstacles      []obstacles.Obstacle
	FestivalConfig *models.FestivalConfig
	debug          bool
	hardDebug      bool
}

// Initialize the simulation with the given number of drones, crowd members, and obstacles
func NewSimulation(numDrones, numCrowdMembers, numObstacles int) *Simulation {
	s := &Simulation{
		Map:        GetMap(30, 20),
		DroneRange: 2,
		MoveChan:   make(chan models.MovementRequest),
		DeadChan:   make(chan models.DeadRequest),
		debug:      false,
		hardDebug:  false,
	}
	s.Initialize(numDrones, numCrowdMembers, numObstacles)
	go s.handleMovementRequests()
	go s.handleDeadPerson()
	return s
}

func (s *Simulation) handleMovementRequests() {
	for req := range s.MoveChan {
		if !s.Map.IsBlocked(req.NewPosition) {
			if req.MemberType != "persons" && req.MemberType != "drones" {
				req.ResponseChan <- models.MovementResponse{Authorized: false}
				continue
			}

			var entity interface{}

			if req.MemberType == "drones" {
				for _, drone := range s.Map.Drones {
					if drone.ID == req.MemberID {
						entity = drone
						break
					}
				}
			}

			if req.MemberType == "persons" {
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

func (s *Simulation) handleDeadPerson() {
	for req := range s.DeadChan {
		if req.MemberType != "persons" {
			req.ResponseChan <- models.DeadResponse{Authorized: false}
			continue
		}

		var entity interface{}

		for _, person := range s.Map.Persons {
			if person.ID == req.MemberID {
				entity = person
				break
			}
		}

		s.Map.RemoveEntity(entity)
		req.ResponseChan <- models.DeadResponse{Authorized: true}
	}
}

func (s *Simulation) Initialize(nDrones int, nCrowd int, nObstacles int) {
	// Try to load festival configuration
	configPath := "configs/festival_layout.json"
	config, err := LoadFestivalConfig(configPath)
	if err != nil {
		fmt.Printf("Warning: Could not load festival config from %s: %v\n", configPath, err)
		fmt.Println("Using default obstacle initialization...")
		// Create default obstacles if no config is available
		for i := 0; i < nObstacles; i++ {
			randomPOIType := models.POIType(rand.Intn(8))
			defaultCapacity := 10
			switch randomPOIType {
			case models.MedicalTent:
				defaultCapacity = 15
			case models.ChargingStation:
				defaultCapacity = 5
			case models.MainStage:
				defaultCapacity = 100
			}

			obstacle := obstacles.NewObstacle(
				i,
				models.Position{
					X: float64(rand.Intn(s.Map.Width)),
					Y: float64(rand.Intn(s.Map.Height)),
				},
				randomPOIType,
				defaultCapacity,
			)
			s.Obstacles = append(s.Obstacles, obstacle)
			s.Map.AddObstacle(&s.Obstacles[len(s.Obstacles)-1])
		}
	} else {
		fmt.Println("Successfully loaded festival configuration!")
		fmt.Printf("Found %d POI locations in config\n", len(config.POILocations))
		s.FestivalConfig = config
		fmt.Println("Applying festival configuration...")
		err = s.Map.ApplyFestivalConfig(config)
		if err != nil {
			fmt.Printf("Error applying festival config: %v\nFalling back to default initialization\n", err)
			// Fall back to default initialization
			s.Initialize(nDrones, nCrowd, nObstacles)
			return
		}
		fmt.Println("Successfully applied festival configuration")
	}

	// Initialize drones and crowd members
	s.createDrones(nDrones)

	for i := 0; i < nCrowd; i++ {
		member := persons.NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20, s.Map.Width, s.Map.Height, s.MoveChan, s.DeadChan)
		s.Persons = append(s.Persons, member)
		s.Map.AddCrowdMember(&s.Persons[len(s.Persons)-1])
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

	droneSeeFunction := func(d *drones.Drone) []*persons.Person {
		// Get the current cell of the drones
		currentCell := d.Position
		rangeDrone := s.DroneRange

		Vector := models.Vector{X: currentCell.X, Y: currentCell.Y}
		_, valuesInt := Vector.GenerateCircleValues(rangeDrone)

		droneInformations := make([]*persons.Person, 0)

		probs := d.DetectionPrecisionFunc()

		if s.hardDebug {
			fmt.Println("Detection probabilities & Int values")
			fmt.Println(valuesInt)
			fmt.Println(probs)
		}

		// Get The crowd members around the drones in the range of the drones
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
							fmt.Printf("ATTENTION -- ACCES MEMBRE (%d) ET CELLULE NON MEMBRE --- MEMBRE : %.2f, %.2f --- CELLEULE %.2f, %.2f \n", member.ID, member.Position.X, member.Position.Y, position.X, position.Y)
						}
						if rand.Float64() < probs[int(distance)] {
							droneInformations = append(droneInformations, member)
						}
					}
				}
			}
		}
		return droneInformations
	}

	for i := 0; i < n; i++ {
		d := drones.NewSurveillanceDrone(i, models.Position{X: 0, Y: 0}, 100.0, detectionFunc, droneSeeFunction, s.MoveChan)
		s.Drones = append(s.Drones, d)
		s.Map.AddDrone(&s.Drones[len(s.Drones)-1])
	}
}

func (s *Simulation) Update() {
	fmt.Println("New Tick")
	var wg sync.WaitGroup

	updatedPersons := make(map[int]struct{})
	updatedDrones := make(map[int]struct{})

	// Create indexes slice and shuffle it
	indexes := make([]int, len(s.Persons))
	for i := range indexes {
		indexes[i] = i
	}
	rand.Shuffle(len(indexes), func(i, j int) {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	})

	// Update persons in random order
	for _, idx := range indexes {
		if _, exists := updatedPersons[s.Persons[idx].ID]; !exists {
			updatedPersons[s.Persons[idx].ID] = struct{}{}
			wg.Add(1)
			go func(p *persons.Person) {
				defer wg.Done()
				p.Myturn()
			}(&s.Persons[idx])
		}
	}
	wg.Wait()

	// Shuffle and update drones
	indexes = make([]int, len(s.Drones))
	for i := range indexes {
		indexes[i] = i
	}
	rand.Shuffle(len(indexes), func(i, j int) {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	})

	for _, idx := range indexes {
		if _, exists := updatedDrones[s.Drones[idx].ID]; !exists {
			wg.Add(1)
			updatedDrones[s.Drones[idx].ID] = struct{}{}
			go func(d *drones.Drone) {
				defer wg.Done()
				d.Myturn()
			}(&s.Drones[idx])
		}
	}
	wg.Wait()

	for index, cell := range s.Map.Cells {
		for _, member := range cell.Persons {
			fmt.Printf("%v - Person %d is at position (%.2f, %.2f) -- Current Cell = (%.2f, %.2f) \n", index, member.ID, member.Position.X, member.Position.Y, cell.Position.X, cell.Position.Y)
		}
	}

	fmt.Println("End of the tick")
}

func (s *Simulation) UpdateCrowdSize(newSize int) {
	currentSize := 0
	for _, cell := range s.Map.Cells {
		currentSize += len(cell.Persons)
	}

	if newSize > currentSize {
		for i := currentSize; i < newSize; i++ {
			member := persons.NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20, s.Map.Width, s.Map.Height, s.MoveChan, s.DeadChan)
			s.Persons = append(s.Persons, member)
			s.Map.AddCrowdMember(&s.Persons[len(s.Persons)-1])
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
