package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/entities/rescue"
	"UTC_IA04/pkg/models"
	"fmt"
	"sync"
)

func (d *Drone) initProtocol4() {
	//Calculate Patrol Path.
	d.Memory.DronePatrolPath = append(d.Memory.DronePatrolPath, models.Position{X: d.MyWatch.CornerBottomLeft.X, Y: d.MyWatch.CornerTopRight.Y})
	d.Memory.DroneActualTarget = models.Position{X: d.MyWatch.CornerBottomLeft.X, Y: d.MyWatch.CornerTopRight.Y}

	fmt.Printf("[DRONES] - Succeffuly terminated Protocole 3 init.\n")
}

/*

Fonctionnement du protocole 4 :

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
		- J'essaye de voir si je peux envoyer l'information à un drone qui est dans mon network.
			- Un network est un sous-ensemble de drones qui peuvent communiquer entre eux, ils sont chainées et ils forment un sous-graphe.
		- Si je ne peux pas, je prends le drone le plus proche dans mon network en terme de distance d'un RP et je lui transfère la resposnabilité de sauver les personnes.

Step 3 :
- Je bouge vers le rescue point si je suis le drone le plus proche.
*/

func (d *Drone) ThinkProtocol4() models.Position {
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
					fmt.Printf("[DRONE %d] Person %d will not be rescued by RescuePoint %d -- ERROR : %v\n",
						d.ID, person.ID, response.RescuePointID, response.Error)
				}
				return true
			})
		}

		dronesDistancesInRP := make(map[*Drone]float64)
		dronesDistancesInRP[d] = d.Position.CalculateDistance(rp.Position)

		if !canCommunicate {
			responsabilityTransfered := false
			for index := range d.DroneNetwork {
				friend := d.DroneNetwork[index]
				rpFriend := d.GetRescuePoint(friend.Position)
				friendRpCalculatePosition := rpFriend.Position.CalculateDistance(friend.Position)
				friendCanCommunicate := friendRpCalculatePosition <= float64(d.DroneCommRange)
				dronesDistancesInRP[friend] = friendRpCalculatePosition
				if friendCanCommunicate {
					d.Memory.Persons.PersonsToSave.Range(func(key, value interface{}) bool {
						friend.Memory.Persons.PersonsToSave.Store(key, value)
						return true
					})
					d.Memory.Persons.PersonsToSave = sync.Map{}
					responsabilityTransfered = true
					break
				}
				// Si le drone ne peut pas communiquer, regarder si un drone voisin à lui peut communiquer.
			}

			if !responsabilityTransfered {
				// Find the closest drone to the RP
				closestDrone := d
				for drone, distance := range dronesDistancesInRP {
					if distance < dronesDistancesInRP[closestDrone] {
						closestDrone = drone
					}
				}
				if closestDrone.ID == d.ID {
					fmt.Printf("[DRONE %d] Responsability not transfered to any drone, moving to RP %d\n", d.ID, rp.ID)
					return d.nextStepToPos(rp.Position)
				}

				fmt.Printf("[DRONE %d] Responsability transfered to drone %d, moving to RP %d\n", d.ID, closestDrone.ID, rp.ID)
				d.Memory.Persons.PersonsToSave.Range(func(key, value interface{}) bool {
					closestDrone.Memory.Persons.PersonsToSave.Store(key, value)
					return true
				})
				d.Memory.Persons.PersonsToSave = sync.Map{}
				responsabilityTransfered = true
			}
		}

	}

	fmt.Printf("[DRONE-WARNING] - Cannot find any RP. Is your Map Config correct?")
	return d.patrolMovementLogic()
}
