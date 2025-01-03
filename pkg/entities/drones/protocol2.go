package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/entities/rescue"
	"UTC_IA04/pkg/models"
	"fmt"
	"sync"
)

func (d *Drone) initProtocol2() {
	//Calculate Patrol Path.
	d.Memory.DronePatrolPath = append(d.Memory.DronePatrolPath, models.Position{X: d.MyWatch.CornerBottomLeft.X, Y: d.MyWatch.CornerTopRight.Y})
	d.Memory.DroneActualTarget = models.Position{X: d.MyWatch.CornerBottomLeft.X, Y: d.MyWatch.CornerTopRight.Y}

	fmt.Printf("[DRONES] - Succeffuly terminated Protocole 2 init.\n")
}

/*

Fonctionnement du protocole 2 :

Step 0 :

- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescurer à ma place.
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.

Step 1 :
- Je scanne les personnes en danger
- Si je vois une personne en danger, je la sauvegarde.

Step 2 :
- J'essaye de communiquer avec un RP si un RP est dans mon rayon de communication.
   - Si aucun RP n'est dans mon rayon de communication.
		- J'essaye de voir si je peux envoyer l'information à un drone qui est en n+1 de mon rayon de communication.
		- Si je ne peux pas, je bouge vers le rescue point le plus proche.
- Je bouge vers le rescue point si je ne peux pas communiquer.
*/

func (d *Drone) ThinkProtocol2() models.Position {
	if d.IsCharging {
		// Drone AFK quand il charge car il est docké.
		return d.Position
	}

	for _, person := range d.SeenPeople {
		if person.IsInDistress() {
			d.Memory.Persons.PersonsToSave.Store(person.ID, person)
		}
	}

	isEmpty := true
	d.Memory.Persons.PersonsToSave.Range(func(_, _ interface{}) bool {
		isEmpty = false
		return false
	})

	if isEmpty {
		// Je patrouille
		return d.patrolMovementLogic()
	}

	if rp := d.GetRescuePoint(d.Position); rp != nil {
		canCommunicate := rp.Position.CalculateDistance(d.Position) <= float64(d.DroneCommRange)
		if canCommunicate {
			//var idPersonsToDelete []int
			d.Memory.Persons.PersonsToSave.Range(func(key, value interface{}) bool {
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
					d.Memory.Persons.PersonsToSave.Delete(person.ID)
				} else {
					if d.debug {
						fmt.Printf("[DRONE %d] Person %d will not be rescued by RescuePoint %d -- ERROR : %v\n",
							d.ID, person.ID, response.RescuePointID, response.Error)
					}
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
					d.Memory.Persons.PersonsToSave.Range(func(key, value interface{}) bool {
						friend.Memory.Persons.PersonsToSave.Store(key, value)
						return true
					})
					d.Memory.Persons.PersonsToSave = sync.Map{}
					responsabilityTransfered = true
					break
				}

			}

			if !responsabilityTransfered {
				if d.debug {
					fmt.Printf("[DRONE %d] Responsability not transfered to any drone, moving to RP %d\n", d.ID, rp.ID)
				}
				return d.nextStepToPos(rp.Position)
			}
		}

	}
	if d.debug {
		fmt.Printf("[DRONE-WARNING] - Cannot find any RP. Is your Map Config correct?")
	}
	return d.patrolMovementLogic()
}
