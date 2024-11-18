package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// Simulation parameters
var (
	numDrones       = 5
	numCrowdMembers = 10
	numObstacles    = 3
	mapWidth        = 20
	mapHeight       = 20
)

// CellContent represents what is in a cell
type CellContent struct {
	Drones       int
	CrowdMembers int
	Obstacles    int
}

// MapCell represents a single cell on the map
type MapCell struct {
	content *CellContent
	rect    *canvas.Rectangle
	label   *widget.Label // Changed to widget.Label for better text handling
}

// Create the grid once and initialize cells
func createInitialMap(grid *fyne.Container, contents [][]*CellContent) {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			cell := &MapCell{
				content: contents[y][x],
				rect:    canvas.NewRectangle(color.RGBA{220, 220, 220, 255}),
				label:   widget.NewLabel(""), // Using widget.Label instead of canvas.Text
			}

			cell.rect.SetMinSize(fyne.NewSize(40, 40))
			cell.label.Alignment = fyne.TextAlignCenter

			// Create a centered container for the label
			labelContainer := container.NewCenter(cell.label)

			// Layer the components
			cellContainer := container.NewStack(
				cell.rect,
				labelContainer,
			)

			grid.Add(cellContainer)
		}
	}
}

// Update map cells without recreating the grid
func updateMap(contents [][]*CellContent, grid *fyne.Container) {
	for y := 0; y < mapHeight; y++ {
		for x := 0; x < mapWidth; x++ {
			cell := grid.Objects[y*mapWidth+x].(*fyne.Container)
			content := contents[y][x]

			// Update color based on content
			rect := cell.Objects[0].(*canvas.Rectangle)
			label := cell.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)

			switch {
			case content.Drones > 0:
				rect.FillColor = color.RGBA{0, 0, 255, 255} // Blue for drones
			case content.CrowdMembers > 0:
				rect.FillColor = color.RGBA{255, 0, 0, 255} // Red for crowd members
			case content.Obstacles > 0:
				rect.FillColor = color.RGBA{0, 0, 0, 255} // Black for obstacles
			default:
				rect.FillColor = color.RGBA{220, 220, 220, 255} // Default light gray
			}

			// Update label with better formatting
			if content.Drones > 0 || content.CrowdMembers > 0 || content.Obstacles > 0 {
				label.SetText(fmt.Sprintf("D:%d\nC:%d\nO:%d",
					content.Drones, content.CrowdMembers, content.Obstacles))
			} else {
				label.SetText("") // Empty cell
			}

			rect.Refresh()
		}
	}
}

// Generate random content
func generateRandomContent() [][]*CellContent {
	rand.Seed(time.Now().UnixNano())
	contents := make([][]*CellContent, mapHeight)

	// Initialize empty grid
	for y := 0; y < mapHeight; y++ {
		row := make([]*CellContent, mapWidth)
		for x := 0; x < mapWidth; x++ {
			row[x] = &CellContent{}
		}
		contents[y] = row
	}

	// Place entities
	placeEntities := func(count int, updateFunc func(cell *CellContent)) {
		for i := 0; i < count; i++ {
			x, y := rand.Intn(mapWidth), rand.Intn(mapHeight)
			updateFunc(contents[y][x])
		}
	}

	placeEntities(numDrones, func(cell *CellContent) { cell.Drones++ })
	placeEntities(numCrowdMembers, func(cell *CellContent) { cell.CrowdMembers++ })
	placeEntities(numObstacles, func(cell *CellContent) { cell.Obstacles++ })

	return contents
}

// Create settings panel with input validation
func createSettingsPanel(grid *fyne.Container, contents [][]*CellContent) *fyne.Container {
	droneEntry := widget.NewEntry()
	droneEntry.SetPlaceHolder("Enter # of Drones")

	crowdEntry := widget.NewEntry()
	crowdEntry.SetPlaceHolder("Enter # of Crowd Members")

	obstacleEntry := widget.NewEntry()
	obstacleEntry.SetPlaceHolder("Enter # of Obstacles")

	applyButton := widget.NewButton("Apply Settings", func() {
		// Parse and validate input values
		if d, err := strconv.Atoi(droneEntry.Text); err == nil && d >= 0 {
			numDrones = d
		}
		if c, err := strconv.Atoi(crowdEntry.Text); err == nil && c >= 0 {
			numCrowdMembers = c
		}
		if o, err := strconv.Atoi(obstacleEntry.Text); err == nil && o >= 0 {
			numObstacles = o
		}

		// Generate new random content and update
		contents = generateRandomContent()
		updateMap(contents, grid)
		grid.Refresh()
	})

	return container.NewVBox(
		widget.NewLabel("Settings"),
		droneEntry,
		crowdEntry,
		obstacleEntry,
		applyButton,
		layout.NewSpacer(),
	)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Enhanced Drone Simulation")

	grid := container.NewGridWithColumns(mapWidth)
	contents := generateRandomContent()
	createInitialMap(grid, contents)

	settingsPanel := createSettingsPanel(grid, contents)

	content := container.NewBorder(
		nil, nil, nil, settingsPanel,
		grid,
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(1000, 700))
	myWindow.ShowAndRun()
}