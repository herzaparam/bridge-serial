package main

import (
	"bridge-serial/config"
	"bridge-serial/internal/runner"
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

	app, err := runner.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	app.Run()

}
