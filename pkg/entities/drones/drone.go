package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/entities/rescue"
	"UTC_IA04/pkg/models"
	"fmt"
	"sync"
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
	MyWatch                 models.MyWatch
	ProtocolStruct          struct {
		PersonsToSave sync.Map // map[int]*persons.Person // PersonID -> isBeingRescued
	}
	GetRescuePoint func(pos models.Position) *rescue.RescuePoint
	// Simulation              interface {
	// 	GetRescuePoint(pos models.Position) *rescue.RescuePoint
	// }
}

// NewSurveillanceDrone crée un nouveau drone
func NewSurveillanceDrone(id int,
	position models.Position,
	myWatch models.MyWatch,
	battery float64, droneSeeRange int,
	droneCommunicationRange int,
	droneSeeFunc func(d *Drone) []*persons.Person,
	droneInComRange func(d *Drone) []*Drone,
	getRescuePoint func(pos models.Position) *rescue.RescuePoint,
	moveChan chan models.MovementRequest,
	mapPoi map[models.POIType][]models.Position,
	chargingChan chan models.ChargingRequest,
	medicalDeliveryChan chan models.MedicalDeliveryRequest,
	savePersonChan chan models.SavePersonRequest,
	protocolMode int,
	savePersonByRescuer chan models.RescuePeopleRequest,
	MapWidth int,
	MapHeight int,

) Drone {
	fmt.Printf("[DRONE %d] And now My watch at %v begin !\n", id, myWatch)
	return Drone{
		ID:                      id,
		Position:                position,
		MyWatch:                 myWatch,
		Battery:                 battery,
		DroneSeeRange:           droneSeeRange,
		DroneCommRange:          droneCommunicationRange,
		DroneSeeFunction:        droneSeeFunc,
		DroneInComRangeFunc:     droneInComRange,
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
		GetRescuePoint:          getRescuePoint,
	}
}

func (d *Drone) InitProtocol() {
	switch d.ProtocolMode {
	case 1:
		d.initProtocol1()
	case 2:
		d.initProtocol2()
	case 3:
		d.initProtocol3()
	}
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
			d.Battery -= dechargingStep
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

func (d *Drone) Think() models.Position {
	// Handle battery management first
	pos, goCharging := d.BatteryManagement()
	if goCharging {
		return pos
	}

	switch d.ProtocolMode {
	case 1:
		return d.ThinkProtocol1()
	case 2:
		return d.ThinkProtocol2()
	case 3:
		return d.ThinkProtocol3()
	default:
		return d.randomMovement()
	}
}

func (d *Drone) Myturn() {
	// Cannot communicate with other drones if charging
	d.ReceiveInfo()

	if d.tryCharging() {
		return
	}

	target := d.Think()
	if target.X == d.Position.X && target.Y == d.Position.Y {
		d.Battery -= 0.25
		return
	}
	moved := d.Move(target)
	if !moved {
		fmt.Printf("Drone %d could not move to %v\n", d.ID, target)
	}
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
