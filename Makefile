.PHONY: build run test clean docker-build docker-up docker-down migrate seed

# Build the application
build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

# Run the API server
run-api:
	go run ./cmd/api

# Run the worker
run-worker:
	go run ./cmd/worker

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build Docker images
docker-build:
	docker build -t hopper-api:latest --target api .
	docker build -t hopper-worker:latest --target worker .

# Start Docker Compose
docker-up:
	docker-compose up -d

# Stop Docker Compose
docker-down:
	docker-compose down

# View Docker logs
docker-logs:
	docker-compose logs -f

# Run database migrations
migrate:
	@echo "Running database migrations..."
	psql postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE) < migrations/000001_initial_schema.up.sql

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	psql postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE) < migrations/000001_initial_schema.down.sql

# Seed database with sample data
seed:
	@echo "Seeding database..."
	psql postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE) < seeds/seed.sql

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run ./...

# Run all checks
check: fmt lint test

# Generate OpenAPI docs
openapi:
	@echo "OpenAPI spec already generated in openapi/openapi.yaml"
