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
	Dead                    bool
	StillInSim              bool
	DistressProbability     float64
	Lifespan                int
	CurrentDistressDuration int
	width                   int
	height                  int
	MoveChan                chan models.MovementRequest
	DeadChan                chan models.DeadRequest
	ExitChan                chan models.ExitRequest
	Profile                 PersonProfile
	State                   StateData
	MovementPattern         MovementPattern
	ZonePreference          ZonePreference
	EntryTime               time.Time
	CurrentPath             []models.Position
	CurrentPOI              *models.POIType
	TargetPOIPosition       *models.Position
	TimeAtPOI               time.Duration
	LastZoneChange          time.Time
	debug                   bool
	hardDebug               bool
	HasReceivedMedical      bool
	SeekingExit             bool
	TreatmentTime           time.Duration
	AssignedDroneID         *int
}

func NewCrowdMember(id int, position models.Position, distressProbability float64, lifespan int, width int, height int, moveChan chan models.MovementRequest, deadChan chan models.DeadRequest, exitChan chan models.ExitRequest) Person {
	profileType := ProfileType(rand.Intn(4))
	movementPattern := MovementPattern(rand.Intn(5))
	zonePreference := GetZonePreference(movementPattern)
	now := time.Now()

	p := Person{
		ID:                      id,
		Position:                position,
		InDistress:              false,
		Dead:                    false,
		StillInSim:              true,
		SeekingExit:             false,
		DistressProbability:     distressProbability,
		Lifespan:                lifespan,
		width:                   width,
		height:                  height,
		CurrentDistressDuration: 0,
		MoveChan:                moveChan,
		DeadChan:                deadChan,
		ExitChan:                exitChan,
		Profile:                 NewPersonProfile(profileType),
		State:                   NewStateData(),
		MovementPattern:         movementPattern,
		ZonePreference:          zonePreference,
		EntryTime:               now,
		CurrentPath:             make([]models.Position, 0),
		CurrentPOI:              nil,
		TargetPOIPosition:       nil,
		TimeAtPOI:               0,
		LastZoneChange:          now,
		debug:                   false,
		hardDebug:               false,
		HasReceivedMedical:      false,
		TreatmentTime:           0,
	}
	return p
}

func (p *Person) IsAssigned() bool {
	return p.AssignedDroneID != nil
}

func (c *Person) Myturn() {
	if c.SeekingExit && !c.InDistress {
		if len(c.CurrentPath) == 0 {
			exitPos := models.Position{X: (float64(c.width)/10)*9 + 0.1, Y: c.Position.Y}
			if exitPos.CalculateDistance(c.Position) < 1 {
				c.Exit()
				return
			}

			path := models.FindPath(c.Position, exitPos, c.width, c.height, make(map[models.Position]bool))
			c.CurrentPath = path
		}
		c.goTo()
		return
	}

	if c.hardDebug {
		fmt.Printf("Person %d executing turn - Current State: %v, Position: %v\n",
			c.ID, c.State.CurrentState, c.Position)
	}
	if c.InDistress {
		if c.hardDebug {
			fmt.Printf("Person %d is in distress, not moving\n", c.ID)
		}
		c.UpdateHealth()
		return
	}
	c.State.UpdateState(c)
	c.UpdateHealth()

	if c.GetCurrentZone() == "exit" {
		c.Exit()
		return
	}

	obstacles := make(map[models.Position]bool)

	switch c.State.CurrentState {
	case Exploring:
		c.UpdatePosition(obstacles)
	case SeekingPOI:
		if c.CurrentPOI == nil {
			for poiType := range c.ZonePreference.POIPreferences {
				if c.ZonePreference.ShouldVisitPOI(poiType) {
					c.CurrentPOI = &poiType
					break
				}
			}
		}
		c.UpdatePosition(obstacles)
	case Resting:
		// Don't move while resting
		c.TimeAtPOI += time.Second
		if c.Profile.StaminaLevel > 0.8 {
			c.State.CurrentState = Exploring
			c.TimeAtPOI = 0
			c.CurrentPOI = nil
			c.TargetPOIPosition = nil
		}
	}
}

