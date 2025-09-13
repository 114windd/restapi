# Makefile for Hybrid REST + gRPC Service

.PHONY: help build run test clean proto docker-build docker-run

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  proto        - Generate protobuf code"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"

# Build the application
build: proto
	@echo "Building hybrid service..."
	go build -o bin/hybrid-api cmd/server/main.go

# Run the application
run: build
	@echo "Starting hybrid service..."
	./bin/hybrid-api

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	protoc --go_out=pkg/proto --go-grpc_out=pkg/proto pkg/proto/user.proto

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t hybrid-api .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Install development tools
dev-tools:
	@echo "Installing development tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Run the test script
test-script: build
	@echo "Running test script..."
	./scripts/test_complete.sh
