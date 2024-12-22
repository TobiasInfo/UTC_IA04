package rescue

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"time"
)

type Rescuer struct {
	ID        int
	Position  models.Position
	Person    *persons.Person
	HomePoint models.Position
	State     RescuerState
	Active    bool
}

type RescuerState int

const (
	Idle RescuerState = iota
	MovingToPerson
	ReturningToBase
)

func (rp *RescuePoint) updateRescuers() {
	for {
		for index := range rp.Rescuers {
			rescuer := rp.Rescuers[index]
			if rescuer.State == MovingToPerson {
				// Faire bouger jusqu'à la personne et mettre le rescuer en inactif
				if rescuer.Position.CalculateDistance(rescuer.Person.Position) <= 1 {
					rescueResponse := make(chan models.RescuePeopleResponse)
					rp.SavePersonByRescuer <- models.RescuePeopleRequest{
						PersonID:      rescuer.Person.ID,
						RescuerID:     rescuer.ID,
						RescuePointID: rp.ID,
						ResponseChan:  rescueResponse,
					}

					select {
					case response := <-rescueResponse:
						if response.Authorized {
							fmt.Printf("[RESCUER] Successfully healed person %d\n", rescuer.Person.ID)
						}
					case <-time.After(1 * time.Second):
						fmt.Printf("[RESCUER] Timeout while waiting for response for person %d\n", rescuer.Person.ID)
					}

					personID := rescuer.Person.ID

					rp.ActiveMissions.Delete(rescuer.Person.ID)
					for i, _ := range rp.Rescuers {
						tempRescuer := rp.Rescuers[i]
						if tempRescuer.Person != nil {
							if tempRescuer.Person.ID == personID {
								tempRescuer.Person = nil
								tempRescuer.State = ReturningToBase
							}
						}
					}
				} else {
					// Move one step closer to person
					rescuer.Position = stepTowards(rescuer.Position, rescuer.Person.Position)
				}
			}
			if rescuer.State == ReturningToBase {
				// Faire bouger jusqu'à la base et mettre le rescuer en inactif
				if rescuer.Position.CalculateDistance(rescuer.HomePoint) <= 1 {
					rescuer.State = Idle
					rescuer.Person = nil
					rescuer.Position = models.Position{X: rescuer.HomePoint.X, Y: rescuer.HomePoint.Y}
					rescuer.Active = false
				} else {
					rescuer.Position = stepTowards(rescuer.Position, rescuer.HomePoint)
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func stepTowards(from models.Position, to models.Position) models.Position {
	direction := models.Position{
		X: to.X - from.X,
		Y: to.Y - from.Y,
	}

	nextStep := models.Position{
		X: from.X,
		Y: from.Y,
	}

	if direction.X > 0 {
		nextStep.X += 1
	} else if direction.X < 0 {
		nextStep.X -= 1
	}

	if direction.Y > 0 {
		nextStep.Y += 1
	} else if direction.Y < 0 {
		nextStep.Y -= 1
	}

	return nextStep
}
