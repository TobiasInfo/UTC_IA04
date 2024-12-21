package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
	"math/rand"
)

type DroneState int

const (
	NoDefinedState DroneState = iota
	GoingToCharge
)

type Drone struct {
	ID                      int
	DroneSeeRange           int
	DroneCommRange          int
	Position                models.Position
	Battery                 float64
	DroneSeeFunction        func(d *Drone) []*persons.Person
	DroneInComRangeFunc     func(d *Drone) []*Drone
	ReportedZonesByCentrale []models.Position
	SeenPeople              []*persons.Person
	DroneInComRange         []*Drone
	MoveChan                chan models.MovementRequest
	MapPoi                  map[models.POIType][]models.Position
	ChargingChan            chan models.ChargingRequest
	IsCharging              bool
	MedicalDeliveryChan     chan models.MedicalDeliveryRequest
	MedicalTentTimer        int
	DeploymentTimer         int
	PeopleToSave            *persons.Person
	Objectif                models.Position
	HasMedicalGear          bool
	SavePersonChan          chan models.SavePersonRequest
	ProtocolMode            int      // 1 = protocol 1, 2 = protocol 2, 3 = protocol 3
	Rescuer                 *Rescuer // For protocol 3, the spawned rescuer
	SavePersonByRescuer     chan models.RescuePeopleRequest
	MapWidth                int
	MapHeight               int
	DroneState              DroneState
}

type Rescuer struct {
	ID          int
	Position    models.Position
	Person      *persons.Person
	MedicalTent models.Position
	State       int  // 0 = going to person, 1 = returning to tent
	Active      bool // Tracks if Rescuer is currently on a mission
}

// NewSurveillanceDrone crée un nouveau drone
func NewSurveillanceDrone(id int,
	position models.Position,
	battery float64, droneSeeRange int,
	droneCommunicationRange int,
	droneSeeFunc func(d *Drone) []*persons.Person,
	DroneInComRange func(d *Drone) []*Drone,
	moveChan chan models.MovementRequest,
	mapPoi map[models.POIType][]models.Position,
	chargingChan chan models.ChargingRequest,
	medicalDeliveryChan chan models.MedicalDeliveryRequest,
	savePersonChan chan models.SavePersonRequest,
	protocolMode int,
	savePersonByRescuer chan models.RescuePeopleRequest,
	MapWidth int,
	MapHeight int) Drone {
	return Drone{
		ID:                      id,
		Position:                position,
		Battery:                 battery,
		DroneSeeRange:           droneSeeRange,
		DroneCommRange:          droneCommunicationRange,
		DroneSeeFunction:        droneSeeFunc,
		DroneInComRangeFunc:     DroneInComRange,
		ReportedZonesByCentrale: []models.Position{},
		SeenPeople:              []*persons.Person{},
		DroneInComRange:         []*Drone{},
		MoveChan:                moveChan,
		MapPoi:                  mapPoi,
		ChargingChan:            chargingChan,
		IsCharging:              false,
		MedicalDeliveryChan:     medicalDeliveryChan,
		MedicalTentTimer:        0,
		DeploymentTimer:         1,
		PeopleToSave:            nil,
		Objectif:                models.Position{},
		HasMedicalGear:          false,
		SavePersonChan:          savePersonChan,
		ProtocolMode:            protocolMode,
		Rescuer:                 nil,
		SavePersonByRescuer:     savePersonByRescuer,
		MapWidth:                MapWidth,
		MapHeight:               MapHeight,
		DroneState:              NoDefinedState,
	}
}

func (d *Drone) tryCharging() bool {
	if d.DroneState != GoingToCharge && !d.IsCharging {
		return false
	}

	if d.IsCharging {
		d.Battery += 5
		if d.Battery >= 80+rand.Float64()*20 {
			d.IsCharging = false
			d.DroneState = NoDefinedState
			return false
		}
		return true
	}

	responseChan := make(chan models.ChargingResponse)
	d.ChargingChan <- models.ChargingRequest{
		DroneID:      d.ID,
		Position:     d.Position,
		ResponseChan: responseChan,
	}

	response := <-responseChan
	if response.Authorized {
		fmt.Printf("[DRONE %d] Starting to charge at (%.0f, %.0f)\n", d.ID, d.Position.X, d.Position.Y)
		d.IsCharging = true
		d.Battery += 5
		return true
	}
	return false
}

