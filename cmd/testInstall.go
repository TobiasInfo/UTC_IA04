package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Test Fyne")

	hello := widget.NewLabel("Hello Fyne!")
	btn := widget.NewButton("Quitter", func() {
		a.Quit()
	})

	w.SetContent(container.NewVBox(
		hello,
		btn,
	))

	w.Resize(fyne.NewSize(300, 200))
	w.Show()
	a.Run()
}
