package simulation

import (
	"UTC_IA04/pkg/entities/drones"
	"UTC_IA04/pkg/entities/obstacles"
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"reflect"
	"sync"
)

// MapCell represents a single cell on the map
type MapCell struct {
	Position  models.Position
	Obstacles []*obstacles.Obstacle
	Persons   []*persons.Person
	Drones    []*drones.Drone
}

// Map represents the entire simulation environment
type Map struct {
	Width     int
	Height    int
	Persons   []*persons.Person
	Drones    []*drones.Drone
	Obstacles []*obstacles.Obstacle
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
				Obstacles: []*obstacles.Obstacle{},
				Persons:   []*persons.Person{},
				Drones:    []*drones.Drone{},
			}
		}
	}

	return &Map{
		Width:   width,
		Height:  height,
		Cells:   cells,
		Persons: []*persons.Person{},
		Drones:  []*drones.Drone{},
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

// GetCells returns a pointer to the map cells
func (m *Map) GetCells() map[models.Position]*MapCell {
	return m.Cells
}

// GetDrones returns the drones at a specific position on the map
func (m *Map) GetDrones(position models.Position) []*drones.Drone {
	cell, exists := m.Cells[position]
	if !exists {
		return nil
	}
	return cell.Drones
}

// AddObstacle adds an obstacles to a specific position on the map
func (m *Map) AddObstacle(obstacle *obstacles.Obstacle) {
	cell := m.Cells[obstacle.Position]
	cell.Obstacles = append(cell.Obstacles, obstacle)
	m.Obstacles = append(m.Obstacles, obstacle)
}

// AddCrowdMember adds a crowd member to a specific position on the map
func (m *Map) AddCrowdMember(member *persons.Person) {
	cell := m.Cells[member.Position]
	cell.Persons = append(cell.Persons, member)
	m.Persons = append(m.Persons, member)
}

// AddDrone adds a drones to a specific position on the map
func (m *Map) AddDrone(drone *drones.Drone) {
	cell := m.Cells[drone.Position]
	cell.Drones = append(cell.Drones, drone)
	m.Drones = append(m.Drones, drone)
}

// MoveEntity updates the position of an entity (e.g., drones, crowd member)
func (m *Map) MoveEntity(entity interface{}, newPosition models.Position) {
	var currentCell, newCell *MapCell
	switch e := entity.(type) {
	case *drones.Drone:
		currentCell = m.Cells[e.Position]
		newCell = m.Cells[newPosition]
		removeDroneFromCell(currentCell, e)
		newCell.Drones = append(newCell.Drones, e)
	case *persons.Person:
		currentCell = m.Cells[e.Position]
		newCell = m.Cells[newPosition]
		removeCrowdMemberFromCell(currentCell, e)
		newCell.Persons = append(newCell.Persons, e)
	default:
		fmt.Println("Unknown entity type; cannot move entity")
		//Print type of entity
		fmt.Println(reflect.TypeOf(entity))
		//fmt.Println("Unknown entity type")
	}
}

func (m *Map) RemoveEntity(entity interface{}) {
	var currentCell *MapCell
	switch e := entity.(type) {
	case *drones.Drone:
		currentCell = m.Cells[e.Position]
		removeDroneFromCell(currentCell, e)
	case *persons.Person:
		currentCell = m.Cells[e.Position]
		removeCrowdMemberFromCell(currentCell, e)
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

// removeDroneFromCell removes a drones from a map cell
func removeDroneFromCell(cell *MapCell, drone *drones.Drone) {
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
func removeCrowdMemberFromCell(cell *MapCell, member *persons.Person) {
	moved := false
	fmt.Printf("Removing crowd member %d from cell (%.2f, %.2f) \n", member.ID, cell.Position.X, cell.Position.Y)
	for i, m := range cell.Persons {
		if m.ID == member.ID {
			cell.Persons = append(cell.Persons[:i], cell.Persons[i+1:]...)
			moved = true
			for _, c := range cell.Persons {
				fmt.Printf("Remaining member %d in cell\n", c.ID)
			}
			break
		}
	}
	if !moved {
		fmt.Println("Crowd member not found in cell")
	}
}