func (d *Drone) Move(target models.Position) bool {
	if d.Battery <= 0 {
		return false
	}
	//fmt.Printf("[DRONE %d] Moving from (%.0f, %.0f) to (%.0f, %.0f)\n", d.ID, d.Position.X, d.Position.Y, target.X, target.Y)

	responseChan := make(chan models.MovementResponse)
	d.MoveChan <- models.MovementRequest{MemberID: d.ID, MemberType: "drone", NewPosition: target, ResponseChan: responseChan}
	response := <-responseChan

	if response.Authorized {
		dechargingStep := 0.5
		if d.Battery >= dechargingStep {
			//d.Battery -= dechargingStep
		} else {
			d.Battery = 0.0
		}
		d.Position = target
		return true
	}

	return false
}

func (d *Drone) ReceiveInfo() {
	// Le code est le même pour protocole 1, 2 ou 3, on récupère juste les infos
	seenPeople := d.DroneSeeFunction(d)
	droneInComRange := d.DroneInComRangeFunc(d)

	d.SeenPeople = seenPeople
	d.DroneInComRange = droneInComRange
}

func FindBestDroneForRescue(drones []*Drone, person *persons.Person) *Drone {
	var bestDrone *Drone
	minCost := math.Inf(1)

	for _, dr := range drones {
		if dr.PeopleToSave != nil {
			continue
		}

		medicalTentPos, _ := dr.closestPOI(models.MedicalTent)
		_, distanceToCharging := dr.closestPOI(models.ChargingStation)
		// On utilise la distance de Manhattan ici comme dans la version actuelle
		distanceToTent := dr.Position.CalculateManhattanDistance(medicalTentPos)
		distanceTentToPerson := medicalTentPos.CalculateManhattanDistance(person.Position)
		totalDistance := distanceToTent + distanceTentToPerson + distanceToCharging + 2
		if dr.Battery >= totalDistance && totalDistance < minCost {
			bestDrone = dr
			minCost = totalDistance
		}
	}
	return bestDrone
}

func (d *Drone) GetAllReachableDrones() []*Drone {
	visited := make(map[int]bool)
	queue := []*Drone{d}
	var result []*Drone

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.ID] {
			continue
		}
		visited[current.ID] = true
		result = append(result, current)

		neighbors := current.DroneInComRangeFunc(current)
		for _, neigh := range neighbors {
			if !visited[neigh.ID] {
				queue = append(queue, neigh)
			}
		}
	}
	return result
}

func (d *Drone) closestPOI(poiType models.POIType) (models.Position, float64) {
	pois := d.MapPoi[poiType]
	minDistance := math.Inf(1)
	var closestPOI models.Position

	for _, poi := range pois {
		distance := d.Position.CalculateManhattanDistance(poi)
		if distance < minDistance {
			minDistance = distance
			closestPOI = poi
		}
	}
	return closestPOI, minDistance
}

func (d *Drone) nextStepToPos(pos models.Position) models.Position {
	return stepTowards(d.Position, pos)
}

func (d *Drone) BatteryManagement() (models.Position, bool) {
	closestStation, minDistance := d.closestPOI(models.ChargingStation)
	if d.Battery <= minDistance+5 || d.DroneState == GoingToCharge {
		step := d.nextStepToPos(closestStation)
		d.DroneState = GoingToCharge
		fmt.Printf("[DRONE %d] Low battery, Drone position (%.0f, %0.f), moving to charging station at (%.0f, %.0f) -- Next-Step (%.0f, %.0f)\n", d.ID, d.Position.X, d.Position.Y, closestStation.X, closestStation.Y, step.X, step.Y)
		return step, true
	}
	return models.Position{}, false
}

