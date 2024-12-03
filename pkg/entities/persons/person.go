package persons

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"math"
	"math/rand"
	"time"
)

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
	Profile                 PersonProfile
	State                   StateData
	MovementPattern         MovementPattern
	ZonePreference          ZonePreference
}

func NewCrowdMember(id int, position models.Position, distressProbability float64, lifespan int, width int, height int, moveChan chan models.MovementRequest, deadChan chan models.DeadRequest) Person {
	profileType := ProfileType(rand.Intn(4))
	movementPattern := MovementPattern(rand.Intn(5))
	zonePreference := GetZonePreference(movementPattern)

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
		MovementPattern:         movementPattern,
		ZonePreference:          zonePreference,
	}
}

func (c *Person) MoveRandom() {
	if c.Position.X == -1 && c.Position.Y == -1 || c.InDistress {
		return
	}

	rand.Seed(time.Now().UnixNano())

	currentZone := c.determineCurrentZone()
	newPosition := c.calculateNextPosition(currentZone)

	if c.Position.X == newPosition.X && c.Position.Y == newPosition.Y {
		return
	}

	responseChan := make(chan models.MovementResponse)
	c.MoveChan <- models.MovementRequest{
		MemberID:     c.ID,
		MemberType:   "persons",
		NewPosition:  newPosition,
		ResponseChan: responseChan,
	}

	response := <-responseChan
	if response.Authorized {
		c.Position = newPosition
		fmt.Printf("Person %d moved to %v (Zone: %s)\n", c.ID, c.Position, currentZone)
	}
}

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

func (c *Person) HasReachedPOI() bool {
	if c.State.TargetPOI == nil {
		return false
	}
	return false // Placeholder until POI tracking is implemented
}

func (c *Person) UpdateHealth() {
	if c.State.CurrentState == Resting {
		c.Profile.StaminaLevel += 0.01
		return
	}

	staminaReduction := 0.001
	if c.State.CurrentState == SeekingPOI {
		staminaReduction = 0.002
	}
	c.Profile.StaminaLevel -= staminaReduction

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

func (c *Person) IsAlive() bool {
	return c.Position.X >= 0 && c.Position.Y >= 0
}

func (c *Person) Myturn() {
	if c.InDistress {
		c.UpdateHealth()
		return
	}

	c.State.UpdateState(c)
	c.UpdateHealth()

	switch c.State.CurrentState {
	case Exploring:
		c.MoveRandom()
	case SeekingPOI:
		c.MoveRandom() // Will implement POI seeking movement later
	case Resting:
		// Don't move while resting
	case InQueue:
		// Minimal movement in queue
	}
}

func (c *Person) determineCurrentZone() string {
	zoneWidth := float64(c.width) / 3
	if c.Position.X < zoneWidth {
		return "entrance"
	} else if c.Position.X < zoneWidth*2 {
		return "main"
	}
	return "exit"
}

func (c *Person) calculateNextPosition(currentZone string) models.Position {
	var deltaX, deltaY float64

	deltaY = float64(rand.Intn(3) - 1)

	if c.ZonePreference.ShouldMoveToZone(currentZone) {
		switch c.MovementPattern {
		case EarlyExiter:
			deltaX = 1
		case LateArrival:
			if currentZone == "entrance" {
				deltaX = 0
			} else {
				deltaX = float64(rand.Intn(3) - 1)
			}
		case MainEventFocused:
			if currentZone == "entrance" {
				deltaX = 1
			} else if currentZone == "exit" {
				deltaX = -1
			}
		case FoodEnthusiast:
			deltaX = c.moveTowardsNearestPOI(models.FoodStand)
		default:
			deltaX = float64(rand.Intn(3) - 1)
		}
	} else {
		deltaX = float64(rand.Intn(3) - 1)
	}

	newX := math.Max(0, math.Min(float64(c.width), c.Position.X+deltaX))
	newY := math.Max(0, math.Min(float64(c.height), c.Position.Y+deltaY))

	return models.Position{X: newX, Y: newY}
}

func (c *Person) moveTowardsNearestPOI(poiType models.POIType) float64 {
	if rand.Float64() < c.ZonePreference.GetPOIPreference(poiType) {
		return float64(rand.Intn(3) - 1)
	}
	return 0
}
