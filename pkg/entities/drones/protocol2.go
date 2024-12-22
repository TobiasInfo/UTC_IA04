package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"sync"
)

type Protocol2 struct {
	PersonsToSave sync.Map // map[int]bool // PersonID -> isBeingRescued
}

func (d *Drone) initProtocol2() {
	d.ProtocolStruct = Protocol2{}
}

func (d *Drone) ThinkProtocol2() models.Position {
	if d.hasActiveObjective() {
		return d.handleActiveObjective()
	}

	for _, person := range d.SeenPeople {
		if d.shouldHandlePerson(person) {
			return d.coordinateRescue(person)
		}
	}

	return d.randomMovement()
}

func (d *Drone) hasActiveObjective() bool {
	return d.Objectif != (models.Position{}) && d.PeopleToSave != nil
}

func (d *Drone) coordinateRescue(person *persons.Person) models.Position {
	allDrones := d.GetAllReachableDrones()
	bestDrone := findBestDroneForRescue(allDrones, person)

	if bestDrone == nil {
		return d.randomMovement()
	}

	if bestDrone.ID == d.ID {
		return d.initializeRescueMission(person)
	}

	d.assignRescueToOtherDrone(bestDrone, person)
	return d.randomMovement()
}