func (d *Drone) UpdateRescuer() {
	if d.Rescuer == nil {
		return
	}
	rescuer := d.Rescuer

	if !rescuer.Active {
		return
	}

	if rescuer.State == 0 {
		// Moving towards person
		// 13.0, 7.0 -- 13.5, 7.5
		if rescuer.Position.CalculateDistance(rescuer.Person.Position) <= 1 {
			// Save the person
			fmt.Printf("[RESCUER] Saving person %d\n", rescuer.Person.ID)
			rescuer.Person.AssignedDroneID = nil

			rescueResponse := make(chan models.RescuePeopleResponse)
			d.SavePersonByRescuer <- models.RescuePeopleRequest{
				PersonID:     rescuer.Person.ID,
				RescuerID:    d.ID,
				ResponseChan: rescueResponse,
			}

			response := <-rescueResponse
			if response.Authorized {
				fmt.Printf("[RESCUER] Successfully treated person %d\n", rescuer.Person.ID)
			} else {
				fmt.Printf("[RESCUER] Failed to treat person %d: %s\n", rescuer.Person.ID, response.Reason)
			}

			// Start return journey
			rescuer.State = 1
		} else {
			// Move one step closer to person
			rescuer.Position = stepTowards(rescuer.Position, rescuer.Person.Position)
		}
	} else if rescuer.State == 1 {
		// Returning to medical tent
		if rescuer.Position.CalculateDistance(rescuer.MedicalTent) <= 1 {
			// Mission complete, rescuer disappears
			fmt.Printf("[RESCUER] Mission complete, returning to tent.\n")
			rescuer.Active = false
			d.Rescuer = nil
			d.PeopleToSave = nil
		} else {
			// Move one step closer to tent
			rescuer.Position = stepTowards(rescuer.Position, rescuer.MedicalTent)
		}
	}
}

// stepTowards calcule un pas vers une cible
func stepTowards(from, to models.Position) models.Position {
	dx := to.X - from.X
	dy := to.Y - from.Y
	step := from

	threshold := 0.5

	if math.Abs(dx) > threshold {
		if dx > 0 {
			step.X += 1
		} else {
			step.X -= 1
		}
	}
	if math.Abs(dy) > threshold {
		if dy > 0 {
			step.Y += 1
		} else {
			step.Y -= 1
		}
	}

	return step
}

