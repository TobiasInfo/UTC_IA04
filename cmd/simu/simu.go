package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"UTC_IA04/cmd/ui"
	"UTC_IA04/cmd/ui/assets"
	"UTC_IA04/pkg/entities/drones"
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/models"
	"UTC_IA04/pkg/simulation"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Mode int

const (
	Menu Mode = iota
	Simulation
	SimulationDebug
)

type Game struct {
	Mode              Mode
	StartButton       ui.Button
	StartButtonDebug  ui.Button
	PauseButton       ui.Button
	SimButton         ui.Button
	DroneField        ui.TextField
	PeopleField       ui.TextField
	DropdownMap       ui.Dropdown
	DropdownProtocole ui.Dropdown
	Sim               *simulation.Simulation
	StaticLayer       *ebiten.Image
	DynamicLayer      *ebiten.Image
	Paused            bool
	DroneCount        int
	PeopleCount       int
	ObstacleCount     int
	DroneImage        *ebiten.Image
	PoiImages         map[models.POIType]*ebiten.Image
	hoveredPos        *models.Position
}

// Zone colors
var (
	EntranceZoneColor = color.RGBA{135, 206, 235, 180} // Light blue
	MainZoneColor     = color.RGBA{144, 238, 144, 180} // Light green
	ExitZoneColor     = color.RGBA{255, 182, 193, 180} // Light pink

	// POI colors
	MedicalColor   = color.RGBA{255, 0, 0, 255}     // Red
	ChargingColor  = color.RGBA{255, 255, 0, 255}   // Yellow
	ToiletColor    = color.RGBA{128, 128, 128, 255} // Gray
	DrinkColor     = color.RGBA{0, 191, 255, 255}   // Deep sky blue
	FoodColor      = color.RGBA{255, 165, 0, 255}   // Orange
	MainStageColor = color.RGBA{148, 0, 211, 255}   // Purple
	SecondaryColor = color.RGBA{186, 85, 211, 255}  // Medium purple
	RestAreaColor  = color.RGBA{46, 139, 87, 255}   // Sea green
)

func NewGame(droneCount, peopleCount, obstacleCount int) *Game {
	g := &Game{
		Mode:          Menu,
		DroneCount:    droneCount,
		PeopleCount:   peopleCount,
		ObstacleCount: obstacleCount,
		StaticLayer:   ebiten.NewImage(800, 600),
		DynamicLayer:  ebiten.NewImage(800, 600),
		Sim:           simulation.NewSimulation(droneCount, peopleCount, obstacleCount),
	}

	// Load the drone image once
	g.DroneImage = loadImage("img/drone.png")
	g.PoiImages = map[models.POIType]*ebiten.Image{
		0: loadImage(assets.POIIcon(0)),
		1: loadImage(assets.POIIcon(1)),
		2: loadImage(assets.POIIcon(2)),
		3: loadImage(assets.POIIcon(3)),
		4: loadImage(assets.POIIcon(4)),
		5: loadImage(assets.POIIcon(5)),
		6: loadImage(assets.POIIcon(6)),
		7: loadImage(assets.POIIcon(7)),
	}

	return g
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Maintain minimum size while allowing larger sizes
	width := math.Max(800, float64(outsideWidth))
	height := math.Max(600, float64(outsideHeight))
	return int(width), int(height)
}

func (g *Game) Update() error {
	mx, my := ebiten.CursorPosition()
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// Create a slice to hold input runes
	var inputRunes []rune
	inputRunes = ebiten.AppendInputChars(inputRunes)

	switch g.Mode {
	case Menu:
		g.StartButton.Update(float64(mx), float64(my), mousePressed)
		g.StartButtonDebug.Update(float64(mx), float64(my), mousePressed)
		g.DroneField.Update(float64(mx), float64(my), mousePressed, inputRunes, ebiten.IsKeyPressed(ebiten.KeyEnter))
		g.PeopleField.Update(float64(mx), float64(my), mousePressed, inputRunes, ebiten.IsKeyPressed(ebiten.KeyEnter))
		g.DropdownMap.Update(float64(mx), float64(my), mousePressed)
		g.DropdownProtocole.Update(float64(mx), float64(my), mousePressed)
	case Simulation:
		g.SimButton.Update(float64(mx), float64(my), mousePressed)
		g.PauseButton.Update(float64(mx), float64(my), mousePressed)
		g.updatePOIHover(float64(mx), float64(my))
		if g.Paused {
			return nil
		}
		g.Sim.Update()
	case SimulationDebug:
		g.SimButton.Update(float64(mx), float64(my), mousePressed)
		g.PauseButton.Update(float64(mx), float64(my), mousePressed)
		g.updatePOIHover(float64(mx), float64(my))
		if g.Paused {
			return nil
		}
	}

	return nil
}