func (c *Person) UpdatePosition(obstacles map[models.Position]bool) bool {
	if len(c.CurrentPath) == 0 {
		c.generateNewPath(obstacles)
	}

	if c.HasReachedPOI() {
		c.State.CurrentState = Resting
		c.State.TimeInState = 0
		return false
	}

	if len(c.CurrentPath) > 0 {
		nextPos := c.CurrentPath[0]
		if c.tryMove(nextPos) {
			c.CurrentPath = c.CurrentPath[1:]
			return true
		} else {
			c.CurrentPath = []models.Position{}
			return false
		}
	}

	fmt.Printf("Person %d has no valid moves\n", c.ID)
	return false
}

func (c *Person) goTo() bool {
	if len(c.CurrentPath) > 0 {
		nextPos := c.CurrentPath[0]

		if c.tryMove(nextPos) {
			c.CurrentPath = c.CurrentPath[1:]
			return true
		} else {
			c.CurrentPath = []models.Position{}
			return false
		}
	}

	fmt.Printf("Person %d has no valid moves\n", c.ID)
	return false
}

// MoveTo updates the Person's position towards a target position
func (c *Person) tryMove(target models.Position) bool {
	if c.Position.X == -1 && c.Position.Y == -1 {
		return false
	}

	for {
		if c.Position.X == target.X && c.Position.Y == target.Y {
			return false
		}

		responseChan := make(chan models.MovementResponse)
		c.MoveChan <- models.MovementRequest{MemberID: c.ID, MemberType: "persons", NewPosition: target, ResponseChan: responseChan}
		response := <-responseChan

		if response.Authorized {
			c.Position.X = target.X
			c.Position.Y = target.Y
			return true
		}

		if response.Reason == "Position is blocked" {
			return false
		}
	}
}

func (c *Person) shouldLeavePOI() bool {
	minTime := 30 * time.Second
	maxTime := 5 * time.Minute

	if c.TimeAtPOI < minTime {
		return false
	}

	// Probability of leaving increases with time
	timeRatio := float64(c.TimeAtPOI) / float64(maxTime)
	if timeRatio > 1 {
		return true
	}

	// 1. Low stamina
	if c.Profile.StaminaLevel < 0.3 {
		return rand.Float64() < 0.8
	}
	// 2. Close to their exit time
	if c.GetTimeSinceEntry() > c.ZonePreference.ExitTime-15*time.Minute {
		return rand.Float64() < 0.9
	}

	return rand.Float64() < timeRatio
}

func (c *Person) hoverNearPOI(obstacles map[models.Position]bool) bool {
	if c.TargetPOIPosition == nil {
		return false
	}

	// Stay within a certain radius of the POI
	radius := 2.0
	angle := rand.Float64() * 2 * math.Pi

	newX := c.TargetPOIPosition.X + math.Cos(angle)*radius
	newY := c.TargetPOIPosition.Y + math.Sin(angle)*radius

	// Ensure within bounds
	newX = math.Max(0, math.Min(float64(c.width), newX))
	newY = math.Max(0, math.Min(float64(c.height), newY))

	newPos := models.Position{X: newX, Y: newY}

	// Only move if not blocked
	if !obstacles[newPos] {
		moved := c.tryMove(newPos)
		if moved {
			fmt.Printf("Person %d hovering near POI\n", c.ID)
		} else {
			fmt.Printf("Person %d failed to hover near POI\n", c.ID)
		}
		return moved
	}
	return false
}

func (c *Person) generateNewPath(obstacles map[models.Position]bool) {
	var targetPos models.Position

	if c.CurrentPOI != nil && c.TargetPOIPosition != nil {
		targetPos = *c.TargetPOIPosition
	} else {
		currentZone := c.determineCurrentZone()
		targetZone := c.ZonePreference.GetNextZone(currentZone, c.EntryTime)
		if targetZone == currentZone {
			targetPos = c.getRandomZonePosition(targetZone)
		} else {
			targetPos = c.getZoneEntryPoint(targetZone)
		}
	}
	path := models.FindPath(c.Position, targetPos, c.width, c.height, obstacles)
	c.CurrentPath = path
}

