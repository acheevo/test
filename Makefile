.PHONY: help run build test test-integration test-unit clean deps lint fmt docker-build docker-run docker-stop docker-clean

# Default target
help:
	@echo "Available commands:"
	@echo "  run                - Run the application"
	@echo "  build              - Build the application"
	@echo "  test               - Run all tests"
	@echo "  test-unit          - Run unit tests"
	@echo "  test-integration   - Run integration tests"
	@echo "  lint               - Run linter"
	@echo "  fmt                - Format code"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Install dependencies"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run with Docker Compose"
	@echo "  docker-stop        - Stop Docker containers"
	@echo "  docker-clean       - Clean Docker resources"

# Run the application
run:
	go run cmd/api/main.go

# Build the application
build:
	mkdir -p bin
	go build -o bin/api cmd/api/main.go

# Run all tests
test:
	go test -v -race ./...

# Run unit tests
test-unit:
	go test -v -race ./internal/...

# Run integration tests
test-integration:
	go test -v -race ./tests/integration/...

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Format code
fmt:
	go fmt ./...
	goimports -w -local github.com/acheevo/test .

# Clean build artifacts
clean:
	rm -rf bin/
	docker system prune -f
	docker volume prune -f

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Docker commands
docker-build:
	docker build -t test-api .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

docker-clean:
	docker-compose down -v
	docker system prune -f
	docker volume prune -f

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Pre-commit checks
pre-commit: fmt lint test

# Build for production
build-prod:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/api cmd/api/main.go