package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/entities/rescue"
	"UTC_IA04/pkg/models"
	"fmt"
)

func (d *Drone) initProtocol1() {}

/*

Fonctionnement du protocole 1 :

Step 1 :
- Je scanne les personnes en danger
- Si je vois une personne en danger, je la sauvegarde.

Step 2 :
- Dès que ma liste est supérieur > 1 je m'en vais vers le RP le + Proche pour régler les problémes.
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescurer à ma place.
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.
*/

func (d *Drone) ThinkProtocol1() models.Position {
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
		return d.randomMovement()
	}

	if rp := d.GetRescuePoint(d.Position); rp != nil {
		canCommunicate := rp.Position.CalculateDistance(d.Position) <= float64(d.DroneCommRange)
		if canCommunicate {
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

		if !canCommunicate {
			fmt.Printf("[DRONE %d] Responsability not transfered to any drone, moving to RP %d\n", d.ID, rp.ID)
			return d.nextStepToPos(rp.Position)
		}

	}

	fmt.Printf("[DRONE-WARNING] - Cannot find any RP. Is your Map Config correct?")
	return d.randomMovement()
}
