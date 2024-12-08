package main

import (
	game "UTC_IA04/cmd/simu"
	"UTC_IA04/cmd/ui"
	"github.com/hajimehoshi/ebiten/v2"
	"strconv"
)

func main() {
	g := game.NewGame(
		5,  // default drone count
		10, // default people count
		5,  // default obstacle count
	)
	println("Debut simu dans GUI")
	//fmt.Printf("Simulation Details: %+v\n", g.Sim.GetAvailablePOIs())

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
			// BIG BIG ERREUR DE TOBIAS ICI, POUR LE SHAMER NOUS ALLONS LAISSER LE CODE D'ORIGINE
			// SHAME ??, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔
			//g.Sim = simulation.NewSimulation(g.DroneCount, g.PeopleCount, g.ObstacleCount)

			g.Sim.UpdateCrowdSize(g.PeopleCount)
			g.Sim.UpdateDroneSize(g.DroneCount)

			g.Mode = game.Simulation
		},
	}

	g.StartButtonDebug = ui.Button{
		X: 400 - 100, Y: 450, Width: 300, Height: 50, Text: "Start Simulation (Debug Mode)",
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
			// BIG BIG ERREUR DE TOBIAS ICI, POUR LE SHAMER NOUS ALLONS LAISSER LE CODE D'ORIGINE
			// SHAME ??, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔, SHAME🔔
			//g.Sim = simulation.NewSimulation(g.DroneCount, g.PeopleCount, g.ObstacleCount)

			g.Sim.UpdateCrowdSize(g.PeopleCount)
			g.Sim.UpdateDroneSize(g.DroneCount)

			g.Mode = game.SimulationDebug
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

	g.SimButton = ui.Button{
		X: 600, Y: 250, Width: 150, Height: 40, Text: "Update Sim",
		OnClick: func() {
			g.Sim.Update()
		},
	}

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Simulation Drones")
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
