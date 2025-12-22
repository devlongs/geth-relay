.PHONY: build run test clean install-deps

build:
	@echo "Building geth-relay..."
	go build -o bin/geth-relay ./cmd/geth-relay
	@echo "Done building."
	@echo "Run \"./bin/geth-relay\" to launch geth-relay."

run:
	@echo "Running geth-relay..."
	go run ./cmd/geth-relay/main.go

run-config:
	@echo "Running geth-relay with config..."
	go run ./cmd/geth-relay/main.go -config configs/config.yaml

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

install-deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run
