package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Drone struct {
	ID                      int
	visionCapacity          int
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
	LastSharedUpdate        time.Time
	DroneComm               chan DroneUpdate
}

type RescueIntent struct {
	DroneID    int
	PersonID   int
	PathLength float64
	Battery    float64
}

type DroneUpdate struct {
	DroneID          int
	Position         models.Position
	AssignedPersonID int
	Battery          float64
	HasMedicalGear   bool
	TotalPathLength  float64
	MessageType      string
	ResponseNeeded   bool // New field to indicate if response is expected
}

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
	savePersonChan chan models.SavePersonRequest) Drone {
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
		DroneComm:               make(chan DroneUpdate, 10),
		LastSharedUpdate:        time.Now(),
	}
}

func (intent RescueIntent) IsBetterThan(other RescueIntent) bool {
	if math.Abs(intent.PathLength-other.PathLength) < 0.1 {
		return intent.DroneID < other.DroneID
	}
	return intent.PathLength < other.PathLength
}

func (d *Drone) calculateRescuePath(person *persons.Person) (float64, models.Position) {
	medicalTentPos, _ := d.closestPOI(models.MedicalTent)

	// Calculate total path: current -> medical tent -> person
	distanceToTent := d.Position.CalculateDistance(medicalTentPos)
	distanceFromTentToPerson := medicalTentPos.CalculateDistance(person.Position)

	totalDistance := distanceToTent + distanceFromTentToPerson

	return totalDistance, medicalTentPos
}

func (d *Drone) ShareStatus(nearbyDrones []*Drone) {
	if time.Since(d.LastSharedUpdate) < 500*time.Millisecond {
		return
	}

	assignedID := -1
	if d.PeopleToSave != nil {
		assignedID = d.PeopleToSave.ID
	}

	update := DroneUpdate{
		DroneID:          d.ID,
		Position:         d.Position,
		AssignedPersonID: assignedID,
		Battery:          d.Battery,
		HasMedicalGear:   d.HasMedicalGear,
		TotalPathLength:  0,
		MessageType:      "STATUS",
	}

	for _, otherDrone := range nearbyDrones {
		if d.ID != otherDrone.ID {
			select {
			case otherDrone.DroneComm <- update:
			default:
			}
		}
	}

	d.LastSharedUpdate = time.Now()
}

