package rescue

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"sync"
	"time"
)

type RescuePoint struct {
	ID                   int
	Position             models.Position
	Rescuers             map[int]*Rescuer
	RequestChan          chan RescueRequest
	RPRequestChan        chan RescueRequest
	IsPersonBeingRescued chan RescueRequest
	ResponseChan         chan RescueResponse
	SavePersonByRescuer  chan models.RescuePeopleRequest
	ActiveMissions       sync.Map // map[int]bool // PersonID -> isBeingRescued
	AllRescuePoints      []*RescuePoint
}

type RescueRequest struct {
	PersonID      int
	Position      models.Position
	DroneSenderID int
	ResponseChan  chan RescueResponse
}

type RescueResponse struct {
	Accepted      bool
	RescuePointID int
	Error         error
}

func NewRescuePoint(id int, position models.Position, savePersonByRescuer chan models.RescuePeopleRequest) *RescuePoint {
	fmt.Printf("[RP] New RescuePoint created at position (%.0f, %.0f)\n", position.X, position.Y)
	return &RescuePoint{
		ID:                   id,
		Position:             position,
		Rescuers:             make(map[int]*Rescuer),
		RequestChan:          make(chan RescueRequest),
		RPRequestChan:        make(chan RescueRequest),
		IsPersonBeingRescued: make(chan RescueRequest),
		ResponseChan:         make(chan RescueResponse),
		AllRescuePoints:      make([]*RescuePoint, 0),
		SavePersonByRescuer:  savePersonByRescuer,
	}
}

func (rp *RescuePoint) Start() {
	rp.startWithRecover(rp.handleRequests, "handleRequests")
	rp.startWithRecover(rp.isPersonBeingRescuedByThisRp, "isPersonBeingRescuedByThisRp")
	rp.startWithRecover(rp.updateRescuers, "updateRescuers")
	rp.startWithRecover(rp.handleRPRequests, "handleRPRequests")
}

func (rp *RescuePoint) startWithRecover(f func(), name string) {
	go func() {
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("[ERROR] %s panicked: %v - Restarting...\n", name, r)
						time.Sleep(time.Second) // Petit délai avant redémarrage
					}
				}()
				f()
			}()
		}
	}()
}

func (rp *RescuePoint) isPersonBeingRescuedByThisRp() {
	for req := range rp.IsPersonBeingRescued {
		if _, exists := rp.ActiveMissions.Load(req.PersonID); exists {
			req.ResponseChan <- RescueResponse{
				Accepted: true,
				Error:    fmt.Errorf("person already being rescued"),
			}
		} else {
			req.ResponseChan <- RescueResponse{
				Accepted: false,
			}
		}
	}
}

func (rp *RescuePoint) handleRequests() {
	for req := range rp.RequestChan {
		// Vérifier si la personne est déjà en cours de sauvetage
		if _, exists := rp.ActiveMissions.Load(req.PersonID); exists {
			req.ResponseChan <- RescueResponse{
				Accepted: false,
				Error:    fmt.Errorf("person already being rescued by this rescue point"),
			}
			continue
		}

		personAlreadyBeingRescued := false

		for index := range rp.AllRescuePoints {
			rpTemp := rp.AllRescuePoints[index]
			if rp.ID != rpTemp.ID {
				tempReq := RescueRequest{
					PersonID:     req.PersonID,
					Position:     req.Position,
					ResponseChan: make(chan RescueResponse),
				}
				rpTemp.IsPersonBeingRescued <- tempReq
				response := <-tempReq.ResponseChan
				if response.Accepted {
					req.ResponseChan <- RescueResponse{
						Accepted: false,
						Error:    fmt.Errorf("person already being rescued by another rescue point (%d)", rpTemp.ID),
					}
					personAlreadyBeingRescued = true
					break
				}
			}
		}

		if personAlreadyBeingRescued {
			continue
		}

		// Trouver le point de sauvetage le plus proche
		closestRP := rp.findClosestRescuePoint(req.Position)
		if closestRP.ID != rp.ID {
			closestRP.RPRequestChan <- req
			continue
		}

		// Vérifier la disponibilité d'un rescuer
		rescuer := rp.getAvailableRescuer()
		if rescuer == nil {
			req.ResponseChan <- RescueResponse{
				Accepted: false,
				Error:    fmt.Errorf("no available rescuers"),
			}
			continue
		}

		// Assigner la mission
		rp.assignMission(rescuer, req)
		rp.ActiveMissions.Store(req.PersonID, true)

		req.ResponseChan <- RescueResponse{
			Accepted:      true,
			RescuePointID: rp.ID,
		}

		fmt.Printf("[RP] Mission assigned to RescuePoint %d to Rescue Person : %d by Drone %d\n", rp.ID, req.PersonID, req.DroneSenderID)
	}
}

