package main

import (
	"UTC_IA04/pkg/simulation"
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Mode int

const (
	Menu Mode = iota
	Simulation
)

type Button struct {
	X, Y, Width, Height float64
	Text                string
	OnClick             func()
	lastClicked         time.Time
}

func (b *Button) Draw(screen *ebiten.Image) {
	buttonImage := ebiten.NewImage(int(b.Width), int(b.Height))
	buttonImage.Fill(color.RGBA{0, 128, 255, 255}) // Blue background
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(b.X, b.Y)
	screen.DrawImage(buttonImage, opts)

	textX := int(b.X + b.Width/4)
	textY := int(b.Y + b.Height/3)
	ebitenutil.DebugPrintAt(screen, b.Text, textX, textY)
}

func (b *Button) Update(mx, my float64, pressed bool) bool {
	if time.Since(b.lastClicked) < 500*time.Millisecond {
		return false
	}

	if pressed && mx >= b.X && mx <= b.X+b.Width && my >= b.Y && my <= b.Y+b.Height {
		b.lastClicked = time.Now()
		if b.OnClick != nil {
			b.OnClick()
		}
		return true
	}
	return false
}

type Game struct {
	StaticLayer   *ebiten.Image
	DynamicLayer  *ebiten.Image
	Sim           *simulation.Simulation
	DroneImage    *ebiten.Image
	ObstacleImage *ebiten.Image
	Sliders       []Slider
	PauseButton   Button
	StartButton   Button
	Paused        bool
	Mode          Mode
}

type Slider struct {
	X, Y, Width, Height float64
	Min, Max, Value     int
	Label               string
	OnChange            func(value int)
}

func (s *Slider) Draw(screen *ebiten.Image) {
	bar := ebiten.NewImage(int(s.Width), int(s.Height))
	bar.Fill(color.RGBA{150, 150, 150, 255}) // Gray background
	barOpts := &ebiten.DrawImageOptions{}
	barOpts.GeoM.Translate(s.X, s.Y)
	screen.DrawImage(bar, barOpts)

	cursorX := s.X + (float64(s.Value-s.Min)/float64(s.Max-s.Min))*s.Width - 5
	cursor := ebiten.NewImage(10, int(s.Height))
	cursor.Fill(color.RGBA{255, 0, 0, 255}) // Red cursor
	cursorOpts := &ebiten.DrawImageOptions{}
	cursorOpts.GeoM.Translate(cursorX, s.Y)
	screen.DrawImage(cursor, cursorOpts)

	labelText := fmt.Sprintf("%s: %d", s.Label, s.Value)
	ebitenutil.DebugPrintAt(screen, labelText, int(s.X), int(s.Y)-20)
}

func (s *Slider) Update(mx, my float64, pressed bool) bool {
	if pressed && mx >= s.X && mx <= s.X+s.Width && my >= s.Y && my <= s.Y+s.Height {
		newValue := int(((mx-s.X)/s.Width)*float64(s.Max-s.Min)) + s.Min
		if newValue < s.Min {
			newValue = s.Min
		}
		if newValue > s.Max {
			newValue = s.Max
		}
		if newValue != s.Value {
			s.Value = newValue
			if s.OnChange != nil {
				s.OnChange(s.Value)
			}
		}
		return true
	}
	return false
}

func (g *Game) Update() error {
	mx, my := ebiten.CursorPosition()
	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	switch g.Mode {
	case Menu:
		g.StartButton.Update(float64(mx), float64(my), mousePressed)
	case Simulation:
		g.PauseButton.Update(float64(mx), float64(my), mousePressed)

		if g.Paused {
			return nil
		}

		for _, slider := range g.Sliders {
			slider.Update(float64(mx), float64(my), mousePressed)
		}

		g.Sim.Update()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.Mode {
	case Menu:
		screen.Fill(color.RGBA{0, 0, 0, 255}) // Black background
		g.StartButton.Draw(screen)
		ebitenutil.DebugPrint(screen, "Welcome to the Simulation!\nPress Start to begin.")
	case Simulation:
		g.drawStaticLayer()
		screen.DrawImage(g.StaticLayer, nil)

		g.drawDynamicLayer()
		screen.DrawImage(g.DynamicLayer, nil)

		g.drawMetricsWindow(screen)
		g.drawSliders(screen)

		g.PauseButton.Draw(screen)
	}
}

func (g *Game) drawStaticLayer() {
	g.StaticLayer.Fill(color.RGBA{34, 139, 34, 255})
	for _, obstacle := range g.Sim.Obstacles {
		drawObstacle(g.StaticLayer, obstacle.Position.X*30, obstacle.Position.Y*30, 30, 30, g.ObstacleImage)
	}
}

func (g *Game) drawDynamicLayer() {
	g.DynamicLayer.Clear()
	for _, person := range g.Sim.Persons {
		color := color.RGBA{255, 0, 0, 255}
		drawPerson(g.DynamicLayer, person.Position.X*30, person.Position.Y*30, 10, color)
	}
	for _, drone := range g.Sim.Drones {
		drawVisionCircle(g.DynamicLayer, drone.Position.X*30, drone.Position.Y*30, 60)
		drawDrone(g.DynamicLayer, drone.Position.X*30, drone.Position.Y*30, 20, g.DroneImage)
	}
}

func (g *Game) drawMetricsWindow(screen *ebiten.Image) {
	metrics := ebiten.NewImage(200, 200)
	metrics.Fill(color.RGBA{0, 0, 0, 200})
	text := fmt.Sprintf("Metrics:\n- Drones: %d\n- People: %d\n- Obstacles: %d",
		len(g.Sim.Drones), len(g.Sim.Persons), len(g.Sim.Obstacles))
	ebitenutil.DebugPrint(metrics, text)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(820, 50)
	screen.DrawImage(metrics, opts)
}

func (g *Game) drawSliders(screen *ebiten.Image) {
	for _, slider := range g.Sliders {
		slider.Draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1050, 600
}

func drawObstacle(img *ebiten.Image, x, y, width, height float64, obstacleImg *ebiten.Image) {
	if obstacleImg == nil {
		return
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(width/float64(obstacleImg.Bounds().Dx()), height/float64(obstacleImg.Bounds().Dy()))
	opts.GeoM.Translate(x, y)
	img.DrawImage(obstacleImg, opts)
}

func drawPerson(img *ebiten.Image, x, y, radius float64, clr color.Color) {
	circle := ebiten.NewImage(int(2*radius), int(2*radius))
	drawFilledCircle(circle, clr)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x-radius, y-radius)
	img.DrawImage(circle, opts)
}

func drawDrone(img *ebiten.Image, x, y, size float64, droneImg *ebiten.Image) {
	if droneImg == nil {
		return
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(size/float64(droneImg.Bounds().Dx()), size/float64(droneImg.Bounds().Dy()))
	opts.GeoM.Translate(x-size/2, y-size/2)
	img.DrawImage(droneImg, opts)
}

func drawVisionCircle(img *ebiten.Image, x, y, radius float64) {
	circle := ebiten.NewImage(int(2*radius), int(2*radius))
	drawFilledCircle(circle, color.RGBA{0, 255, 0, 64})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x-radius, y-radius)
	img.DrawImage(circle, opts)
}

func drawFilledCircle(img *ebiten.Image, clr color.Color) {
	w, h := img.Size()
	cx, cy := float64(w)/2, float64(h)/2
	radius := float64(w) / 2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := float64(x)-cx, float64(y)-cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, clr)
			}
		}
	}
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

