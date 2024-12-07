package drones

import (
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
	"math/rand"
)

//TODO il faut un channel par drones pour que la centrale puisse communiquer de manière
//précise avec ses drones, jsp comment on peut implémenter ça

// SurveillanceDrone represents a drones in the simulation
type Drone struct {
	ID                      int
	visionCapacity          int
	Position                models.Position
	Battery                 float64
	DetectionPrecisionFunc  func() []float64
	DroneSeeFunction        func(d *Drone) []*persons.Person
	ReportedZonesByCentrale []models.Position
	SeenPeople              []*persons.Person
	MoveChan                chan models.MovementRequest
}

// NewSurveillanceDrone creates a new instance of SurveillanceDrone
func NewSurveillanceDrone(id int, position models.Position, battery float64, detectionFunc func() []float64, droneSeeFunc func(d *Drone) []*persons.Person, moveChan chan models.MovementRequest) Drone {
	return Drone{
		ID:                      id,
		Position:                position,
		Battery:                 battery,
		DetectionPrecisionFunc:  detectionFunc,
		DroneSeeFunction:        droneSeeFunc,
		ReportedZonesByCentrale: []models.Position{},
		SeenPeople:              []*persons.Person{},
		MoveChan:                moveChan,
	}
}

// Move updates the drones's position to the destination
func (d *Drone) Move(target models.Position) bool {
	if d.Battery <= 0 {
		//fmt.Printf("Drone %d cannot move. Battery is empty.\n", d.ID)
		return false
	}

	//fmt.Printf("Trying to move drones %d to %v\n", d.ID, target)

	if d.Position.X == target.X && d.Position.Y == target.Y {
		return false
	}

	responseChan := make(chan models.MovementResponse)
	d.MoveChan <- models.MovementRequest{MemberID: d.ID, MemberType: "drone", NewPosition: target, ResponseChan: responseChan}
	response := <-responseChan

	if response.Authorized {
		d.Battery -= 1.0
		d.Position.X = target.X
		d.Position.Y = target.Y
		//fmt.Printf("Drone %d moved to %v\n", d.ID, d.Position)
		return true
	}

	return false
}

// DetectIncident simulates the detection of incidents in the drones's vicinity
func (d *Drone) DetectIncident() map[models.Position][2]int {
	detectedIncidents := make(map[models.Position][2]int)

	//A voir comment on gère le radius de vision
	radius := 3

	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			position := models.Position{
				X: d.Position.X + float64(x),
				Y: d.Position.Y + float64(y),
			}

			distress := rand.Intn(5) // Random number of people in distress
			//TODO
			//distres := CheckMapDistress(position, d.DetectionPrecisionFunc)

			people := rand.Intn(20) // Random number of people in the zone
			//TODO
			//people := CheckMapCrowdMember(position)

			detectedIncidents[position] = [2]int{distress, people}
		}
	}

	//fmt.Printf("Drone %d detected incidents: %v\n", d.ID, detectedIncidents)
	return detectedIncidents
}

func (d *Drone) ReceiveInfo() {
	//Recupère les informations qui lui ont été envoyées lors des tours précédents
	//J'imagine qu'on fait marcher ça avec un channel associé à chaque drones pour la réception

	//Lire les informations sur le channel jusqu'à ce qu'il soit vide, garder la dernière carte reçue
	var infoReception []models.Position
	//infoReception = readChannel()

	infos := d.DroneSeeFunction(d)
	seenPeople := make([]*persons.Person, 0)

	for _, info := range infos {
		seenPeople = append(seenPeople, info)
	}

	d.SeenPeople = seenPeople

	d.ReportedZonesByCentrale = infoReception

}

func (d *Drone) Think() models.Position {
	directions := []models.Position{
		{X: 0, Y: -1}, // Up
		{X: 0, Y: 1},  // Down
		{X: -1, Y: 0}, // Left
		{X: 1, Y: 0},  // Right
	}

	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		target := models.Position{
			X: d.Position.X + dir.X,
			Y: d.Position.Y + dir.Y,
		}
		if target.X >= 0 && target.Y >= 0 && target.X < 30 && target.Y < 20 {
			//fmt.Printf("Drone %d Thinks Target: (%f, %f)\n", d.ID, target.X, target.Y)
			return target
		}
	}
	//fmt.Printf("Drone %d has no valid moves, staying at (%f, %f)\n", d.ID, d.Position.X, d.Position.Y)
	return d.Position
}

func (d *Drone) Myturn() {
	// Get the next position to move to
	target := d.Think()

	// Try to move to the calculated target
	moved := d.Move(target)

	if !moved {
		fmt.Printf("Drone %d could not move to %v\n", d.ID, target)
	}
}

//ICI METHODE DE CALCUL, on cherche le centre selon des poids qu'on peut fixer, d'intéret
//c'est une somme pondérée des coordonnées des zones d'intérêt. C'est pas idéal car si tout est autour de lui mais equidistant, il reste sur place
//Si il n'a pas d'info de la centrale comme actuellement, il reste sur place

type WeightedParameters struct {
	DistanceWeight   float64 // Poids pour la distance aux zones
	BatteryWeight    float64 // Poids pour considérer la consommation de batterie
	ClusteringWeight float64 // Poids pour favoriser les zones avec plus de points proches
}

func (d *Drone) CalculateOptimalPosition(params WeightedParameters) models.Position {
	if len(d.ReportedZonesByCentrale) == 0 {
		return d.Position // Si pas de zones, reste sur place
	}

	// Calculer le centre de gravité pondéré des zones reportées
	var sumX, sumY, totalWeight float64

	for _, zone := range d.ReportedZonesByCentrale {
		// Calculer le poids pour cette zone
		weight := calculateZoneWeight(d, zone, params)

		sumX += zone.X * weight
		sumY += zone.Y * weight
		totalWeight += weight
	}

	// Éviter la division par zéro
	if totalWeight == 0 {
		return d.Position
	}

	return models.Position{
		X: sumX / totalWeight,
		Y: sumY / totalWeight,
	}
}

// calculateZoneWeight calcule le poids d'une zone spécifique
func calculateZoneWeight(d *Drone, zone models.Position, params WeightedParameters) float64 {
	// Distance entre le drones et la zones
	distance := calculateDistance(d.Position, zone)

	// Facteur de distance inversé (plus proche = plus important)
	distanceFactor := 1.0 / (1.0 + distance)

	// Facteur de batterie (considère la batterie nécessaire pour atteindre la zone)
	batteryFactor := 1.0 - (distance * 1) // Simplifié: 1 unité de batterie par unité de distance
	if batteryFactor < 0 {
		batteryFactor = 0
	}

	// Facteur de clustering (nombre de zones proches)
	clusterFactor := calculateClusterFactor(d, zone)

	// Combiner tous les facteurs avec leurs poids
	return (distanceFactor * params.DistanceWeight) +
		(batteryFactor * params.BatteryWeight) +
		(clusterFactor * params.ClusteringWeight)
}

// calculateDistance calcule la distance euclidienne entre deux positions
func calculateDistance(pos1, pos2 models.Position) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// calculateClusterFactor évalue combien de zones sont proches de la zone donnée
func calculateClusterFactor(d *Drone, targetZone models.Position) float64 {
	const proximityThreshold = 2.0 // Distance considérée comme "proche"
	nearbyZones := 0.0

	for _, zone := range d.ReportedZonesByCentrale {
		if calculateDistance(targetZone, zone) < proximityThreshold {
			nearbyZones += 1.0
		}
	}

	return nearbyZones / float64(len(d.ReportedZonesByCentrale))
}