func (g *Game) updatePOIHover(mx, my float64) {
	g.hoveredPos = nil

	// Convert screen coordinates back to game coordinates
	width := float64(g.StaticLayer.Bounds().Dx())
	height := float64(g.StaticLayer.Bounds().Dy())

	gameX := mx * float64(g.Sim.Map.Width) / width
	gameY := my * float64(g.Sim.Map.Height) / height

	// Get current POI map
	poiMap := g.Sim.GetAvailablePOIs()

	// Check each POI position
	for _, positions := range poiMap {
		for _, pos := range positions {
			if math.Abs(gameX-pos.X) <= 1 && math.Abs(gameY-pos.Y) <= 1 {
				g.hoveredPos = &pos
				return
			}
		}
	}
}

func (g *Game) getHoveredPerson(mx, my float64) *persons.Person {
	// Convert screen coordinates back to game coordinates
	width := float64(g.StaticLayer.Bounds().Dx())
	height := float64(g.StaticLayer.Bounds().Dy())

	gameX := mx * float64(g.Sim.Map.Width) / width
	gameY := my * float64(g.Sim.Map.Height) / height

	for _, person := range g.Sim.Persons {
		if math.Abs(gameX-person.Position.X) <= 0.3 && math.Abs(gameY-person.Position.Y) <= 0.3 {
			return &person
		}
	}
	return nil
}

