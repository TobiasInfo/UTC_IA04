package simulation

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"sync"
)

// MapCell represents a single cell on the map
type MapCell struct {
	Position     models.Position
	Obstacles    []*Obstacle
	CrowdMembers []*CrowdMember
	Drones       []*SurveillanceDrone
}

// Map represents the entire simulation environment
type Map struct {
	Width  int
	Height int
	Cells  map[models.Position]*MapCell
}

var (
	instance *Map      // The singleton instance
	once     sync.Once // Ensures Map is initialized only once
)

// NewMap creates a new map with the given dimensions, but we will only use it once due to the singleton pattern.
func newMap(width, height int) *Map {
	cells := make(map[models.Position]*MapCell)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pos := models.Position{X: float64(x), Y: float64(y)}
			cells[pos] = &MapCell{
				Position:     pos,
				Obstacles:    []*Obstacle{},
				CrowdMembers: []*CrowdMember{},
				Drones:       []*SurveillanceDrone{},
			}
		}
	}

	return &Map{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}

// GetMap returns the singleton instance of the Map
func GetMap(width, height int) *Map {
	once.Do(func() {
		// Initialize the singleton instance only once
		instance = newMap(width, height)
	})
	return instance
}

// GetDrones returns the drones at a specific position on the map
func (m *Map) GetDrones(position models.Position) []*SurveillanceDrone {
	cell, exists := m.Cells[position]
	if !exists {
		return nil
	}
	return cell.Drones
}

// AddObstacle adds an obstacle to a specific position on the map
func (m *Map) AddObstacle(obstacle *Obstacle) {
	cell := m.Cells[obstacle.Position]
	cell.Obstacles = append(cell.Obstacles, obstacle)
}

// AddCrowdMember adds a crowd member to a specific position on the map
func (m *Map) AddCrowdMember(member *CrowdMember) {
	cell := m.Cells[member.Position]
	cell.CrowdMembers = append(cell.CrowdMembers, member)
}

// AddDrone adds a drone to a specific position on the map
func (m *Map) AddDrone(drone *SurveillanceDrone) {
	cell := m.Cells[drone.Position]
	cell.Drones = append(cell.Drones, drone)
}

// MoveEntity updates the position of an entity (e.g., drone, crowd member)
func (m *Map) MoveEntity(entity interface{}, newPosition models.Position) {
	var currentCell, newCell *MapCell
	switch e := entity.(type) {
	case *SurveillanceDrone:
		currentCell = m.Cells[e.Position]
		newCell = m.Cells[newPosition]
		removeDroneFromCell(currentCell, e)
		e.Position = newPosition
		newCell.Drones = append(newCell.Drones, e)
	case *CrowdMember:
		currentCell = m.Cells[e.Position]
		newCell = m.Cells[newPosition]
		removeCrowdMemberFromCell(currentCell, e)
		e.Position = newPosition
		newCell.CrowdMembers = append(newCell.CrowdMembers, e)
	default:
		fmt.Println("Unknown entity type")
	}
}

// IsBlocked checks if a position is blocked by obstacles
func (m *Map) IsBlocked(position models.Position) bool {
	cell, exists := m.Cells[position]
	if !exists {
		return true // Outside the map boundaries
	}
	state := len(cell.Obstacles) > 0
	fmt.Println("Position", position, "is blocked:", state)
	return state
}

// CountCrowdMembers returns the total number of crowd members on the map
func (m *Map) CountCrowdMembers() int {
	count := 0
	for _, cell := range m.Cells {
		count += len(cell.CrowdMembers)
	}
	return count
}

// CountDrones returns the total number of drones on the map
func (m *Map) CountDrones() int {
	count := 0
	for _, cell := range m.Cells {
		count += len(cell.Drones)
	}
	return count
}

// removeDroneFromCell removes a drone from a map cell
func removeDroneFromCell(cell *MapCell, drone *SurveillanceDrone) {
	for i, d := range cell.Drones {
		if d == drone {
			cell.Drones = append(cell.Drones[:i], cell.Drones[i+1:]...)
			break
		}
	}
}

// removeCrowdMemberFromCell removes a crowd member from a map cell
func removeCrowdMemberFromCell(cell *MapCell, member *CrowdMember) {
	for i, m := range cell.CrowdMembers {
		if m == member {
			cell.CrowdMembers = append(cell.CrowdMembers[:i], cell.CrowdMembers[i+1:]...)
			break
		}
	}
}
