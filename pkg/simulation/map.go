package simulation

import (
	"UTC_IA04/pkg/entities/drones"
	"UTC_IA04/pkg/entities/obstacles"
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"fmt"
	"sync"
)

type MapCell struct {
	Position  models.Position
	Obstacles []*obstacles.Obstacle
	Persons   []*persons.Person
	Drones    []*drones.Drone
	//mu        sync.RWMutex
}

type Map struct {
	Width     int
	Height    int
	Obstacles []*obstacles.Obstacle
	Cells     map[models.Position]*MapCell
	debug     bool
	hardDebug bool
	mu        sync.RWMutex
}

var (
	instance *Map      // The singleton instance
	once     sync.Once // Ensures Map is initialized only once
)

func newMap(width, height int) *Map {
	cells := make(map[models.Position]*MapCell)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			for i := 0; i < 10; i++ {
				for j := 0; j < 10; j++ {
					pos := models.Position{X: float64(x) + (float64(i) / 10), Y: float64(y) + (float64(j) / 10)}
					cells[pos] = &MapCell{
						Position:  pos,
						Obstacles: []*obstacles.Obstacle{},
						Persons:   []*persons.Person{},
						Drones:    []*drones.Drone{},
					}
				}
			}
		}
	}

	cimitierePos := models.Position{X: -10, Y: -10}
	cells[cimitierePos] = &MapCell{
		Position:  cimitierePos,
		Obstacles: []*obstacles.Obstacle{},
		Persons:   []*persons.Person{},
		Drones:    []*drones.Drone{},
	}

	return &Map{
		Width:     width,
		Height:    height,
		Cells:     cells,
		debug:     false,
		hardDebug: false,
		mu:        sync.RWMutex{},
	}
}

func GetMap(width, height int) *Map {
	once.Do(func() {
		instance = newMap(width, height)
	})
	return instance
}

func (m *Map) AddObstacle(obstacle *obstacles.Obstacle) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cell := m.Cells[obstacle.Position]
	cell.Obstacles = append(cell.Obstacles, obstacle)
	m.Obstacles = append(m.Obstacles, obstacle)
}

func (m *Map) AddCrowdMember(member *persons.Person) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cell := m.Cells[member.Position]
	cell.Persons = append(cell.Persons, member)
}

func (m *Map) AddDrone(drone *drones.Drone) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cell := m.Cells[drone.Position]
	cell.Drones = append(cell.Drones, drone)
}

func (m *Map) MoveEntity(entity interface{}, newPosition models.Position) {
	var currentCell, newCell *MapCell

	m.mu.RLock()
	switch e := entity.(type) {
	case *drones.Drone:
		currentCell = m.Cells[e.Position]
		newCell = m.Cells[newPosition]
		removeDroneFromCell(currentCell, e)
		newCell.Drones = append(newCell.Drones, e)
		e.Position = newPosition

	case *persons.Person:
		currentCell = m.Cells[e.Position]
		newCell = m.Cells[newPosition]

		if m.hardDebug {
			fmt.Printf("Moving person %d from %v to %v\n", e.ID, e.Position, newPosition)
		}

		if currentCell.Position.X < newCell.Position.X ||
			(currentCell.Position.X == newCell.Position.X && currentCell.Position.Y < newCell.Position.Y) {
		}

		removeCrowdMemberFromCell(currentCell, e)
		newCell.Persons = append(newCell.Persons, e)
		e.Position = newPosition

	default:
		fmt.Println("Unknown entity type; cannot move entity")
	}
	m.mu.RUnlock()
}

func (m *Map) RemoveEntity(entity interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *Map) IsBlocked(position models.Position) bool {
	m.mu.RLock()
	cell, exists := m.Cells[position]
	m.mu.RUnlock()

	if !exists {
		return true
	}
	return len(cell.Obstacles) > 0
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

func removeCrowdMemberFromCell(cell *MapCell, member *persons.Person) {
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
