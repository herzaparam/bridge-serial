package runner

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (r *App) createMainInterface() *fyne.Container {
	// Title
	title := widget.NewLabel("Bridge RS-232")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// Control buttons
	r.startButton = widget.NewButton("Start Bridge", r.onStartClick)
	r.stopButton = widget.NewButton("Stop Bridge", r.onStopClick)
	r.stopButton.Disable()

	// Status display
	r.statusDisplay = widget.NewLabel("")
	r.statusDisplay.Alignment = fyne.TextAlignCenter
	r.statusDisplay.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		widget.NewSeparator(),
		container.NewHBox(r.startButton, r.stopButton),
		r.statusDisplay,
	)
	return container.NewPadded(content)
}
