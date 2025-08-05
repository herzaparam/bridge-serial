package runner

import (
	"bridge-serial/config"
	"bridge-serial/internal/bridge"
	"bridge-serial/pkg/logger"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	config *config.Config

	bridgeManager *bridge.BridgeManager

	// ui related
	window fyne.Window
	// portSelect    *widget.Select
	userInput     *widget.Entry
	passwordInput *widget.Entry
	statusDisplay *widget.Label
	startButton   *widget.Button
	stopButton    *widget.Button
	verifyButton  *widget.Button
}

func NewApp(cfg *config.Config) (*App, error) {
	myApp := app.New()
	myApp.SetIcon(nil)

	window := myApp.NewWindow(cfg.App.WindowTitle)
	window.Resize(fyne.NewSize(400, 100))

	return &App{
		config: cfg,
		window: window,
	}, nil
}

func (a *App) Run() {
	var content *fyne.Container

	if a.hasValidCredentials() {
		logger.Debug("Processing request: %s", "main interface")
		content = a.createMainInterface()
		a.bridgeManager = bridge.NewBridgeManager(a.config)
	} else {
		logger.Debug("Processing request: %s", "form interface")
		content = a.createFormInterface()
	}

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

func (a *App) onVerifyClick() {
	logger.Info("Verify button clicked, validating user input")

	if err := a.validateUser(); err != nil {
		logger.Error("Failed to validate user: %v", err)
		dialog.ShowError(err, a.window)
		return
	}

	logger.Info("User validation passed, writing configuration")

	logger.Info("Configuration saved successfully, showing success dialog")
	dialog.ShowInformation("Success", "Configuration saved successfully!", a.window)

	// Switch to main interface after successful configuration
	logger.Info("Switching to main interface")
	content := a.createMainInterface()
	a.window.SetContent(content)
}

func (a *App) hasValidCredentials() bool {
	if !a.config.IsConfigExist() {
		return false
	}

	config, err := a.config.ReadConfig()
	if err != nil {
		logger.Error("Failed to read config for validation: %v", err)
		return false
	}

	// Check if both username and password are not empty
	return config.User != "" && config.Password != ""
}

func (a *App) validateUser() error {
	if a.userInput.Text == "" {
		return fmt.Errorf("please add a username")
	}

	if a.passwordInput.Text == "" {
		return fmt.Errorf("please add a password")
	}

	return nil
}
