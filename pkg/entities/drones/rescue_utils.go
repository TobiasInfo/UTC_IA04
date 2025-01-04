package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
)

type Rescuer struct {
	ID          int
	Position    models.Position
	Person      *persons.Person
	MedicalTent models.Position
	State       int  // 0 = going to person, 1 = returning to tent
	Active      bool // Tracks if Rescuer is currently on a mission
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
		if rescuer.Position.CalculateDistance(rescuer.Person.Position) <= 1 {
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