func (d *Drone) Think() models.Position {
	// Handle battery management first
	pos, goCharging := d.BatteryManagement()
	if goCharging {
		return pos
	}

	if d.ProtocolMode == 3 {
		// Update existing rescuer if one exists
		if d.Rescuer != nil {
			d.UpdateRescuer()
			return d.randomMovement()
		}

		// Check if we have a person to save
		if d.PeopleToSave != nil {
			medicalTentPos, _ := d.closestPOI(models.MedicalTent)
			distToTent := d.Position.CalculateManhattanDistance(medicalTentPos)

			if distToTent <= float64(d.DroneCommRange) {
				// Within communication range, spawn rescuer
				fmt.Printf("[DRONE %d] Within medical tent range, spawning rescuer for person %d\n",
					d.ID, d.PeopleToSave.ID)
				d.Rescuer = &Rescuer{
					ID:          d.ID,
					Position:    medicalTentPos,
					Person:      d.PeopleToSave,
					MedicalTent: medicalTentPos,
					State:       0,
					Active:      true,
				}
				// Clear mission and return to patrol
				d.PeopleToSave = nil
				return d.randomMovement()
			} else {
				// Move towards medical tent
				return d.nextStepToPos(medicalTentPos)
			}
		}

		// Look for new people in distress
		for _, person := range d.SeenPeople {
			if person.IsInDistress() {
				if person.IsAssigned() && (d.PeopleToSave == nil || d.PeopleToSave.ID != person.ID) {
					continue
				}

				fmt.Printf("[DRONE %d] Detected person in distress (ID: %d) at (%.0f, %.0f)\n",
					d.ID, person.ID, person.Position.X, person.Position.Y)

				// Get all reachable drones
				allDrones := d.GetAllReachableDrones()
				for _, dr := range allDrones {
					if dr.ID != d.ID {
						fmt.Printf("[DRONE %d] Informing DRONE %d about person in distress (ID: %d)\n",
							d.ID, dr.ID, person.ID)
					}
				}

				// Find best drone for rescue
				bestDrone := FindBestDroneForRescue(allDrones, person)
				if bestDrone == nil {
					fmt.Printf("[DRONE %d] No drone available for person %d (insufficient battery or all busy)\n",
						d.ID, person.ID)
					break
				}

				if bestDrone.ID == d.ID {
					fmt.Printf("[DRONE %d] Taking responsibility for person %d (protocol 3)\n",
						d.ID, person.ID)
					person.AssignedDroneID = &d.ID
					d.PeopleToSave = person
					medicalTentPos, _ := d.closestPOI(models.MedicalTent)
					return d.nextStepToPos(medicalTentPos)
				} else {
					fmt.Printf("[DRONE %d] Not handling person %d, DRONE %d will handle it.\n",
						d.ID, person.ID, bestDrone.ID)
					person.AssignedDroneID = &bestDrone.ID
					bestDrone.PeopleToSave = person
				}
			}
		}

		// No person in distress, continue patrol
		return d.randomMovement()
	}
	if d.ProtocolMode == 2 {
		if d.Objectif != (models.Position{}) && d.PeopleToSave != nil {
			if d.Position.X == d.Objectif.X && d.Position.Y == d.Objectif.Y {
				medicalTentPos, _ := d.closestPOI(models.MedicalTent)
				if d.Position == medicalTentPos && !d.HasMedicalGear {
					responseChan := make(chan models.MedicalDeliveryResponse)
					d.MedicalDeliveryChan <- models.MedicalDeliveryRequest{
						PersonID:     d.PeopleToSave.ID,
						DroneID:      d.ID,
						ResponseChan: responseChan,
					}
					rep := <-responseChan
					if rep.Authorized {
						fmt.Printf("[DRONE %d] A récupéré le matériel médical pour la personne %d.\n", d.ID, d.PeopleToSave.ID)
						d.Objectif = models.Position{X: math.Round(d.PeopleToSave.Position.X), Y: math.Round(d.PeopleToSave.Position.Y)}
						d.HasMedicalGear = true
					}
				} else if d.Position.X == math.Round(d.PeopleToSave.Position.X) && d.Position.Y == math.Round(d.PeopleToSave.Position.Y) {
					responseSave := make(chan models.SavePersonResponse)
					d.SavePersonChan <- models.SavePersonRequest{
						PersonID:     d.PeopleToSave.ID,
						DroneID:      d.ID,
						ResponseChan: responseSave,
					}
					rep := <-responseSave
					if rep.Authorized {
						fmt.Printf("[DRONE %d] A sauvé la personne %d !\n", d.ID, d.PeopleToSave.ID)
						d.PeopleToSave.AssignedDroneID = nil
						d.PeopleToSave = nil
						d.Objectif = models.Position{}
						d.HasMedicalGear = false
					} else {
						fmt.Printf("[DRONE %d] n'a pas pu sauver la personne %d !\n", d.ID, d.PeopleToSave.ID)
						d.PeopleToSave.AssignedDroneID = nil
						d.PeopleToSave = nil
						d.Objectif = models.Position{}
					}
				}
			}
			step := d.nextStepToPos(d.Objectif)
			return step
		}

		for _, person := range d.SeenPeople {
			if person.IsInDistress() {
				if person.IsAssigned() && (d.PeopleToSave == nil || d.PeopleToSave.ID != person.ID) {
					continue
				}

				fmt.Printf("[DRONE %d] A détecté une personne en détresse (ID: %d) à (%.0f, %.0f)\n",
					d.ID, person.ID, person.Position.X, person.Position.Y)

				allDrones := d.GetAllReachableDrones()
				for _, dr := range allDrones {
					if dr.ID != d.ID {
						fmt.Printf("[DRONE %d] Informe DRONE %d d'une personne en détresse (ID: %d)\n",
							d.ID, dr.ID, person.ID)
					}
				}

				bestDrone := FindBestDroneForRescue(allDrones, person)
				if bestDrone == nil {
					fmt.Printf("[DRONE %d] Aucun drone ne peut gérer la personne %d (batterie insuffisante ou tous occupés)\n", d.ID, person.ID)
					break
				}
				if bestDrone.ID == d.ID {
					fmt.Printf("[DRONE %d] Prend en charge la personne %d car c'est le mieux placé.\n", d.ID, person.ID)
					person.AssignedDroneID = &d.ID
					d.PeopleToSave = person
					medicalTentPos, _ := d.closestPOI(models.MedicalTent)
					d.Objectif = medicalTentPos
					return d.nextStepToPos(d.Objectif)
				} else {
					fmt.Printf("[DRONE %d] Ne gère pas la personne %d, c'est le DRONE %d qui s'en charge.\n", d.ID, person.ID, bestDrone.ID)
					person.AssignedDroneID = &bestDrone.ID
					medicalTentPos, _ := bestDrone.closestPOI(models.MedicalTent)
					bestDrone.PeopleToSave = person
					bestDrone.Objectif = medicalTentPos
					bestDrone.HasMedicalGear = false
				}
			}
		}
		return d.randomMovement()
	}
	if d.ProtocolMode == 1 {
		if d.Objectif != (models.Position{}) {
			fmt.Printf("[DRONE %d] Objectif : (%.0f, %.0f)\n", d.ID, d.Objectif.X, d.Objectif.Y)
			if d.Position.X == d.Objectif.X && d.Position.Y == d.Objectif.Y {
				medicalTentPos, _ := d.closestPOI(models.MedicalTent)
				if d.Position.X == medicalTentPos.X && d.Position.Y == medicalTentPos.Y {
					fmt.Println("MEDICAL : Drone", d.ID, "is at medical tent")
					responseChan := make(chan models.MedicalDeliveryResponse)
					d.MedicalDeliveryChan <- models.MedicalDeliveryRequest{
						PersonID:     d.PeopleToSave.ID,
						DroneID:      d.ID,
						ResponseChan: responseChan,
					}
					rep := <-responseChan
					if rep.Authorized {
						d.Objectif = models.Position{X: math.Round(d.PeopleToSave.Position.X), Y: math.Round(d.PeopleToSave.Position.Y)}
						d.HasMedicalGear = true
					}
				}
				if d.Position.X == math.Round(d.PeopleToSave.Position.X) && d.Position.Y == math.Round(d.PeopleToSave.Position.Y) {
					fmt.Println("MEDICAL : Drone", d.ID, "is delivering medical supplies to person", d.PeopleToSave.ID)
					responseSave := make(chan models.SavePersonResponse)
					d.SavePersonChan <- models.SavePersonRequest{
						PersonID:     d.PeopleToSave.ID,
						DroneID:      d.ID,
						ResponseChan: responseSave,
					}
					rep := <-responseSave
					if rep.Authorized {
						d.PeopleToSave = nil
						d.Objectif = models.Position{}
						d.HasMedicalGear = false
					} else {
						fmt.Printf("[DRONE %d] n'a pas pu sauver la personne %d !\n", d.ID, d.PeopleToSave.ID)
						d.PeopleToSave.AssignedDroneID = nil
						d.PeopleToSave = nil
						d.Objectif = models.Position{}
					}
				}
			}
			step := d.nextStepToPos(d.Objectif)
			return step
		}

		for _, person := range d.SeenPeople {
			if person.IsInDistress() {
				if person.IsAssigned() && (d.PeopleToSave == nil || d.PeopleToSave.ID != person.ID) {
					continue
				}

				d.PeopleToSave = person
				if !person.IsAssigned() {
					person.AssignedDroneID = &d.ID
				}
				medicalTentPos, _ := d.closestPOI(models.MedicalTent)
				_, distanceToCharging := d.closestPOI(models.ChargingStation)
				distanceToTent := d.Position.CalculateManhattanDistance(medicalTentPos)
				distancePersonToTent := person.Position.CalculateManhattanDistance(medicalTentPos)
				totalBatteryNeeded := distanceToTent + distancePersonToTent + distanceToCharging + 2
				if d.Battery >= totalBatteryNeeded {
					fmt.Println("MEDICAL : Drone", d.ID, "has enough battery to complete the mission")
					fmt.Println("MEDICAL : Drone", d.ID, "detected person", person.ID, "in distress")
					d.Objectif = medicalTentPos
					step := d.nextStepToPos(d.Objectif)
					return step
				}
			}
		}
		return d.randomMovement()
	}

	// En théorie, ce cas est unreachable.
	return d.randomMovement()
}