func main() {
	rand.Seed(time.Now().UnixNano())

	game := &Game{
		StaticLayer:   ebiten.NewImage(800, 600),
		DynamicLayer:  ebiten.NewImage(800, 600),
		Sim:           simulation.NewSimulation(5, 20, 10),
		DroneImage:    loadImage("img/drone.png"),
		ObstacleImage: loadImage("img/chapiteau.png"),
		Paused:        false,
		Mode:          Menu,
	}

	// Initialize the Start button
	game.StartButton = Button{
		X: 400, Y: 300, Width: 200, Height: 50,
		Text: "Start Simulation",
		OnClick: func() {
			game.Mode = Simulation
		},
	}

	// Initialize the Pause button
	game.PauseButton = Button{
		X: 830, Y: 450, Width: 150, Height: 40,
		Text: "Pause",
		OnClick: func() {
			game.Paused = !game.Paused
			if game.Paused {
				game.PauseButton.Text = "Resume"
			} else {
				game.PauseButton.Text = "Pause"
			}
		},
	}

	game.Sliders = []Slider{
		{830, 360, 150, 20, 0, 20, 3, "People", func(value int) { game.Sim.UpdateCrowdSize(value) }},
		{830, 400, 150, 20, 0, 20, 2, "Drones", func(value int) { game.Sim.UpdateDroneSize(value) }},
	}

	ebiten.SetWindowSize(1050, 600)
	ebiten.SetWindowTitle("Simulation with Menu and Pause")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
