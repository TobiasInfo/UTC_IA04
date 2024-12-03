package main

import (
	"UTC_IA04/pkg/models"
	"UTC_IA04/pkg/simulation"
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	// Basic colors
	colorEmpty  = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	colorDrone  = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	colorCrowd  = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	colorBorder = color.NRGBA{R: 100, G: 100, B: 100, A: 255}
	colorSeen   = color.NRGBA{R: 80, G: 255, B: 20, A: 255}

	// POI Colors map
	poiColors = map[models.POIType]color.NRGBA{
		models.MedicalTent:     {R: 255, G: 255, B: 255, A: 255}, // White
		models.ChargingStation: {R: 255, G: 255, B: 0, A: 255},   // Yellow
		models.Toilet:          {R: 128, G: 0, B: 128, A: 255},   // Purple
		models.DrinkStand:      {R: 0, G: 255, B: 255, A: 255},   // Cyan
		models.FoodStand:       {R: 255, G: 165, B: 0, A: 255},   // Orange
		models.MainStage:       {R: 255, G: 20, B: 147, A: 255},  // Pink
		models.SecondaryStage:  {R: 219, G: 112, B: 147, A: 255}, // PaleVioletRed
		models.RestArea:        {R: 34, G: 139, B: 34, A: 255},   // ForestGreen
	}

	// POI Names map
	poiNames = map[models.POIType]string{
		models.MedicalTent:     "Medical Tent",
		models.ChargingStation: "Charging Station",
		models.Toilet:          "Toilet",
		models.DrinkStand:      "Drink Stand",
		models.FoodStand:       "Food Stand",
		models.MainStage:       "Main Stage",
		models.SecondaryStage:  "Secondary Stage",
		models.RestArea:        "Rest Area",
	}
)

// Simulation parameters
var (
	numDrones               = 5
	numCrowdMembers         = 2
	numObstacles            = 3
	cellSize        float32 = 10.0
)

const (
	mapWidth  = 5
	mapHeight = 5
)

type SimulationGUI struct {
	sim           *simulation.Simulation
	grid          *fyne.Container
	isRunning     bool
	stopChannel   chan bool
	mutex         sync.Mutex
	distressLabel *widget.Label
	statsLabel    *widget.Label
}

func normalizeCoordinates(x, y float64) (int, int) {
	normX := ((int(x) % mapWidth) + mapWidth) % mapWidth
	normY := ((int(y) % mapHeight) + mapHeight) % mapHeight
	return normX, normY
}

func createCell() *canvas.Rectangle {
	cell := canvas.NewRectangle(colorEmpty)
	cell.Resize(fyne.NewSize(cellSize, cellSize))
	cell.StrokeWidth = 1
	cell.StrokeColor = colorBorder
	return cell
}

func initGrid(grid *fyne.Container) {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			cell := createCell()
			grid.Add(cell)
		}
	}
	grid.Refresh()
}

func updateGridFromSimulation(sim *simulation.Simulation, grid *fyne.Container) {
	// Reset all cells
	for i := range grid.Objects {
		if cell, ok := grid.Objects[i].(*canvas.Rectangle); ok {
			cell.FillColor = colorEmpty
		}
	}

	// Draw POIs first (background)
	for _, cell := range sim.Map.Cells {
		for _, obstacle := range cell.Obstacles {
			x, y := normalizeCoordinates(obstacle.Position.X, obstacle.Position.Y)
			idx := y*mapWidth + x
			if idx >= 0 && idx < len(grid.Objects) {
				if cell, ok := grid.Objects[idx].(*canvas.Rectangle); ok {
					if color, exists := poiColors[obstacle.POIType]; exists {
						cell.FillColor = color
					}
				}
			}
		}
	}

	// Draw people
	for _, cell := range sim.Map.Cells {
		for _, person := range cell.Persons {
			x, y := normalizeCoordinates(person.Position.X, person.Position.Y)
			idx := y*mapWidth + x
			if idx >= 0 && idx < len(grid.Objects) {
				if cell, ok := grid.Objects[idx].(*canvas.Rectangle); ok {
					if person.InDistress {
						cell.FillColor = colorSeen
					} else {
						cell.FillColor = colorCrowd
					}
				}
			}
		}
	}

	// Draw drones (top layer)
	for _, cell := range sim.Map.Cells {
		for _, drone := range cell.Drones {
			x, y := normalizeCoordinates(drone.Position.X, drone.Position.Y)
			idx := y*mapWidth + x
			if idx >= 0 && idx < len(grid.Objects) {
				if cell, ok := grid.Objects[idx].(*canvas.Rectangle); ok {
					cell.FillColor = colorDrone
				}
			}
		}
	}

	grid.Refresh()
}

