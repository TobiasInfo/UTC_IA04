package simulation

import (
	"UTC_IA04/pkg/models"
	"fmt"
	"sync"
)

// MapCell represents a single cell on the map
type MapCell struct {
	Position  models.Position
	Obstacles []*Obstacle
	Persons   []*Person
	Drones    []*Drone
}

// Map represents the entire simulation environment
type Map struct {
	Width     int
	Height    int
	Persons   []*Person
	Drones    []*Drone
	Obstacles []*Obstacle
	Cells     map[models.Position]*MapCell
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
				Position:  pos,
				Obstacles: []*Obstacle{},
				Persons:   []*Person{},
				Drones:    []*Drone{},
			}
		}
	}

	return &Map{
		Width:   width,
		Height:  height,
		Cells:   cells,
		Persons: []*Person{},
		Drones:  []*Drone{},
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
func (m *Map) GetDrones(position models.Position) []*Drone {
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
	m.Obstacles = append(m.Obstacles, obstacle)
}

// AddCrowdMember adds a crowd member to a specific position on the map
func (m *Map) AddCrowdMember(member *Person) {
	cell := m.Cells[member.Position]
	cell.Persons = append(cell.Persons, member)
	m.Persons = append(m.Persons, member)
}

// AddDrone adds a drone to a specific position on the map
func (m *Map) AddDrone(drone *Drone) {
	cell := m.Cells[drone.Position]
	cell.Drones = append(cell.Drones, drone)
	m.Drones = append(m.Drones, drone)
}

// MoveEntity updates the position of an entity (e.g., drone, crowd member)
func (m *Map) MoveEntity(entity interface{}, newPosition models.Position) {
	var currentCell, newCell *MapCell
	switch e := entity.(type) {
	case *Drone:
		currentCell = m.Cells[e.Position]
		newCell = m.Cells[newPosition]
		removeDroneFromCell(currentCell, e)
		newCell.Drones = append(newCell.Drones, e)
		e.Position = newPosition
	case *Person:
		currentCell = m.Cells[e.Position]
		newCell = m.Cells[newPosition]
		removeCrowdMemberFromCell(currentCell, e)
		newCell.Persons = append(newCell.Persons, e)
		e.Position = newPosition
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
	//fmt.Println("Position", position, "is blocked:", state)
	return state
}

// CountCrowdMembers returns the total number of crowd members on the map
func (m *Map) CountCrowdMembers() int {
	count := 0
	for _, cell := range m.Cells {
		count += len(cell.Persons)
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
func removeDroneFromCell(cell *MapCell, drone *Drone) {
	moved := false
	for i, d := range cell.Drones {
		if d.ID == drone.ID {
			cell.Drones = append(cell.Drones[:i], cell.Drones[i+1:]...)
			moved = true
			break
		}
	}
	if !moved {
		fmt.Println("Drone not found in cell")
	}
}

// removeCrowdMemberFromCell removes a crowd member from a map cell
func removeCrowdMemberFromCell(cell *MapCell, member *Person) {
	moved := false
	for i, m := range cell.Persons {
		if m.ID == member.ID {
			cell.Persons = append(cell.Persons[:i], cell.Persons[i+1:]...)
			moved = true
			break
		}
	}
	if !moved {
		fmt.Println("Crowd member not found in cell")
	}
}
