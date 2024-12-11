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
	DroneSeeRange           int
	DroneCommRange          int
	Position                models.Position
	Battery                 float64
	DroneSeeFunction        func(d *Drone) []*persons.Person
	DroneInComRangeFunc     func(d *Drone) []*Drone
	ReportedZonesByCentrale []models.Position
	SeenPeople              []*persons.Person
	DroneInComRange         []*Drone
	MoveChan                chan models.MovementRequest
	MapPoi                  map[models.POIType][]models.Position
	ChargingChan            chan models.ChargingRequest
	IsCharging              bool
	MedicalDeliveryChan     chan models.MedicalDeliveryRequest
	MedicalTentTimer        int
	DeploymentTimer         int
}

// NewSurveillanceDrone creates a new instance of SurveillanceDrone
func NewSurveillanceDrone(id int,
	position models.Position,
	battery float64, droneSeeRange int,
	droneCommunicationRange int,
	droneSeeFunc func(d *Drone) []*persons.Person,
	DroneInComRange func(d *Drone) []*Drone,
	moveChan chan models.MovementRequest,
	mapPoi map[models.POIType][]models.Position,
	chargingChan chan models.ChargingRequest,
	medicalDeliveryChan chan models.MedicalDeliveryRequest) Drone {
	return Drone{
		ID:                      id,
		Position:                position,
		Battery:                 battery,
		DroneSeeRange:           droneSeeRange,
		DroneCommRange:          droneCommunicationRange,
		DroneSeeFunction:        droneSeeFunc,
		DroneInComRangeFunc:     DroneInComRange,
		ReportedZonesByCentrale: []models.Position{},
		SeenPeople:              []*persons.Person{},
		DroneInComRange:         []*Drone{},
		MoveChan:                moveChan,
		MapPoi:                  mapPoi,
		ChargingChan:            chargingChan,
		IsCharging:              false,
		MedicalDeliveryChan:     medicalDeliveryChan,
		MedicalTentTimer:        0,
		DeploymentTimer:         1,
	}
}

// Add charging check method
func (d *Drone) tryCharging() bool {
	if d.IsCharging {
		// Already charging, continue
		d.Battery += 5
		if d.Battery >= 80+rand.Float64()*20 { // Random value between 80 and 100
			d.IsCharging = false
			return false
		}
		return true
	}

	// Try to start charging
	responseChan := make(chan models.ChargingResponse)
	d.ChargingChan <- models.ChargingRequest{
		DroneID:      d.ID,
		Position:     d.Position,
		ResponseChan: responseChan,
	}

	response := <-responseChan
	if response.Authorized {
		d.IsCharging = true
		d.Battery += 5
		return true
	}
	return false
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
		// TODO : Adjust this value
		dechargingStep := 0.0
		if d.Battery >= dechargingStep {
			d.Battery -= dechargingStep
		} else {
			d.Battery = 0.0
		}

		d.Position.X = target.X
		d.Position.Y = target.Y
		//fmt.Printf("Drone %d moved to %v\n", d.ID, d.Position)
		return true
	}

	return false
}

// DetectIncident simulates the detection of incidents in the drones's vicinity
// func (d *Drone) DetectIncident() map[models.Position][2]int {
// 	detectedIncidents := make(map[models.Position][2]int)

// 	//A voir comment on gère le radius de vision
// 	radius := 3

// 	for x := -radius; x <= radius; x++ {
// 		for y := -radius; y <= radius; y++ {
// 			position := models.Position{
// 				X: d.Position.X + float64(x),
// 				Y: d.Position.Y + float64(y),
// 			}

// 			distress := rand.Intn(5) // Random number of people in distress
// 			//TODO
// 			//distres := CheckMapDistress(position, d.DetectionPrecisionFunc)

// 			people := rand.Intn(20) // Random number of people in the zone
// 			//TODO
// 			//people := CheckMapCrowdMember(position)

// 			detectedIncidents[position] = [2]int{distress, people}
// 		}
// 	}

// 	//fmt.Printf("Drone %d detected incidents: %v\n", d.ID, detectedIncidents)
// 	return detectedIncidents
// }