func (g *Game) getHoveredDrone(mx, my float64) *drones.Drone {
	// Convert screen coordinates back to game coordinates
	width := float64(g.StaticLayer.Bounds().Dx())
	height := float64(g.StaticLayer.Bounds().Dy())

	gameX := mx * float64(g.Sim.Map.Width) / width
	gameY := my * float64(g.Sim.Map.Height) / height

	for _, drone := range g.Sim.Drones {
		if math.Abs(gameX-drone.Position.X) <= 0.3 && math.Abs(gameY-drone.Position.Y) <= 0.3 {
			return &drone
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.Mode {
	case Menu:
		g.drawMenu(screen)
	case Simulation:
		g.drawSimulation(screen)
	case SimulationDebug:
		g.drawSimulation(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 50, 255})
	title := "Festival Surveillance Simulation"
	ebitenutil.DebugPrintAt(screen, title, 250, 50)

	instructions := "Configure simulation parameters:\n" +
		"Click fields to edit, press Enter to confirm."
	ebitenutil.DebugPrintAt(screen, instructions, 200, 100)

	ebitenutil.DebugPrintAt(screen, "Number of Drones:", 100, 183)
	g.DroneField.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Number of People:", 350, 183)
	g.PeopleField.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Map Selection:", 100, 253)
	g.DropdownMap.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Protocole Selection:", 350, 253)
	g.DropdownProtocole.Draw(screen)

	g.StartButton.Draw(screen)
	g.StartButtonDebug.Draw(screen)
}

func (g *Game) drawSimulation(screen *ebiten.Image) {
	// Get screen dimensions
	screenWidth := float64(screen.Bounds().Dx())
	screenHeight := float64(screen.Bounds().Dy())

	// Calculate scaling factors
	scaleX := screenWidth / float64(g.Sim.Map.Width)
	scaleY := screenHeight / float64(g.Sim.Map.Height)

	// Use the smaller scale to maintain aspect ratio
	scale := math.Min(scaleX, scaleY)

	// Calculate offsets to center the map
	offsetX := (screenWidth - float64(g.Sim.Map.Width)*scale) / 2
	offsetY := (screenHeight - float64(g.Sim.Map.Height)*scale) / 2

	g.drawStaticLayer()
	g.drawDynamicLayer()

	// Draw both layers with scaling
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale/30, scale/30)
	op.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(g.StaticLayer, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale/30, scale/30)
	op.GeoM.Translate(offsetX, offsetY)
	screen.DrawImage(g.DynamicLayer, op)

	g.drawMetricsWindow(screen)
	g.PauseButton.Draw(screen)
	g.SimButton.Draw(screen)

	// Handle POI hover information with scaled coordinates
	if g.hoveredPos != nil {
		mx, my := ebiten.CursorPosition()
		personsAtPOI := 0
		for _, person := range g.Sim.Persons {
			if person.HasReachedPOI() &&
				person.TargetPOIPosition != nil &&
				*person.TargetPOIPosition == *g.hoveredPos {
				personsAtPOI++
			}
		}
		info := fmt.Sprintf("Visitors: %d", personsAtPOI)
		ebitenutil.DebugPrintAt(screen, info, mx+10, my+10)
	}

	mx, my := ebiten.CursorPosition()
	simX := (float64(mx) - offsetX) / (scale / 30)
	simY := (float64(my) - offsetY) / (scale / 30)

	// Check for hovering over rescuers with scaled coordinates
	for _, drone := range g.Sim.Drones {
		if drone.Rescuer != nil {
			rescuerPos := drone.Rescuer.Position
			if math.Abs(simX-rescuerPos.X*30) <= 9 &&
				math.Abs(simY-rescuerPos.Y*30) <= 9 {
				rescuerInfo := fmt.Sprintf(
					"Rescuer Info:\n"+
						"From Drone: %d\n"+
						"Status: %s\n"+
						"Target Person: %d\n"+
						"Position: (%.1f, %.1f)",
					drone.ID,
					map[int]string{0: "Going to Person", 1: "Returning to Tent"}[drone.Rescuer.State],
					drone.Rescuer.Person.ID,
					rescuerPos.X,
					rescuerPos.Y,
				)
				ebitenutil.DebugPrintAt(screen, rescuerInfo, mx+10, my+10)
				return
			}
		}
	}

	// Check for hovering over people with scaled coordinates
	if hoveredPerson := g.getHoveredPerson(simX/30, simY/30); hoveredPerson != nil {
		//mapPos := g.Sim.GetPersonPauseInMap(hoveredPerson)
		personInfo := fmt.Sprintf(
			"Person Info\n"+
				"ID: %d\n"+
				"In Distress: %t\n"+
				"Has Reached POI: %t\n"+
				"Position: (%.1f, %.1f)\n"+
				"CurrentDistressDuration : %d",

			hoveredPerson.ID,
			hoveredPerson.InDistress,
			hoveredPerson.HasReachedPOI(),
			hoveredPerson.Position.X,
			hoveredPerson.Position.Y,
			hoveredPerson.CurrentDistressDuration,
		)
		ebitenutil.DebugPrintAt(screen, personInfo, mx+10, my+10)
	}

	// Check for hovering over drones with scaled coordinates
	if hoveredDrone := g.getHoveredDrone(simX/30, simY/30); hoveredDrone != nil {
		//dronePosInMap := g.Sim.GetDronePauseInMap(hoveredDrone)
		droneInfo := fmt.Sprintf(
			"Drone Info\n"+
				"ID: %d\n"+
				"Position: (%.1f, %.1f) \n"+
				"Battery: %.1f\n"+
				"Number of seen people: %d\n"+
				"Is charging: %t\n",
			// "Has medical Gear: %t\n"+
			// "Objectif: (%.1f, %.1f)",

			hoveredDrone.ID,
			hoveredDrone.Position.X,
			hoveredDrone.Position.Y,
			hoveredDrone.Battery,
			len(hoveredDrone.SeenPeople),
			hoveredDrone.IsCharging,
			// hoveredDrone.HasMedicalGear,
			// hoveredDrone.Objectif.X,
			// hoveredDrone.Objectif.Y,
		)
		ebitenutil.DebugPrintAt(screen, droneInfo, mx+10, my+10)
	}
}

func (g *Game) drawStaticLayer() {
	g.StaticLayer.Clear()
	width := g.StaticLayer.Bounds().Dx()
	height := g.StaticLayer.Bounds().Dy()

	// Draw zones with correct proportions (10-80-10 split)
	entranceWidth := int(float64(width) * 0.1)
	mainWidth := int(float64(width) * 0.8)
	exitWidth := width - entranceWidth - mainWidth

	// Draw the zones
	drawRectangle(g.StaticLayer, 0, 0, float64(entranceWidth), float64(height), EntranceZoneColor)
	drawRectangle(g.StaticLayer, float64(entranceWidth), 0, float64(mainWidth), float64(height), MainZoneColor)
	drawRectangle(g.StaticLayer, float64(entranceWidth+mainWidth), 0, float64(exitWidth), float64(height), ExitZoneColor)

	// Draw POIs
	poiMap := g.Sim.GetAvailablePOIs()
	for poiType, positions := range poiMap {
		for _, pos := range positions {
			if g.PoiImages[poiType] != nil {
				bounds := g.PoiImages[poiType].Bounds()
				w, h := bounds.Dx(), bounds.Dy()

				op := &ebiten.DrawImageOptions{}
				iconScale := 0.07
				op.GeoM.Scale(iconScale, iconScale)
				op.GeoM.Translate(-float64(w)*iconScale/2, -float64(h)*iconScale/2)

				// Scale position to fit screen
				screenX := pos.X * float64(width) / float64(g.Sim.Map.Width)
				screenY := pos.Y * float64(height) / float64(g.Sim.Map.Height)
				op.GeoM.Translate(screenX, screenY)

				g.StaticLayer.DrawImage(g.PoiImages[poiType], op)
			}
		}
	}
}

func (g *Game) drawDynamicLayer() {
	g.DynamicLayer.Clear()
	width := float64(g.DynamicLayer.Bounds().Dx())
	height := float64(g.DynamicLayer.Bounds().Dy())

	seenPeople := make(map[int]bool)

	// Draw rescuers
	for _, drone := range g.Sim.Drones {
		if drone.Rescuer != nil {
			rescuerPos := drone.Rescuer.Position
			screenX := rescuerPos.X * width / float64(g.Sim.Map.Width)
			screenY := rescuerPos.Y * height / float64(g.Sim.Map.Height)

			// Draw base circle for rescuer
			drawCircle(g.DynamicLayer, screenX, screenY, 6, color.RGBA{0, 255, 0, 255})

			// Draw red cross
			crossSize := 4.0
			drawRectangle(g.DynamicLayer,
				screenX-crossSize, screenY-1,
				crossSize*2, 2,
				color.RGBA{255, 0, 0, 255})
			drawRectangle(g.DynamicLayer,
				screenX-1, screenY-crossSize,
				2, crossSize*2,
				color.RGBA{255, 0, 0, 255})

			// Draw path line
			if drone.Rescuer.State == 0 {
				targetX := drone.Rescuer.Person.Position.X * width / float64(g.Sim.Map.Width)
				targetY := drone.Rescuer.Person.Position.Y * height / float64(g.Sim.Map.Height)
				ebitenutil.DrawLine(g.DynamicLayer,
					screenX, screenY,
					targetX, targetY,
					color.RGBA{0, 255, 0, 128})
			} else {
				tentX := drone.Rescuer.MedicalTent.X * width / float64(g.Sim.Map.Width)
				tentY := drone.Rescuer.MedicalTent.Y * height / float64(g.Sim.Map.Height)
				ebitenutil.DrawLine(g.DynamicLayer,
					screenX, screenY,
					tentX, tentY,
					color.RGBA{0, 255, 0, 128})
			}
		}

		// Draw drone and its vision range
		droneX := drone.Position.X * width / float64(g.Sim.Map.Width)
		droneY := drone.Position.Y * height / float64(g.Sim.Map.Height)
		seeRange := float64(g.Sim.DroneSeeRange) * width / float64(g.Sim.Map.Width)

		drawTranslucentCircle(g.DynamicLayer, droneX, droneY, seeRange, color.RGBA{0, 0, 0, 32})

		if g.DroneImage != nil {
			bounds := g.DroneImage.Bounds()
			w, h := bounds.Dx(), bounds.Dy()

			op := &ebiten.DrawImageOptions{}
			scale := 0.04
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(-float64(w)*scale/2, -float64(h)*scale/2)
			op.GeoM.Translate(droneX, droneY)

			g.DynamicLayer.DrawImage(g.DroneImage, op)

			for _, person := range drone.SeenPeople {
				personX := person.Position.X * width / float64(g.Sim.Map.Width)
				personY := person.Position.Y * height / float64(g.Sim.Map.Height)
				drawRectangle(g.DynamicLayer, personX, personY, 5, 5, color.RGBA{255, 255, 0, 255})
				seenPeople[person.ID] = true
			}
		} else {
			drawCircle(g.DynamicLayer, droneX, droneY, 10, color.RGBA{0, 0, 255, 255})
		}
	}

	// Draw people
	for _, person := range g.Sim.Persons {
		if person.IsDead() {
			continue
		}

		if _, ok := seenPeople[person.ID]; ok {
			continue
		}

		personX := person.Position.X * width / float64(g.Sim.Map.Width)
		personY := person.Position.Y * height / float64(g.Sim.Map.Height)

		couleur := color.RGBA{255, 0, 0, 255}
		if person.HasReachedPOI() {
			couleur = color.RGBA{0, 255, 0, 255}
		}
		if person.IsInDistress() {
			couleur = color.RGBA{0, 0, 0, 255}
		}
		drawCircle(g.DynamicLayer, personX, personY, 3, couleur)
	}
}

func (g *Game) drawMetricsWindow(screen *ebiten.Image) {
	// Get screen dimensions
	screenWidth := float64(screen.Bounds().Dx())
	screenHeight := float64(screen.Bounds().Dy())

	// Define metrics window dimensions (proportional to screen)
	metricsWidth := screenWidth * 0.2   // 20% of screen width
	metricsHeight := screenHeight * 0.3 // 30% of screen height

	// Position metrics window in top right corner with padding
	padding := 20.0
	metrics := ebiten.NewImage(int(metricsWidth), int(metricsHeight))
	metrics.Fill(color.RGBA{30, 30, 30, 200})

	stats := g.Sim.GetStatistics()
	text := fmt.Sprintf(
		"Simulation Metrics\n"+
			"Total People: %d\n"+
			"In Distress: %d\n"+
			"Cases Treated: %d\n"+
			"Avg Battery: %.1f%%\n"+
			"Area Coverage: %.1f%%\n"+
			"Avg Comms Range: %.1f",
		stats.TotalPeople,
		stats.InDistress,
		stats.CasesTreated,
		stats.AverageBattery,
		stats.AverageCoverage,
		stats.AverageCommsRange,
	)
	ebitenutil.DebugPrintAt(metrics, text, 10, 10)

	// Draw metrics window
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(screenWidth-metricsWidth-padding, padding)
	screen.DrawImage(metrics, opts)

	// Update button positions to be relative to screen dimensions
	// Position buttons below metrics window
	buttonWidth := metricsWidth
	buttonHeight := 40.0
	buttonSpacing := 10.0

	// Update PauseButton position and dimensions
	g.PauseButton.Width = buttonWidth
	g.PauseButton.Height = buttonHeight
	g.PauseButton.X = screenWidth - metricsWidth - padding
	g.PauseButton.Y = padding + metricsHeight + buttonSpacing

	// Update SimButton position and dimensions
	g.SimButton.Width = buttonWidth
	g.SimButton.Height = buttonHeight
	g.SimButton.X = screenWidth - metricsWidth - padding
	g.SimButton.Y = g.PauseButton.Y + buttonHeight + buttonSpacing

	// Draw buttons
	g.PauseButton.Draw(screen)
	g.SimButton.Draw(screen)
}

func loadImage(path string) *ebiten.Image {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error loading image:", err)
		return nil
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return nil
	}
	return ebiten.NewImageFromImage(img)
}
