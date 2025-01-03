package simulation

import (
	"UTC_IA04/pkg/entities/drones"
	"UTC_IA04/pkg/entities/obstacles"
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/entities/rescue"
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	DEFAULT_DISTRESS_PROBABILITY = 0.1
	DEFAULT_PROTOCOL_MODE        = 3
)

type FestivalState int

const (
	Active FestivalState = iota
	Ended
)

type Simulation struct {
	Map                        *Map
	DroneSeeRange              int
	DroneCommRange             int
	MoveChan                   chan models.MovementRequest
	DeadChan                   chan models.DeadRequest
	ExitChan                   chan models.ExitRequest
	ChargingChan               chan models.ChargingRequest
	MedicalDeliveryChan        chan models.MedicalDeliveryRequest
	SavePersonChan             chan models.SavePersonRequest
	SavePeopleByRescuerChan    chan models.RescuePeopleRequest
	Persons                    []persons.Person
	Drones                     []drones.Drone
	Obstacles                  []obstacles.Obstacle
	FestivalConfig             *models.FestivalConfig
	debug                      bool
	hardDebug                  bool
	currentTick                int
	festivalTotalTicks         int
	DefaultDistressProbability float64
	festivalTime               *FestivalTime
	poiMap                     map[models.POIType][]models.Position
	mu                         sync.RWMutex
	treatedCases               int
	deadCases                  int
	RescuePoints               map[models.Position]*rescue.RescuePoint
	FestivalState              FestivalState
}

type SimulationStatistics struct {
	TotalPeople     int
	InDistress      int
	CasesTreated    int
	CasesDead       int
	AverageBattery  float64
	AverageCoverage float64
	PeopleDensity   models.DensityGrid
	DroneNetwork    models.DroneNetwork
}

// CIMITIERE DES PERSONNES MORTES EN (-10, -10)

func NewSimulation(numDrones, numCrowdMembers, numObstacles int) *Simulation {
	s := &Simulation{
		Map:                     GetMap(30, 20),
		DroneSeeRange:           4,
		DroneCommRange:          6,
		MoveChan:                make(chan models.MovementRequest),
		DeadChan:                make(chan models.DeadRequest),
		ExitChan:                make(chan models.ExitRequest),
		ChargingChan:            make(chan models.ChargingRequest),
		SavePersonChan:          make(chan models.SavePersonRequest),
		debug:                   false,
		hardDebug:               false,
		currentTick:             0,
		festivalTotalTicks:      1000,
		deadCases:               0,
		festivalTime:            NewFestivalTime(),
		poiMap:                  make(map[models.POIType][]models.Position),
		MedicalDeliveryChan:     make(chan models.MedicalDeliveryRequest),
		SavePeopleByRescuerChan: make(chan models.RescuePeopleRequest),
		RescuePoints:            make(map[models.Position]*rescue.RescuePoint),
		FestivalState:           Active,
	}
	s.Initialize(numDrones, numCrowdMembers, numObstacles)
	go s.handleMovementRequests()
	go s.handleDeadPerson()
	go s.handleExitRequest()
	go s.handleChargingRequests()
	go s.handleMedicalDelivery()
	go s.handleSavePerson()
	go s.handleSavePersonByRescuer()
	return s
}