func (rp *RescuePoint) handleRPRequests() {
	for req := range rp.RPRequestChan {
		// Vérifier la disponibilité d'un rescuer
		rescuer := rp.getAvailableRescuer()
		if rescuer == nil {
			req.ResponseChan <- RescueResponse{
				Accepted: false,
				Error:    fmt.Errorf("no available rescuers"),
			}
			continue
		}

		// Assigner la mission
		rp.assignMission(rescuer, req)
		rp.ActiveMissions.Store(req.PersonID, true)

		req.ResponseChan <- RescueResponse{
			Accepted:      true,
			RescuePointID: rp.ID,
		}

		fmt.Printf("[RP] Mission assigned to RescuePoint %d to Rescue Person : %d by Drone %d\n", rp.ID, req.PersonID, req.DroneSenderID)
	}
}

func (rp *RescuePoint) findClosestRescuePoint(pos models.Position) *RescuePoint {
	if len(rp.AllRescuePoints) == 0 {
		return rp
	}

	// Calculer la distance pour le point actuel
	minDist := rp.Position.CalculateManhattanDistance(pos)
	closestPoints := []*RescuePoint{rp}

	// Parcourir tous les autres points
	for index := range rp.AllRescuePoints {
		point := rp.AllRescuePoints[index]
		dist := point.Position.CalculateManhattanDistance(pos)

		if dist < minDist {
			// Nouvelle distance minimale trouvée
			minDist = dist
			closestPoints = []*RescuePoint{point}
		} else if dist == minDist {
			// Point équidistant trouvé
			closestPoints = append(closestPoints, point)
		}
	}

	// S'il y a plusieurs points à égale distance, prendre celui avec le plus petit ID
	if len(closestPoints) > 1 {
		selectedRP := closestPoints[0]
		for _, point := range closestPoints[1:] {
			if point.ID < selectedRP.ID {
				selectedRP = point
			}
		}
		return selectedRP
	}

	return closestPoints[0]
}

func (rp *RescuePoint) getAvailableRescuer() *Rescuer {
	// D'abord, chercher un rescuer inactif
	for _, rescuer := range rp.Rescuers {
		if !rescuer.Active {
			rescuer.Active = false
			rescuer.State = Idle
			rescuer.Position = rp.Position
			rescuer.HomePoint = rp.Position
			return rescuer
		}
	}

	// Si aucun rescuer n'est disponible, en créer un nouveau
	newRescuerID := len(rp.Rescuers)
	newRescuer := &Rescuer{
		ID:        newRescuerID,
		Position:  rp.Position, // Le nouveau rescuer commence à la position du rescue point
		HomePoint: rp.Position,
		State:     Idle,
		Active:    false,
	}

	rp.Rescuers[newRescuerID] = newRescuer
	return newRescuer
}

func (rp *RescuePoint) assignMission(rescuer *Rescuer, req RescueRequest) {
	rescuer.Active = true
	rescuer.State = MovingToPerson
	rescuer.Person = &persons.Person{
		ID:       req.PersonID,
		Position: req.Position,
	}
	fmt.Printf("[RESCUE POINT %d] Rescuer %d assigned to person %d at position (%.0f, %.0f)\n",
		rp.ID, rescuer.ID, req.PersonID, req.Position.X, req.Position.Y)
}
