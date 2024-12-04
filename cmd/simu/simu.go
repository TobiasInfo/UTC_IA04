package game

import (
	"fmt"
	"image/color"

	"UTC_IA04/cmd/ui"
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
}

func NewGame(droneCount, peopleCount, obstacleCount int) *Game {
	return &Game{
		Mode:          Menu,
		DroneCount:    droneCount,
		PeopleCount:   peopleCount,
		ObstacleCount: obstacleCount,
		StaticLayer:   ebiten.NewImage(800, 600),
		DynamicLayer:  ebiten.NewImage(800, 600),
		Sim:           simulation.NewSimulation(droneCount, peopleCount, obstacleCount),
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func (g *Game) Update() error {
	mx, my := ebiten.CursorPosition()
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	switch g.Mode {
	case Menu:
		g.StartButton.Update(float64(mx), float64(my), mousePressed)
		g.DroneField.Update(float64(mx), float64(my), mousePressed, ebiten.InputChars(), ebiten.IsKeyPressed(ebiten.KeyEnter))
		g.PeopleField.Update(float64(mx), float64(my), mousePressed, ebiten.InputChars(), ebiten.IsKeyPressed(ebiten.KeyEnter))
		g.ObstacleField.Update(float64(mx), float64(my), mousePressed, ebiten.InputChars(), ebiten.IsKeyPressed(ebiten.KeyEnter))

	case Simulation:
		g.PauseButton.Update(float64(mx), float64(my), mousePressed)
		if g.Paused {
			return nil
		}
		g.Sim.Update()
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
	// Fill background with a calming color
	screen.Fill(color.RGBA{30, 30, 50, 255})

	// Add a title
	title := "Welcome to the Simulation!"
	ebitenutil.DebugPrintAt(screen, title, 250, 50)

	// Instructions
	instructions := "Use the fields below to set parameters.\n" +
		"Click on a field, type the number, and press Enter.\n" +
		"Then click 'Start Simulation' to begin."
	ebitenutil.DebugPrintAt(screen, instructions, 200, 100)

	// Labels and fields
	ebitenutil.DebugPrintAt(screen, "Number of Drones:", 200, 200)
	g.DroneField.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Number of People:", 200, 250)
	g.PeopleField.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Number of Obstacles:", 200, 300)
	g.ObstacleField.Draw(screen)

	g.StartButton.Draw(screen)
}

func (g *Game) drawSimulation(screen *ebiten.Image) {
	// Draw the environment
	g.drawStaticLayer()
	screen.DrawImage(g.StaticLayer, nil)

	// Draw moving entities
	g.drawDynamicLayer()
	screen.DrawImage(g.DynamicLayer, nil)

	// Draw the metrics window
	g.drawMetricsWindow(screen)

	// Draw the pause/resume button
	g.PauseButton.Draw(screen)
}

func (g *Game) drawStaticLayer() {
	g.StaticLayer.Fill(color.RGBA{34, 139, 34, 255}) // Green background
	for _, obstacle := range g.Sim.Obstacles {
		drawRectangle(g.StaticLayer, obstacle.Position.X*30, obstacle.Position.Y*30, 30, 30, color.RGBA{0, 0, 0, 255})
	}
}

func (g *Game) drawDynamicLayer() {
	g.DynamicLayer.Clear()

	// Draw people
	for _, person := range g.Sim.Persons {
		drawCircle(g.DynamicLayer, person.Position.X*30, person.Position.Y*30, 10, color.RGBA{255, 0, 0, 255})
	}

	// Draw drones and their vision
	for _, drone := range g.Sim.Drones {
		drawTranslucentCircle(g.DynamicLayer, drone.Position.X*30, drone.Position.Y*30, 60, color.RGBA{0, 255, 0, 64})
		drawCircle(g.DynamicLayer, drone.Position.X*30, drone.Position.Y*30, 10, color.RGBA{0, 0, 255, 255})
	}
}

func (g *Game) drawMetricsWindow(screen *ebiten.Image) {
	metricsWidth, metricsHeight := 200, 120
	metrics := ebiten.NewImage(metricsWidth, metricsHeight)
	metrics.Fill(color.RGBA{30, 30, 30, 200}) // Semi-transparent dark background for metrics

	// Add a nice title and spacing
	title := "Simulation Metrics"
	ebitenutil.DebugPrintAt(metrics, title, 10, 10)
	// Use formatted strings to make metrics more readable
	text := fmt.Sprintf(
		"Drones: %d\nPeople: %d\nObstacles: %d",
		len(g.Sim.Drones), len(g.Sim.Persons), len(g.Sim.Obstacles),
	)
	ebitenutil.DebugPrintAt(metrics, text, 10, 30)

	opts := &ebiten.DrawImageOptions{}
	// Place it in the top-right corner, below the control panel
	opts.GeoM.Translate(580, 50)
	screen.DrawImage(metrics, opts)
}
