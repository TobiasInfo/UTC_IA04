// main_gui.go
// Replace entire file contents in cmd/gui/main_gui.go

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
	colorEmpty    = color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	colorDrone    = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	colorCrowd    = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	colorDistress = color.NRGBA{R: 255, G: 165, B: 0, A: 255} // Orange for distress
	colorBorder   = color.NRGBA{R: 100, G: 100, B: 100, A: 255}

	// Zone colors
	colorEntrance = color.NRGBA{R: 220, G: 240, B: 220, A: 255} // Light green
	colorMain     = color.NRGBA{R: 240, G: 220, B: 220, A: 255} // Light red
	colorExit     = color.NRGBA{R: 220, G: 220, B: 240, A: 255} // Light blue

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
)

const (
	cellSize      float32 = 25.0
	defaultWidth          = 30
	defaultHeight         = 20
	initialDrones         = 5
	initialCrowd          = 20
)

type SimulationGUI struct {
	sim            *simulation.Simulation
	grid           *fyne.Container
	isRunning      bool
	stopChannel    chan bool
	mutex          sync.Mutex
	timeLabel      *widget.Label
	phaseLabel     *widget.Label
	statsLabel     *widget.Label
	zoneStatsLabel *widget.Label
	timeScale      float64
}

func createCell(bgColor color.Color) *canvas.Rectangle {
	cell := canvas.NewRectangle(bgColor)
	cell.Resize(fyne.NewSize(cellSize, cellSize))
	cell.StrokeWidth = 0.5
	cell.StrokeColor = colorBorder
	return cell
}

func initGrid(grid *fyne.Container, width, height int) {
	zoneWidth := width / 3

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var bgColor color.Color

			// Set zone background colors
			if x < zoneWidth {
				bgColor = colorEntrance
			} else if x < zoneWidth*2 {
				bgColor = colorMain
			} else {
				bgColor = colorExit
			}

			cell := createCell(bgColor)
			grid.Add(cell)
		}
	}
	grid.Refresh()
}

func updateGridFromSimulation(sim *simulation.Simulation, grid *fyne.Container) {
	width, height := defaultWidth, defaultHeight
	zoneWidth := width / 3

	// First, reset all cells to their zone colors
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			if cell, ok := grid.Objects[idx].(*canvas.Rectangle); ok {
				if x < zoneWidth {
					cell.FillColor = colorEntrance
				} else if x < zoneWidth*2 {
					cell.FillColor = colorMain
				} else {
					cell.FillColor = colorExit
				}
			}
		}
	}

	// Draw POIs
	for _, cell := range sim.Map.Cells {
		for _, obstacle := range cell.Obstacles {
			x, y := int(obstacle.Position.X), int(obstacle.Position.Y)
			idx := y*width + x
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
			x, y := int(person.Position.X), int(person.Position.Y)
			idx := y*width + x
			if idx >= 0 && idx < len(grid.Objects) {
				if cell, ok := grid.Objects[idx].(*canvas.Rectangle); ok {
					if person.InDistress {
						cell.FillColor = colorDistress
					} else {
						cell.FillColor = colorCrowd
					}
				}
			}
		}
	}

	// Draw drones
	for _, cell := range sim.Map.Cells {
		for range cell.Drones {
			x, y := int(cell.Position.X), int(cell.Position.Y)
			idx := y*width + x
			if idx >= 0 && idx < len(grid.Objects) {
				if cell, ok := grid.Objects[idx].(*canvas.Rectangle); ok {
					cell.FillColor = colorDrone
				}
			}
		}
	}

	grid.Refresh()
}

func createControlPanel(sim *SimulationGUI) *fyne.Container {
	// Time scale control
	timeScaleSlider := widget.NewSlider(1, 120)
	timeScaleSlider.Value = sim.timeScale
	timeScaleLabel := widget.NewLabel(fmt.Sprintf("Time Scale: %.0fx", sim.timeScale))

	timeScaleSlider.OnChanged = func(value float64) {
		sim.timeScale = value
		timeScaleLabel.SetText(fmt.Sprintf("Time Scale: %.0fx", value))
		sim.sim.GetFestivalTime().SetTimeScale(value)
	}

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

	// Statistics labels
	sim.timeLabel = widget.NewLabel("Time: 00:00:00")
	sim.phaseLabel = widget.NewLabel("Phase: Setup")
	sim.statsLabel = widget.NewLabel("Simulation Statistics")
	sim.zoneStatsLabel = widget.NewLabel("Zone Statistics")

	return container.NewVBox(
		widget.NewLabel("Control Panel"),
		widget.NewSeparator(),
		timeScaleLabel,
		timeScaleSlider,
		widget.NewSeparator(),
		crowdLabel,
		crowdSlider,
		widget.NewSeparator(),
		droneLabel,
		droneSlider,
		widget.NewSeparator(),
		sim.timeLabel,
		sim.phaseLabel,
		widget.NewSeparator(),
		sim.statsLabel,
		widget.NewSeparator(),
		sim.zoneStatsLabel,
	)
}

