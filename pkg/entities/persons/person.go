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
	EntryTime               time.Time
	CurrentPath             []models.Position
	CurrentPOI              *models.POIType
	TargetPOIPosition       *models.Position
	TimeAtPOI               time.Duration
	LastZoneChange          time.Time
}

func NewCrowdMember(id int, position models.Position, distressProbability float64, lifespan int, width int, height int, moveChan chan models.MovementRequest, deadChan chan models.DeadRequest) Person {
	profileType := ProfileType(rand.Intn(4))
	movementPattern := MovementPattern(rand.Intn(5))
	zonePreference := GetZonePreference(movementPattern)
	now := time.Now()

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
		EntryTime:               now,
		CurrentPath:             make([]models.Position, 0),
		CurrentPOI:              nil,
		TargetPOIPosition:       nil,
		TimeAtPOI:               0,
		LastZoneChange:          now,
	}
}

func (c *Person) HasReachedPOI() bool {
	if c.CurrentPOI == nil || c.TargetPOIPosition == nil {
		return false
	}

	// Check if we're within interaction range of the POI (2 units)
	dist := c.Position.CalculateDistance(*c.TargetPOIPosition)
	return dist <= 2.0
}

func (c *Person) SetTargetPOI(poiType models.POIType, position models.Position) {
	c.CurrentPOI = &poiType
	c.TargetPOIPosition = &position
	c.TimeAtPOI = 0
}

func (c *Person) UpdatePosition(obstacles map[models.Position]bool) bool {
	fmt.Printf("Person %d UpdatePosition starting with path length: %d\n",
		c.ID, len(c.CurrentPath))

	if len(c.CurrentPath) > 0 {
		nextPos := c.CurrentPath[0]
		fmt.Printf("Person %d attempting next position in path: {%.2f, %.2f}\n",
			c.ID, nextPos.X, nextPos.Y)
	}
	if c.HasReachedPOI() {
		fmt.Printf("Person %d at POI, handling hover\n", c.ID)
		return c.hoverNearPOI(obstacles)
	}

	if len(c.CurrentPath) == 0 {
		fmt.Printf("Person %d generating new path\n", c.ID)
		c.generateNewPath(obstacles)
	}

	if len(c.CurrentPath) > 0 {
		nextPos := c.CurrentPath[0]
		fmt.Printf("Person %d attempting to move to {%.2f, %.2f}\n",
			c.ID, nextPos.X, nextPos.Y)

		if c.tryMove(nextPos) {
			fmt.Printf("Person %d successfully moved\n", c.ID)
			c.CurrentPath = c.CurrentPath[1:]
			return true
		} else {
			fmt.Printf("Person %d movement failed, clearing path\n", c.ID)
			c.CurrentPath = []models.Position{}
			return false
		}
	}

	fmt.Printf("Person %d has no valid moves\n", c.ID)
	return false
}

func (c *Person) tryMove(newPos models.Position) bool {
	fmt.Printf("Person %d attempting move from {%.2f, %.2f} to {%.2f, %.2f}\n",
		c.ID, c.Position.X, c.Position.Y, newPos.X, newPos.Y)

	responseChan := make(chan models.MovementResponse)
	c.MoveChan <- models.MovementRequest{
		MemberID:     c.ID,
		MemberType:   "persons",
		NewPosition:  newPos,
		ResponseChan: responseChan,
	}
	fmt.Printf("Person %d waiting for movement response\n", c.ID)

	response := <-responseChan
	if response.Authorized {
		prevZone := c.determineCurrentZone()
		c.Position = newPos
		newZone := c.determineCurrentZone()

		if prevZone != newZone {
			c.LastZoneChange = time.Now()
		}
		fmt.Printf("Person %d movement authorized, now at {%.2f, %.2f}\n",
			c.ID, c.Position.X, c.Position.Y)
		return true
	}
	fmt.Printf("Person %d movement denied\n", c.ID)
	return false
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

	// Additional factors that might make someone leave earlier:
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
		return c.tryMove(newPos)
	}
	return false
}

func (c *Person) generateNewPath(obstacles map[models.Position]bool) {
	fmt.Printf("Person %d generating path - POI: %v, Target POI Pos: %v\n",
		c.ID, c.CurrentPOI, c.TargetPOIPosition)

	var targetPos models.Position

	if c.CurrentPOI != nil && c.TargetPOIPosition != nil {
		targetPos = *c.TargetPOIPosition
		fmt.Printf("Person %d targeting POI at {%.2f, %.2f}\n", c.ID, targetPos.X, targetPos.Y)
	} else {
		currentZone := c.determineCurrentZone()
		targetZone := c.ZonePreference.GetNextZone(currentZone, c.EntryTime)
		fmt.Printf("Person %d in zone %s, targeting zone %s\n", c.ID, currentZone, targetZone)

		if targetZone == currentZone {
			targetPos = c.getRandomZonePosition(targetZone)
		} else {
			targetPos = c.getZoneEntryPoint(targetZone)
		}
		fmt.Printf("Person %d generated target position {%.2f, %.2f}\n", c.ID, targetPos.X, targetPos.Y)
	}

	path := models.FindPath(c.Position, targetPos, c.width, c.height, obstacles)
	fmt.Printf("Person %d path generated with %d steps\n", c.ID, len(path))
	c.CurrentPath = path
}

