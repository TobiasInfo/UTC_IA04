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
		screen.Fill(color.RGBA{0, 0, 0, 255})
		ebitenutil.DebugPrint(screen, "Welcome to the Simulation!\nSet parameters and press Start.")

		g.StartButton.Draw(screen)

		ebitenutil.DebugPrintAt(screen, "Drones:", 50, 150)
		g.DroneField.Draw(screen)

		ebitenutil.DebugPrintAt(screen, "People:", 50, 200)
		g.PeopleField.Draw(screen)

		ebitenutil.DebugPrintAt(screen, "Obstacles:", 50, 250)
		g.ObstacleField.Draw(screen)

	case Simulation:
		g.drawStaticLayer()
		screen.DrawImage(g.StaticLayer, nil)

		g.drawDynamicLayer()
		screen.DrawImage(g.DynamicLayer, nil)

		g.drawMetricsWindow(screen)
		g.PauseButton.Draw(screen)
	}
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
	metrics := ebiten.NewImage(200, 200)
	metrics.Fill(color.RGBA{0, 0, 0, 200})
	text := fmt.Sprintf("Metrics:\nDrones: %d\nPeople: %d\nObstacles: %d",
		len(g.Sim.Drones), len(g.Sim.Persons), len(g.Sim.Obstacles))
	ebitenutil.DebugPrint(metrics, text)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(600, 50)
	screen.DrawImage(metrics, opts)
}
