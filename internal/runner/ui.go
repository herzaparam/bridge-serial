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

func (a *App) createFormInterface() *fyne.Container {
	// Title
	title := widget.NewLabel("Rapier Bridge RS-232")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// User credentials
	userLabel := widget.NewLabel("Username:")
	a.userInput = widget.NewEntry()
	a.userInput.SetPlaceHolder("Enter username")

	passwordLabel := widget.NewLabel("Password:")
	a.passwordInput = widget.NewPasswordEntry()
	a.passwordInput.SetPlaceHolder("Enter password")

	a.verifyButton = widget.NewButton("Verify", a.onVerifyClick)

	// Main layout - everything in one window
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		userLabel,
		a.userInput,
		passwordLabel,
		a.passwordInput,
		widget.NewSeparator(),
		a.verifyButton,
	)
	return container.NewPadded(content)
}
