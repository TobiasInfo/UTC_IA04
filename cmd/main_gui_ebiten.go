package main

import (
	"UTC_IA04/pkg/simulation"
	"fmt"
	"image/color"
	"math/rand"
	"strconv"
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

func (b *Button) Update(mx, my float64, pressed bool) {
	if time.Since(b.lastClicked) < 500*time.Millisecond {
		return
	}

	if pressed && mx >= b.X && mx <= b.X+b.Width && my >= b.Y && my <= b.Y+b.Height {
		b.lastClicked = time.Now()
		if b.OnClick != nil {
			b.OnClick()
		}
	}
}

type TextField struct {
	X, Y, Width, Height float64
	Text                string
	IsActive            bool
	OnEnter             func(value int)
}

func (tf *TextField) Draw(screen *ebiten.Image) {
	field := ebiten.NewImage(int(tf.Width), int(tf.Height))
	if tf.IsActive {
		field.Fill(color.RGBA{200, 200, 255, 255}) // Light blue when active
	} else {
		field.Fill(color.RGBA{255, 255, 255, 255}) // White otherwise
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(tf.X, tf.Y)
	screen.DrawImage(field, opts)

	ebitenutil.DebugPrintAt(screen, tf.Text, int(tf.X+5), int(tf.Y+5))
}

func (tf *TextField) Update(mx, my float64, pressed bool, inputChars []rune, enterPressed bool) {
	if pressed && mx >= tf.X && mx <= tf.X+tf.Width && my >= tf.Y && my <= tf.Y+tf.Height {
		tf.IsActive = true
	} else if pressed {
		tf.IsActive = false
	}

	if tf.IsActive {
		for _, char := range inputChars {
			if char >= '0' && char <= '9' {
				tf.Text += string(char)
			}
		}

		if ebiten.IsKeyPressed(ebiten.KeyBackspace) && len(tf.Text) > 0 {
			tf.Text = tf.Text[:len(tf.Text)-1]
		}

		if enterPressed {
			value, err := strconv.Atoi(tf.Text)
			if err == nil && tf.OnEnter != nil {
				tf.OnEnter(value)
			}
			tf.IsActive = false
		}
	}
}

type Game struct {
	Mode          Mode
	StartButton   Button
	PauseButton   Button
	DroneField    TextField
	PeopleField   TextField
	ObstacleField TextField
	Sim           *simulation.Simulation
	StaticLayer   *ebiten.Image
	DynamicLayer  *ebiten.Image
	Paused        bool
	DroneCount    int
	PeopleCount   int
	ObstacleCount int
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

	// Dessin des personnes
	for _, person := range g.Sim.Persons {
		drawCircle(g.DynamicLayer, person.Position.X*30, person.Position.Y*30, 10, color.RGBA{255, 0, 0, 255})
	}

	// Dessin des drones et de leur champ de vision
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

func drawRectangle(img *ebiten.Image, x, y, width, height float64, clr color.Color) {
	rect := ebiten.NewImage(int(width), int(height))
	rect.Fill(clr)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x, y)
	img.DrawImage(rect, opts)
}

func drawCircle(img *ebiten.Image, x, y, radius float64, clr color.Color) {
	circle := ebiten.NewImage(int(2*radius), int(2*radius))
	for cy := 0; cy < int(2*radius); cy++ {
		for cx := 0; cx < int(2*radius); cx++ {
			dx := float64(cx) - radius
			dy := float64(cy) - radius
			if dx*dx+dy*dy <= radius*radius {
				circle.Set(cx, cy, clr)
			}
		}
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x-radius, y-radius)
	img.DrawImage(circle, opts)
}

func drawTranslucentCircle(img *ebiten.Image, x, y, radius float64, clr color.Color) {
	circle := ebiten.NewImage(int(2*radius), int(2*radius))
	for cy := 0; cy < int(2*radius); cy++ {
		for cx := 0; cx < int(2*radius); cx++ {
			dx := float64(cx) - radius
			dy := float64(cy) - radius
			if dx*dx+dy*dy <= radius*radius {
				circle.Set(cx, cy, clr)
			}
		}
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(x-radius, y-radius)
	img.DrawImage(circle, opts)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	game := &Game{
		Mode:          Menu,
		DroneCount:    5,
		PeopleCount:   10,
		ObstacleCount: 5,
		StaticLayer:   ebiten.NewImage(800, 600),
		DynamicLayer:  ebiten.NewImage(800, 600),
	}

	game.DroneField = TextField{
		X: 150, Y: 150, Width: 200, Height: 30, Text: "5",
		OnEnter: func(value int) {
			game.DroneCount = value
		},
	}
	game.PeopleField = TextField{
		X: 150, Y: 200, Width: 200, Height: 30, Text: "10",
		OnEnter: func(value int) {
			game.PeopleCount = value
		},
	}
	game.ObstacleField = TextField{
		X: 150, Y: 250, Width: 200, Height: 30, Text: "5",
		OnEnter: func(value int) {
			game.ObstacleCount = value
		},
	}

	game.StartButton = Button{
		X: 300, Y: 350, Width: 200, Height: 50, Text: "Start Simulation",
		OnClick: func() {
			game.Sim = simulation.NewSimulation(game.DroneCount, game.PeopleCount, game.ObstacleCount)
			game.Mode = Simulation
		},
	}

	game.PauseButton = Button{
		X: 300, Y: 400, Width: 150, Height: 40, Text: "Pause",
		OnClick: func() {
			game.Paused = !game.Paused
			if game.Paused {
				game.PauseButton.Text = "Resume"
			} else {
				game.PauseButton.Text = "Pause"
			}
		},
	}

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Simulation with Vision Circles")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