func (c *Person) getRandomZonePosition(zone string) models.Position {
	var x, y float64

	switch zone {
	case "entrance":
		x = rand.Float64() * float64(c.width) / 3
	case "main":
		x = float64(c.width)/3 + rand.Float64()*float64(c.width)/3
	case "exit":
		x = float64(c.width)*2/3 + rand.Float64()*float64(c.width)/3
	}

	y = rand.Float64() * float64(c.height)
	return models.Position{X: x, Y: y}
}

func (c *Person) getZoneEntryPoint(zone string) models.Position {
	y := rand.Float64() * float64(c.height)

	switch zone {
	case "main":
		return models.Position{X: float64(c.width) / 3, Y: y}
	case "exit":
		return models.Position{X: float64(c.width) * 2 / 3, Y: y}
	default:
		return models.Position{X: 0, Y: y}
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

func (c *Person) UpdateHealth() {
	if c.State.CurrentState == Resting {
		c.Profile.StaminaLevel += 0.01
		if c.Profile.StaminaLevel > 1.0 {
			c.Profile.StaminaLevel = 1.0
		}
		return
	}

	staminaReduction := 0.001
	if c.State.CurrentState == SeekingPOI {
		staminaReduction = 0.002
	}
	c.Profile.StaminaLevel -= staminaReduction
	if c.Profile.StaminaLevel < 0 {
		c.Profile.StaminaLevel = 0
	}

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
	if !c.IsAlive() {
		return
	}

	c.InDistress = false
	c.CurrentDistressDuration = 0

	responseChan := make(chan models.DeadResponse)
	c.DeadChan <- models.DeadRequest{
		MemberID:     c.ID,
		MemberType:   "persons",
		ResponseChan: responseChan,
	}

	fmt.Printf("Person %d requesting removal from map\n", c.ID)

	response := <-responseChan
	if !response.Authorized {
		fmt.Printf("Person %d removal not authorized\n", c.ID)
		return
	}

	c.Position.X = -1
	c.Position.Y = -1
	fmt.Printf("Person %d has been removed from simulation\n", c.ID)
}

func (c *Person) IsAlive() bool {
	return c.Position.X >= 0 && c.Position.Y >= 0
}

func (c *Person) Myturn() {
	fmt.Printf("Person %d executing turn - Current State: %v, Position: %v\n",
		c.ID, c.State.CurrentState, c.Position)
	if c.InDistress {
		fmt.Printf("Person %d is in distress, not moving\n", c.ID)
		c.UpdateHealth()
		return
	}

	c.State.UpdateState(c)
	c.UpdateHealth()

	// Create obstacles map for pathfinding
	obstacles := make(map[models.Position]bool)
	// This would typically be populated with actual obstacle positions from the simulation

	switch c.State.CurrentState {
	case Exploring:
		fmt.Printf("Person %d is exploring\n", c.ID)
		moved := c.UpdatePosition(obstacles)
		fmt.Printf("Person %d movement result: %v\n", c.ID, moved)
	case SeekingPOI:
		fmt.Printf("Person %d is seeking POI\n", c.ID)
		if c.CurrentPOI == nil {
			for poiType := range c.ZonePreference.POIPreferences {
				if c.ZonePreference.ShouldVisitPOI(poiType) {
					// The actual POI position will be set by the simulation
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

func (c *Person) GetTimeSinceEntry() time.Duration {
	return time.Since(c.EntryTime)
}

func (c *Person) GetTimeInCurrentZone() time.Duration {
	return time.Since(c.LastZoneChange)
}

func (c *Person) IsReadyToLeave() bool {
	return c.GetTimeSinceEntry() >= c.ZonePreference.ExitTime
}

func (c *Person) GetCurrentZone() string {
	return c.determineCurrentZone()
}

func (c *Person) GetID() int {
	return c.ID
}

func (c *Person) GetPosition() models.Position {
	return c.Position
}

func (c *Person) IsInDistress() bool {
	return c.InDistress
}

func (c *Person) GetStamina() float64 {
	return c.Profile.StaminaLevel
}