func createInfoPanel() *widget.Label {
	info := "Festival Simulation\n\n" +
		"Entities:\n" +
		"■ Drone (Blue): Surveillance units\n" +
		"■ Person (Red): Festival attendee\n" +
		"■ Distress (Green): Person needing assistance\n\n" +
		"Facilities:\n"

	for poiType, name := range poiNames {
		if color, exists := poiColors[poiType]; exists {
			info += fmt.Sprintf("■ %s (RGB: %d,%d,%d)\n", name, color.R, color.G, color.B)
		}
	}

	return widget.NewLabel(info)
}

func createControlPanel(sim *SimulationGUI) *fyne.Container {
	// Crowd control
	crowdSlider := widget.NewSlider(1, 500)
	crowdSlider.Value = float64(sim.sim.Map.CountCrowdMembers())
	crowdLabel := widget.NewLabel(fmt.Sprintf("People: %d", int(crowdSlider.Value)))

	crowdSlider.OnChanged = func(value float64) {
		crowdLabel.SetText(fmt.Sprintf("People: %d", int(value)))
		sim.mutex.Lock()
		sim.sim.UpdateCrowdSize(int(value))
		sim.mutex.Unlock()
	}

	// Drone control
	droneSlider := widget.NewSlider(1, 20)
	droneSlider.Value = float64(sim.sim.Map.CountDrones())
	droneLabel := widget.NewLabel(fmt.Sprintf("Drones: %d", int(droneSlider.Value)))

	droneSlider.OnChanged = func(value float64) {
		droneLabel.SetText(fmt.Sprintf("Drones: %d", int(value)))
		sim.mutex.Lock()
		sim.sim.UpdateDroneSize(int(value))
		sim.mutex.Unlock()
	}

	// Statistics
	sim.distressLabel = widget.NewLabel(fmt.Sprintf("People in distress: %d", sim.sim.CountCrowdMembersInDistress()))
	sim.statsLabel = widget.NewLabel("Simulation Statistics")

	// Info panel
	infoPanel := createInfoPanel()

	return container.NewVBox(
		widget.NewLabel("Control Panel"),
		widget.NewSeparator(),
		crowdLabel,
		crowdSlider,
		widget.NewSeparator(),
		droneLabel,
		droneSlider,
		widget.NewSeparator(),
		sim.distressLabel,
		sim.statsLabel,
		widget.NewSeparator(),
		container.NewHScroll(infoPanel),
	)
}

func (s *SimulationGUI) runSimulation() {
	for {
		select {
		case <-s.stopChannel:
			return
		default:
			s.mutex.Lock()
			if !s.isRunning {
				s.mutex.Unlock()
				return
			}
			s.sim.Update()
			s.mutex.Unlock()

			if window := fyne.CurrentApp().Driver().CanvasForObject(s.grid); window != nil {
				s.mutex.Lock()
				distressCount := s.sim.CountCrowdMembersInDistress()
				s.distressLabel.SetText(fmt.Sprintf("People in distress: %d", distressCount))

				totalPeople := s.sim.Map.CountCrowdMembers()
				distressRate := 0.0
				if totalPeople > 0 {
					distressRate = float64(distressCount) / float64(totalPeople) * 100
				}

				s.statsLabel.SetText(fmt.Sprintf(
					"Statistics:\n"+
						"Total People: %d\n"+
						"Active Drones: %d\n"+
						"Distress Rate: %.1f%%",
					totalPeople,
					s.sim.Map.CountDrones(),
					distressRate,
				))

				updateGridFromSimulation(s.sim, s.grid)
				window.Refresh(s.grid)
				s.mutex.Unlock()
			}

			time.Sleep(2000 * time.Millisecond)
		}
	}
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Festival Safety Simulation")

	grid := container.NewGridWithColumns(mapWidth)
	initGrid(grid)

	simGUI := &SimulationGUI{
		sim:         simulation.NewSimulation(numDrones, numCrowdMembers, numObstacles),
		grid:        grid,
		isRunning:   false,
		stopChannel: make(chan bool),
	}

	controlPanel := createControlPanel(simGUI)

	controls := container.NewHBox(
		widget.NewButton("Start", func() {
			if !simGUI.isRunning {
				simGUI.isRunning = true
				go simGUI.runSimulation()
			}
		}),
		widget.NewButton("Stop", func() {
			simGUI.mutex.Lock()
			simGUI.isRunning = false
			simGUI.mutex.Unlock()
			simGUI.stopChannel <- true
		}),
	)

	content := container.NewBorder(
		nil,
		controls,
		nil,
		controlPanel,
		container.NewPadded(grid),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(1400, 800))
	myWindow.ShowAndRun()
}
