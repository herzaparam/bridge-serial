package runner

import (
	"bridge-serial/config"
	"bridge-serial/internal/bridge"
	"bridge-serial/pkg/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	config *config.Config

	bridgeManager *bridge.BridgeManager

	// ui related
	window        fyne.Window
	statusDisplay *widget.Label
	startButton   *widget.Button
	stopButton    *widget.Button
}

func NewApp(cfg *config.Config) (*App, error) {
	myApp := app.New()
	myApp.SetIcon(nil)

	window := myApp.NewWindow(cfg.App.WindowTitle)
	window.Resize(fyne.NewSize(400, 200))

	return &App{
		config:        cfg,
		window:        window,
		bridgeManager: bridge.NewBridgeManager(cfg),
	}, nil
}

func (a *App) Run() {
	content := a.createMainInterface()
	a.window.SetContent(content)
	a.window.ShowAndRun()
	logger.Info("Application started")
}

func (a *App) onStartClick() {
	err := a.bridgeManager.Start()
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}

	a.startButton.Disable()
	a.stopButton.Enable()
	a.statusDisplay.SetText("running")
	logger.Info("Bridge started successfully")
}

func (a *App) onStopClick() {
	if a.bridgeManager != nil {
		err := a.bridgeManager.Stop()
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
	}

	a.startButton.Enable()
	a.stopButton.Disable()
	a.statusDisplay.SetText("stopped")
	logger.Info("Bridge stopped successfully")
}
