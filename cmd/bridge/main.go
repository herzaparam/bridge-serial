package main

import (
	"bridge-serial/config"
	"bridge-serial/internal/bridge"
	"bridge-serial/pkg/logger"
	"flag"
	"log"
)

func main() {
	mode := flag.String("mode", "production", "mode of the application")
	cfg, err := config.LoadConfig(*mode)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = logger.Init(logger.INFO, "./logs")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	bManager := bridge.NewBridgeManager(cfg)
	if err := bManager.Start(); err != nil {
		log.Fatalf("Failed to start bridge manager: %v", err)
	}

}