func (d *Drone) tryCharging() bool {
	if d.IsCharging {
		d.Battery += 5
		if d.Battery >= 80+rand.Float64()*20 {
			d.IsCharging = false
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

	if d.Position.X == target.X && d.Position.Y == target.Y {
		return false
	}

	responseChan := make(chan models.MovementResponse)
	d.MoveChan <- models.MovementRequest{MemberID: d.ID, MemberType: "drone", NewPosition: target, ResponseChan: responseChan}
	response := <-responseChan

	if response.Authorized {
		dechargingStep := 0.25
		if d.Battery >= dechargingStep {
			d.Battery -= dechargingStep
		} else {
			d.Battery = 0.0
		}

		d.Position.X = target.X
		d.Position.Y = target.Y
		return true
	}

	return false
}

func (d *Drone) ReceiveInfo() {
	seenPeople := d.DroneSeeFunction(d)
	droneInComRange := d.DroneInComRangeFunc(d)

	d.SeenPeople = seenPeople
	d.DroneInComRange = droneInComRange
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
	dx := pos.X - d.Position.X
	dy := pos.Y - d.Position.Y

	var step models.Position
	if math.Abs(dx) > math.Abs(dy) {
		if dx > 0 {
			step = models.Position{X: d.Position.X + 1, Y: d.Position.Y}
		} else {
			step = models.Position{X: d.Position.X - 1, Y: d.Position.Y}
		}
	} else {
		if dy > 0 {
			step = models.Position{X: d.Position.X, Y: d.Position.Y + 1}
		} else {
			step = models.Position{X: d.Position.X, Y: d.Position.Y - 1}
		}
	}

	return step
}

func (d *Drone) BatteryManagement() (models.Position, bool) {
	closestStation, minDistance := d.closestPOI(models.ChargingStation)
	if d.Battery <= minDistance+5 {
		step := d.nextStepToPos(closestStation)

		if step.X >= 0 && step.Y >= 0 && step.X < 30 && step.Y < 20 {
			return step, true
		}

		return d.Position, true
	}
	return models.Position{}, false
}

func (d *Drone) Think() models.Position {
	pos, goCharging := d.BatteryManagement()
	if goCharging {
		fmt.Printf("[DRONE-%d] Low battery (%.1f%%), heading to charge\n", d.ID, d.Battery)
		return pos
	}

	nearbyDrones := d.DroneInComRangeFunc(d)
	d.ShareStatus(nearbyDrones)

	// Process any pending updates
	for {
		select {
		case update := <-d.DroneComm:
			d.handleDroneUpdate(update)
		default:
			goto processLoop
		}
	}
processLoop:

	if d.Objectif != (models.Position{}) {
		if d.Position.X == d.Objectif.X && d.Position.Y == d.Objectif.Y {
			medicalTentPos, _ := d.closestPOI(models.MedicalTent)
			if d.Position.X == medicalTentPos.X && d.Position.Y == medicalTentPos.Y {
				fmt.Printf("[DRONE-%d] At medical tent\n", d.ID)
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
				fmt.Printf("[DRONE-%d] Delivering medical supplies to person %d\n", d.ID, d.PeopleToSave.ID)
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
				}
			}
		}
		step := d.nextStepToPos(d.Objectif)
		return step
	}

	// Track people we've already considered to avoid duplicate evaluations
	consideredPeople := make(map[int]bool)

	for _, person := range d.SeenPeople {
		if person.IsInDistress() && !consideredPeople[person.ID] {
			consideredPeople[person.ID] = true

			if d.isPersonBeingHelped(person, nearbyDrones) {
				fmt.Printf("[DRONE-%d] Person %d is already being helped by another drone\n",
					d.ID, person.ID)
				continue
			}

			pathLength, medicalTentPos := d.calculateRescuePath(person)
			_, distanceToCharging := d.closestPOI(models.ChargingStation)
			totalBatteryNeeded := pathLength + distanceToCharging + 2

			fmt.Printf("[DRONE-%d] Evaluating rescue of person %d:\n", d.ID, person.ID)
			fmt.Printf("[DRONE-%d]  - Total path length: %.1f\n", d.ID, pathLength)
			fmt.Printf("[DRONE-%d]  - Total battery needed: %.1f (current: %.1f)\n", d.ID,
				totalBatteryNeeded, d.Battery)

			if d.Battery >= totalBatteryNeeded {
				if d.broadcastRescueIntent(person, nearbyDrones) {
					fmt.Printf("[DRONE-%d] Taking responsibility for person %d\n", d.ID, person.ID)
					d.PeopleToSave = person
					d.Objectif = medicalTentPos
					return d.nextStepToPos(d.Objectif)
				}
			}
		}
	}

	return d.calculatePatrolStep()
}

func (d *Drone) calculateRescuePriority(pathLength float64) float64 {
	// Combine path length and drone ID for a unique priority
	// Lower value = higher priority
	return pathLength + (float64(d.ID) * 0.001) // Add small drone ID factor to break ties
}

func (d *Drone) handleDroneUpdate(update DroneUpdate) {
    // Only care about COMMIT messages if we were considering the same person
    if update.MessageType == "COMMIT" && 
       d.PeopleToSave != nil && 
       update.AssignedPersonID == d.PeopleToSave.ID {
        fmt.Printf("[DRONE-%d] Dropping rescue of person %d - Drone %d has committed\n",
            d.ID, update.AssignedPersonID, update.DroneID)
        d.PeopleToSave = nil
        d.Objectif = models.Position{}
        d.HasMedicalGear = false
    }
}

func (d *Drone) isDroneProbablyOutOfRange(droneID int) bool {
	for _, drone := range d.DroneInComRange {
		if drone.ID == droneID {
			return false
		}
	}
	return true
}

func (d *Drone) broadcastRescueIntent(person *persons.Person, nearbyDrones []*Drone) bool {
    pathLength, _ := d.calculateRescuePath(person)
    
    // First broadcast our intention
    update := DroneUpdate{
        DroneID:          d.ID,
        AssignedPersonID: person.ID,
        TotalPathLength:  pathLength,
        MessageType:      "INTENT",
    }

    // Send to all nearby drones
    for _, otherDrone := range nearbyDrones {
        if d.ID != otherDrone.ID {
            select {
            case otherDrone.DroneComm <- update:
            default:
            }
        }
    }

    // Wait briefly for any responses that show a shorter path
    timeout := time.After(50 * time.Millisecond)
    for {
        select {
        case response := <-d.DroneComm:
            if response.AssignedPersonID == person.ID && response.TotalPathLength < pathLength {
                fmt.Printf("[DRONE-%d] Backing off from person %d - Drone %d has shorter path (%.1f vs %.1f)\n",
                    d.ID, person.ID, response.DroneID, response.TotalPathLength, pathLength)
                return false
            }
        case <-timeout:
            goto timeoutExit
        }
    }
timeoutExit:

    // If we're still here, we have the shortest path - take responsibility
    commitUpdate := DroneUpdate{
        DroneID:          d.ID,
        AssignedPersonID: person.ID,
        MessageType:      "COMMIT",
    }

    // Tell others we're taking it
    for _, otherDrone := range nearbyDrones {
        if d.ID != otherDrone.ID {
            select {
            case otherDrone.DroneComm <- commitUpdate:
            default:
            }
        }
    }

    fmt.Printf("[DRONE-%d] Taking responsibility for person %d (path length: %.1f)\n",
        d.ID, person.ID, pathLength)
    return true
}

func (d *Drone) isPersonBeingHelped(person *persons.Person, nearbyDrones []*Drone) bool {
	for _, drone := range nearbyDrones {
		if drone.PeopleToSave != nil && drone.PeopleToSave.ID == person.ID && drone.ID != d.ID {
			return true
		}
	}
	return false
}

func (d *Drone) initiateHelp(person *persons.Person) models.Position {
	medicalTentPos, _ := d.closestPOI(models.MedicalTent)
	_, distanceToCharging := d.closestPOI(models.ChargingStation)
	distanceToTent := d.Position.CalculateDistance(medicalTentPos)
	distancePersonToTent := person.Position.CalculateDistance(medicalTentPos)
	totalBatteryNeeded := distanceToTent + distancePersonToTent + distanceToCharging + 2

	if d.Battery >= totalBatteryNeeded {
		d.PeopleToSave = person
		d.Objectif = medicalTentPos
		return d.nextStepToPos(d.Objectif)
	}

	return d.calculatePatrolStep()
}

func (d *Drone) calculatePatrolStep() models.Position {
	directions := []models.Position{
		{X: 0, Y: -1},
		{X: 0, Y: 1},
		{X: -1, Y: 0},
		{X: 1, Y: 0},
	}

	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		target := models.Position{
			X: d.Position.X + dir.X,
			Y: d.Position.Y + dir.Y,
		}
		if target.X >= 0 && target.Y >= 0 && target.X < 30 && target.Y < 20 {
			return target
		}
	}

	return d.Position
}