func (d *Drone) randomMovement() models.Position {
	// Define all possible directions
	directions := []models.Position{
		{X: 0, Y: -1},
		{X: 0, Y: 1},
		{X: -1, Y: 0},
		{X: -1, Y: -1},
		{X: 1, Y: 1},
		{X: 1, Y: -1},
		{X: -1, Y: 1},
		{X: 1, Y: 0},
	}

	// Si des personnes sont visibles, privilégier ces directions
	if len(d.SeenPeople) > 0 {
		// Calculer les scores pour chaque direction en fonction de la proximité avec les personnes
		directionScores := make(map[models.Position]float64)
		for _, dir := range directions {
			potentialPos := models.Position{
				X: d.Position.X + dir.X,
				Y: d.Position.Y + dir.Y,
			}

			// Vérifier les limites de la carte
			if potentialPos.X <= 0 || potentialPos.Y <= 0 ||
				math.Round(potentialPos.X) >= float64(d.MapWidth) ||
				potentialPos.Y >= float64(d.MapHeight) {
				continue
			}

			// Calculer un score basé sur la distance aux personnes visibles
			score := 0.0
			for _, person := range d.SeenPeople {
				dist := potentialPos.CalculateDistance(person.Position)
				score += 1.0 / (dist + 1) // Plus la distance est faible, plus le score est élevé
			}
			directionScores[dir] = score
		}

		// Trouver la direction avec le meilleur score
		var bestDir models.Position
		bestScore := -1.0
		for dir, score := range directionScores {
			if score > bestScore {
				bestScore = score
				bestDir = dir
			}
		}

		if bestScore > 0 {
			return models.Position{
				X: d.Position.X + bestDir.X,
				Y: d.Position.Y + bestDir.Y,
			}
		}
	}

	// Si aucune personne n'est visible ou si aucune direction valide n'a été trouvée,
	// utiliser le comportement aléatoire original
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		target := models.Position{
			X: d.Position.X + dir.X,
			Y: d.Position.Y + dir.Y,
		}

		if target.X <= 0 || target.Y <= 0 ||
			math.Round(target.X) >= float64(d.MapWidth) ||
			math.Round(target.Y) >= float64(d.MapHeight) {
			continue
		}

		return target
	}

	return d.Position
}

