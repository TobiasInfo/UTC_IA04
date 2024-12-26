package game

import (
	"UTC_IA04/cmd/ui"
	"UTC_IA04/cmd/ui/assets"
	"UTC_IA04/pkg/entities/drones"
	"UTC_IA04/pkg/entities/persons"
	"UTC_IA04/pkg/entities/rescue"
	"UTC_IA04/pkg/models"
	"UTC_IA04/pkg/simulation"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Zone colors
var (
	EntranceZoneColor = color.RGBA{135, 206, 235, 1} // Light blue
	MainZoneColor     = color.RGBA{0, 0, 0, 0}       // Light green
	ExitZoneColor     = color.RGBA{255, 182, 193, 1} // Light pink

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

// WorldTransform handles coordinate conversion between simulation and screen space
type WorldTransform struct {
	screenWidth, screenHeight float64
	worldWidth, worldHeight   float64
	scale                     float64
	offsetX, offsetY          float64
	debug                     bool
}

func NewWorldTransform(screenW, screenH, worldW, worldH float64) *WorldTransform {
	t := &WorldTransform{
		screenWidth:  screenW,
		screenHeight: screenH,
		worldWidth:   worldW,
		worldHeight:  worldH,
		debug:        false,
	}
	t.calculateTransform()
	return t
}

func (t *WorldTransform) calculateTransform() {
	scaleX := t.screenWidth / t.worldWidth
	scaleY := t.screenHeight / t.worldHeight
	t.scale = math.Min(scaleX, scaleY)

	t.offsetX = (t.screenWidth - (t.worldWidth * t.scale)) / 2
	t.offsetY = (t.screenHeight - (t.worldHeight * t.scale)) / 2

	if t.debug {
		fmt.Printf("Transform calculated: scale=%f, offset=(%f,%f)\n",
			t.scale, t.offsetX, t.offsetY)
	}
}

func (t *WorldTransform) WorldToScreen(wx, wy float64) (float64, float64) {
	sx := (wx * t.scale) + t.offsetX
	sy := (wy * t.scale) + t.offsetY

	if t.debug {
		fmt.Printf("World(%f,%f) -> Screen(%f,%f)\n", wx, wy, sx, sy)
	}
	return sx, sy
}

func (t *WorldTransform) ScreenToWorld(sx, sy float64) (float64, float64) {
	wx := (sx - t.offsetX) / t.scale
	wy := (sy - t.offsetY) / t.scale

	if t.debug {
		fmt.Printf("Screen(%f,%f) -> World(%f,%f)\n", sx, sy, wx, wy)
	}
	return wx, wy
}

type Mode int

const (
	Menu Mode = iota
	Simulation
	SimulationDebug
)

type Game struct {
	Mode                  Mode
	StartButton           ui.Button
	StartButtonDebug      ui.Button
	PauseButton           ui.Button
	SimButton             ui.Button
	DroneField            ui.TextField
	PeopleField           ui.TextField
	DropdownMap           ui.Dropdown
	DropdownProtocole     ui.Dropdown
	Sim                   *simulation.Simulation
	StaticLayer           *ebiten.Image
	DynamicLayer          *ebiten.Image
	Paused                bool
	DroneCount            int
	PeopleCount           int
	ObstacleCount         int
	DroneImage            *ebiten.Image
	PoiImages             map[models.POIType]*ebiten.Image
	hoveredPos            *models.Position
	transform             *WorldTransform
	GrassImage            *ebiten.Image
	TiledFloorImage       *ebiten.Image
	AttendeeImage         *ebiten.Image
	AttendeeDeadImage     *ebiten.Image
	RescuerLookLeftImage  *ebiten.Image
	RescuerLookRightImage *ebiten.Image
}

func NewGame(droneCount, peopleCount, obstacleCount int) *Game {
	g := &Game{
		Mode:          Menu,
		DroneCount:    droneCount,
		PeopleCount:   peopleCount,
		ObstacleCount: obstacleCount,
		StaticLayer:   ebiten.NewImage(1000, 700),
		DynamicLayer:  ebiten.NewImage(1000, 700),
		Sim:           simulation.NewSimulation(droneCount, peopleCount, obstacleCount),
		transform:     NewWorldTransform(1000, 700, 30, 20),
	}

	g.DroneImage = loadImage("img/drone-real.png")
	g.GrassImage = loadImage("img/grass.png")
	g.TiledFloorImage = loadImage("img/tiledfloor-preview.png")
	g.AttendeeImage = loadImage("img/attendee-default.png")
	g.AttendeeDeadImage = loadImage("img/attendee-dead.png")
	g.RescuerLookLeftImage = loadImage("img/pompier-real-look-left.png")
	g.RescuerLookRightImage = loadImage("img/pompier-real-look-right.png")
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
	width := math.Max(800, float64(outsideWidth))
	height := math.Max(600, float64(outsideHeight))

	g.transform = NewWorldTransform(width, height, float64(g.Sim.Map.Width), float64(g.Sim.Map.Height))

	return int(width), int(height)
}

func (g *Game) Update() error {
	mx, my := ebiten.CursorPosition()
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

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

		worldX, worldY := g.transform.ScreenToWorld(float64(mx), float64(my))
		g.updatePOIHover(worldX, worldY)

		if g.Paused {
			return nil
		}
		g.Sim.Update()
	case SimulationDebug:
		g.SimButton.Update(float64(mx), float64(my), mousePressed)
		g.PauseButton.Update(float64(mx), float64(my), mousePressed)

		worldX, worldY := g.transform.ScreenToWorld(float64(mx), float64(my))
		g.updatePOIHover(worldX, worldY)

		if g.Paused {
			return nil
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

func (g *Game) updatePOIHover(worldX, worldY float64) {
	g.hoveredPos = nil

	poiMap := g.Sim.GetAvailablePOIs()
	for _, positions := range poiMap {
		for _, pos := range positions {
			if math.Abs(worldX-pos.X) <= 1 && math.Abs(worldY-pos.Y) <= 1 {
				g.hoveredPos = &pos
				if g.transform.debug {
					fmt.Printf("Hovering POI at world pos: (%f,%f)\n", pos.X, pos.Y)
				}
				return
			}
		}
	}
}

func (g *Game) getHoveredPerson(worldX, worldY float64) *persons.Person {
	for _, person := range g.Sim.Persons {
		if math.Abs(worldX-person.Position.X) <= 0.3 && math.Abs(worldY-person.Position.Y) <= 0.3 {
			return &person
		}
	}
	return nil
}

func (g *Game) getHoveredDrone(worldX, worldY float64) *drones.Drone {
	for _, drone := range g.Sim.Drones {
		if math.Abs(worldX-drone.Position.X) <= 0.3 && math.Abs(worldY-drone.Position.Y) <= 0.3 {
			return &drone
		}
	}
	return nil
}
func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 50, 255})
	title := "Festival Surveillance Simulation"
	ebitenutil.DebugPrintAt(screen, title, 400, 50)

	instructions := "Configure simulation parameters:\n" +
		"Click fields to edit, press Enter to confirm."
	ebitenutil.DebugPrintAt(screen, instructions, 400, 100)

	ebitenutil.DebugPrintAt(screen, "Number of Drones:", 255, 183)
	g.DroneField.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Number of People:", 625, 183)
	g.PeopleField.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Map Selection:", 255, 290)
	g.DropdownMap.Draw(screen)

	ebitenutil.DebugPrintAt(screen, "Protocole Selection:", 625, 290)
	g.DropdownProtocole.Draw(screen)

	g.StartButton.Draw(screen)
	g.StartButtonDebug.Draw(screen)
}

func (g *Game) drawStaticLayer() {
	g.StaticLayer.Clear()

	staticW, staticH := g.StaticLayer.Size() // normalement 800,600
	grassW := g.GrassImage.Bounds().Dx()
	grassH := g.GrassImage.Bounds().Dy()

	for y := 0; y < staticH; y += grassH {
		for x := 0; x < staticW; x += grassW {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			g.StaticLayer.DrawImage(g.GrassImage, op)
		}
	}

	// tileW := g.TiledFloorImage.Bounds().Dx()
	// tileH := g.TiledFloorImage.Bounds().Dy()

	// Draw zones using world coordinates
	worldWidth := float64(g.Sim.Map.Width)
	worldHeight := float64(g.Sim.Map.Height)
	entranceX1, y1 := g.transform.WorldToScreen(0, 0)
	entranceX2, y2 := g.transform.WorldToScreen(worldWidth*0.1, 0)
	mainX2, _ := g.transform.WorldToScreen(worldWidth*0.9, 0)
	exitX2, y2 := g.transform.WorldToScreen(worldWidth, worldHeight)

	if g.transform.debug {
		fmt.Printf("Drawing zones - Entrance: (%f,%f)->(%f,%f), Main: ->(%f), Exit: ->(%f)\n",
			entranceX1, y1, entranceX2, y2, mainX2, exitX2)
	}
	//On dessine l'entrée
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(entranceX1, y1)
	g.StaticLayer.DrawImage(g.TiledFloorImage, op)
	op.GeoM.Translate(entranceX1, y1+200.0)
	g.StaticLayer.DrawImage(g.TiledFloorImage, op)

	// drawRectangle(g.StaticLayer, entranceX1, y1, entranceX2-entranceX1, y2-y1, EntranceZoneColor)
	// drawRectangle(g.StaticLayer, entranceX2, y1, mainX2-entranceX2, y2-y1, MainZoneColor)
	// drawRectangle(g.StaticLayer, mainX2, y1, exitX2-mainX2, y2-y1, ExitZoneColor)

	// Draw POIs using world coordinates
	poiMap := g.Sim.GetAvailablePOIs()
	for poiType, positions := range poiMap {
		for _, pos := range positions {
			if g.PoiImages[poiType] != nil {
				bounds := g.PoiImages[poiType].Bounds()
				w, h := float64(bounds.Dx()), float64(bounds.Dy())

				screenX, screenY := g.transform.WorldToScreen(pos.X, pos.Y)

				op := &ebiten.DrawImageOptions{}
				iconScale := assets.PoiScale(poiType)
				op.GeoM.Scale(iconScale, iconScale)
				op.GeoM.Translate(-w*iconScale/2, -h*iconScale/2)
				op.GeoM.Translate(screenX, screenY)

				g.StaticLayer.DrawImage(g.PoiImages[poiType], op)

				if g.transform.debug {
					fmt.Printf("Drawing POI type %d at world(%f,%f) -> screen(%f,%f)\n",
						poiType, pos.X, pos.Y, screenX, screenY)
				}
			}
		}
	}
}
func (g *Game) drawDynamicLayer() {
	g.DynamicLayer.Clear()

	seenPeople := make(map[int]bool)

	// Draw rescuers
	for _, drone := range g.Sim.Drones {
		// Draw drone and its vision range
		droneScreenX, droneScreenY := g.transform.WorldToScreen(drone.Position.X, drone.Position.Y)
		seeRangeScreen := g.transform.scale * float64(g.Sim.DroneSeeRange)

		if !drone.IsCharging {
			drawTranslucentCircle(g.DynamicLayer, droneScreenX, droneScreenY, seeRangeScreen, color.RGBA{0, 0, 0, 32})
		}

		if g.DroneImage != nil {
			bounds := g.DroneImage.Bounds()
			w, h := float64(bounds.Dx()), float64(bounds.Dy())

			op := &ebiten.DrawImageOptions{}
			scale := 0.08
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(-w*scale/2, -h*scale/2)
			op.GeoM.Translate(droneScreenX, droneScreenY)

			g.DynamicLayer.DrawImage(g.DroneImage, op)

			// for _, person := range drone.SeenPeople {
			// 	personScreenX, personScreenY := g.transform.WorldToScreen(person.Position.X, person.Position.Y)
			// 	if person.IsInDistress() {
			// 		drawRectangle(g.DynamicLayer, personScreenX-2.5, personScreenY-2.5, 5, 5, color.RGBA{148, 0, 211, 255})
			// 	} else {
			// 		drawRectangle(g.DynamicLayer, personScreenX-2.5, personScreenY-2.5, 5, 5, color.RGBA{255, 255, 0, 255})
			// 	}
			// 	seenPeople[person.ID] = true
			// }
		} else {
			drawCircle(g.DynamicLayer, droneScreenX, droneScreenY, 10, color.RGBA{0, 0, 255, 255})
		}
	}

	for _, rp := range g.Sim.RescuePoints {
		for _, rescuer := range rp.Rescuers {
			if !rescuer.Active {
				continue
			}

			screenX, screenY := g.transform.WorldToScreen(rescuer.Position.X, rescuer.Position.Y)

			if rescuer.State == rescue.MovingToPerson && rescuer.Person != nil {
				if rescuer.Person.Position.X > rescuer.Position.X {
					if g.RescuerLookRightImage != nil {
						bounds := g.RescuerLookRightImage.Bounds()
						w, h := float64(bounds.Dx()), float64(bounds.Dy())

						op := &ebiten.DrawImageOptions{}
						scale := 0.1
						op.GeoM.Scale(scale, scale)
						op.GeoM.Translate(-w*scale/2, -h*scale/2)
						op.GeoM.Translate(screenX, screenY)

						g.DynamicLayer.DrawImage(g.RescuerLookRightImage, op)
					} else {
						drawCircle(g.DynamicLayer, screenX, screenY, 6, color.RGBA{0, 255, 0, 255})

						crossSize := 4.0
						drawRectangle(g.DynamicLayer,
							screenX-crossSize, screenY-1,
							crossSize*2, 2,
							color.RGBA{255, 0, 0, 255})
						drawRectangle(g.DynamicLayer,
							screenX-1, screenY-crossSize,
							2, crossSize*2,
							color.RGBA{255, 0, 0, 255})
					}

				} else {
					if g.RescuerLookLeftImage != nil {
						bounds := g.RescuerLookLeftImage.Bounds()
						w, h := float64(bounds.Dx()), float64(bounds.Dy())

						op := &ebiten.DrawImageOptions{}
						scale := 0.1
						op.GeoM.Scale(scale, scale)
						op.GeoM.Translate(-w*scale/2, -h*scale/2)
						op.GeoM.Translate(screenX, screenY)

						g.DynamicLayer.DrawImage(g.RescuerLookLeftImage, op)
					} else {
						drawCircle(g.DynamicLayer, screenX, screenY, 6, color.RGBA{0, 255, 0, 255})

						crossSize := 4.0
						drawRectangle(g.DynamicLayer,
							screenX-crossSize, screenY-1,
							crossSize*2, 2,
							color.RGBA{255, 0, 0, 255})
						drawRectangle(g.DynamicLayer,
							screenX-1, screenY-crossSize,
							2, crossSize*2,
							color.RGBA{255, 0, 0, 255})
					}

				}

			} else {
				if rescuer.State == rescue.ReturningToBase && (rescuer.HomePoint.X != 0 || rescuer.HomePoint.Y != 0) {
					if rescuer.HomePoint.X > rescuer.Position.X {
						if g.RescuerLookRightImage != nil {
							bounds := g.RescuerLookRightImage.Bounds()
							w, h := float64(bounds.Dx()), float64(bounds.Dy())

							op := &ebiten.DrawImageOptions{}
							scale := 0.1
							op.GeoM.Scale(scale, scale)
							op.GeoM.Translate(-w*scale/2, -h*scale/2)
							op.GeoM.Translate(screenX, screenY)

							g.DynamicLayer.DrawImage(g.RescuerLookRightImage, op)
						} else {
							drawCircle(g.DynamicLayer, screenX, screenY, 6, color.RGBA{0, 255, 0, 255})

							crossSize := 4.0
							drawRectangle(g.DynamicLayer,
								screenX-crossSize, screenY-1,
								crossSize*2, 2,
								color.RGBA{255, 0, 0, 255})
							drawRectangle(g.DynamicLayer,
								screenX-1, screenY-crossSize,
								2, crossSize*2,
								color.RGBA{255, 0, 0, 255})
						}

					} else {
						if g.RescuerLookLeftImage != nil {
							bounds := g.RescuerLookLeftImage.Bounds()
							w, h := float64(bounds.Dx()), float64(bounds.Dy())

							op := &ebiten.DrawImageOptions{}
							scale := 0.1
							op.GeoM.Scale(scale, scale)
							op.GeoM.Translate(-w*scale/2, -h*scale/2)
							op.GeoM.Translate(screenX, screenY)

							g.DynamicLayer.DrawImage(g.RescuerLookLeftImage, op)
						} else {
							drawCircle(g.DynamicLayer, screenX, screenY, 6, color.RGBA{0, 255, 0, 255})

							crossSize := 4.0
							drawRectangle(g.DynamicLayer,
								screenX-crossSize, screenY-1,
								crossSize*2, 2,
								color.RGBA{255, 0, 0, 255})
							drawRectangle(g.DynamicLayer,
								screenX-1, screenY-crossSize,
								2, crossSize*2,
								color.RGBA{255, 0, 0, 255})
						}

					}
				}
			}

			// if g.RescuerImage != nil {
			// 	bounds := g.RescuerImage.Bounds()
			// 	w, h := float64(bounds.Dx()), float64(bounds.Dy())

			// 	op := &ebiten.DrawImageOptions{}
			// 	scale := 0.1
			// 	op.GeoM.Scale(scale, scale)
			// 	op.GeoM.Translate(-w*scale/2, -h*scale/2)
			// 	op.GeoM.Translate(screenX, screenY)

			// 	g.DynamicLayer.DrawImage(g.RescuerImage, op)

			// } else {
			// 	drawCircle(g.DynamicLayer, screenX, screenY, 6, color.RGBA{0, 255, 0, 255})

			// 	crossSize := 4.0
			// 	drawRectangle(g.DynamicLayer,
			// 		screenX-crossSize, screenY-1,
			// 		crossSize*2, 2,
			// 		color.RGBA{255, 0, 0, 255})
			// 	drawRectangle(g.DynamicLayer,
			// 		screenX-1, screenY-crossSize,
			// 		2, crossSize*2,
			// 		color.RGBA{255, 0, 0, 255})
			// }

			if rescuer.State == rescue.MovingToPerson && rescuer.Person != nil {
				targetScreenX, targetScreenY := g.transform.WorldToScreen(
					rescuer.Person.Position.X,
					rescuer.Person.Position.Y)
				ebitenutil.DrawLine(g.DynamicLayer,
					screenX, screenY,
					targetScreenX, targetScreenY,
					color.RGBA{0, 255, 0, 128})
			}

			if rescuer.State == rescue.ReturningToBase && rescuer.HomePoint.X != 0 || rescuer.HomePoint.Y != 0 {
				tentScreenX, tentScreenY := g.transform.WorldToScreen(
					rescuer.HomePoint.X,
					rescuer.HomePoint.Y)
				ebitenutil.DrawLine(g.DynamicLayer,
					screenX, screenY,
					tentScreenX, tentScreenY,
					color.RGBA{0, 255, 0, 128})
			}
		}
	}

	// Draw people
	for _, person := range g.Sim.Persons {
		if person.IsDead() || person.Position.X < 0 {
			continue
		}

		if _, ok := seenPeople[person.ID]; ok {
			continue
		}

		screenX, screenY := g.transform.WorldToScreen(person.Position.X, person.Position.Y)

		// personColor := color.RGBA{255, 0, 0, 255} // Default red
		// if person.HasReachedPOI() {
		// 	personColor = color.RGBA{0, 255, 0, 255} // Green for at POI
		// }
		// if person.IsInDistress() {
		// 	personColor = color.RGBA{0, 0, 0, 255} // Black for distress
		// }

		// drawCircle(g.DynamicLayer, screenX, screenY, 3, personColor)

		if !person.IsInDistress() {
			if g.AttendeeImage != nil {
				bounds := g.AttendeeImage.Bounds()
				w, h := float64(bounds.Dx()), float64(bounds.Dy())

				op := &ebiten.DrawImageOptions{}
				scale := 0.05
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(-w*scale/2, -h*scale/2)
				op.GeoM.Translate(screenX, screenY)

				g.DynamicLayer.DrawImage(g.AttendeeImage, op)

			} else {
				drawCircle(g.DynamicLayer, screenX, screenY, 10, color.RGBA{0, 0, 255, 255})
			}

		} else {
			if g.AttendeeDeadImage != nil {
				bounds := g.AttendeeDeadImage.Bounds()
				w, h := float64(bounds.Dx()), float64(bounds.Dy())

				op := &ebiten.DrawImageOptions{}
				scale := 0.05
				op.GeoM.Scale(scale, scale)
				op.GeoM.Translate(-w*scale/2, -h*scale/2)
				op.GeoM.Translate(screenX, screenY)

				g.DynamicLayer.DrawImage(g.AttendeeDeadImage, op)

			} else {
				drawCircle(g.DynamicLayer, screenX, screenY, 10, color.RGBA{0, 0, 0, 255})
			}

		}

	}

	if g.transform.debug {
		g.drawDebugGrid()
	}
}

func (g *Game) drawDebugGrid() {
	gridColor := color.RGBA{100, 100, 100, 100}

	// Vertical lines
	for x := 0.0; x < float64(g.Sim.Map.Width); x++ {
		screenX1, screenY1 := g.transform.WorldToScreen(x, 0)
		screenX2, screenY2 := g.transform.WorldToScreen(x, float64(g.Sim.Map.Height))
		ebitenutil.DrawLine(g.DynamicLayer, screenX1, screenY1, screenX2, screenY2, gridColor)
	}

	// Horizontal lines
	for y := 0.0; y < float64(g.Sim.Map.Height); y++ {
		screenX1, screenY1 := g.transform.WorldToScreen(0, y)
		screenX2, screenY2 := g.transform.WorldToScreen(float64(g.Sim.Map.Width), y)
		ebitenutil.DrawLine(g.DynamicLayer, screenX1, screenY1, screenX2, screenY2, gridColor)
	}
}
func (g *Game) drawSimulation(screen *ebiten.Image) {
	// Draw the base layers
	g.drawStaticLayer()
	g.drawDynamicLayer()

	// Draw both layers to screen
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(g.StaticLayer, op)
	screen.DrawImage(g.DynamicLayer, op)

	// Draw metrics window
	g.drawMetricsWindow(screen)

	// Draw buttons
	g.PauseButton.Draw(screen)
	g.SimButton.Draw(screen)

	// Handle hover information
	mx, my := ebiten.CursorPosition()
	worldX, worldY := g.transform.ScreenToWorld(float64(mx), float64(my))

	// Draw coordinate debug if enabled
	if g.transform.debug {
		debugInfo := fmt.Sprintf(
			"Screen: (%d,%d)\nWorld: (%.2f,%.2f)",
			mx, my, worldX, worldY,
		)
		ebitenutil.DebugPrintAt(screen, debugInfo, 10, 10)
	}

	// Draw POI hover information
	if g.hoveredPos != nil {
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

	// Handle hover information for rescuers
	for _, drone := range g.Sim.Drones {
		if drone.Rescuer != nil {
			rescuerWorldPos := drone.Rescuer.Position
			rescuerScreenX, rescuerScreenY := g.transform.WorldToScreen(rescuerWorldPos.X, rescuerWorldPos.Y)

			if math.Abs(float64(mx)-rescuerScreenX) <= 9 &&
				math.Abs(float64(my)-rescuerScreenY) <= 9 {
				rescuerInfo := fmt.Sprintf(
					"Rescuer Info:\n"+
						"From Drone: %d\n"+
						"Status: %s\n"+
						"Target Person: %d\n"+
						"Position: (%.1f, %.1f)",
					drone.ID,
					map[int]string{0: "Going to Person", 1: "Returning to Tent"}[drone.Rescuer.State],
					drone.Rescuer.Person.ID,
					rescuerWorldPos.X,
					rescuerWorldPos.Y,
				)
				ebitenutil.DebugPrintAt(screen, rescuerInfo, mx+10, my+10)
				return
			}
		}
	}

	// Handle hover information for persons
	if hoveredPerson := g.getHoveredPerson(worldX, worldY); hoveredPerson != nil {
		personInfo := fmt.Sprintf(
			"Person Info:\n"+
				"ID: %d\n"+
				"In Distress: %t\n"+
				"Has Reached POI: %t\n"+
				"Position: (%.1f, %.1f)\n"+
				"CurrentDistressDuration: %d",
			hoveredPerson.ID,
			hoveredPerson.InDistress,
			hoveredPerson.HasReachedPOI(),
			hoveredPerson.Position.X,
			hoveredPerson.Position.Y,
			hoveredPerson.CurrentDistressDuration,
		)
		ebitenutil.DebugPrintAt(screen, personInfo, mx+10, my+10)
	}

	// Handle hover information for drones
	if hoveredDrone := g.getHoveredDrone(worldX, worldY); hoveredDrone != nil {
		droneInfo := fmt.Sprintf(
			"Drone Info:\n"+
				"ID: %d\n"+
				"Position: (%.1f, %.1f)\n"+
				"Watch Bounds: (%.1f, %.1f) - (%.1f, %.1f)\n"+
				"Battery: %.1f\n"+
				"Number of seen people: %d\n"+
				"Is charging: %t\n",
			hoveredDrone.ID,
			hoveredDrone.Position.X,
			hoveredDrone.Position.Y,
			hoveredDrone.MyWatch.CornerBottomLeft.X,
			hoveredDrone.MyWatch.CornerBottomLeft.Y,
			hoveredDrone.MyWatch.CornerTopRight.X,
			hoveredDrone.MyWatch.CornerTopRight.Y,
			hoveredDrone.Battery,
			len(hoveredDrone.SeenPeople),
			hoveredDrone.IsCharging,
		)
		ebitenutil.DebugPrintAt(screen, droneInfo, mx+10, my+10)
	}
}

func (g *Game) drawMetricsWindow(screen *ebiten.Image) {
	screenWidth := float64(screen.Bounds().Dx())
	screenHeight := float64(screen.Bounds().Dy())

	metricsWidth := screenWidth * 0.2
	metricsHeight := screenHeight * 0.3

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
			"Area Coverage: %.1f%%\n",
		stats.TotalPeople,
		stats.InDistress,
		stats.CasesTreated,
		stats.AverageBattery,
		stats.AverageCoverage,
	)
	ebitenutil.DebugPrintAt(metrics, text, 10, 10)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(screenWidth-metricsWidth-padding, padding)
	screen.DrawImage(metrics, opts)

	buttonWidth := metricsWidth
	buttonHeight := 40.0
	buttonSpacing := 10.0

	g.PauseButton.Width = buttonWidth
	g.PauseButton.Height = buttonHeight
	g.PauseButton.X = screenWidth - metricsWidth - padding
	g.PauseButton.Y = padding + metricsHeight + buttonSpacing

	g.SimButton.Width = buttonWidth
	g.SimButton.Height = buttonHeight
	g.SimButton.X = screenWidth - metricsWidth - padding
	g.SimButton.Y = g.PauseButton.Y + buttonHeight + buttonSpacing
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
