.PHONY: run build test clean migrate-up migrate-down generate-mocks

# Build the application (for local development)
build:
	@go build -o bin/server ./cmd/server

# Build for Docker (creates binary in build/ folder)
build-docker:
	@mkdir -p build
	@CGO_ENABLED=1 go build -o build/server ./cmd/server
	@echo "✓ Build successful! Binary created at: build/server"

# Run the application
run:
	@go run ./cmd/server

# Run with CGO disabled (for testing, won't work with PostgreSQL/MongoDB)
run-no-cgo:
	@CGO_ENABLED=0 go run ./cmd/server

# Run using Docker Compose (includes databases)
run-docker:
	@docker-compose up -d postgres mongodb
	@echo "Databases started. Run backend with: go run ./cmd/server"

# Run everything in Docker
run-docker-all:
	@docker-compose up

# Run tests
test:
	@go test ./... -v

# Run tests with coverage
test-coverage:
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out

# Run integration tests
test-integration:
	@go test ./tests/integration/... -v

# Clean build artifacts
clean:
	@rm -rf bin/
	@rm -f coverage.out

# Run database migrations
migrate-up:
	@echo "Running migrations..."
	@# TODO: Add migration command

# Rollback database migrations
migrate-down:
	@echo "Rolling back migrations..."
	@# TODO: Add rollback command

# Generate mocks
generate-mocks:
	@echo "Generating mocks..."
	@# TODO: Add mock generation command

# Install dependencies
deps:
	@go mod download
	@go mod tidy

# Format code
fmt:
	@go fmt ./...

# Lint code
lint:
	@golangci-lint run

# Run all checks
check: fmt lint test

