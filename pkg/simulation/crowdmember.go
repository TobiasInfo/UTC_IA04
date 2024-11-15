package simulation

import (
	"UTC_IA04/pkg/models"
	"math/rand"
)

// CrowdMember represents a person in the simulation
type CrowdMember struct {
	ID                      int
	Position                models.Position
	InDistress              bool
	DistressProbability     float64
	Lifespan                int // Maximum duration in distress before "death"
	CurrentDistressDuration int
}

// NewCrowdMember creates a new instance of a CrowdMember
func NewCrowdMember(id int, position models.Position, distressProbability float64, lifespan int) *CrowdMember {
	return &CrowdMember{
		ID:                      id,
		Position:                position,
		InDistress:              false,
		DistressProbability:     distressProbability,
		Lifespan:                lifespan,
		CurrentDistressDuration: 0,
	}
}

// Move randomly updates the CrowdMember's position
func (c *CrowdMember) MoveRandom() {
	c.Position.X += float64(rand.Intn(3) - 1) // Move randomly in X direction (-1, 0, 1)
	c.Position.Y += float64(rand.Intn(3) - 1) // Move randomly in Y direction (-1, 0, 1)
}

// MoveTo updates the CrowdMember's position towards a target position
func (c *CrowdMember) MoveTo(position models.Position) {
	if c.Position.X < position.X {
		c.Position.X++
	} else if c.Position.X > position.X {
		c.Position.X--
	}

	if c.Position.Y < position.Y {
		c.Position.Y++
	} else if c.Position.Y > position.Y {
		c.Position.Y--
	}
}

// UpdateHealth updates the health state of the CrowdMember
func (c *CrowdMember) UpdateHealth(neighborCount int) {
	effectiveDistressProbability := c.DistressProbability * float64(neighborCount)

	if rand.Float64() < effectiveDistressProbability {
		c.InDistress = true
	}

	// Update distress duration or reset if no distress
	if c.InDistress {
		c.CurrentDistressDuration++
		if c.CurrentDistressDuration >= c.Lifespan {
			c.Die()
			// Add a new CrowdMember to the map with random values
			// TODO : Change the parameters to be more realistic
			//simulation.Map.AddCrowdMember(NewCrowdMember(rand.Intn(1000), models.Position{X: rand.Float64() * float64(simulation.Map.Width), Y: rand.Float64() * float64(simulation.Map.Height)}, 0.1, 10))
		}
	} else {
		c.CurrentDistressDuration = 0
	}
}

// CountNeighbors counts the number of neighboring CrowdMembers in the threshold distance
func (c *CrowdMember) CountNeighbors(crowd []*CrowdMember, threshold float64) int {
	count := 0
	for _, neighbor := range crowd {
		if c != neighbor && c.Position.CalculateDistance(neighbor.Position) <= threshold { // Neighbor threshold
			count++
		}
	}
	return count
}

// Die handles the death of a CrowdMember
func (c *CrowdMember) Die() {

	// TODO : CHeck how to recover the map size properly
	c.InDistress = false
	c.CurrentDistressDuration = 0
	// Descroy the curent crowd member
	// TODO : Destroy propoerly the crowd member using map instance
	// simulation.Map.DestroyCrowdMember(c)

	// For now, I just set the position to -1 to indicate that the crowd member is dead
	c.Position.X = -1
	c.Position.Y = -1
}

// IsAlive checks if the CrowdMember is still alive
func (c *CrowdMember) IsAlive() bool {
	return c.Position.X >= 0 && c.Position.Y >= 0
}