func (s *Simulation) handleSavePersonByRescuer() {
	for req := range s.SavePeopleByRescuerChan {
		// Trouver le drone correspondant au RescuerID
		var rp *rescue.RescuePoint
		for i := range s.RescuePoints {
			if s.RescuePoints[i].ID == req.RescuePointID {
				rp = s.RescuePoints[i]
				break
			}
		}

		if rp == nil {
			req.ResponseChan <- models.RescuePeopleResponse{
				Authorized: false,
				Reason:     "Rescue point not found",
			}
			continue
		}

		rescuer, exists := rp.Rescuers[req.RescuerID]

		if !exists {
			req.ResponseChan <- models.RescuePeopleResponse{
				Authorized: false,
				Reason:     "Rescuer not found",
			}
			continue
		}

		var personToSave *persons.Person
		for i := range s.Persons {
			if s.Persons[i].ID == req.PersonID {
				personToSave = &s.Persons[i]
				break
			}
		}

		if personToSave == nil {
			req.ResponseChan <- models.RescuePeopleResponse{
				Authorized: false,
				Reason:     "Person not found",
			}
			continue
		}

		if rescuer.Person.Position.CalculateDistance(personToSave.Position) > 1 {
			req.ResponseChan <- models.RescuePeopleResponse{
				Authorized: false,
				Reason:     "Rescuer not at person's position",
			}
			continue
		}

		personToSave.InDistress = false
		s.mu.Lock()
		s.treatedCases++
		s.mu.Unlock()
		personToSave.CurrentDistressDuration = 0
		personToSave.State.CurrentState = persons.Resting
		personToSave.Profile.StaminaLevel = 1.0
		personToSave.State.UpdateState(personToSave)

		req.ResponseChan <- models.RescuePeopleResponse{
			Authorized: true,
			Reason:     "Person saved",
		}
	}
}

func (s *Simulation) handleSavePerson() {
	for req := range s.SavePersonChan {
		authorized := false
		for _, drone := range s.Drones {
			if drone.ID == req.DroneID {
				if !drone.HasMedicalGear {
					break
				}
				for i := range s.Persons {
					person := &s.Persons[i]
					if person.ID == req.PersonID {
						if math.Round(person.Position.X) == drone.Position.X && math.Round(person.Position.Y) == drone.Position.Y {
							if person.InDistress {
								authorized = true
								person.InDistress = false
								s.mu.Lock()
								s.treatedCases++
								s.mu.Unlock()
								person.CurrentDistressDuration = 0
								person.State.CurrentState = 2
								person.Profile.StaminaLevel = 1.0
								person.State.UpdateState(person)
								if s.debug {
									fmt.Printf("Person %d has been healed!\n", person.ID)
								}
								break
							}
						}
					}
				}
			}
		}
		req.ResponseChan <- models.SavePersonResponse{
			Authorized: authorized,
			Reason:     map[bool]string{true: "Person saved", false: "Save failed"}[authorized],
		}
	}
}

func (s *Simulation) handleMedicalDelivery() {
	for req := range s.MedicalDeliveryChan {
		// s.mu.Lock()
		authorized := false
		for _, drone := range s.Drones {
			if drone.ID == req.DroneID {
				for _, pos := range s.poiMap[models.MedicalTent] {
					if pos.X == drone.Position.X && pos.Y == drone.Position.Y {
						authorized = true
						break
					}
				}
			}
		}
		// s.mu.Unlock()
		req.ResponseChan <- models.MedicalDeliveryResponse{
			Authorized: authorized,
			Reason:     map[bool]string{true: "Medical delivered", false: "Delivery failed"}[authorized],
		}
	}
}

func (s *Simulation) handleChargingRequests() {
	for req := range s.ChargingChan {
		s.mu.RLock()
		// Check if position is actually a charging station
		isChargingStation := false
		for _, pos := range s.poiMap[models.ChargingStation] {
			if pos == req.Position {
				isChargingStation = true
				break
			}
		}
		s.mu.RUnlock()

		if isChargingStation {
			req.ResponseChan <- models.ChargingResponse{
				Authorized: true,
				Reason:     "Charging station available",
			}
		} else {
			req.ResponseChan <- models.ChargingResponse{
				Authorized: false,
				Reason:     "Not at a charging station",
			}
		}
	}
}

