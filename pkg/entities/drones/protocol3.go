package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/entities/rescue"
	"UTC_IA04/pkg/models"
	"fmt"
	"sync"
)

type Protocol3 struct {
	PersonsToSave sync.Map // map[int]bool // *Person -> isBeingRescued
}

func (d *Drone) initProtocol3() {
	d.ProtocolStruct = Protocol3{
		PersonsToSave: sync.Map{},
	}
}

/*

Fonctionnement du protocole 3 :

Step 1 :
- Je scanne les personnes en danger
- Si je vois une personne en danger, je la sauvegarde.

Step 2 :
- J'essaye de communiquer avec un RP si un RP est dans mon rayon de communication.
   - Si aucun RP n'est dans mon rayon de communication, je bouge vers le rescue point le plus proche, à chaque mouvement j'essaye de transmettre si un drone est à côté d'un RP.
- Je bouge vers le rescue point si je ne peux pas communiquer.
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescurer à ma place.
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.
*/

func (d *Drone) ThinkProtocol3() models.Position {
	if d.IsCharging {
		// Drone AFK quand il charge car il est docké.
		return d.Position
	}

	for _, person := range d.SeenPeople {
		if person.IsInDistress() {
			d.ProtocolStruct.PersonsToSave.Store(person.ID, person)
		}
	}

	isEmpty := true
	d.ProtocolStruct.PersonsToSave.Range(func(_, _ interface{}) bool {
		isEmpty = false
		return false
	})

	if isEmpty {
		// Je patrouille
		return d.randomMovement()
	}

	if rp := d.GetRescuePoint(d.Position); rp != nil {
		canCommunicate := rp.Position.CalculateDistance(d.Position) <= float64(d.DroneCommRange)
		if canCommunicate {
			//var idPersonsToDelete []int
			d.ProtocolStruct.PersonsToSave.Range(func(key, value interface{}) bool {
				person := value.(*persons.Person)
				respChan := make(chan rescue.RescueResponse)
				rp.RequestChan <- rescue.RescueRequest{
					PersonID:      person.ID,
					Position:      person.Position,
					DroneSenderID: d.ID,
					ResponseChan:  respChan,
				}
				response := <-respChan
				if response.Accepted {
					d.ProtocolStruct.PersonsToSave.Delete(person.ID)
				} else {
					fmt.Printf("[DRONE %d] Person %d will not be rescued by RescuePoint %d -- ERROR : %v\n",
						d.ID, person.ID, response.RescuePointID, response.Error)
				}
				return true
			})
		}

		if !canCommunicate {
			responsabilityTransfered := false
			for index := range d.DroneInComRange {
				friend := d.DroneInComRange[index]
				rpFriend := d.GetRescuePoint(friend.Position)
				friendCanCommunicate := rpFriend.Position.CalculateDistance(friend.Position) <= float64(d.DroneCommRange)
				if friendCanCommunicate {
					d.ProtocolStruct.PersonsToSave.Range(func(key, value interface{}) bool {
						friend.ProtocolStruct.PersonsToSave.Store(key, value)
						return true
					})
					d.ProtocolStruct.PersonsToSave = sync.Map{}
					responsabilityTransfered = true
					break
				}

			}

			if !responsabilityTransfered {
				fmt.Printf("[DRONE %d] Responsability not transfered to any drone, moving to RP %d\n", d.ID, rp.ID)
				return d.nextStepToPos(rp.Position)
			}
		}

	}

	// fmt.Printf("[DRONE %d] Person %d is in distress and not assigned\n", d.ID, person.ID)
	// if rp := d.GetRescuePoint(person.Position); rp != nil {
	// 	// Je bouge vers le rescue point si je ne peux pas communiquer.
	// 	respChan := make(chan rescue.RescueResponse)
	// 	rp.RequestChan <- rescue.RescueRequest{
	// 		PersonID:     person.ID,
	// 		Position:     person.Position,
	// 		ResponseChan: respChan,
	// 	}

	// 	fmt.Printf("[RP] Request sent to RescuePoint %d\n", rp.ID)

	// 	response := <-respChan

	// 	fmt.Printf("[RP] Response received from RescuePoint %d: %v\n", rp.ID, response)
	// 	if response.Accepted {
	// 		person.AssignedDroneID = &d.ID // Marquer comme pris en charge
	// 		fmt.Printf("[DRONE %d] Person %d will be rescued by RescuePoint %d\n",
	// 			d.ID, person.ID, response.RescuePointID)
	// 	} else {
	// 		fmt.Printf("[DRONE %d] Person %d will not be rescued by RescuePoint %d\n",
	// 			d.ID, person.ID, response.RescuePointID)
	// 	}
	// }

	return d.randomMovement()
}
