package config

import (
	"bridge-serial/pkg/logger"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.bug.st/serial"
)

type Config struct {
	App          AppConfig
	SerialBridge SerialBridgeConfig
	HTTPClient   HTTPClientConfig
	SocketConfig SocketConfig

	User     string
	Password string
}

type AppConfig struct {
	AppName     string
	WindowTitle string
	Mode        string
}

type SerialBridgeConfig struct {
	DataBits int
	Parity   serial.Parity
	StopBits serial.StopBits
	Timeout  time.Duration
	BaudRate int
}

type HTTPClientConfig struct {
	BaseURL string
}

type SocketConfig struct {
	Port          string
	RetryInterval time.Duration
}

func LoadConfig(mode string) (*Config, error) {
	return &Config{
		App: AppConfig{
			AppName:     "rapier-bridge",
			WindowTitle: "Rapier Bridge Serial",
			Mode:        mode,
		},
		SerialBridge: SerialBridgeConfig{
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: serial.OneStopBit,
			Timeout:  10 * time.Second,
			BaudRate: 9600,
		},
		HTTPClient: HTTPClientConfig{
			BaseURL: "http://localhost:8080",
		},
		SocketConfig: SocketConfig{
			Port:          ":8001",
			RetryInterval: 5 * time.Second,
		},
	}, nil
}

func (c *Config) GetDefaultConfigPath() string {
	return filepath.Join(getConfigDir(c.App.AppName), "config.json")
}

func getConfigDir(appName string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get user home directory: %v", err)
		return "."
	}
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), appName)
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", appName)
	case "linux":
		return filepath.Join(homeDir, ".config", appName)
	default:
		return filepath.Join(homeDir, ".config", appName)
	}
}

func (c *Config) IsConfigExist() bool {
	configPath := c.GetDefaultConfigPath()
	_, err := os.Stat(configPath)
	if err != nil {
		logger.Error("Failed to check if config exists: %v", err)
	}
	return err == nil
}

func (c *Config) ReadConfig() (*Config, error) {
	configPath := c.GetDefaultConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return c, nil
}