func (s *Simulation) handleMovementRequests() {
	for req := range s.MoveChan {
		// Vérifie si la position cible est valide sur la carte
		if req.MemberType != "persons" && req.MemberType != "drone" {
			req.ResponseChan <- models.MovementResponse{Authorized: false, Reason: "Invalid member type"}
			continue
		}

		var entity interface{}
		s.mu.RLock()
		if req.MemberType == "drone" {
			for i := range s.Drones {
				if s.Drones[i].ID == req.MemberID {
					entity = &s.Drones[i]
					break
				}
			}
		}

		if req.MemberType == "persons" {
			for i := range s.Persons {
				if s.Persons[i].ID == req.MemberID {
					entity = &s.Persons[i]
					break
				}
			}
		}
		s.mu.RUnlock()

		if entity == nil {
			req.ResponseChan <- models.MovementResponse{Authorized: false, Reason: "Member not found"}
			continue
		}

		if req.NewPosition.X < 0 || req.NewPosition.Y < 0 || req.NewPosition.X >= float64(s.Map.Width) || req.NewPosition.Y >= float64(s.Map.Height) {
			req.ResponseChan <- models.MovementResponse{Authorized: false, Reason: "Position is out of bounds"}
			continue
		}

		if req.MemberType == "drone" {
			// Pour les drones, ignorer les obstacles
			s.mu.Lock()
			s.Map.MoveEntity(entity, req.NewPosition)
			s.mu.Unlock()
			req.ResponseChan <- models.MovementResponse{Authorized: true, Reason: "Drones can move above obstacles"}
		} else {
			// Pour les personnes, vérifier les obstacles
			if !s.Map.IsBlocked(req.NewPosition) {
				s.mu.Lock()
				s.Map.MoveEntity(entity, req.NewPosition)
				s.mu.Unlock()
				req.ResponseChan <- models.MovementResponse{Authorized: true, Reason: "OK"}
			} else {
				req.ResponseChan <- models.MovementResponse{Authorized: false, Reason: "Position is blocked"}
			}
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
		for _, person := range s.Persons {
			if person.ID == req.MemberID {
				entity = &person
				break
			}
		}
		s.mu.RUnlock()

		if entity != nil {
			s.mu.Lock()
			s.Map.MoveEntity(entity, models.Position{X: -10, Y: -10})
			s.deadCases++
			s.mu.Unlock()
			req.ResponseChan <- models.DeadResponse{Authorized: true}
		} else {
			req.ResponseChan <- models.DeadResponse{Authorized: false}
		}
	}
}

func (s *Simulation) handleExitRequest() {
	for req := range s.ExitChan {
		if req.MemberType != "persons" {
			req.ResponseChan <- models.ExitResponse{Authorized: false}
			continue
		}

		var entity interface{}
		s.mu.RLock()
		for _, person := range s.Persons {
			if person.ID == req.MemberID {
				entity = &person
				break
			}
		}
		s.mu.RUnlock()

		if entity != nil {
			s.mu.Lock()
			s.Map.RemoveEntity(entity)
			s.mu.Unlock()
			req.ResponseChan <- models.ExitResponse{Authorized: true}
		} else {
			req.ResponseChan <- models.ExitResponse{Authorized: false}
		}
	}
}

func (s *Simulation) Initialize(numDrones, numCrowdMembers, numObstacles int) {
	fmt.Println("Initializing simulation")
	// @TODO : Récupérer la distress depuis la config.
	s.DefaultDistressProbability = DEFAULT_DISTRESS_PROBABILITY

	configPath := "configs/empty_layout.json"
	config, err := LoadFestivalConfig(configPath)
	if err != nil {
		fmt.Printf("Warning: Could not load empty config from %s: %v\n", configPath, err)
		fmt.Println("Using default obstacle initialization...")
		s.initializeDefaultObstacles(numObstacles)
	} else {
		fmt.Println("Successfully loaded empty configuration!")
		s.FestivalConfig = config
		err = s.Map.ApplyFestivalConfig(config)
		if err != nil {
			fmt.Printf("Error applying empty config: %v\nFalling back to default initialization\n", err)
			s.initializeDefaultObstacles(numObstacles)
		} else {
			fmt.Println("Successfully applied empty configuration")
			s.buildPOIMap()
		}
	}

	// Initialiser les POIs
	//s.initializePOIs()

	// Initialiser les RescuePoints après les POIs
	s.InitializeRescuePoints()

	s.createDrones(numDrones)
	s.createInitialCrowd(numCrowdMembers)
	s.festivalTime.Start()
}

func (s *Simulation) UpdateMap(nomConfig string) {
	fmt.Println("Loading Map")

	configPath := "configs/" + nomConfig + ".json"
	config, err := LoadFestivalConfig(configPath)
	if err != nil {
		fmt.Printf("Warning: Could not load festival config from %s: %v\n", configPath, err)
		fmt.Println("Using default obstacle initialization...")
		s.initializeDefaultObstacles(0)
	} else {
		fmt.Println("Successfully loaded festival configuration!")
		s.FestivalConfig = config
		err = s.Map.ApplyFestivalConfig(config)
		if err != nil {
			fmt.Printf("Error applying festival config: %v\nFalling back to default initialization\n", err)
			s.initializeDefaultObstacles(0)
		} else {
			fmt.Println("Successfully applied festival configuration")
			s.buildPOIMap()
		}
		s.InitializeRescuePoints()
	}
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

// func (s *Simulation) createDrones(numDrones int) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	for i := 0; i < numDrones; i++ {
// 		watch := models.MyWatch{
// 			CornerBottomLeft: models.Position{X: float64(i * s.Map.Width / numDrones), Y: 0},
// 			CornerTopRight: models.Position{
// 				X: float64((i + 1) * s.Map.Width / numDrones),
// 				Y: float64(s.Map.Height),
// 			},
// 		}

// 		drone := drones.NewSurveillanceDrone(
// 			i,
// 			models.Position{
// 				X: float64(i*s.Map.Width/numDrones) + float64(s.Map.Width/(2*numDrones)),
// 				Y: float64(s.Map.Height / 2),
// 			},
// 			watch,
// 			100.0,
// 			s.DroneSeeRange,
// 			s.DroneCommRange,
// 			s.getDroneSeeFunction(i),
// 			s.getDroneInComRangeFunction(i),
// 			s.MoveChan,
// 			s.poiMap,
// 			s.ChargingChan,
// 			s.MedicalDeliveryChan,
// 			s.SavePersonChan,
// 			DEFAULT_PROTOCOL_MODE,
// 			s.SavePeopleByRescuerChan,
// 			s.Map.Width,
// 			s.Map.Height,
// 			s, // Passer la simulation elle-même
// 		)

// 		s.Drones = append(s.Drones, drone)
// 		s.Map.AddDrone(&s.Drones[len(s.Drones)-1])
// 	}
// }

func (s *Simulation) createDrones(n int) {
	droneSeeFunction := func(d *drones.Drone) []*persons.Person {
		currentCell := d.Position
		rangeDrone := s.DroneSeeRange

		//fmt.Printf("RangeDrone : %d\n", rangeDrone)

		Vector := models.Vector{X: currentCell.X, Y: currentCell.Y}
		cercleValuesFloat, _ := Vector.GenerateCircleValues(rangeDrone)

		droneInformations := make([]*persons.Person, 0)
		nbPersDetected := 0

		for z := 0; z < len(cercleValuesFloat); z++ {
			positionInCercle := cercleValuesFloat[z]
			position := models.Position{X: d.Position.X + positionInCercle.X, Y: d.Position.Y + positionInCercle.Y}
			distance := currentCell.CalculateDistance(position)
			if cell, exists := s.Map.Cells[position]; exists {
				if s.debug {
					fmt.Printf("Distance : %.2f\n", distance)
				}
				//fmt.Println("Position : ", position)
				for _, member := range cell.Persons {
					//probaDetection := 1.0
					probaDetection := max(0, 1.0/float64(s.DroneSeeRange)-(float64(nbPersDetected)*0.03))
					if rand.Float64() < probaDetection {
						//fmt.Printf("Drone %d (%.2f, %.2f) sees person %d (%.2f, %.2f) \n", d.ID, d.Position.X, d.Position.Y, member.ID, member.Position.X, member.Position.Y)
						droneInformations = append(droneInformations, member)
						nbPersDetected++
					}
				}
			}
		}
		return droneInformations
	}

	droneInComRange := func(d *drones.Drone) []*drones.Drone {
		rangeDrone := s.DroneCommRange
		droneInformations := make([]*drones.Drone, 0)

		for i := range s.Drones {
			drone := &s.Drones[i]
			if drone == d {
				continue
			}

			dist := drone.Position.CalculateDistance(d.Position)

			if dist <= float64(rangeDrone) {
				droneInformations = append(droneInformations, drone)
			}
		}

		return droneInformations
	}

	droneGetRescuePoint := func(pos models.Position) *rescue.RescuePoint {
		var closest *rescue.RescuePoint
		minDist := math.Inf(1)

		for _, rp := range s.RescuePoints {
			dist := pos.CalculateDistance(rp.Position)
			if dist < minDist {
				minDist = dist
				closest = rp
			}
		}

		return closest
	}

	getDroneNetwork := func(d *drones.Drone) drones.DroneEffectiveNetwork {
		return s.calculateSingleDroneNetwork(d)
	}

	positionsDrone := goDronesZones(n, s.Map.Width, s.Map.Height)

	for i := 0; i < n; i++ {
		zone := positionsDrone[i]
		battery := 60 + rand.Float64()*(100-60)
		d := drones.NewSurveillanceDrone(i, models.Position{X: float64((zone[0][0] + zone[1][0]) / 2), Y: float64((zone[0][1] + zone[1][1]) / 2)},
			models.MyWatch{CornerBottomLeft: models.Position{X: float64(zone[0][0]), Y: float64(zone[0][1])}, CornerTopRight: models.Position{X: float64(zone[1][0]), Y: float64(zone[1][1])}},
			battery, s.DroneSeeRange, s.DroneCommRange,
			droneSeeFunction, droneInComRange, droneGetRescuePoint, getDroneNetwork,
			s.MoveChan, s.poiMap, s.ChargingChan, s.MedicalDeliveryChan,
			s.SavePersonChan, DEFAULT_PROTOCOL_MODE, // Use the constant here
			s.SavePeopleByRescuerChan, s.Map.Width, s.Map.Height,
			s.debug)
		d.InitProtocol()
		s.Drones = append(s.Drones, d)
		s.Map.AddDrone(&s.Drones[len(s.Drones)-1])
	}

}

func goDronesZones(N int, W, H int) [][2][2]int {
	if N <= 0 {
		return [][2][2]int{}
	}

	Nx := int(math.Floor(math.Sqrt(float64(N))))
	Ny := int(math.Ceil(float64(N) / float64(Nx)))
	dx := float64(W) / float64(Nx)
	dy := float64(H) / float64(Ny)

	zones := make([][2][2]int, N)
	for k := 0; k < N; k++ {
		i := k % Nx
		j := k / Nx
		x1 := int(math.Round(float64(i) * dx))
		y1 := int(math.Round(float64(j) * dy))
		x2 := int(math.Round(float64(i+1) * dx))
		y2 := int(math.Round(float64(j+1) * dy))
		zones[k] = [2][2]int{{x1, y1}, {x2, y2}}
	}

	return zones
}

func (s *Simulation) createInitialCrowd(n int) {
	fmt.Println("Creating initial crowd")
	for i := 0; i < n; i++ {
		member := persons.NewCrowdMember(i,
			models.Position{X: 0, Y: float64(rand.Intn(s.Map.Height))},
			s.DefaultDistressProbability, 200, s.Map.Width, s.Map.Height, s.MoveChan, s.DeadChan, s.ExitChan)
		s.Persons = append(s.Persons, member)
		s.Map.AddCrowdMember(&s.Persons[len(s.Persons)-1])
	}
}

func (s *Simulation) Update() {
	// SI 1000 Ticks fin du festival
	if s.currentTick == s.festivalTotalTicks {
		fmt.Println("End of festival")

		for i := range s.Persons {
			p := &s.Persons[i]
			path := models.FindPath(p.Position, models.Position{X: (float64(s.Map.Width)/10)*9 + 0.1, Y: p.Position.Y}, s.Map.Width, s.Map.Height, make(map[models.Position]bool))
			p.CurrentPath = path
			s.Persons[i].SeekingExit = true
		}
	}

	if s.FestivalState == Ended {
		return
	}

	time.Sleep(200 * time.Millisecond)
	if s.hardDebug {
		fmt.Println("New Tick")
	}
	s.currentTick++
	var wg sync.WaitGroup

	// Update persons every 10 ticks
	if s.currentTick%1 == 0 {
		//fmt.Printf("\n=== TICK %d: PEOPLE SHOULD MOVE ===\n", s.currentTick)
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
					if p.IsAlive() && p.StillInSim {
						if p.CurrentPOI != nil && p.TargetPOIPosition == nil {
							if pos := s.getNearestPOI(p.Position, *p.CurrentPOI); pos != nil {
								p.SetTargetPOI(*p.CurrentPOI, *pos)
							} else {
								p.CurrentPOI = nil
							}
						}
						p.Myturn()
					}
				}(&s.Persons[idx])
			}
		}
	}

	allPeopleAreOut := true
	for i := range s.Persons {
		if s.Persons[i].StillInSim {
			allPeopleAreOut = false
			break
		}
	}

	allDronesAreCharging := true
	if allPeopleAreOut {
		for i := range s.Drones {
			s.Drones[i].DroneState = drones.FinalGoingToDock
			if !s.Drones[i].IsCharging {
				allDronesAreCharging = false
			}
		}

	}

	if allDronesAreCharging && allPeopleAreOut {
		s.FestivalState = Ended
	}

	wg.Wait()

	var wgDroneRecive sync.WaitGroup

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
			wgDroneRecive.Add(1)
			updatedDrones[s.Drones[idx].ID] = struct{}{}
			go func(d *drones.Drone) {
				defer wgDroneRecive.Done()
				d.ReceiveInfo()
			}(&s.Drones[idx])
		}
	}

	wgDroneRecive.Wait()

	var wgDrone sync.WaitGroup
	updatedDrones = make(map[int]struct{})

	for _, idx := range droneIndexes {
		if _, exists := updatedDrones[s.Drones[idx].ID]; !exists {
			wgDrone.Add(1)
			updatedDrones[s.Drones[idx].ID] = struct{}{}
			go func(d *drones.Drone) {
				defer wgDrone.Done()
				d.Myturn()
			}(&s.Drones[idx])
		}
	}

	wgDrone.Wait()

	var rpWg sync.WaitGroup

	for i, _ := range s.RescuePoints {
		rpWg.Add(1)
		go func(rp *rescue.RescuePoint) {
			defer rpWg.Done()
			rp.UpdateRescuers()
		}(s.RescuePoints[i])
	}

	rpWg.Wait()

	if s.debug {
		for index, cell := range s.Map.Cells {
			for _, member := range cell.Persons {
				fmt.Printf("%v - Person %d is at position (%.2f, %.2f) -- Current Cell = (%.2f, %.2f) \n",
					index, member.ID, member.Position.X, member.Position.Y, cell.Position.X, cell.Position.Y)
			}
		}
	}

}

