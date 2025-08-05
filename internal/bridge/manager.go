package bridge

import (
	"bridge-serial/config"
	"bridge-serial/internal/model"
	"bridge-serial/internal/serial"
	"bridge-serial/internal/socket"
	"bridge-serial/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type BridgeManager struct {
	config     *config.Config
	serial     *serial.SerialBridge
	wsServer   *socket.Server
	httpServer *http.Server
	stopChan   chan bool
	isRunning  bool
	wg         sync.WaitGroup
	mu         sync.Mutex
}

func NewBridgeManager(config *config.Config) *BridgeManager {
	return &BridgeManager{
		config:     config,
		serial:     serial.NewSerialBridge(&config.SerialBridge),
		wsServer:   socket.NewServer(),
		httpServer: nil,
		stopChan:   make(chan bool),
	}
}

func (bm *BridgeManager) createHTTPServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", bm.wsServer.ServeWS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		clientCount := bm.wsServer.GetConnectedClientsCount()
		fmt.Fprintf(w, `{"status":"ok","connected_clients":%d}`, clientCount)
	})

	return &http.Server{
		Addr:    bm.config.SocketConfig.Port,
		Handler: mux,
	}
}

func (bm *BridgeManager) Start() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.isRunning {
		logger.Error("bridge is already running")
		return fmt.Errorf("bridge is already running")
	}

	bm.stopChan = make(chan bool)

	bm.httpServer = bm.createHTTPServer()

	bm.wsServer.Start()
	logger.Info("WebSocket server started on %s", bm.config.SocketConfig.Port)

	bm.wg.Add(1)
	go func() {
		defer bm.wg.Done()
		logger.Info("WebSocket endpoint: ws://localhost%s/ws", bm.config.SocketConfig.Port)
		logger.Info("Health check: http://localhost%s/health", bm.config.SocketConfig.Port)

		if err := bm.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error: %v", err)
		}
		logger.Info("HTTP server goroutine stopped")
	}()

	err := bm.serial.Connect()
	if err != nil {
		logger.Error("failed to connect to serial port: %v", err)
		bm.wsServer.Stop()
		if bm.httpServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			bm.httpServer.Shutdown(ctx)
			cancel()
			bm.httpServer = nil
		}
		return fmt.Errorf("failed to connect to serial port: %v", err)
	}

	bm.isRunning = true
	bm.wg.Add(1)
	go bm.run()

	logger.Info("bridge started successfully")
	return nil
}

func (bm *BridgeManager) Stop() error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if !bm.isRunning {
		logger.Error("bridge is not running")
		return fmt.Errorf("bridge is not running")
	}

	logger.Info("Stopping bridge...")
	bm.isRunning = false

	if bm.stopChan != nil {
		close(bm.stopChan)
		bm.stopChan = nil
	}

	done := make(chan struct{})
	go func() {
		bm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Goroutines stopped successfully")
	case <-time.After(3 * time.Second):
		logger.Error("Timeout waiting for goroutines to stop")
	}

	bm.wsServer.Stop()

	if bm.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := bm.httpServer.Shutdown(ctx); err != nil {
			logger.Error("HTTP server forced to shutdown: %v", err)
		}
		bm.httpServer = nil // Clear reference
	}

	err := bm.serial.Disconnect()
	if err != nil {
		logger.Error("error disconnecting from serial port: %v", err)
	}

	logger.Info("bridge stopped")
	return nil
}

func (bm *BridgeManager) IsRunning() bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return bm.isRunning
}

func (bm *BridgeManager) run() {
	defer bm.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-bm.stopChan:
			logger.Info("Stop signal received, exiting run loop")
			return

		case <-ticker.C:
			if !bm.serial.IsConnected() {
				logger.Debug("Serial port not connected, skipping read")
				continue
			}

			data, err := bm.serial.ReadData()
			if err != nil {
				logger.Debug("no data from serial port: %v", err)
				continue
			}

			processedData, err := bm.processScaleData(data)
			if err != nil {
				logger.Error("error processing scale data: %v", err)
				continue
			}

			logger.Info("sending data to socket server: %s", data)
			err = bm.sendDataViaSocket(processedData, data)
			if err != nil {
				logger.Error("error sending data to socket server: %v", err)
				continue
			}

			logger.Info("successfully processed and sent scale data - Value: %.2f %s, Type: %s", processedData.Value, processedData.Unit, processedData.Type)
		}
	}
}

func (bm *BridgeManager) sendDataViaSocket(scaleData *model.ScaleDataRequest, rawData string) error {
	payload := map[string]interface{}{
		"scale_data": scaleData,
		"raw_data":   rawData,
		"timestamp":  time.Now().Unix(),
		"port":       bm.serial.GetPortName(),
	}

	bm.wsServer.BroadcastMessage("scale_data", payload)
	logger.Info("Broadcasted scale data to %d connected clients", bm.wsServer.GetConnectedClientsCount())

	return nil
}

func (b *BridgeManager) processScaleData(rawData string) (*model.ScaleDataRequest, error) {
	logger.Info("processing scale data: %s", rawData)
	var request model.ScaleDataRequest
	// Parse the raw data format: "WTST   12.11   g" or "WTUS    0.84   g"
	// Format: [PREFIX][SPACES][VALUE][SPACES][UNIT]
	fields := strings.Fields(rawData)
	if len(fields) < 2 {
		logger.Error("invalid scale data format: %s", rawData)
		return nil, fmt.Errorf("invalid scale data format: %s", rawData)
	}

	valueStr := fields[len(fields)-2]
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		logger.Error("failed to parse value '%s' from scale data: %v", valueStr, err)
		return nil, fmt.Errorf("failed to parse value '%s' from scale data: %v", valueStr, err)
	}

	unit := fields[len(fields)-1]
	dataType := fields[0]

	request = model.ScaleDataRequest{
		Value: value,
		Unit:  unit,
		Type:  dataType,
	}

	return &request, nil
}
