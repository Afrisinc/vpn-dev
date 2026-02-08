.PHONY: help build run test docker-build docker-up docker-down docker-logs lint fmt clean

# Variables
GO := go
DOCKER_COMPOSE := docker-compose
APP_NAME := vpn-dev
BINARY_PATH := ./cmd/myapp
BUILD_OUTPUT := ./vpn-dev.exe

help:
	@echo "VPN Management API - Available Commands"
	@echo "======================================="
	@echo "Local Development:"
	@echo "  make build          - Build the application locally"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run all tests"
	@echo "  make coverage       - Generate test coverage report"
	@echo "  make lint           - Run linters (fmt, vet, golangci-lint)"
	@echo "  make fmt            - Format code"
	@echo ""
	@echo "Docker Development:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start containers (docker-compose up)"
	@echo "  make docker-down    - Stop containers (docker-compose down)"
	@echo "  make docker-logs    - View container logs"
	@echo "  make docker-clean   - Remove containers, networks, and volumes"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate-up  - Run pending migrations"
	@echo "  make db-migrate-down- Rollback last migration"
	@echo "  make db-seed        - Seed database with test data"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make swagger        - Generate Swagger documentation"

# Local Build and Run
build:
	@echo "Building $(APP_NAME)..."
	cd $(BINARY_PATH) && $(GO) build -o ../../$(BUILD_OUTPUT) .
	@echo "✅ Build complete: $(BUILD_OUTPUT)"

run: build
	@echo "Running $(APP_NAME)..."
	./$(BUILD_OUTPUT)

test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

lint: fmt
	@echo "Running linters..."
	$(GO) vet ./...
	@command -v golangci-lint >/dev/null 2>&1 || (echo "Installing golangci-lint..." && $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	gofmt -s -w .

# Docker Operations
docker-build:
	@echo "Building Docker image..."
	$(DOCKER_COMPOSE) build

docker-up:
	@echo "Starting containers..."
	$(DOCKER_COMPOSE) up -d
	@echo "✅ Containers started"
	@echo "API: http://localhost:8080"
	@echo "Database: localhost:5432"

docker-down:
	@echo "Stopping containers..."
	$(DOCKER_COMPOSE) down

docker-logs:
	$(DOCKER_COMPOSE) logs -f vpn-api

docker-clean:
	@echo "Cleaning Docker resources..."
	$(DOCKER_COMPOSE) down -v
	@echo "✅ Cleanup complete"

docker-shell:
	$(DOCKER_COMPOSE) exec vpn-api sh

# Database Operations
db-migrate-up:
	@echo "Running migrations..."
	cd cmd/migrate && $(GO) run main.go up

db-migrate-down:
	@echo "Rolling back migrations..."
	cd cmd/migrate && $(GO) run main.go down

db-seed:
	@echo "Seeding database..."
	cd cmd/seed && $(GO) run main.go

# Swagger Documentation
swagger:
	@echo "Generating Swagger documentation..."
	@command -v swag >/dev/null 2>&1 || (echo "Installing swag..." && $(GO) install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/myapp/main.go -o docs

# Cleanup
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BUILD_OUTPUT)
	rm -f coverage.out coverage.html
	$(GO) clean ./...
	@echo "✅ Cleanup complete"

# Development Helper
dev: docker-down docker-build docker-up db-migrate-up
	@echo "✅ Development environment ready!"
	@echo "API running at http://localhost:8080"
	@echo "Postgres at localhost:5432"
	@echo ""
	@echo "View logs with: make docker-logs"
	@echo "Stop with: make docker-down"

# CI/CD Testing
ci-test: lint test
	@echo "✅ All CI checks passed!"