func (c *Person) getRandomZonePosition(zone string) models.Position {
	var x, y float64

	switch zone {
	case "entrance":
		x = rand.Float64() * float64(c.width) / 10
	case "main":
		x = float64(c.width)/10 + rand.Float64()*float64(c.width)*8/10
	case "exit":
		x = float64(c.width)*8/10 + rand.Float64()*float64(c.width)/10
	}

	y = rand.Float64() * float64(c.height)
	return models.Position{X: x, Y: y}
}

func (c *Person) getZoneEntryPoint(zone string) models.Position {
	y := rand.Float64() * float64(c.height)

	switch zone {
	case "main":
		return models.Position{X: float64(c.width) / 10, Y: y}
	case "exit":
		return models.Position{X: float64(c.width) * 9 / 10, Y: y}
	default:
		return models.Position{X: 0, Y: y}
	}
}

func (c *Person) determineCurrentZone() string {
	zoneWidth := float64(c.width) / 10
	if c.Position.X < zoneWidth {
		return "entrance"
	} else if c.Position.X < zoneWidth*9 {
		return "main"
	}
	return "exit"
}

func (c *Person) UpdateHealth() {

	if c.State.CurrentState == Resting {
		c.Profile.StaminaLevel += 0.01
		if c.Profile.StaminaLevel > 1.0 {
			c.Profile.StaminaLevel = 1.0
		}
	} else {
		staminaReduction := 0.001
		if c.State.CurrentState == SeekingPOI {
			staminaReduction = 0.002
		}
		c.Profile.StaminaLevel -= staminaReduction
		if c.Profile.StaminaLevel < 0 {
			c.Profile.StaminaLevel = 0
		}
	}

	if c.InDistress {
		c.CurrentDistressDuration++
		if c.CurrentDistressDuration >= c.Lifespan {
			c.Die()
		}
	} else {
		effectiveProbability := c.DistressProbability *
			(1.0 - c.Profile.MalaiseResistance) *
			(1.0 - c.Profile.StaminaLevel)

		randNum := rand.Float64() * 20

		if randNum < effectiveProbability {
			c.InDistress = true
		}
		c.CurrentDistressDuration = 0
	}
}

func (c *Person) Die() {
	if c.IsDead() {
		return
	}

	c.InDistress = false
	c.Dead = true
	c.CurrentDistressDuration = 0

	responseChan := make(chan models.DeadResponse)
	c.DeadChan <- models.DeadRequest{
		MemberID:     c.ID,
		MemberType:   "persons",
		ResponseChan: responseChan,
	}

	response := <-responseChan
	if !response.Authorized {
		return
	}

	c.Position.X = -10
	c.Position.Y = -10
	c.StillInSim = false
}

func (c *Person) Exit() {
	responseChan := make(chan models.ExitResponse)
	c.ExitChan <- models.ExitRequest{
		MemberID:     c.ID,
		MemberType:   "persons",
		ResponseChan: responseChan,
	}

	response := <-responseChan
	if !response.Authorized {
		return
	}

	c.Position.X = -1
	c.Position.Y = -1
	c.StillInSim = false
}

func (c *Person) HasReachedPOI() bool {
	if c.CurrentPOI == nil || c.TargetPOIPosition == nil {
		return false
	}
	dist := c.Position.CalculateDistance(*c.TargetPOIPosition)
	return dist <= 2.0
}

func (c *Person) SetTargetPOI(poiType models.POIType, position models.Position) {
	c.CurrentPOI = &poiType
	c.TargetPOIPosition = &position
	c.TimeAtPOI = 0
}

func (c *Person) IsDead() bool {
	return c.Dead
}

func (c *Person) GetTimeSinceEntry() time.Duration {
	return time.Since(c.EntryTime)
}

func (c *Person) GetCurrentZone() string {
	return c.determineCurrentZone()
}

func (c *Person) IsInDistress() bool {
	return c.InDistress
}

func (c *Person) GetStamina() float64 {
	return c.Profile.StaminaLevel
}
