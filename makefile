ENTRY=cmd/app/main.go
OUTPUT=bin/rapier-bridge

run:
	go run $(ENTRY) --mode="development"

run-socket:
	go run cmd/websocket-server/main.go

build:
	go build -o $(OUTPUT) $(ENTRY)

clean:
	rm -rf bin/*

build-distribution:
	./script/build-windows.sh
	./script/build-macos.sh

install:
	./script/install.sh

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  run - Run the application in development mode"
	@echo "  build - Build the application"
	@echo "  install - Install the application"
	@echo "  help - Show this help message"

.PHONY: run build install help run-socket clean build-distribution