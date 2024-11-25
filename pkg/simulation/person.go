package simulation

import (
	"UTC_IA04/pkg/models"
	"math/rand"
)

// CrowdMember represents a person in the simulation
type Person struct {
	ID                      int
	Position                models.Position
	InDistress              bool
	DistressProbability     float64
	Lifespan                int // Maximum duration in distress before "death"
	CurrentDistressDuration int
	width                   int
	height                  int
	MoveChan                chan models.MovementRequest
}

// NewCrowdMember creates a new instance of a CrowdMember
func NewCrowdMember(id int, position models.Position, distressProbability float64, lifespan int, width int, height int, moveChan chan models.MovementRequest) *Person {
	return &Person{
		ID:                      id,
		Position:                position,
		InDistress:              false,
		DistressProbability:     distressProbability,
		Lifespan:                lifespan,
		CurrentDistressDuration: 0,
		width:                   width,
		height:                  height,
		MoveChan:                moveChan,
	}
}

// Move randomly updates the CrowdMember's position
func (c *Person) MoveRandom() {
	for {
		if c.Position.X == -1 && c.Position.Y == -1 {
			return
		}

		newX := c.Position.X + float64(rand.Intn(3)-1)
		newY := c.Position.Y + float64(rand.Intn(3)-1)

		if newX < 0 {
			newX = 0
		}
		if newY < 0 {
			newY = 0
		}
		if newX > float64(c.width) {
			newX = float64(c.width)
		}
		if newY > float64(c.height) {
			newY = float64(c.height)
		}

		newPosition := models.Position{X: newX, Y: newY}
		responseChan := make(chan models.MovementResponse)
		c.MoveChan <- models.MovementRequest{MemberID: c.ID, MemberType: "person", NewPosition: newPosition, ResponseChan: responseChan}
		response := <-responseChan

		if response.Authorized {
			break
		}
	}
}

// MoveTo updates the CrowdMember's position towards a target position, at random speed
func (c *Person) MoveTo(position models.Position) {

	if c.Position.X == -1 && c.Position.Y == -1 {
		//The Member is Dead, so no moving
		return

	}

	speed := rand.Float64()
	if c.Position.X < position.X {
		c.Position.X += speed
	} else if c.Position.X > position.X {
		c.Position.X -= speed
	}

	if c.Position.Y < position.Y {
		c.Position.Y += speed
	} else if c.Position.Y > position.Y {
		c.Position.Y -= speed
	}
}

// UpdateHealth updates the health state of the CrowdMember
func (c *Person) UpdateHealth() {

	//TODO use a function to communicate with the global map to know how many people are around our crowdMember
	neighborCount := 3

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
func (c *Person) CountNeighbors(crowd []*Person, threshold float64) int {
	count := 0
	for _, neighbor := range crowd {
		if c != neighbor && c.Position.CalculateDistance(neighbor.Position) <= threshold { // Neighbor threshold
			count++
		}
	}
	return count
}

// Die handles the death of a CrowdMember
func (c *Person) Die() {

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
func (c *Person) IsAlive() bool {
	return c.Position.X >= 0 && c.Position.Y >= 0
}

func (c *Person) Myturn() {

	c.MoveRandom()

	c.UpdateHealth()

	return

	//alternative :
	//TODO si on veut que les humains se dirigent vers des points d'intérêts on aura besoin d'un attribut objectif en plus
	//TODO trouver le bon facteur de taille de carte

	//TODO implémenter les colisions avec les obstacles et les bords de map

	//sizeofmap := 20.
	//but := models.Position{rand.Float64()*sizeofmap, rand.Float64()*sizeofmap}
	//c.MoveTo(but)
	//return
}
