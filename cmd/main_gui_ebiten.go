package main

import (
	"math/rand"
	"strconv"
	"time"

	game "UTC_IA04/cmd/simu"
	"UTC_IA04/cmd/ui"
	"UTC_IA04/pkg/simulation"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	g := game.NewGame(
		5,   // default drone count
		100, // default people count
		5,   // default obstacle count
	)

	// Centering and making the UI look nicer
	fieldWidth := 200.0
	fieldHeight := 30.0
	fieldX := 400.0 - fieldWidth/2

	g.DroneField = ui.TextField{
		X: fieldX, Y: 200, Width: fieldWidth, Height: fieldHeight, Text: "5",
		OnEnter: func(value int) {
			g.DroneCount = value
		},
	}
	g.PeopleField = ui.TextField{
		X: fieldX, Y: 250, Width: fieldWidth, Height: fieldHeight, Text: "10",
		OnEnter: func(value int) {
			g.Sim.UpdateCrowdSize(value)
			g.PeopleCount = value
		},
	}
	g.ObstacleField = ui.TextField{
		X: fieldX, Y: 300, Width: fieldWidth, Height: fieldHeight, Text: "5",
		OnEnter: func(value int) {
			g.ObstacleCount = value
		},
	}

	g.StartButton = ui.Button{
		X: 400 - 100, Y: 380, Width: 200, Height: 50, Text: "Start Simulation",
		OnClick: func() {
			// Parse current values from the text fields to ensure they are up-to-date
			if val, err := strconv.Atoi(g.DroneField.Text); err == nil {
				g.DroneCount = val
			}
			if val, err := strconv.Atoi(g.PeopleField.Text); err == nil {
				g.PeopleCount = val
			}
			if val, err := strconv.Atoi(g.ObstacleField.Text); err == nil {
				g.ObstacleCount = val
			}

			// Now start the simulation with the updated values
			g.Sim = simulation.NewSimulation(g.DroneCount, g.PeopleCount, g.ObstacleCount)
			g.Mode = game.Simulation
		},
	}

	g.PauseButton = ui.Button{
		X: 600, Y: 180, Width: 150, Height: 40, Text: "Pause",
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
