package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
)

type Rescuer struct {
	ID          int
	Position    models.Position
	Person      *persons.Person
	MedicalTent models.Position
	State       int  // 0 = going to person, 1 = returning to tent
	Active      bool // Tracks if Rescuer is currently on a mission
}

// initializeRescueMission initialise une mission de sauvetage pour une personne
func (d *Drone) initializeRescueMission(person *persons.Person) models.Position {
	fmt.Printf("[DRONE %d] Prend en charge la personne %d\n", d.ID, person.ID)
	person.AssignedDroneID = &d.ID
	d.PeopleToSave = person
	medicalTentPos, _ := d.closestPOI(models.MedicalTent)
	d.Objectif = medicalTentPos
	return d.nextStepToPos(d.Objectif)
}

// handleMedicalTentArrival gère l'arrivée du drone à la tente médicale
func (d *Drone) handleMedicalTentArrival() models.Position {
	responseChan := make(chan models.MedicalDeliveryResponse)
	d.MedicalDeliveryChan <- models.MedicalDeliveryRequest{
		PersonID:     d.PeopleToSave.ID,
		DroneID:      d.ID,
		ResponseChan: responseChan,
	}
	rep := <-responseChan
	if rep.Authorized {
		fmt.Printf("[DRONE %d] A récupéré le matériel médical pour la personne %d\n", d.ID, d.PeopleToSave.ID)
		d.Objectif = models.Position{X: math.Round(d.PeopleToSave.Position.X), Y: math.Round(d.PeopleToSave.Position.Y)}
		d.HasMedicalGear = true
	}
	return d.Position
}

// handlePersonRescue gère le sauvetage d'une personne
func (d *Drone) handlePersonRescue() models.Position {
	responseSave := make(chan models.SavePersonResponse)
	d.SavePersonChan <- models.SavePersonRequest{
		PersonID:     d.PeopleToSave.ID,
		DroneID:      d.ID,
		ResponseChan: responseSave,
	}
	rep := <-responseSave
	if rep.Authorized {
		fmt.Printf("[DRONE %d] A sauvé la personne %d\n", d.ID, d.PeopleToSave.ID)
		d.PeopleToSave.AssignedDroneID = nil
		d.PeopleToSave = nil
		d.Objectif = models.Position{}
		d.HasMedicalGear = false
	} else {
		fmt.Printf("[DRONE %d] n'a pas pu sauver la personne %d\n", d.ID, d.PeopleToSave.ID)
		d.PeopleToSave.AssignedDroneID = nil
		d.PeopleToSave = nil
		d.Objectif = models.Position{}
	}
	return d.Position
}

// handleActiveObjective gère un objectif actif
func (d *Drone) handleActiveObjective() models.Position {
	if d.Position.X == d.Objectif.X && d.Position.Y == d.Objectif.Y {
		medicalTentPos, _ := d.closestPOI(models.MedicalTent)
		if d.Position == medicalTentPos && !d.HasMedicalGear {
			return d.handleMedicalTentArrival()
		} else if d.Position.X == math.Round(d.PeopleToSave.Position.X) &&
			d.Position.Y == math.Round(d.PeopleToSave.Position.Y) {
			return d.handlePersonRescue()
		}
	}
	return d.nextStepToPos(d.Objectif)
}

// assignRescueToOtherDrone assigne une mission de sauvetage à un autre drone
func (d *Drone) assignRescueToOtherDrone(bestDrone *Drone, person *persons.Person) {
	fmt.Printf("[DRONE %d] Ne gère pas la personne %d, c'est le DRONE %d qui s'en charge\n",
		d.ID, person.ID, bestDrone.ID)
	person.AssignedDroneID = &bestDrone.ID
	medicalTentPos, _ := bestDrone.closestPOI(models.MedicalTent)
	bestDrone.PeopleToSave = person
	bestDrone.Objectif = medicalTentPos
	bestDrone.HasMedicalGear = false
}

// coordinateRescueWithRescuer coordonne le sauvetage avec un sauveteur
func (d *Drone) coordinateRescueWithRescuer(person *persons.Person) models.Position {
	// Si on est déjà assigné à cette personne, continuer la mission
	if d.PeopleToSave != nil && d.PeopleToSave.ID == person.ID {
		medicalTentPos, _ := d.closestPOI(models.MedicalTent)
		return d.nextStepToPos(medicalTentPos)
	}

	fmt.Printf("[DRONE %d] Detected person in distress (ID: %d) at (%.0f, %.0f)\n",
		d.ID, person.ID, person.Position.X, person.Position.Y)

	// Ne considérer que les personnes non assignées
	if person.IsAssigned() {
		return d.randomMovement()
	}

	allDrones := d.GetAllReachableDrones()
	for _, dr := range allDrones {
		if dr.ID != d.ID {
			fmt.Printf("[DRONE %d] Informing DRONE %d about person in distress (ID: %d)\n",
				d.ID, dr.ID, person.ID)
		}
	}

	bestDrone := findBestDroneForRescue(allDrones, person)
	if bestDrone == nil {
		fmt.Printf("[DRONE %d] No drone available for person %d\n", d.ID, person.ID)
		return d.randomMovement()
	}

	if bestDrone.ID == d.ID {
		// Si on est le meilleur drone, on prend la responsabilité
		fmt.Printf("[DRONE %d] Taking responsibility for person %d\n", d.ID, person.ID)
		person.AssignedDroneID = &d.ID
		d.PeopleToSave = person
		medicalTentPos, _ := d.closestPOI(models.MedicalTent)
		return d.nextStepToPos(medicalTentPos)
	}

	// Si un autre drone est meilleur, on le laisse gérer
	fmt.Printf("[DRONE %d] Not handling person %d, DRONE %d will handle it\n",
		d.ID, person.ID, bestDrone.ID)
	return d.randomMovement()
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

func (d *Drone) shouldHandlePerson(person *persons.Person) bool {
	return person.IsInDistress() &&
		(!person.IsAssigned() ||
			(d.PeopleToSave != nil && d.PeopleToSave.ID == person.ID))
}