func (s *SimulationGUI) updateStatistics() {
	stats := s.sim.GetStatistics()
	festivalTime := s.sim.GetFestivalTime()

	currentTime := time.Now().Add(festivalTime.GetElapsedTime())
	s.timeLabel.SetText(fmt.Sprintf("Time: %02d:%02d:%02d",
		currentTime.Hour(), currentTime.Minute(), currentTime.Second()))

	s.phaseLabel.SetText(fmt.Sprintf("Phase: %s", stats.CurrentPhase))

	s.statsLabel.SetText(fmt.Sprintf(
		"Statistics:\n"+
			"Total People: %d\n"+
			"People in Distress: %d (%.1f%%)\n"+
			"Active Drones: %d\n"+
			"Time Remaining: %v",
		stats.TotalPeople,
		stats.PeopleInDistress,
		float64(stats.PeopleInDistress)/float64(stats.TotalPeople)*100,
		stats.ActiveDrones,
		stats.RemainingTime.Round(time.Second),
	))

	s.zoneStatsLabel.SetText(fmt.Sprintf(
		"Zone Distribution:\n"+
			"Entrance: %d (%.1f%%)\n"+
			"Main: %d (%.1f%%)\n"+
			"Exit: %d (%.1f%%)",
		stats.ZoneStatistics["entrance"],
		float64(stats.ZoneStatistics["entrance"])/float64(stats.TotalPeople)*100,
		stats.ZoneStatistics["main"],
		float64(stats.ZoneStatistics["main"])/float64(stats.TotalPeople)*100,
		stats.ZoneStatistics["exit"],
		float64(stats.ZoneStatistics["exit"])/float64(stats.TotalPeople)*100,
	))
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
			s.updateStatistics()
			s.mutex.Unlock()

			if window := fyne.CurrentApp().Driver().CanvasForObject(s.grid); window != nil {
				s.mutex.Lock()
				updateGridFromSimulation(s.sim, s.grid)
				window.Refresh(s.grid)
				s.mutex.Unlock()
			}

			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Festival Safety Simulation")

	grid := container.NewGridWithColumns(defaultWidth)
	initGrid(grid, defaultWidth, defaultHeight)

	simGUI := &SimulationGUI{
		sim:         simulation.NewSimulation(initialDrones, initialCrowd, 0), // Using POIs from config
		grid:        grid,
		isRunning:   false,
		stopChannel: make(chan bool),
		timeScale:   1.0,
	}

	controlPanel := createControlPanel(simGUI)

	// Create toolbar with buttons
	controls := container.NewHBox(
		widget.NewButton("Start", func() {
			if !simGUI.isRunning {
				simGUI.mutex.Lock()
				simGUI.isRunning = true
				simGUI.mutex.Unlock()
				go simGUI.runSimulation()
			}
		}),
		widget.NewButton("Stop", func() {
			simGUI.mutex.Lock()
			simGUI.isRunning = false
			simGUI.mutex.Unlock()
			simGUI.stopChannel <- true
		}),
		widget.NewButton("Reset", func() {
			simGUI.mutex.Lock()
			wasRunning := simGUI.isRunning
			if wasRunning {
				simGUI.isRunning = false
				simGUI.stopChannel <- true
			}
			simGUI.sim = simulation.NewSimulation(initialDrones, initialCrowd, 0)
			if wasRunning {
				simGUI.isRunning = true
				go simGUI.runSimulation()
			}
			simGUI.mutex.Unlock()
		}),
	)

	// Create legend for colors
	legend := widget.NewTextGrid()
	legendText := "Legend:\n" +
		"■ Crowd (Red)\n" +
		"■ Distress (Orange)\n" +
		"■ Drone (Blue)\n" +
		"■ Medical (White)\n" +
		"■ Charging (Yellow)\n" +
		"■ Toilets (Purple)\n" +
		"■ Drinks (Cyan)\n" +
		"■ Food (Orange)\n" +
		"■ Main Stage (Pink)\n" +
		"■ Secondary (Rose)\n" +
		"■ Rest Area (Green)\n\n" +
		"Zones:\n" +
		"■ Entrance (Light Green)\n" +
		"■ Main (Light Red)\n" +
		"■ Exit (Light Blue)"
	legend.SetText(legendText)

	// Layout setup
	rightPanel := container.NewVBox(controlPanel, widget.NewSeparator(), legend)
	content := container.NewBorder(
		nil,
		controls,
		nil,
		rightPanel,
		container.NewPadded(grid),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(1400, 800))
	myWindow.ShowAndRun()
}
