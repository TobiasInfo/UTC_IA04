package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
	"math/rand"
)

// RandomMovement calcule le prochain mouvement aléatoire pour un drone
func (d *Drone) randomMovement() models.Position {
	// Define all possible directions
	directions := []models.Position{
		{X: 0, Y: -1}, {X: 0, Y: 1},
		{X: -1, Y: 0}, {X: 1, Y: 0},
		{X: -1, Y: -1}, {X: 1, Y: 1},
		{X: 1, Y: -1}, {X: -1, Y: 1},
	}

	// Si des personnes sont visibles, privilégier ces directions
	if len(d.SeenPeople) > 0 {
		directionScores := make(map[models.Position]float64)
		for _, dir := range directions {
			potentialPos := models.Position{
				X: d.Position.X + dir.X,
				Y: d.Position.Y + dir.Y,
			}

			if !d.isValidPosition(potentialPos) {
				continue
			}

			score := calculateDirectionScore(d, potentialPos)
			directionScores[dir] = score
		}

		if bestDir, bestScore := findBestDirection(directionScores); bestScore > 0 {
			return models.Position{
				X: d.Position.X + bestDir.X,
				Y: d.Position.Y + bestDir.Y,
			}
		}
	}

	// Mouvement aléatoire si aucune personne n'est visible
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		target := models.Position{
			X: d.Position.X + dir.X,
			Y: d.Position.Y + dir.Y,
		}

		if d.isValidPosition(target) {
			return target
		}
	}

	fmt.Printf("Drone %d is at (%.0f, %.0f) and he's stuck (Oh no step-brother I am stuck in the washing machine :c)\n", d.ID, d.Position.X, d.Position.Y)
	return d.nextStepToPos(models.Position{X: d.MyWatch.CornerBottomLeft.X, Y: d.MyWatch.CornerBottomLeft.Y})
}

func (d *Drone) isValidPosition(pos models.Position) bool {
	// Vérifier les limites de la carte
	if pos.X <= 0 || pos.Y <= 0 ||
		math.Round(pos.X) > float64(d.MapWidth) ||
		math.Round(pos.Y) > float64(d.MapHeight) {
		return false
	}

	// Vérifier les limites de la watch
	if pos.X <= d.MyWatch.CornerBottomLeft.X || pos.Y <= d.MyWatch.CornerBottomLeft.Y ||
		pos.X >= d.MyWatch.CornerTopRight.X || pos.Y >= d.MyWatch.CornerTopRight.Y {
		return false
	}

	return true
}

func calculateDirectionScore(d *Drone, pos models.Position) float64 {
	score := 0.0
	for _, person := range d.SeenPeople {
		dist := pos.CalculateDistance(person.Position)
		score += 1.0 / (dist + 1)
		if person.InDistress {
			score += 3.0
		}
	}

	//for _, drone := range d.DroneInComRange {
	//	score -= 1.0 / (pos.CalculateDistance(drone.Position) + 1)
	//}
	return score
}

func findBestDirection(scores map[models.Position]float64) (models.Position, float64) {
	var bestDir models.Position
	bestScore := -1.0
	for dir, score := range scores {
		if score > bestScore {
			bestScore = score
			bestDir = dir
		}
	}
	return bestDir, bestScore
}

// FindBestDroneForRescue trouve le meilleur drone pour sauver une personne
func findBestDroneForRescue(drones []*Drone, person *persons.Person) *Drone {
	var bestDrone *Drone
	minCost := math.Inf(1)

	for _, dr := range drones {
		if dr.PeopleToSave != nil {
			continue
		}

		medicalTentPos, _ := dr.closestPOI(models.MedicalTent)
		_, distanceToCharging := dr.closestPOI(models.ChargingStation)
		distanceToTent := dr.Position.CalculateManhattanDistance(medicalTentPos)
		distanceTentToPerson := medicalTentPos.CalculateManhattanDistance(person.Position)
		totalDistance := distanceToTent + distanceTentToPerson + distanceToCharging + 2

		if dr.Battery >= totalDistance && totalDistance < minCost {
			bestDrone = dr
			minCost = totalDistance
		}
	}
	return bestDrone
}

func (d *Drone) nextStepToPos(pos models.Position) models.Position {
	return stepTowards(d.Position, pos)
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

func (d *Drone) closestPOI(poiType models.POIType) (models.Position, float64) {
	pois := d.MapPoi[poiType]
	minDistance := math.Inf(1)
	var closestPOI models.Position

	for _, poi := range pois {
		distance := d.Position.CalculateManhattanDistance(poi)
		if distance < minDistance {
			minDistance = distance
			closestPOI = poi
		}
	}
	return closestPOI, minDistance
}
