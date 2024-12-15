package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type DetectedPersonInfo struct {
	PersonID       int
	Position       models.Position
	DetectionTime  time.Time
	DetectingDrone int
	InDistress     bool
}

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
	DetectedPeople          map[int]DetectedPersonInfo
	LastVisitedPositions    []models.Position
	maxPositionHistory      int
	mu                      sync.RWMutex
}

func NewSurveillanceDrone(
	id int,
	position models.Position,
	battery float64,
	droneSeeRange int,
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
		DetectedPeople:          make(map[int]DetectedPersonInfo),
		LastVisitedPositions:    make([]models.Position, 0),
		maxPositionHistory:      20,
	}
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
		dechargingStep := 1.0
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

func (d *Drone) ShareDetectedPeople() {
	d.mu.Lock()
	for _, person := range d.SeenPeople {
		if person.IsInDistress() {
			info := DetectedPersonInfo{
				PersonID:       person.ID,
				Position:       person.Position,
				DetectionTime:  time.Now(),
				DetectingDrone: d.ID,
				InDistress:     true,
			}
			d.DetectedPeople[person.ID] = info
			
			for _, nearbyDrone := range d.DroneInComRange {
				nearbyDrone.ReceivePersonInfo(info)
			}
		}
	}
	d.mu.Unlock()
}

func (d *Drone) ReceivePersonInfo(info DetectedPersonInfo) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if existing, exists := d.DetectedPeople[info.PersonID]; !exists || 
		existing.DetectionTime.Before(info.DetectionTime) {
		d.DetectedPeople[info.PersonID] = info
	}
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

func (d *Drone) calculateWeightedMove() models.Position {
	type cellWeight struct {
		pos    models.Position
		weight float64
	}
	
	var possibleMoves []cellWeight
	currentPos := d.Position
	
	for x := -1.0; x <= 1.0; x++ {
		for y := -1.0; y <= 1.0; y++ {
			if x == 0 && y == 0 {
				continue
			}
			
			newPos := models.Position{
				X: currentPos.X + x,
				Y: currentPos.Y + y,
			}
			
			if newPos.X < 0 || newPos.Y < 0 || newPos.X >= 30 || newPos.Y >= 20 {
				continue
			}
			
			weight := d.calculatePositionWeight(newPos)
			possibleMoves = append(possibleMoves, cellWeight{
				pos:    newPos,
				weight: weight,
			})
		}
	}
	
	if len(possibleMoves) == 0 {
		return d.Position
	}
	
	sort.Slice(possibleMoves, func(i, j int) bool {
		return possibleMoves[i].weight > possibleMoves[j].weight
	})
	
	d.mu.Lock()
	d.LastVisitedPositions = append(d.LastVisitedPositions, d.Position)
	if len(d.LastVisitedPositions) > d.maxPositionHistory {
		d.LastVisitedPositions = d.LastVisitedPositions[1:]
	}
	d.mu.Unlock()
	
	return possibleMoves[0].pos
}

func (d *Drone) calculatePositionWeight(pos models.Position) float64 {
	var weight float64 = 1.0
	
	distanceWeight := 1.0 / (1.0 + pos.CalculateDistance(d.Position))
	weight *= distanceWeight
	
	d.mu.RLock()
	for i, visitedPos := range d.LastVisitedPositions {
		if visitedPos.CalculateDistance(pos) < 2.0 {
			recencyPenalty := float64(len(d.LastVisitedPositions)-i) / float64(len(d.LastVisitedPositions))
			weight *= (1.0 - recencyPenalty*0.5)
		}
	}
	d.mu.RUnlock()
	
	for _, otherDrone := range d.DroneInComRange {
		if otherDrone.ID != d.ID {
			droneDistance := pos.CalculateDistance(otherDrone.Position)
			if droneDistance < float64(d.DroneCommRange) {
				weight *= droneDistance / float64(d.DroneCommRange)
			}
		}
	}
	
	chargerPos, minDistance := d.closestPOI(models.ChargingStation)
	distanceToCharger := pos.CalculateDistance(chargerPos)
	if d.Battery < 40.0 && distanceToCharger < minDistance {
		weight *= 1.5
	}
	
	weight *= 0.9 + rand.Float64()*0.2
	
	return weight
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
		return pos
	}

	if d.Objectif != (models.Position{}) {
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
				}
			}
		}
		step := d.nextStepToPos(d.Objectif)
		return step
	}

	var step models.Position
	for _, person := range d.SeenPeople {
		if person.IsInDistress() {
			d.PeopleToSave = person
			medicalTentPos, _ := d.closestPOI(models.MedicalTent)
			_, distanceToCharging := d.closestPOI(models.ChargingStation)
			distanceToTent := d.Position.CalculateDistance(medicalTentPos)
			distancePersonToTent := person.Position.CalculateDistance(medicalTentPos)
			totalBatteryNeeded := distanceToTent + distancePersonToTent + distanceToCharging + 2
			if d.Battery >= totalBatteryNeeded {
				fmt.Println("MEDICAL : Drone", d.ID, "has enough battery to complete the mission")
				fmt.Println("MEDICAL : Drone", d.ID, "detected person", person.ID, "in distress")
				d.Objectif = medicalTentPos
				step = d.nextStepToPos(d.Objectif)
				return step
			}
		}
	}
	return d.calculateWeightedMove()
}

func (d *Drone) cleanupOldDetections() {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	threshold := time.Now().Add(-5 * time.Minute)
	for id, info := range d.DetectedPeople {
		if info.DetectionTime.Before(threshold) {
			delete(d.DetectedPeople, id)
		}
	}
}

func (d *Drone) Myturn() {
	d.SeenPeople = []*persons.Person{}
	d.DroneInComRange = []*Drone{}

	if d.tryCharging() {
		d.ReceiveInfo()
		return
	}
	d.ReceiveInfo()

	// Share detected people info with nearby drones
	d.ShareDetectedPeople()
	
	// Clean up old detected people information
	d.cleanupOldDetections()

	target := d.Think()

	moved := d.Move(target)

	if !moved {
		fmt.Printf("Drone %d could not move to %v\n", d.ID, target)
	}
}