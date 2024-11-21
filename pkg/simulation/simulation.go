package simulation

import (
	"UTC_IA04/pkg/models"
	"fmt"
)

type Simulation struct {
	drones []*SurveillanceDrone
	crowd  []*CrowdMember
	// Add fields for performance metrics and system parameters
}

func NewSimulation() *Simulation {
	return &Simulation{
		drones: []*SurveillanceDrone{},
		crowd:  []*CrowdMember{},
	}
}

func (s *Simulation) Initialize(nDrones int, nCrowd int) {

	//Exemple de fonction de détection qui avec un radius n , qui retourne une liste des pourcentages (en int) de détection selon la distance par rapport au centre
	detectionFunc := func(n int) []int {
		pourcentageArray := make([]int, n)
		for i := 0; i < n; i++ {
			pourcentageArray[i] = 100 * (1 - int(float64(i)/float64(n)))
		}
		return pourcentageArray
	}

	for i := 0; i < nDrones; i++ {
		s.drones = append(s.drones, NewSurveillanceDrone(i, models.Position{X: 0, Y: 0}, 100.0, detectionFunc))
	}

	for i := 0; i < nCrowd; i++ {
		s.crowd = append(s.crowd, NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20))
	}

}

func (s *Simulation) StartSimulation(numberIteration int) {
	fmt.Println("Simulation started")
	// Logic to initialize and run the simulation

	s.Initialize(5, 40)

	for tick := 0; tick < numberIteration; tick++ {
		// Mise à jour des drones

		for _, drone := range s.drones {
			// Mise à jour de la position du drone
			drone.Myturn()
		}

		// Mise à jour des membres de la foule
		for _, member := range s.crowd {
			// Mise à jour de la position du membre
			member.Myturn()
		}

	}

	for _, drone := range s.drones {
		fmt.Printf("Position du drone %d : X = %f , Y = %f", drone.ID, drone.Position.X, drone.Position.Y)
	}

	// Mise à jour des membres de la foule
	for _, member := range s.crowd {
		fmt.Printf("Position du membre %d : X = %f , Y = %f", member.ID, member.Position.X, member.Position.Y)
	}

}
