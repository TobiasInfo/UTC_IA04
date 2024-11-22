package simulation

import (
	"UTC_IA04/pkg/models"
	"fmt"
)

type Simulation struct {
	Drones    []*SurveillanceDrone
	Crowd     []*CrowdMember
	Obstacles []*Obstacle
	// Add fields for performance metrics and system parameters
}

// func NewSimulation() *Simulation {
// 	return &Simulation{
// 		Drones:    []*SurveillanceDrone{},
// 		Crowd:     []*CrowdMember{},
// 		Obstacles: []*Obstacle{},
// 	}
// }

// Initialize the simulation with the given number of drones, crowd members, and obstacles
func NewSimulation(numDrones, numCrowdMembers, numObstacles int) *Simulation {
	s := &Simulation{
		Drones:    []*SurveillanceDrone{},
		Crowd:     []*CrowdMember{},
		Obstacles: []*Obstacle{},
	}
	s.Initialize(numDrones, numCrowdMembers, numObstacles)
	return s
}

func (s *Simulation) Initialize(nDrones int, nCrowd int, nObstacles int) {

	//Exemple de fonction de détection qui avec un radius n , qui retourne une liste des pourcentages (en int) de détection selon la distance par rapport au centre
	detectionFunc := func(n int) []int {
		pourcentageArray := make([]int, n)
		for i := 0; i < n; i++ {
			pourcentageArray[i] = 100 * (1 - int(float64(i)/float64(n)))
		}
		return pourcentageArray
	}

	for i := 0; i < nDrones; i++ {
		s.Drones = append(s.Drones, NewSurveillanceDrone(i, models.Position{X: 0, Y: 0}, 100.0, detectionFunc))
	}

	for i := 0; i < nCrowd; i++ {
		s.Crowd = append(s.Crowd, NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20))
	}
	for i := 0; i < nObstacles; i++ {
		s.Obstacles = append(s.Obstacles, NewObstacle(models.Position{X: 0, Y: 0}))
	}

}

func (s *Simulation) StartSimulation(numberIteration int) {
	fmt.Println("Simulation started")
	// Logic to initialize and run the simulation

	s.Initialize(5, 40, 3)

	for tick := 0; tick < numberIteration; tick++ {
		// Mise à jour des drones

		for _, drone := range s.Drones {
			// Mise à jour de la position du drone
			drone.Myturn()
		}

		// Mise à jour des membres de la foule
		for _, member := range s.Crowd {
			// Mise à jour de la position du membre
			member.Myturn()
		}

	}

	for _, drone := range s.Drones {
		fmt.Printf("Position du drone %d : X = %f , Y = %f", drone.ID, drone.Position.X, drone.Position.Y)
	}

	// Mise à jour des membres de la foule
	for _, member := range s.Crowd {
		fmt.Printf("Position du membre %d : X = %f , Y = %f", member.ID, member.Position.X, member.Position.Y)
	}

}

// Update the simulation state
func (s *Simulation) Update() {
	// Update the drones
	for _, drone := range s.Drones {
		drone.Myturn()
	}

	// Update the crowd members
	for _, member := range s.Crowd {
		member.Myturn()
	}
}

func (s *Simulation) UpdateCrowdSize(newSize int) {

	currentSize := len(s.Crowd)
	if newSize > currentSize {
		for i := currentSize; i < newSize; i++ {
			s.Crowd = append(s.Crowd, NewCrowdMember(i, models.Position{X: 0, Y: 0}, 0.001, 20))
		}
	} else if newSize < currentSize {
		s.Crowd = s.Crowd[:newSize]
	}

}

func (s *Simulation) UpdateDroneSize(newSize int) {

	currentSize := len(s.Drones)
	if newSize > currentSize {
		for i := currentSize; i < newSize; i++ {
			s.Drones = append(s.Drones, NewSurveillanceDrone(i, models.Position{X: 0, Y: 0}, 100.0, nil))
		}
	} else if newSize < currentSize {
		s.Drones = s.Drones[:newSize]
	}

}

func (s *Simulation) CountCrowdMembersInDistress() int {
	count := 0
	for _, member := range s.Crowd {
		if member.InDistress {
			count++
		}
	}
	return count
}