func (s *Simulation) UpdateDroneSize(newSize int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if newSize == len(s.Drones) {
		fmt.Println("[DRONES] - No change needed, current size is equal to new size")
		return
	}

	currentSize := len(s.Drones)

	if newSize > currentSize {
		// Ajout des nouveaux drones
		s.createDrones(newSize - currentSize)
	} else if newSize < currentSize {
		dronesToRemove := currentSize - newSize
		for i := 0; i < dronesToRemove; i++ {
			if len(s.Drones) == 0 {
				break
			}

			// On prend la dernière personne de la liste
			droneToRemove := s.Drones[len(s.Drones)-1]

			// On la retire de la carte
			s.Map.RemoveEntity(droneToRemove)
			//s.Map.DeleteEntity(droneToRemove)

			// On la retire de la liste des personnes
			s.Drones = s.Drones[:len(s.Drones)-1]
		}
	}
}

func (s *Simulation) UpdateDroneProtocole(newprot int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Drones {
		s.Drones[i].UpdateProtocole(newprot)
	}
}

func (s *Simulation) UpdateCrowdSize(newSize int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if newSize == len(s.Persons) {
		fmt.Println("[CrowMember] - No change needed, current size is equal to new size")
		return
	}

	currentSize := len(s.Persons)

	if newSize > currentSize {
		for i := currentSize; i < newSize; i++ {
			member := persons.NewCrowdMember(i,
				models.Position{X: 0, Y: float64(rand.Intn(s.Map.Height))},
				s.DefaultDistressProbability, 20, s.Map.Width, s.Map.Height, s.MoveChan, s.DeadChan, s.ExitChan)
			s.Persons = append(s.Persons, member)
			s.Map.AddCrowdMember(&s.Persons[len(s.Persons)-1])
		}
	} else if newSize < currentSize {
		// Retrait des personnes en excès
		personsToRemove := currentSize - newSize

		// On commence par la fin de la liste pour éviter les problèmes d'index
		for i := 0; i < personsToRemove; i++ {
			if len(s.Persons) == 0 {
				break
			}

			// On prend la dernière personne de la liste
			personToRemove := s.Persons[len(s.Persons)-1]

			// On la retire de la carte
			s.Map.RemoveEntity(personToRemove)
			//s.Map.DeleteEntity(personToRemove)

			// On la retire de la liste des personnes
			s.Persons = s.Persons[:len(s.Persons)-1]
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

func (s *Simulation) calculatePeopleDensity() models.DensityGrid {
	const gridSize = 10 // 10x10 grid
	grid := make([][]float64, gridSize)
	for i := range grid {
		grid[i] = make([]float64, gridSize)
	}

	cellWidth := float64(s.Map.Width) / float64(gridSize)
	cellHeight := float64(s.Map.Height) / float64(gridSize)

	for _, person := range s.Persons {
		if person.Position.X < 0 || person.Position.Y < 0 {
			continue
		}
		gridX := int(person.Position.X / cellWidth)
		gridY := int(person.Position.Y / cellHeight)
		if gridX >= 0 && gridX < gridSize && gridY >= 0 && gridY < gridSize {
			grid[gridY][gridX]++
		}
	}

	maxCount := 0.0
	for _, row := range grid {
		for _, count := range row {
			if count > maxCount {
				maxCount = count
			}
		}
	}

	if maxCount > 0 {
		for i := range grid {
			for j := range grid[i] {
				grid[i][j] /= maxCount
			}
		}
	}

	return models.DensityGrid{
		Grid:     grid,
		Width:    s.Map.Width,
		Height:   s.Map.Height,
		CellSize: gridSize,
	}
}

func (s *Simulation) calculateDroneNetwork() models.DroneNetwork {
	network := models.DroneNetwork{
		DronePositions: make([]models.Position, len(s.Drones)),
	}

	for i, drone := range s.Drones {
		network.DronePositions[i] = drone.Position
	}

	for i, drone1 := range s.Drones {
		for j, drone2 := range s.Drones {
			if i >= j {
				continue
			}
			dist := drone1.Position.CalculateDistance(drone2.Position)
			if dist <= float64(s.DroneCommRange) {
				network.DroneConnections = append(network.DroneConnections, drone1.Position)
				network.DroneConnections = append(network.DroneConnections, drone2.Position)
			}
		}
	}

	for _, drone := range s.Drones {
		for _, rp := range s.RescuePoints {
			dist := drone.Position.CalculateDistance(rp.Position)
			if dist <= float64(s.DroneCommRange) {
				network.RescueConnections = append(network.RescueConnections, drone.Position)
				network.RescueConnections = append(network.RescueConnections, rp.Position)
			}
		}
	}

	return network
}

// Replace the existing GetStatistics method with this updated version
func (s *Simulation) GetStatistics() SimulationStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalPeople := len(s.Persons)
	inDistress := s.CountCrowdMembersInDistress()

	totalBattery := 0.0
	droneCount := len(s.Drones)

	if droneCount > 0 {
		for _, d := range s.Drones {
			if !d.IsCharging {
				totalBattery += d.Battery
			} else {
				droneCount--
			}
		}
	}

	var avgBattery, coverage float64
	if droneCount > 0 {
		avgBattery = totalBattery / float64(droneCount)
		totalArea := float64(s.Map.Width * s.Map.Height)
		droneArea := math.Pi * float64(s.DroneSeeRange*s.DroneSeeRange)
		coverage = math.Min((float64(droneCount)*droneArea/totalArea)*100, 100)
	}

	return SimulationStatistics{
		TotalPeople:     totalPeople,
		InDistress:      inDistress,
		CasesTreated:    s.treatedCases,
		CasesDead:       s.deadCases,
		AverageBattery:  avgBattery,
		AverageCoverage: coverage,
		PeopleDensity:   s.calculatePeopleDensity(),
		DroneNetwork:    s.calculateDroneNetwork(),
	}
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

func (s *Simulation) GetPersonPauseInMap(p *persons.Person) models.Position {
	for _, cell := range s.Map.Cells {
		for _, member := range cell.Persons {
			if member.ID == p.ID {
				return cell.Position
				//fmt.Printf("Person %d is at position (%.2f, %.2f) -- Current Cell = (%.2f, %.2f) \n", member.ID, member.Position.X, member.Position.Y, cell.Position.X, cell.Position.Y)
			}
		}
	}
	return models.Position{X: -1, Y: -1}
}

func (s *Simulation) GetDronePauseInMap(p *drones.Drone) models.Position {
	for _, cell := range s.Map.Cells {
		for _, member := range cell.Drones {
			if member.ID == p.ID {
				return cell.Position
				//fmt.Printf("Person %d is at position (%.2f, %.2f) -- Current Cell = (%.2f, %.2f) \n", member.ID, member.Position.X, member.Position.Y, cell.Position.X, cell.Position.Y)
			}
		}
	}
	return models.Position{X: -1, Y: -1}
}

// InitializeRescuePoints initialise les points de sauvetage pour chaque tente médicale
func (s *Simulation) InitializeRescuePoints() {
	fmt.Printf("[SIMULATION] Initializing RescuePoints\n")
	// Créer un rescue point pour chaque tente médicale
	for i, pos := range s.poiMap[models.MedicalTent] {
		rp := rescue.NewRescuePoint(i, pos, s.SavePeopleByRescuerChan, s.debug)
		s.RescuePoints[pos] = rp
	}

	// Connecter les rescue points entre eux
	allPoints := make([]*rescue.RescuePoint, 0, len(s.RescuePoints))
	for _, rp := range s.RescuePoints {
		allPoints = append(allPoints, rp)
	}

	// Partager la liste complète avec chaque point
	for _, rp := range s.RescuePoints {
		rp.AllRescuePoints = allPoints
		rp.Start() // Démarrer les goroutines de gestion
	}

	fmt.Printf("[SIMULATION] Initialized %d rescue points\n", len(s.RescuePoints))
}

func (s *Simulation) calculateSingleDroneNetwork(targetDrone *drones.Drone) drones.DroneEffectiveNetwork {
	network := drones.DroneEffectiveNetwork{
		Drones: []*drones.Drone{},
	}

	visited := make(map[int]bool)

	// DFS function to find all connected drones
	var dfs func(currentDrone *drones.Drone)
	dfs = func(currentDrone *drones.Drone) {
		visited[currentDrone.ID] = true

		// Add all directly connected drones
		for i, otherDrone := range s.Drones {
			if otherDrone.ID == currentDrone.ID {
				continue
			}

			dist := currentDrone.Position.CalculateDistance(otherDrone.Position)
			if dist <= float64(s.DroneCommRange) && !visited[otherDrone.ID] {
				network.Drones = append(network.Drones, &s.Drones[i])
				dfs(&otherDrone) //Explorer recursivement les drones connectés à ce drone
			}
		}
	}

	dfs(targetDrone)

	return network
}

func (s *Simulation) GetRealFestivalTime() string {
	timeBegin := time.Date(2025, time.June, 1, 12, 0, 0, 0, time.UTC)
	return timeBegin.Add(time.Minute * time.Duration(s.currentTick)).Format("15:04")
}

func (s *Simulation) GetRemaningFestivalTime() string {
	timeBegin := time.Date(2025, time.June, 1, 12, 0, 0, 0, time.UTC)
	timeNow := timeBegin.Add(time.Minute * time.Duration(s.currentTick))
	timeEnd := timeBegin.Add(time.Minute * time.Duration(s.festivalTotalTicks))
	if s.currentTick >= s.festivalTotalTicks {
		return "Festival ended"
	}
	return timeEnd.Sub(timeNow).String()
}
