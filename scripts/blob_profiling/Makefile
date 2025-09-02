# Blob Profiling System Makefile

.PHONY: build test clean run help

# Build configuration
BINARY_NAME=blob_profiler
BUILD_DIR=build
MAIN_PATH=./cmd/blob_profiler

# Go build flags
LDFLAGS=-ldflags "-s -w"
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for specific platform
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

# Cross-compile for all platforms
build-all: build-linux build-darwin
	@echo "Cross-compilation complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests for specific package
test-profiler:
	go test -v ./internal/profiler

test-analysis:
	go test -v ./internal/analysis

test-config:
	go test -v ./internal/config

# Run the profiler
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with custom config
run-custom: build
	@echo "Running with custom config..."
	./$(BUILD_DIR)/$(BINARY_NAME) -config config.json

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-linux    - Build for Linux (amd64)"
	@echo "  build-darwin   - Build for macOS (arm64)"
	@echo "  build-all      - Build for all platforms"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-profiler  - Run profiler tests only"
	@echo "  test-analysis  - Run analysis tests only"
	@echo "  test-config    - Run config tests only"
	@echo "  run            - Build and run the profiler"
	@echo "  run-custom     - Run with custom config"
	@echo "  deps           - Install dependencies"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message"
