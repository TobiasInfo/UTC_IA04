package main

import (
	game "UTC_IA04/cmd/simu"
	"UTC_IA04/cmd/ui"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	g := game.NewGame(
		0, // default drone count
		1, // default people count
		1, // default obstacle count
	)
	println("Debut simu dans GUI")
	//fmt.Printf("Simulation Details: %+v\n", g.Sim.GetAvailablePOIs())

	// Centering and making the UI look nicer
	fieldWidth := 200.0
	fieldHeight := 30.0
	fieldX := 200.0 - fieldWidth/2

	g.DroneField = ui.TextField{
		X: fieldX, Y: 200, Width: fieldWidth, Height: fieldHeight, Text: "1",
		OnEnter: func(value int) {
			g.DroneCount = value
		},
	}
	g.PeopleField = ui.TextField{
		X: fieldX + 250, Y: 200, Width: fieldWidth, Height: fieldHeight, Text: "10",
		OnEnter: func(value int) {
			g.Sim.UpdateCrowdSize(value)
			g.PeopleCount = value
		},
	}
	g.DropdownMap = ui.Dropdown{
		X: fieldX, Y: 270, Width: fieldWidth, Height: fieldHeight,
		Options:       []string{"Carte test 1", "Carte test 2", "Carte test 3"},
		SelectedIndex: 0,
		OnSelect: func(index int) {
			//La correspondance entre index et Carte doit être faite dans les ifelse de OnClick
			println("Index Map Selected:", index)
		},
	}

	g.DropdownProtocole = ui.Dropdown{
		X: fieldX + 250, Y: 270, Width: fieldWidth, Height: fieldHeight,
		Options:       []string{"Protocole test 1", "Protocole test 2", "Protocole test 3"},
		SelectedIndex: 0,
		OnSelect: func(index int) {
			println("Selected Protocol:", index+1)
		},
	}

	g.StartButton = ui.Button{
		X: fieldX, Y: 450, Width: 200, Height: 50, Text: "Start Simulation",
		OnClick: func() {
			// Parse current values from the text fields to ensure they are up-to-date
			if val, err := strconv.Atoi(g.DroneField.Text); err == nil {
				g.DroneCount = val
			}
			if val, err := strconv.Atoi(g.PeopleField.Text); err == nil {
				g.PeopleCount = val
			}

			var chosenMap string
			if g.DropdownMap.SelectedIndex == 0 {
				chosenMap = "festival_layout"
			} else {
				if g.DropdownMap.SelectedIndex == 1 {
					chosenMap = "test_layout"
				} else {
					if g.DropdownMap.SelectedIndex == 2 {
						chosenMap = "test_layout"
					} else {
						chosenMap = "test_layout"
					}
				}
			}
			g.Sim.UpdateMap(chosenMap)

			g.Sim.UpdateCrowdSize(g.PeopleCount)
			g.Sim.UpdateDroneSize(g.DroneCount)
			g.Sim.UpdateDroneProtocole(g.DropdownProtocole.SelectedIndex + 1)

			g.Mode = game.Simulation
		},
	}

	g.StartButtonDebug = ui.Button{
		X: fieldX + 250, Y: 450, Width: 250, Height: 50, Text: "Start Simulation (Debug Mode)", Couleur: color.RGBA{255, 0, 0, 255},
		OnClick: func() {
			// Parse current values from the text fields to ensure they are up-to-date
			if val, err := strconv.Atoi(g.DroneField.Text); err == nil {
				g.DroneCount = val
			}
			if val, err := strconv.Atoi(g.PeopleField.Text); err == nil {
				g.PeopleCount = val
			}

			var chosenMap string
			if g.DropdownMap.SelectedIndex == 0 {
				chosenMap = "festival_layout"
			} else {
				if g.DropdownMap.SelectedIndex == 1 {
					chosenMap = "test_layout"
				} else {
					if g.DropdownMap.SelectedIndex == 2 {
						chosenMap = "test_layout"
					} else {
						chosenMap = "test_layout"
					}
				}
			}
			g.Sim.UpdateMap(chosenMap)
			g.Sim.UpdateCrowdSize(g.PeopleCount)
			g.Sim.UpdateDroneSize(g.DroneCount)
			g.Sim.UpdateDroneProtocole(g.DropdownProtocole.SelectedIndex + 1)

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
