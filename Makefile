.PHONY: all build run test clean docker-build docker-run lint vet fmt deps db-migrate db-rollback db-backup db-restore

# Variables
BINARY_NAME=wechat-service
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)
GO_FILES=$(shell find . -name "*.go" -type f | grep -v "_test.go")

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
	$(GOTEST) -v ./... -coverprofile=coverage.out -race

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -rf logs/*
	rm -rf *.tar.gz

# Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

# Run Docker container
docker-run: docker-build
	docker run -p 8080:8080 \
		-e WECHAT_APPID=$$WECHAT_APPID \
		-e WECHAT_APPSECRET=$$WECHAT_APPSECRET \
		-e WECHAT_TOKEN=$$WECHAT_TOKEN \
		-e REDIS_ADDR=redis:6379 \
		--network=wechat-service_default \
		--name=$(BINARY_NAME)_dev \
		$(BINARY_NAME):latest

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

# Database migrations
db-migrate:
	@echo "Running database migrations..."
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f migrations/001_init_schema.sql

# Database rollback (drop all tables - use with caution!)
db-rollback:
	@echo "Dropping all tables (use with caution!)..."
	@echo "This will delete all data. Type 'yes' to continue."
	@read confirm; if [ "$$confirm" != "yes" ]; then echo "Cancelled"; exit 1; fi
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Database backup
db-backup:
	@echo "Backing up database..."
	@./scripts/db/backup.sh backup

# Database restore
db-restore:
	@echo "Restoring database..."
	@./scripts/db/backup.sh restore $(BACKUP_FILE)

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build & Run:"
	@echo "  build         - Build the binary"
	@echo "  run           - Run the application locally"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo ""
	@echo "Code Quality:"
	@echo "  test          - Run tests with coverage"
	@echo "  lint          - Lint code"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate    - Run database migrations"
	@echo "  db-rollback   - Drop all tables (dangerous!)"
	@echo "  db-backup     - Backup database"
	@echo "  db-restore    - Restore database"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  help          - Show this help"
	@echo ""
	@echo "Environment Variables:"
	@echo "  DB_HOST       - Database host (default: localhost)"
	@echo "  DB_PORT       - Database port (default: 5432)"
	@echo "  DB_USER       - Database user (default: postgres)"
