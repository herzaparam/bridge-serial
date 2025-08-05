ENTRY=cmd/app/main.go
OUTPUT=bin/rapier-bridge

run:
	go run $(ENTRY) --mode="development"

run-socket:
	go run cmd/websocket-server/main.go

build:
	go build -o $(OUTPUT) $(ENTRY)

install:
	./script/install.sh

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  run - Run the application in development mode"
	@echo "  build - Build the application"
	@echo "  install - Install the application"
	@echo "  help - Show this help message"

.PHONY: run build install help run-socket