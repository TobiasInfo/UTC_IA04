package persons

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"math/rand"
	"time"
)

// Person represents a person in the simulation
type Person struct {
	ID                      int
	Position                models.Position
	InDistress              bool
	DistressProbability     float64
	Lifespan                int
	CurrentDistressDuration int
	width                   int
	height                  int
	MoveChan                chan models.MovementRequest
	DeadChan                chan models.DeadRequest
	Profile                 PersonProfile // From profile.go
	State                   StateData     // From state.go
}

// NewCrowdMember creates a new instance of a Person
func NewCrowdMember(id int, position models.Position, distressProbability float64, lifespan int, width int, height int, moveChan chan models.MovementRequest, deadChan chan models.DeadRequest) Person {
	profileType := ProfileType(rand.Intn(4))

	return Person{
		ID:                      id,
		Position:                position,
		InDistress:              false,
		DistressProbability:     distressProbability,
		Lifespan:                lifespan,
		width:                   width,
		height:                  height,
		CurrentDistressDuration: 0,
		MoveChan:                moveChan,
		DeadChan:                deadChan,
		Profile:                 NewPersonProfile(profileType),
		State:                   NewStateData(),
	}
}

// Move randomly updates the Person's position
func (c *Person) MoveRandom() {
	for {
		if c.Position.X == -1 && c.Position.Y == -1 {
			return
		}

		rand.Seed(time.Now().UnixNano())

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
		fmt.Printf("Trying to move person %d to %v\n", c.ID, newPosition)

		if c.Position.X == newPosition.X && c.Position.Y == newPosition.Y {
			return
		}

		responseChan := make(chan models.MovementResponse)
		c.MoveChan <- models.MovementRequest{MemberID: c.ID, MemberType: "persons", NewPosition: newPosition, ResponseChan: responseChan}
		response := <-responseChan

		if response.Authorized {
			c.Position.X = newPosition.X
			c.Position.Y = newPosition.Y
			fmt.Printf("Person %d moved to %v\n", c.ID, c.Position)
			break
		}
	}
}

// MoveTo updates the Person's position towards a target position
func (c *Person) MoveTo(target models.Position) {
	if c.Position.X == -1 && c.Position.Y == -1 {
		return
	}

	for {
		fmt.Printf("Trying to move person %d to %v\n", c.ID, target)

		if c.Position.X == target.X && c.Position.Y == target.Y {
			return
		}

		responseChan := make(chan models.MovementResponse)
		c.MoveChan <- models.MovementRequest{MemberID: c.ID, MemberType: "persons", NewPosition: target, ResponseChan: responseChan}
		response := <-responseChan

		if response.Authorized {
			c.Position.X = target.X
			c.Position.Y = target.Y
			fmt.Printf("Person %d moved to %v\n", c.ID, c.Position)
			break
		}
	}
}

// HasReachedPOI checks if person has reached their target POI
func (c *Person) HasReachedPOI() bool {
	if c.State.TargetPOI == nil {
		return false
	}
	return false // Placeholder until POI tracking is implemented
}

// UpdateHealth updates the health state of the Person
func (c *Person) UpdateHealth() {
	if c.State.CurrentState == Resting {
		c.Profile.StaminaLevel += 0.01
		return
	}

	// Reduce stamina based on current state and movement
	staminaReduction := 0.001
	if c.State.CurrentState == SeekingPOI {
		staminaReduction = 0.002 // Moving with purpose uses more energy
	}
	c.Profile.StaminaLevel -= staminaReduction

	// Calculate malaise probability based on profile and current conditions
	effectiveProbability := c.DistressProbability *
		(1.0 - c.Profile.MalaiseResistance) *
		(1.0 - c.Profile.StaminaLevel)

	if rand.Float64() < effectiveProbability {
		c.InDistress = true
	}

	if c.InDistress {
		c.CurrentDistressDuration++
		if c.CurrentDistressDuration >= c.Lifespan {
			c.Die()
		}
	} else {
		c.CurrentDistressDuration = 0
	}
}

// Die handles the death of a Person
func (c *Person) Die() {
	c.InDistress = false
	c.CurrentDistressDuration = 0

	responseChan := make(chan models.DeadResponse)
	c.DeadChan <- models.DeadRequest{MemberID: c.ID, MemberType: "persons", ResponseChan: responseChan}

	fmt.Printf("Trying to remove person %d from the map\n", c.ID)

	response := <-responseChan

	if !response.Authorized {
		fmt.Printf("Person %d could not be removed from the map\n", c.ID)
		return
	}

	c.Position.X = -1
	c.Position.Y = -1

	fmt.Printf("Person %d died :c\n", c.ID)
}

// IsAlive checks if the Person is still alive
func (c *Person) IsAlive() bool {
	return c.Position.X >= 0 && c.Position.Y >= 0
}

// Myturn handles the person's turn in the simulation
func (c *Person) Myturn() {
	// Update state
	c.State.UpdateState(c)

	// Update health with new profile-based calculations
	c.UpdateHealth()

	// Movement based on current state
	switch c.State.CurrentState {
	case Exploring:
		c.MoveRandom() // Will be replaced with intelligent movement later
	case SeekingPOI:
		c.MoveRandom() // Will implement POI seeking movement later
	case Resting:
		// Don't move while resting
	case InQueue:
		// Minimal movement in queue
	case InDistress:
		// Limited movement during distress
		if rand.Float64() < 0.3 {
			c.MoveRandom()
		}
	}
}
