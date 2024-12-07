package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"UTC_IA04/cmd/ui"
	"UTC_IA04/cmd/ui/assets"
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
)

type Game struct {
	Mode          Mode
	StartButton   ui.Button
	PauseButton   ui.Button
	DroneField    ui.TextField
	PeopleField   ui.TextField
	ObstacleField ui.TextField
	Sim           *simulation.Simulation
	StaticLayer   *ebiten.Image
	DynamicLayer  *ebiten.Image
	Paused        bool
	DroneCount    int
	PeopleCount   int
	ObstacleCount int
	DroneImage    *ebiten.Image
	PoiImages     map[models.POIType]*ebiten.Image
	hoveredPos    *models.Position
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
	return 800, 600
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
		g.DroneField.Update(float64(mx), float64(my), mousePressed, inputRunes, ebiten.IsKeyPressed(ebiten.KeyEnter))
		g.PeopleField.Update(float64(mx), float64(my), mousePressed, inputRunes, ebiten.IsKeyPressed(ebiten.KeyEnter))
		g.ObstacleField.Update(float64(mx), float64(my), mousePressed, inputRunes, ebiten.IsKeyPressed(ebiten.KeyEnter))

	case Simulation:
		g.PauseButton.Update(float64(mx), float64(my), mousePressed)
		g.updatePOIHover(float64(mx), float64(my))
		if g.Paused {
			return nil
		}
		g.Sim.Update()
	}

	return nil
}

func (g *Game) updatePOIHover(mx, my float64) {
	g.hoveredPos = nil

	// Convert mouse coordinates to game coordinates
	gameX := mx / 30
	gameY := my / 30

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
	// Convert mouse coordinates to game coordinates
	gameX := mx / 30
	gameY := my / 30

	// Check each person's position
	for _, person := range g.Sim.Persons {
		if math.Abs(gameX-person.Position.X) <= 0.3 && math.Abs(gameY-person.Position.Y) <= 0.3 {
			return &person
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
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 50, 255})
	title := "Festival Surveillance Simulation"
	ebitenutil.DebugPrintAt(screen, title, 250, 50)

	instructions := "Configure simulation parameters:\n" +
		"Click fields to edit, press Enter to confirm."
	ebitenutil.DebugPrintAt(screen, instructions, 200, 100)

	ebitenutil.DebugPrintAt(screen, "Number of Drones:", 200, 200)
	g.DroneField.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Number of People:", 200, 250)
	g.PeopleField.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Number of Obstacles:", 200, 300)
	g.ObstacleField.Draw(screen)

	g.StartButton.Draw(screen)
}

func (g *Game) drawSimulation(screen *ebiten.Image) {
	g.drawStaticLayer()
	screen.DrawImage(g.StaticLayer, nil)

	g.drawDynamicLayer()
	screen.DrawImage(g.DynamicLayer, nil)

	g.drawMetricsWindow(screen)
	g.PauseButton.Draw(screen)

	// If there's a hovered POI, draw information
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
	if hoveredPerson := g.getHoveredPerson(float64(mx), float64(my)); hoveredPerson != nil {
		personInfo := fmt.Sprintf(
			"Person Info\n"+
				"ID: %d\n"+
				"In Distress: %t\n"+
				"Has Reached POI: %t\n"+
				"Position: (%.1f, %.1f)",
			hoveredPerson.ID,
			hoveredPerson.InDistress,
			hoveredPerson.HasReachedPOI(),
			hoveredPerson.Position.X,
			hoveredPerson.Position.Y,
		)
		ebitenutil.DebugPrintAt(screen, personInfo, mx+10, my+10)
	}
}

func (g *Game) drawStaticLayer() {
	g.StaticLayer.Clear()
	width := g.StaticLayer.Bounds().Dx()
	height := g.StaticLayer.Bounds().Dy()

	// Draw zones with new proportions (20-60-20 split)
	entranceWidth := int(float64(width) * 0.2)
	mainWidth := int(float64(width) * 0.6)
	exitWidth := width - entranceWidth - mainWidth

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
				scale := 0.07
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(-float64(w)*scale/2, -float64(h)*scale/2)
				op.GeoM.Translate(pos.X*30, pos.Y*30)

				g.StaticLayer.DrawImage(g.PoiImages[poiType], op)
			}

		}
	}
}

func (g *Game) drawDynamicLayer() {
	g.DynamicLayer.Clear()

	// Draw people
	for _, person := range g.Sim.Map.Persons {
		if person.Position.X == 0 {
			println("Person at entrance")
			fmt.Printf("Simulation Details: %+v\n", person)

			//println(person)
		}
		couleur := color.RGBA{255, 0, 0, 255}
		if person.HasReachedPOI() {
			couleur = color.RGBA{0, 255, 0, 255} // Green for resting people
		}
		if person.InDistress {
			couleur = color.RGBA{0, 0, 0, 255} // Black for people in distress
		}
		drawCircle(g.DynamicLayer, person.Position.X*30, person.Position.Y*30, 10, couleur)
	}

	// Draw drones
	for _, drone := range g.Sim.Drones {
		drawTranslucentCircle(g.DynamicLayer, drone.Position.X*30, drone.Position.Y*30, 60, color.RGBA{0, 255, 0, 64})

		if g.DroneImage != nil {
			bounds := g.DroneImage.Bounds()
			w, h := bounds.Dx(), bounds.Dy()

			op := &ebiten.DrawImageOptions{}
			scale := 0.07
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(-float64(w)*scale/2, -float64(h)*scale/2)
			op.GeoM.Translate(drone.Position.X*30, drone.Position.Y*30)

			g.DynamicLayer.DrawImage(g.DroneImage, op)
		} else {
			drawCircle(g.DynamicLayer, drone.Position.X*30, drone.Position.Y*30, 10, color.RGBA{0, 0, 255, 255})
		}
	}
}

func (g *Game) drawMetricsWindow(screen *ebiten.Image) {
	stats := g.Sim.GetStatistics()
	metricsWidth, metricsHeight := 200, 150
	metrics := ebiten.NewImage(metricsWidth, metricsHeight)
	metrics.Fill(color.RGBA{30, 30, 30, 200})

	text := fmt.Sprintf(
		"Simulation Metrics\n"+
			"Drones: %d\n"+
			"People: %d\n"+
			"In Distress: %d\n"+
			"Phase: %s\n"+
			"Time Left: %.0fm",
		stats.ActiveDrones,
		stats.TotalPeople,
		stats.PeopleInDistress,
		stats.CurrentPhase,
		stats.RemainingTime.Minutes(),
	)
	ebitenutil.DebugPrintAt(metrics, text, 10, 10)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(580, 50)
	screen.DrawImage(metrics, opts)
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