func (d *Drone) Myturn() {
	// Cannot communicate with other drones if charging
	d.ReceiveInfo()

	if d.tryCharging() {
		return
	}

	target := d.Think()
	if target.X == d.Position.X && target.Y == d.Position.Y {
		//d.Battery -= 0.25
		return
	}
	moved := d.Move(target)
	if !moved {
		fmt.Printf("Drone %d could not move to %v\n", d.ID, target)
	}
}

type WeightedParameters struct {
	DistanceWeight   float64
	BatteryWeight    float64
	ClusteringWeight float64
}

func (d *Drone) CalculateOptimalPosition(params WeightedParameters) models.Position {
	if len(d.ReportedZonesByCentrale) == 0 {
		return d.Position
	}

	var sumX, sumY, totalWeight float64
	for _, zone := range d.ReportedZonesByCentrale {
		weight := calculateZoneWeight(d, zone, params)
		sumX += zone.X * weight
		sumY += zone.Y * weight
		totalWeight += weight
	}
	if totalWeight == 0 {
		return d.Position
	}
	return models.Position{
		X: sumX / totalWeight,
		Y: sumY / totalWeight,
	}
}

func calculateZoneWeight(d *Drone, zone models.Position, params WeightedParameters) float64 {
	distance := zone.CalculateDistance(d.Position)
	distanceFactor := 1.0 / (1.0 + distance)
	batteryFactor := 1.0 - distance
	if batteryFactor < 0 {
		batteryFactor = 0
	}
	clusterFactor := calculateClusterFactor(d, zone)

	return (distanceFactor * params.DistanceWeight) +
		(batteryFactor * params.BatteryWeight) +
		(clusterFactor * params.ClusteringWeight)
}

func calculateClusterFactor(d *Drone, targetZone models.Position) float64 {
	const proximityThreshold = 2.0
	nearbyZones := 0.0
	for _, zone := range d.ReportedZonesByCentrale {
		if zone.CalculateDistance(targetZone) < proximityThreshold {
			nearbyZones += 1.0
		}
	}
	return nearbyZones / float64(len(d.ReportedZonesByCentrale))
}

func (d *Drone) UpdateProtocole(newprot int) {
	if newprot != 1 && newprot != 2 && newprot != 3 {
		return
	}
	if d.ProtocolMode == newprot {
		return
	}
	d.ProtocolMode = newprot
}
