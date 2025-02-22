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
		0,
		1,
		1,
	)

	windowWidth := 1000.0
	windowHeight := 750.0

	fieldWidth := windowWidth * 0.25 // 25% of window width
	fieldHeight := 30.0

	// Calculate positions relative to window width
	leftColumn := windowWidth * 0.25   // 25% from left
	rightColumn := windowWidth * 0.625 // 62.5% from left

	// Row positions from top
	firstRow := windowHeight * 0.3   // 30% down
	secondRow := windowHeight * 0.45 // 45% down
	buttonRow := windowHeight * 0.75 // 75% down

	// Text input fields
	g.DroneField = ui.TextField{
		X:      leftColumn,
		Y:      firstRow,
		Width:  fieldWidth,
		Height: fieldHeight,
		Text:   "1",
		OnEnter: func(value int) {
			g.DroneCount = value
		},
	}

	g.PeopleField = ui.TextField{
		X:      rightColumn,
		Y:      firstRow,
		Width:  fieldWidth,
		Height: fieldHeight,
		Text:   "10",
		OnEnter: func(value int) {
			g.Sim.UpdateCrowdSize(value)
			g.PeopleCount = value
		},
	}

	// Dropdowns
	g.DropdownMap = ui.Dropdown{
		X:             leftColumn,
		Y:             secondRow,
		Width:         fieldWidth,
		Height:        fieldHeight,
		Options:       []string{"Carte test 1", "Carte test 2", "Carte test 3"},
		SelectedIndex: 0,
		OnSelect: func(index int) {
			println("Index Map Selected:", index)
		},
	}

	g.DropdownProtocole = ui.Dropdown{
		X:             rightColumn,
		Y:             secondRow,
		Width:         fieldWidth,
		Height:        fieldHeight,
		Options:       []string{"Protocole test 1", "Protocole test 2", "Protocole test 3", "Protocole test 4"},
		SelectedIndex: 0,
		OnSelect: func(index int) {
			println("Selected Protocol:", index+1)
		},
	}

	buttonWidth := fieldWidth
	g.StartButton = ui.Button{
		X:      leftColumn,
		Y:      buttonRow,
		Width:  buttonWidth,
		Height: fieldHeight * 1.5,
		Text:   "Start Simulation",
		OnClick: func() {
			if val, err := strconv.Atoi(g.DroneField.Text); err == nil {
				g.DroneCount = val
			}
			if val, err := strconv.Atoi(g.PeopleField.Text); err == nil {
				g.PeopleCount = val
			}

			var chosenMap string
			switch g.DropdownMap.SelectedIndex {
			case 0:
				chosenMap = "festival_layout_1"
			case 1:
				chosenMap = "festival_layout_2"
			case 2:
				chosenMap = "festival_layout_3"
			default: //Error
				chosenMap = "festival_layout_new"
			}

			g.Sim.UpdateMap(chosenMap)
			g.Sim.UpdateCrowdSize(g.PeopleCount)
			g.Sim.UpdateDroneSize(g.DroneCount)
			g.Sim.UpdateDroneProtocole(g.DropdownProtocole.SelectedIndex + 1)
			g.Sim.InitDronesProtocols()

			g.Mode = game.Simulation
		},
	}

	g.StartButtonDebug = ui.Button{
		X:       rightColumn,
		Y:       buttonRow,
		Width:   buttonWidth * 1.25,
		Height:  fieldHeight * 1.5,
		Text:    "Start Simulation (Debug Mode)",
		Couleur: color.RGBA{255, 0, 0, 255},
		OnClick: func() {
			if val, err := strconv.Atoi(g.DroneField.Text); err == nil {
				g.DroneCount = val
			}
			if val, err := strconv.Atoi(g.PeopleField.Text); err == nil {
				g.PeopleCount = val
			}

			var chosenMap string
			switch g.DropdownMap.SelectedIndex {
			case 0:
				chosenMap = "festival_layout_1"
			case 1:
				chosenMap = "festival_layout_2"
			case 2:
				chosenMap = "festival_layout_3"
			default:
				chosenMap = "festival_layout_new"
			}
			g.Sim.UpdateMap(chosenMap)
			g.Sim.UpdateCrowdSize(g.PeopleCount)
			g.Sim.UpdateDroneSize(g.DroneCount)
			g.Sim.UpdateDroneProtocole(g.DropdownProtocole.SelectedIndex + 1)
			g.Sim.InitDronesProtocols()

			g.Mode = game.SimulationDebug
		},
	}

	// Initialize simulation buttons
	g.PauseButton = ui.Button{
		X:      windowWidth * 0.75, // 75% from left
		Y:      windowHeight * 0.3, // 30% from top
		Width:  windowWidth * 0.2,  // 20% of window width
		Height: fieldHeight,
		Text:   "Pause",
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
		X:      windowWidth * 0.75, // 75% from left
		Y:      windowHeight * 0.4, // 40% from top
		Width:  windowWidth * 0.2,  // 20% of window width
		Height: fieldHeight,
		Text:   "Update Sim",
		OnClick: func() {
			g.Sim.Update()
		},
	}

	ebiten.SetWindowSize(int(windowWidth), int(windowHeight))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	ebiten.SetWindowTitle("Simulation Drones")
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
