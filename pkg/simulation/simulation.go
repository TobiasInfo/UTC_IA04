// simulation.go
// This is the complete file to replace your current simulation.go

package simulation

import (
	"UTC_IA04/pkg/entities/drones"
	"UTC_IA04/pkg/entities/obstacles"
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"math/rand"
	"sync"
	"time"
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
	currentTick    int
	festivalTime   *FestivalTime
	poiMap         map[models.POIType][]models.Position
	mu             sync.RWMutex
}

func NewSimulation(numDrones, numCrowdMembers, numObstacles int) *Simulation {
	s := &Simulation{
		Map:          GetMap(30, 20),
		DroneRange:   2,
		MoveChan:     make(chan models.MovementRequest),
		DeadChan:     make(chan models.DeadRequest),
		debug:        false,
		hardDebug:    false,
		currentTick:  0,
		festivalTime: NewFestivalTime(),
		poiMap:       make(map[models.POIType][]models.Position),
	}
	s.Initialize(numDrones, numCrowdMembers, numObstacles)
	go s.handleMovementRequests()
	go s.handleDeadPerson()
	return s
}

func (s *Simulation) handleMovementRequests() {
	for req := range s.MoveChan {
		//fmt.Printf("Received movement request for %s %d to position (%.2f, %.2f)\n")
		if !s.Map.IsBlocked(req.NewPosition) {
			if req.MemberType != "persons" && req.MemberType != "drone" {
				req.ResponseChan <- models.MovementResponse{Authorized: false, Reason: "Invalid member type"}
				continue
			}

			var entity interface{}
			s.mu.RLock()
			if req.MemberType == "drone" {
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
			s.mu.RUnlock()

			if entity != nil {
				s.mu.Lock()
				s.Map.MoveEntity(entity, req.NewPosition)
				s.mu.Unlock()
				req.ResponseChan <- models.MovementResponse{Authorized: true, Reason: "OK"}
			} else {
				req.ResponseChan <- models.MovementResponse{Authorized: false, Reason: "Member not found"}
			}
		} else {
			req.ResponseChan <- models.MovementResponse{Authorized: false, Reason: "Position is blocked"}
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
		s.mu.RLock()
		for _, person := range s.Map.Persons {
			if person.ID == req.MemberID {
				entity = person
				break
			}
		}
		s.mu.RUnlock()

		if entity != nil {
			s.mu.Lock()
			s.Map.RemoveEntity(entity)
			s.mu.Unlock()
			req.ResponseChan <- models.DeadResponse{Authorized: true}
		} else {
			req.ResponseChan <- models.DeadResponse{Authorized: false}
		}
	}
}

func (s *Simulation) Initialize(nDrones int, nCrowd int, nObstacles int) {
	configPath := "configs/festival_layout.json"
	config, err := LoadFestivalConfig(configPath)
	if err != nil {
		fmt.Printf("Warning: Could not load festival config from %s: %v\n", configPath, err)
		fmt.Println("Using default obstacle initialization...")
		s.initializeDefaultObstacles(nObstacles)
	} else {
		fmt.Println("Successfully loaded festival configuration!")
		s.FestivalConfig = config
		err = s.Map.ApplyFestivalConfig(config)
		if err != nil {
			fmt.Printf("Error applying festival config: %v\nFalling back to default initialization\n", err)
			s.initializeDefaultObstacles(nObstacles)
		} else {
			fmt.Println("Successfully applied festival configuration")
			s.buildPOIMap()
		}
	}

	s.createDrones(nDrones)
	s.createInitialCrowd(nCrowd)
	s.festivalTime.Start()
}

func (s *Simulation) buildPOIMap() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.poiMap = make(map[models.POIType][]models.Position)
	for _, obstacle := range s.Map.Obstacles {
		poiType := obstacle.GetPOIType()
		s.poiMap[poiType] = append(s.poiMap[poiType], obstacle.Position)
	}
}

func (s *Simulation) getNearestPOI(personPos models.Position, poiType models.POIType) *models.Position {
	s.mu.RLock()
	defer s.mu.RUnlock()

	positions, exists := s.poiMap[poiType]
	if !exists || len(positions) == 0 {
		return nil
	}

	var nearest *models.Position
	minDist := float64(s.Map.Width + s.Map.Height)

	for _, pos := range positions {
		dist := personPos.CalculateDistance(pos)
		if dist < minDist {
			minDist = dist
			posCopy := pos
			nearest = &posCopy
		}
	}

	return nearest
}

func (s *Simulation) initializeDefaultObstacles(nObstacles int) {
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
	s.buildPOIMap()
}

func (s *Simulation) createDrones(n int) {
	detectionFunc := func() []float64 {
		pourcentageArray := make([]float64, s.DroneRange+1)
		for i := 0; i < s.DroneRange+1; i++ {
			pourcentageArray[i] = 1.0 - float64(i)/float64(s.DroneRange)
		}
		return pourcentageArray
	}

	droneSeeFunction := func(d *drones.Drone) []*persons.Person {
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

		for i := 0; i < len(valuesInt); i++ {
			position := valuesInt[i]
			if cell, exists := s.Map.Cells[position]; exists {
				distance := currentCell.CalculateDistance(position)
				if s.debug {
					fmt.Printf("Distance : %.2f\n", distance)
				}
				if !s.Map.IsBlocked(position) {
					for _, member := range cell.Persons {
						if member.Position.X != position.X || member.Position.Y != position.Y {
							fmt.Printf("Warning: Member position mismatch - ID: %d, Position: (%.2f, %.2f), Cell: (%.2f, %.2f)\n",
								member.ID, member.Position.X, member.Position.Y, position.X, position.Y)
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

func (s *Simulation) createInitialCrowd(n int) {
	for i := 0; i < n; i++ {
		member := persons.NewCrowdMember(i,
			models.Position{X: 0, Y: float64(rand.Intn(s.Map.Height))},
			0.001, 20, s.Map.Width, s.Map.Height, s.MoveChan, s.DeadChan)
		s.Persons = append(s.Persons, member)
		s.Map.AddCrowdMember(&s.Persons[len(s.Persons)-1])
	}
}

func (s *Simulation) Update() {
	if s.festivalTime.IsEventEnded() {
		return
	}

	fmt.Println("New Tick")
	s.currentTick++
	var wg sync.WaitGroup

	// Update drones every tick
	updatedDrones := make(map[int]struct{})
	droneIndexes := make([]int, len(s.Drones))
	for i := range droneIndexes {
		droneIndexes[i] = i
	}
	rand.Shuffle(len(droneIndexes), func(i, j int) {
		droneIndexes[i], droneIndexes[j] = droneIndexes[j], droneIndexes[i]
	})

	for _, idx := range droneIndexes {
		if _, exists := updatedDrones[s.Drones[idx].ID]; !exists {
			wg.Add(1)
			updatedDrones[s.Drones[idx].ID] = struct{}{}
			go func(d *drones.Drone) {
				defer wg.Done()
				d.Myturn()
			}(&s.Drones[idx])
		}
	}

	// Update persons every 10 ticks
	if s.currentTick%1 == 0 {
		fmt.Printf("\n=== TICK %d: PEOPLE SHOULD MOVE ===\n", s.currentTick)
		updatedPersons := make(map[int]struct{})
		personIndexes := make([]int, len(s.Persons))
		for i := range personIndexes {
			personIndexes[i] = i
		}
		rand.Shuffle(len(personIndexes), func(i, j int) {
			personIndexes[i], personIndexes[j] = personIndexes[j], personIndexes[i]
		})

		for _, idx := range personIndexes {
			if _, exists := updatedPersons[s.Persons[idx].ID]; !exists {
				updatedPersons[s.Persons[idx].ID] = struct{}{}
				wg.Add(1)
				go func(p *persons.Person) {
					defer wg.Done()
					if p.CurrentPOI != nil && p.TargetPOIPosition == nil {
						if pos := s.getNearestPOI(p.Position, *p.CurrentPOI); pos != nil {
							p.SetTargetPOI(*p.CurrentPOI, *pos)
						} else {
							p.CurrentPOI = nil
						}
					}
					p.Myturn()
				}(&s.Persons[idx])
			}
		}
	}

	wg.Wait()

	if s.debug {
		for index, cell := range s.Map.Cells {
			for _, member := range cell.Persons {
				fmt.Printf("%v - Person %d is at position (%.2f, %.2f) -- Current Cell = (%.2f, %.2f) \n",
					index, member.ID, member.Position.X, member.Position.Y, cell.Position.X, cell.Position.Y)
			}
		}
	}

	fmt.Println("End of the tick")
}

func (s *Simulation) UpdateCrowdSize(newSize int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentSize := 0
	for _, cell := range s.Map.Cells {
		currentSize += len(cell.Persons)
	}

	if newSize > currentSize {
		for i := currentSize; i < newSize; i++ {
			member := persons.NewCrowdMember(i,
				models.Position{X: 0, Y: float64(rand.Intn(s.Map.Height))},
				0.001, 20, s.Map.Width, s.Map.Height, s.MoveChan, s.DeadChan)
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
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *Simulation) GetFestivalTime() *FestivalTime {
	return s.festivalTime
}

func (s *Simulation) SetDebug(debug bool) {
	s.debug = debug
}

func (s *Simulation) SetHardDebug(debug bool) {
	s.hardDebug = debug
}

func (s *Simulation) GetCurrentTick() int {
	return s.currentTick
}

func (s *Simulation) GetStatistics() SimulationStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalPeople := len(s.Persons)
	peopleInDistress := s.CountCrowdMembersInDistress()
	activeDrones := len(s.Drones)

	zoneStats := make(map[string]int)
	zoneStats["entrance"] = 0
	zoneStats["main"] = 0
	zoneStats["exit"] = 0

	for _, person := range s.Persons {
		zone := person.GetCurrentZone()
		zoneStats[zone]++
	}

	return SimulationStatistics{
		TotalPeople:      totalPeople,
		PeopleInDistress: peopleInDistress,
		ActiveDrones:     activeDrones,
		ZoneStatistics:   zoneStats,
		CurrentPhase:     s.festivalTime.GetPhase(),
		ElapsedTime:      s.festivalTime.GetElapsedTime(),
		RemainingTime:    s.festivalTime.GetRemainingTime(),
	}
}

type SimulationStatistics struct {
	TotalPeople      int
	PeopleInDistress int
	ActiveDrones     int
	ZoneStatistics   map[string]int
	CurrentPhase     string
	ElapsedTime      time.Duration
	RemainingTime    time.Duration
}

func (s *Simulation) GetPersonsInRange(center models.Position, radius float64) []*persons.Person {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*persons.Person
	for _, cell := range s.Map.Cells {
		dist := center.CalculateDistance(cell.Position)
		if dist <= radius {
			result = append(result, cell.Persons...)
		}
	}
	return result
}

func (s *Simulation) GetAvailablePOIs() map[models.POIType][]models.Position {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[models.POIType][]models.Position)
	for poiType, positions := range s.poiMap {
		result[poiType] = make([]models.Position, len(positions))
		copy(result[poiType], positions)
	}
	return result
}
