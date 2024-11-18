package main

import (
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
    cellSize        = 40.0 // Size of each cell in pixels
)

// Point represents a position with floating-point coordinates
type Point struct {
    X, Y float64
}

// CellContent represents what is in a cell
type CellContent struct {
    Drones        int
    CrowdMembers  []Point // Now stores points for precise positioning
    Obstacles     int
}

// MapCell represents a single cell on the map
type MapCell struct {
    content    *CellContent
    background *canvas.Rectangle
    entities   *fyne.Container // Container for all shapes in the cell
}

// createShape creates a circle or square based on entity type
func createShape(entityType string, x, y, size float64) fyne.CanvasObject {
    switch entityType {
    case "drone":
        circle := canvas.NewCircle(color.RGBA{0, 0, 255, 255}) // Blue drone
		circle.StrokeWidth = 0
        circle.Resize(fyne.NewSize(float32(size), float32(size)))
        circle.Move(fyne.NewPos(float32(x), float32(y)))
        return circle
    case "crowd":
        circle := canvas.NewCircle(color.RGBA{255, 255, 255, 255}) // White crowd member
		circle.StrokeWidth = 0
        circle.Resize(fyne.NewSize(float32(size), float32(size)))
        circle.Move(fyne.NewPos(float32(x), float32(y)))
        return circle
    case "obstacle":
        square := canvas.NewRectangle(color.RGBA{0, 0, 0, 255}) // Black obstacle
		square.StrokeWidth = 0
        square.Resize(fyne.NewSize(float32(size), float32(size)))
        square.Move(fyne.NewPos(float32(x), float32(y)))
        return square
    default:
        return nil
    }
}

func createInitialMap(grid *fyne.Container, contents [][]*CellContent) {
    for y := 0; y < mapHeight; y++ {
        for x := 0; x < mapWidth; x++ {
            // Create background without border
            background := canvas.NewRectangle(color.RGBA{220, 220, 220, 255})
            background.StrokeWidth = 0 // Remove border
            background.FillColor = color.RGBA{220, 220, 220, 255}
            background.Resize(fyne.NewSize(float32(cellSize), float32(cellSize)))
            
            entities := container.NewWithoutLayout()
            
            _ = &MapCell{
                content:    contents[y][x],
                background: background,
                entities:   entities,
            }

            // Use MaxLayout to ensure background fills the entire cell space
            cellContainer := container.New(layout.NewMaxLayout(), background, entities)
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
            
            // Clear existing entities
            entitiesContainer := cell.Objects[1].(*fyne.Container)
            entitiesContainer.Objects = nil

            // Add drone if present (large circle)
            if content.Drones > 0 {
                droneShape := createShape("drone", 2, 2, cellSize-4)
                entitiesContainer.Add(droneShape)
            }

            // Add obstacles if present (square)
            if content.Obstacles > 0 {
                obstacleShape := createShape("obstacle", 2, 2, cellSize-4)
                entitiesContainer.Add(obstacleShape)
            }

            // Add crowd members (small circles with specific positions)
            for _, point := range content.CrowdMembers {
                // Convert relative coordinates (0-1) to pixel positions within cell
                x := point.X * (cellSize - 6) // Subtract size to keep within bounds
                y := point.Y * (cellSize - 6)
                crowdShape := createShape("crowd", x+2, y+2, 6) // Small 6-pixel circles
                entitiesContainer.Add(crowdShape)
            }

            entitiesContainer.Refresh()
        }
    }
}

// Generate random content with floating-point positions for crowd members
func generateRandomContent() [][]*CellContent {
    rand.Seed(time.Now().UnixNano())
    contents := make([][]*CellContent, mapHeight)

    // Initialize empty grid
    for y := 0; y < mapHeight; y++ {
        row := make([]*CellContent, mapWidth)
        for x := 0; x < mapWidth; x++ {
            row[x] = &CellContent{
                CrowdMembers: make([]Point, 0),
            }
        }
        contents[y] = row
    }

    // Place drones
    for i := 0; i < numDrones; i++ {
        x, y := rand.Intn(mapWidth), rand.Intn(mapHeight)
        contents[y][x].Drones++
    }

    // Place crowd members with random positions within cells
    for i := 0; i < numCrowdMembers; i++ {
        x, y := rand.Intn(mapWidth), rand.Intn(mapHeight)
        point := Point{
            X: rand.Float64(), // Random position within cell (0-1)
            Y: rand.Float64(),
        }
        contents[y][x].CrowdMembers = append(contents[y][x].CrowdMembers, point)
    }

    // Place obstacles
    for i := 0; i < numObstacles; i++ {
        x, y := rand.Intn(mapWidth), rand.Intn(mapHeight)
        contents[y][x].Obstacles++
    }

    return contents
}

// Create settings panel
func createSettingsPanel(grid *fyne.Container, contents [][]*CellContent) *fyne.Container {
    droneEntry := widget.NewEntry()
    droneEntry.SetPlaceHolder("Enter # of Drones")

    crowdEntry := widget.NewEntry()
    crowdEntry.SetPlaceHolder("Enter # of Crowd Members")

    obstacleEntry := widget.NewEntry()
    obstacleEntry.SetPlaceHolder("Enter # of Obstacles")

    applyButton := widget.NewButton("Apply Settings", func() {
        if d, err := strconv.Atoi(droneEntry.Text); err == nil && d >= 0 {
            numDrones = d
        }
        if c, err := strconv.Atoi(crowdEntry.Text); err == nil && c >= 0 {
            numCrowdMembers = c
        }
        if o, err := strconv.Atoi(obstacleEntry.Text); err == nil && o >= 0 {
            numObstacles = o
        }

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