func (d *Drone) Myturn() {
	d.SeenPeople = []*persons.Person{}
	d.DroneInComRange = []*Drone{}

	if d.tryCharging() {
		d.ReceiveInfo()
		return
	}
	d.ReceiveInfo()

	target := d.Think()

	moved := d.Move(target)

	if !moved {
		fmt.Printf("[DRONE-%d] Could not move to target position\n", d.ID)
	}
}

type WeightedParameters struct {
	DistanceWeight   float64 // Poids pour la distance aux zones
	BatteryWeight    float64 // Poids pour considérer la consommation de batterie
	ClusteringWeight float64 // Poids pour favoriser les zones avec plus de points proches
}

func (d *Drone) CalculateOptimalPosition(params WeightedParameters) models.Position {
	if len(d.ReportedZonesByCentrale) == 0 {
		return d.Position // Si pas de zones, reste sur place
	}

	// Calculer le centre de gravité pondéré des zones reportées
	var sumX, sumY, totalWeight float64

	for _, zone := range d.ReportedZonesByCentrale {
		// Calculer le poids pour cette zone
		weight := calculateZoneWeight(d, zone, params)

		sumX += zone.X * weight
		sumY += zone.Y * weight
		totalWeight += weight
	}

	// Éviter la division par zéro
	if totalWeight == 0 {
		return d.Position
	}

	return models.Position{
		X: sumX / totalWeight,
		Y: sumY / totalWeight,
	}
}

// calculateZoneWeight calcule le poids d'une zone spécifique
func calculateZoneWeight(d *Drone, zone models.Position, params WeightedParameters) float64 {
	// Distance entre le drones et la zones
	distance := zone.CalculateDistance(d.Position)

	// Facteur de distance inversé (plus proche = plus important)
	distanceFactor := 1.0 / (1.0 + distance)

	// Facteur de batterie (considère la batterie nécessaire pour atteindre la zone)
	batteryFactor := 1.0 - (distance * 1) // Simplifié: 1 unité de batterie par unité de distance
	if batteryFactor < 0 {
		batteryFactor = 0
	}

	// Facteur de clustering (nombre de zones proches)
	clusterFactor := calculateClusterFactor(d, zone)

	// Combiner tous les facteurs avec leurs poids
	return (distanceFactor * params.DistanceWeight) +
		(batteryFactor * params.BatteryWeight) +
		(clusterFactor * params.ClusteringWeight)
}

// calculateClusterFactor évalue combien de zones sont proches de la zone donnée
func calculateClusterFactor(d *Drone, targetZone models.Position) float64 {
	const proximityThreshold = 2.0 // Distance considérée comme "proche"
	nearbyZones := 0.0

	for _, zone := range d.ReportedZonesByCentrale {
		if zone.CalculateDistance(targetZone) < proximityThreshold {
			nearbyZones += 1.0
		}
	}

	return nearbyZones / float64(len(d.ReportedZonesByCentrale))
}
