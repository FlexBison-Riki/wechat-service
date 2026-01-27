.PHONY: all build run test clean docker-build docker-run lint vet fmt

# Variables
BINARY_NAME=wechat-service
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build the binary
build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/server

# Run the application locally
run: build
	./$(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./... -coverprofile=coverage.out

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -rf logs/*

# Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

# Run Docker container
docker-run: docker-build
	docker run -p 8080:8080 $(BINARY_NAME):$(VERSION)

# Format code
fmt:
	$(GOFMT) -w ./...

# Lint code
lint:
	golangci-lint run ./...

# Vet code
vet:
	$(GOVET) ./...

# Download dependencies
deps:
	$(GOGET) -d ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  run        - Run the application locally"
	@echo "  test       - Run tests with coverage"
	@echo "  clean      - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run  - Run Docker container"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  vet        - Vet code"
	@echo "  deps       - Download dependencies"
