package simulation

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
	"math/rand"
)

//TODO il faut un channel par drone pour que la centrale puisse communiquer de manière
//précise avec ses drones, jsp comment on peut implémenter ça

// SurveillanceDrone represents a drone in the simulation
type SurveillanceDrone struct {
	ID                      int
	visionCapacity          int
	Position                models.Position
	Battery                 float64
	DetectionPrecisionFunc  func() []float64
	droneSeeFunction        func(d *SurveillanceDrone) []CrowdMember
	ReportedZonesByCentrale []models.Position
}

// NewSurveillanceDrone creates a new instance of SurveillanceDrone
func NewSurveillanceDrone(id int, position models.Position, battery float64, detectionFunc func() []float64, droneSeeFunc func(d *SurveillanceDrone) []CrowdMember) *SurveillanceDrone {
	return &SurveillanceDrone{
		ID:                      id,
		Position:                position,
		Battery:                 battery,
		DetectionPrecisionFunc:  detectionFunc,
		droneSeeFunction:        droneSeeFunc,
		ReportedZonesByCentrale: []models.Position{},
	}
}

// Move updates the drone's position to the destination
func (d *SurveillanceDrone) Move(destination models.Position) {
	if d.Battery <= 0 {
		fmt.Printf("Drone %d cannot move. Battery is empty.\n", d.ID)
		return
	}

	fmt.Printf("Drone %d moving from %v to %v\n", d.ID, d.Position, destination)
	d.Position = destination
	d.Battery -= 1.0 // Simulate battery usage
}

// DetectIncident simulates the detection of incidents in the drone's vicinity
func (d *SurveillanceDrone) DetectIncident() map[models.Position][2]int {
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

func (d *SurveillanceDrone) ReceiveInfo() {
	//Recupère les informations qui lui ont été envoyées lors des tours précédents
	//J'imagine qu'on fait marcher ça avec un channel associé à chaque drone pour la réception

	//Lire les informations sur le channel jusqu'à ce qu'il soit vide, garder la dernière carte reçue
	var infoReception []models.Position
	//infoReception = readChannel()

	infos := d.droneSeeFunction(d)
	for _, info := range infos {
		fmt.Printf("Drone %d sees crowd member %d at distance %.2f \n", d.ID, info.ID, d.Position.CalculateDistance(info.Position))
	}

	d.ReportedZonesByCentrale = infoReception

}

func (d *SurveillanceDrone) Think() models.Position {
	//Traite les informations reçues, les compare à ses objectifs, trouve son nouveau but et retourne la case adjacente vers laquelle il veut aller

	if d.Battery < 20 {
		//On va à la base, peu importe la situation
		return models.Position{X: 0, Y: 0}

	}

	zoneASurveiller := d.CalculateOptimalPosition(WeightedParameters{1.0, 0.5, 1.5})

	var nouvellePosition models.Position

	if d.Position.X < zoneASurveiller.X {
		nouvellePosition.X = 1.0
	} else if d.Position.X > zoneASurveiller.X {
		nouvellePosition.X = -1.0
	} else {
		nouvellePosition.X = 0.0
	}
	if d.Position.Y < zoneASurveiller.Y {
		nouvellePosition.Y = 1.0
	} else if d.Position.Y > zoneASurveiller.Y {
		nouvellePosition.Y = -1.0
	} else {
		nouvellePosition.Y = 0.0
	}

	return nouvellePosition

}

// Communicate sends data to a centrale or another protocol
func (d *SurveillanceDrone) SendInfo(protocol string, observation map[models.Position][2]int) {
	switch protocol {
	case "Centrale":
		fmt.Printf("Drone %d communicating with Centrale.\n", d.ID)
		// Add logic for communication
	default:
		fmt.Printf("Drone %d: Unknown protocol '%s'\n", d.ID, protocol)
	}
}

func (d *SurveillanceDrone) Myturn() {

	d.ReceiveInfo()

	nouvelleCase := d.Think()

	d.Move(nouvelleCase)

	observation := d.DetectIncident()

	d.SendInfo("Centrale", observation)

	return

}

//ICI METHODE DE CALCUL, on cherche le centre selon des poids qu'on peut fixer, d'intéret
//c'est une somme pondérée des coordonnées des zones d'intérêt. C'est pas idéal car si tout est autour de lui mais equidistant, il reste sur place
//Si il n'a pas d'info de la centrale comme actuellement, il reste sur place

type WeightedParameters struct {
	DistanceWeight   float64 // Poids pour la distance aux zones
	BatteryWeight    float64 // Poids pour considérer la consommation de batterie
	ClusteringWeight float64 // Poids pour favoriser les zones avec plus de points proches
}

func (d *SurveillanceDrone) CalculateOptimalPosition(params WeightedParameters) models.Position {
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
func calculateZoneWeight(d *SurveillanceDrone, zone models.Position, params WeightedParameters) float64 {
	// Distance entre le drone et la zones
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
func calculateClusterFactor(d *SurveillanceDrone, targetZone models.Position) float64 {
	const proximityThreshold = 2.0 // Distance considérée comme "proche"
	nearbyZones := 0.0

	for _, zone := range d.ReportedZonesByCentrale {
		if calculateDistance(targetZone, zone) < proximityThreshold {
			nearbyZones += 1.0
		}
	}

	return nearbyZones / float64(len(d.ReportedZonesByCentrale))
}
