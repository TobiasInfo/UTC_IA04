package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"math"
	"sync"
)

type Protocol1 struct {
	PersonsToSave sync.Map // map[int]bool // PersonID -> isBeingRescued
}

func (d *Drone) initProtocol1() {
	d.ProtocolStruct = Protocol1{}
}

func (d *Drone) ThinkProtocol1() models.Position {
	if d.Objectif != (models.Position{}) {
		if d.Position == d.Objectif {
			return d.handleObjectiveReached()
		}
		return d.nextStepToPos(d.Objectif)
	}

	// Check for people in distress
	for _, person := range d.SeenPeople {
		if person.IsInDistress() {
			if person.IsAssigned() && (d.PeopleToSave == nil || d.PeopleToSave.ID != person.ID) {
				continue
			}

			if d.canHandleRescue(person) {
				return d.initializeRescueMission(person)
			}
		}
	}

	return d.randomMovement()
}

func (d *Drone) canHandleRescue(person *persons.Person) bool {
	medicalTentPos, _ := d.closestPOI(models.MedicalTent)
	_, distanceToCharging := d.closestPOI(models.ChargingStation)
	distanceToTent := d.Position.CalculateManhattanDistance(medicalTentPos)
	distancePersonToTent := person.Position.CalculateManhattanDistance(medicalTentPos)
	totalBatteryNeeded := distanceToTent + distancePersonToTent + distanceToCharging + 2

	return d.Battery >= totalBatteryNeeded
}

func (d *Drone) handleObjectiveReached() models.Position {
	medicalTentPos, _ := d.closestPOI(models.MedicalTent)

	if d.Position == medicalTentPos && !d.HasMedicalGear {
		return d.handleMedicalTentArrival()
	}

	if d.Position.X == math.Round(d.PeopleToSave.Position.X) &&
		d.Position.Y == math.Round(d.PeopleToSave.Position.Y) {
		return d.handlePersonRescue()
	}

	return d.Position
}
