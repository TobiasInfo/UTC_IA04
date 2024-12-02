package main

import (
	"UTC_IA04/pkg/simulation" // Ajustez le chemin selon votre projet
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
	colorEmpty  = color.RGBA{R: 200, G: 200, B: 200, A: 255}
	colorDrone  = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	colorCrowd  = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	colorBorder = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	colorSeen   = color.RGBA{R: 80, G: 255, B: 20, A: 255} // New color
)

// Simulation parameters
var (
	numDrones       = 5
	numCrowdMembers = 2
	numObstacles    = 3
	cellSize        = 40.0 // Size of each cell in pixels
)

const (
	mapWidth  = 3
	mapHeight = 3
)

type SimulationGUI struct {
	sim           *simulation.Simulation
	grid          *fyne.Container
	isRunning     bool
	stopChannel   chan bool
	mutex         sync.Mutex
	distressLabel *widget.Label
}

// Fonction pour normaliser les coordonnées dans les limites de la grille
func normalizeCoordinates(x, y float64) (int, int) {
	// Gestion des coordonnées négatives avec modulo
	normX := ((int(x) % mapWidth) + mapWidth) % mapWidth
	normY := ((int(y) % mapHeight) + mapHeight) % mapHeight
	return normX, normY
}

// Initialisation de la grille (à faire une seule fois au début)
func initGrid(grid *fyne.Container) {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			cell := createCell()
			grid.Add(cell)
		}
	}
	grid.Refresh()
}

// Mise à jour de la grille (à chaque frame)
func updateGridFromSimulation(sim *simulation.Simulation, grid *fyne.Container) {
	// Réinitialiser toutes les cellules
	for i := range grid.Objects {
		if cell, ok := grid.Objects[i].(*canvas.Rectangle); ok {
			cell.FillColor = colorEmpty
		}
	}

	// Mise à jour des positions des drones
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

	// Mise à jour des positions de la foule
	for _, cell := range sim.Map.Cells {
		for _, member := range cell.Persons {
			x, y := normalizeCoordinates(member.Position.X, member.Position.Y)
			idx := y*mapWidth + x
			if idx >= 0 && idx < len(grid.Objects) {
				if cell, ok := grid.Objects[idx].(*canvas.Rectangle); ok {
					cell.FillColor = colorCrowd
					// TODO : devide the cell to display several points in the same cell
					// Each point should represent a crowd member in the same cell with his float coordinates
				}
			}
		}
	}

	for _, cell := range sim.Map.Cells {
		for _, drone := range cell.Drones {
			for _, member := range drone.SeenPeople {
				x, y := normalizeCoordinates(member.Position.X, member.Position.Y)
				idx := y*mapWidth + x
				if idx >= 0 && idx < len(grid.Objects) {
					if cell, ok := grid.Objects[idx].(*canvas.Rectangle); ok {
						cell.FillColor = colorSeen
					}
				}
			}
		}
	}

	grid.Refresh()
}

func createCell() *canvas.Rectangle {
	cell := canvas.NewRectangle(colorEmpty)
	cell.Resize(fyne.NewSize(20, 20))
	cell.StrokeWidth = 0
	cell.StrokeColor = colorBorder
	return cell
}

func createControlPanel(sim *SimulationGUI) *fyne.Container {
	// Slider pour les membres de la foule
	crowdSlider := widget.NewSlider(1, 500) // Min 1, Max 50 membres
	crowdSlider.Value = float64(sim.sim.Map.CountCrowdMembers())
	crowdLabel := widget.NewLabel("Nombre de membres : " + fmt.Sprintf("%d", int(crowdSlider.Value)))

	crowdSlider.OnChanged = func(value float64) {
		crowdLabel.SetText("Nombre de membres : " + fmt.Sprintf("%d", int(value)))
		// Mettre à jour la simulation avec le nouveau nombre
		sim.mutex.Lock()
		sim.sim.UpdateCrowdSize(int(value))
		sim.mutex.Unlock()
	}

	dronesSlider := widget.NewSlider(1, 20) // Min 1, Max 20 drones
	dronesSlider.Value = float64(sim.sim.Map.CountDrones())
	dronesLabel := widget.NewLabel("Nombre de drones : " + fmt.Sprintf("%d", int(dronesSlider.Value)))

	dronesSlider.OnChanged = func(value float64) {
		dronesLabel.SetText("Nombre de drones : " + fmt.Sprintf("%d", int(value)))
		// Mettre à jour la simulation avec le nouveau nombre
		sim.mutex.Lock()
		sim.sim.UpdateDroneSize(int(value))
		sim.mutex.Unlock()
	}

	// Display the metrics
	// TODO : display the metrics of the simulation
	// Print the number crowd members in destress

	cmInDistress := sim.sim.CountCrowdMembersInDistress()
	sim.distressLabel = widget.NewLabel("Nombre de membres en détresse : " + fmt.Sprintf("%d", cmInDistress))

	return container.NewVBox(
		widget.NewLabel("Contrôles"),
		widget.NewSeparator(),
		crowdLabel,
		crowdSlider,
		widget.NewSeparator(),
		dronesLabel,
		dronesSlider,
		widget.NewSeparator(),
		sim.distressLabel,
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

			// Mise à jour UI thread-safe
			if window := fyne.CurrentApp().Driver().CanvasForObject(s.grid); window != nil {
				s.mutex.Lock()
				cmInDistress := s.sim.CountCrowdMembersInDistress()
				s.distressLabel.SetText(fmt.Sprintf("Nombre de membres en détresse : %d", cmInDistress))
				updateGridFromSimulation(s.sim, s.grid)
				window.Refresh(s.grid)
				s.mutex.Unlock()
			}

			// Vaut mieux gérer le ticking depuis la simulation elle même.

			time.Sleep(2000 * time.Millisecond)
		}
	}
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Simulation")

	// Création de la grille
	grid := container.NewGridWithColumns(mapWidth)
	initGrid(grid)

	simGUI := &SimulationGUI{
		sim:         simulation.NewSimulation(numDrones, numCrowdMembers, numObstacles),
		grid:        grid,
		isRunning:   false,
		stopChannel: make(chan bool),
	}

	// Création du panneau de contrôle
	controlPanel := createControlPanel(simGUI)

	// Boutons de contrôle
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

	// Layout principal avec la grille au centre et le panneau de contrôle à droite
	content := container.NewBorder(
		nil,          // top
		controls,     // bottom
		nil,          // left
		controlPanel, // right
		grid,         // center
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
