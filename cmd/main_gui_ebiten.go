package main

import (
	"math/rand"
	"time"

	game "UTC_IA04/cmd/simu"
	"UTC_IA04/cmd/ui"
	"UTC_IA04/pkg/simulation"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	g := game.NewGame(
		5,  // default drone count
		10, // default people count
		5,  // default obstacle count
	)

	// Initialize UI fields
	g.DroneField = ui.TextField{
		X: 150, Y: 150, Width: 200, Height: 30, Text: "5",
		OnEnter: func(value int) {
			g.DroneCount = value
		},
	}
	g.PeopleField = ui.TextField{
		X: 150, Y: 200, Width: 200, Height: 30, Text: "10",
		OnEnter: func(value int) {
			g.PeopleCount = value
		},
	}
	g.ObstacleField = ui.TextField{
		X: 150, Y: 250, Width: 200, Height: 30, Text: "5",
		OnEnter: func(value int) {
			g.ObstacleCount = value
		},
	}

	g.StartButton = ui.Button{
		X: 300, Y: 350, Width: 200, Height: 50, Text: "Start Simulation",
		OnClick: func() {
			g.Sim = simulation.NewSimulation(g.DroneCount, g.PeopleCount, g.ObstacleCount)
			g.Mode = game.Simulation
		},
	}

	g.PauseButton = ui.Button{
		X: 300, Y: 400, Width: 150, Height: 40, Text: "Pause",
		OnClick: func() {
			g.Paused = !g.Paused
			if g.Paused {
				g.PauseButton.Text = "Resume"
			} else {
				g.PauseButton.Text = "Pause"
			}
		},
	}

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Simulation with Vision Circles")
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
