package serial

import (
	"bridge-serial/config"
	"bridge-serial/pkg/logger"
	"bufio"
	"fmt"
	"strings"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type SerialBridge struct {
	port     serial.Port
	portName string
	reader   *bufio.Reader

	config *config.SerialBridgeConfig
}

func NewSerialBridge(cfg *config.SerialBridgeConfig) *SerialBridge {
	return &SerialBridge{config: cfg}
}

// Connect establishes connection to the serial port
func (s *SerialBridge) Connect() error {
	mode := &serial.Mode{
		BaudRate: s.config.BaudRate,
		DataBits: s.config.DataBits,
		Parity:   s.config.Parity,
		StopBits: s.config.StopBits,
	}

	err := s.getPortDevice()
	if err != nil {
		return fmt.Errorf("failed to get port device: %v", err)
	}

	port, err := serial.Open(s.portName, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port: %v", err)
	}

	// Set read timeout
	err = port.SetReadTimeout(s.config.Timeout)
	if err != nil {
		port.Close()
		return fmt.Errorf("failed to set read timeout: %v", err)
	}

	s.port = port
	s.reader = bufio.NewReader(port)
	logger.Info("connected to serial port: %s", s.portName)
	return nil
}

// Disconnect closes the serial port connection
func (s *SerialBridge) Disconnect() error {
	if s.port != nil {
		err := s.port.Close()
		s.port = nil
		s.reader = nil
		logger.Info("disconnected from serial port: %s", s.portName)
		return err
	}
	logger.Error("failed to disconnect from serial port: %s", s.portName)
	return nil
}

// ReadData reads data from the serial port
func (s *SerialBridge) ReadData() (string, error) {
	if s.reader == nil {
		return "", fmt.Errorf("serial port not connected")
	}
	// Read until newline or timeout
	data, err := s.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read from serial port: %v", err)
	}

	// Clean the data (remove newlines and whitespace)
	data = strings.TrimSpace(data)
	logger.Debug("read data from serial port: %s", data)
	return data, nil
}

func (s *SerialBridge) getPortDevice() error {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		logger.Error("failed to enumerate ports: %v", err)
		return fmt.Errorf("failed to enumerate ports: %v", err)
	}

	for _, port := range ports {
		if port.IsUSB && port.VID == "067B" && port.PID == "2303" {
			s.portName = port.Name
			break
		}
	}
	return nil
}

// IsConnected returns true if the serial port is connected
func (s *SerialBridge) IsConnected() bool {
	return s.port != nil
}

// GetPortName returns the current port name
func (s *SerialBridge) GetPortName() string {
	return s.portName
}