func (d *Drone) ReceiveInfo() {
	seenPeople := d.DroneSeeFunction(d)
	droneInComRange := d.DroneInComRangeFunc(d)

	d.SeenPeople = seenPeople
	d.DroneInComRange = droneInComRange
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

func (d *Drone) Think() models.Position {

	closestStation, minDistance := d.closestPOI(models.ChargingStation)
	// fmt.Printf("Drone %d Battery: %.2f Closest Station: (%f, %f) Min Distance: %.2f\n", d.ID, d.Battery, closestStation.X, closestStation.Y, minDistance)

	// If the drone's battery is low enough that it cannot safely move elsewhere, head towards the station.
	// Instead of returning the station's full coordinates, we return just one step in the right direction.
	if d.Battery <= minDistance+5 {
		// fmt.Printf("Drone %d is heading towards the closest charging station\n", d.ID)

		// Compute the step direction towards the charging station
		dx := closestStation.X - d.Position.X
		dy := closestStation.Y - d.Position.Y

		var step models.Position
		// Choose the axis along which the difference is greater to make the move
		if math.Abs(dx) > math.Abs(dy) {
			// Move along X-axis
			if dx > 0 {
				step = models.Position{X: d.Position.X + 1, Y: d.Position.Y} // Move right
			} else {
				step = models.Position{X: d.Position.X - 1, Y: d.Position.Y} // Move left
			}
		} else {
			// Move along Y-axis
			if dy > 0 {
				step = models.Position{X: d.Position.X, Y: d.Position.Y + 1} // Move down
			} else {
				step = models.Position{X: d.Position.X, Y: d.Position.Y - 1} // Move up
			}
		}

		// Ensure the step is within map boundaries
		if step.X >= 0 && step.Y >= 0 && step.X < 30 && step.Y < 20 {
			return step
		}

		// If the chosen step is invalid, just don't move
		return d.Position
	}
	for _, person := range d.SeenPeople {
		if person.IsInDistress() { //&& !person.HasReceivedMedical {
			fmt.Println("MEDICAL : Drone", d.ID, "detected person", person.ID, "in distress")
			medicalTentPos, _ := d.closestPOI(models.MedicalTent)
			_, distanceToCharging := d.closestPOI(models.ChargingStation)

			// Calculate complete mission battery requirement
			distanceToTent := d.Position.CalculateDistance(medicalTentPos)
			distancePersonToTent := person.Position.CalculateDistance(medicalTentPos) * 2
			totalBatteryNeeded := distanceToTent + distancePersonToTent + distanceToCharging + 2

			if d.Battery > totalBatteryNeeded {
				fmt.Println("MEDICAL : Drone", d.ID, "has enough battery to complete the mission")
				// If at medical tent, wait required time
				if d.Position.X == medicalTentPos.X && d.Position.Y == medicalTentPos.Y {
					fmt.Println("MEDICAL : Drone", d.ID, "is at medical tent")
					if d.MedicalTentTimer > 0 {
						d.MedicalTentTimer--
						return d.Position
					}
					// After waiting, go to person
					d.MedicalTentTimer = 0
					return person.GetPosition()
				}

				// If at person position, deliver medical supplies
				if d.Position.X == person.GetPosition().X && d.Position.Y == person.GetPosition().Y {
					if d.DeploymentTimer > 0 {
						d.DeploymentTimer--
						return d.Position
					}

					responseChan := make(chan models.MedicalDeliveryResponse)
					d.MedicalDeliveryChan <- models.MedicalDeliveryRequest{
						PersonID:     person.ID,
						DroneID:      d.ID,
						ResponseChan: responseChan,
					}
					fmt.Println("MEDICAL : Drone", d.ID, "is delivering medical supplies to person", person.ID)
					response := <-responseChan
					fmt.Println("MEDICAL : Drone", d.ID, "received response:", response)

					if response.Authorized {
						d.DeploymentTimer = 1  // Reset for next time
						d.MedicalTentTimer = 2 // Reset for next time
					}
					return d.Position
				}

				// If not at medical tent yet, go there first
				if d.MedicalTentTimer == 0 {
					d.MedicalTentTimer = 2 // Initialize timer for tent stay
					return medicalTentPos
				}

				// If between tent and person, continue to person
				return person.GetPosition()
			}
		}
	}

	directions := []models.Position{
		{X: 0, Y: -1}, // Up
		{X: 0, Y: 1},  // Down
		{X: -1, Y: 0}, // Left
		{X: 1, Y: 0},  // Right
	}

	// If we are not forced to head to a charging station, move randomly (shuffling directions)
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, dir := range directions {
		target := models.Position{
			X: d.Position.X + dir.X,
			Y: d.Position.Y + dir.Y,
		}
		if target.X >= 0 && target.Y >= 0 && target.X < 30 && target.Y < 20 {
			return target
		}
	}

	// If no valid moves found, stay in the same position
	return d.Position
}

func (d *Drone) Myturn() {
	// Get the next position to move to
	d.SeenPeople = []*persons.Person{}
	d.DroneInComRange = []*Drone{}

	// Check if we're at a charging station and should charge
	if d.tryCharging() {
		d.ReceiveInfo()
		return // Skip movement if charging
	}

	target := d.Think()

	// Try to move to the calculated target
	moved := d.Move(target)

	d.ReceiveInfo()

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
	distance := zone.CalculateDistance(d.Position)

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

// calculateClusterFactor évalue combien de zones sont proches de la zone donnée
func calculateClusterFactor(d *Drone, targetZone models.Position) float64 {
	const proximityThreshold = 2.0 // Distance considérée comme "proche"
	nearbyZones := 0.0

	for _, zone := range d.ReportedZonesByCentrale {
		if zone.CalculateDistance(targetZone) < proximityThreshold {
			nearbyZones += 1.0
		}
	}

	return nearbyZones / float64(len(d.ReportedZonesByCentrale))
}
