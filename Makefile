.PHONY: build run dev templ clean swagger

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/ cmd/main.go

# Run the application
run: build
	@echo "Starting application..."
	@./bin

# Run in development mode with hot reload
dev:
	@echo "Starting development server..."
	@air -c .air.toml

# Generate templ files
templ:
	@echo "Generating templ files..."
	@templ generate

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	@swag init -g cmd/main.go -o docs

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Database operations
db-up:
	@echo "Starting PostgreSQL..."
	@docker-compose up -d postgres

db-down:
	@echo "Stopping PostgreSQL..."
	@docker-compose down

db-logs:
	@echo "Showing PostgreSQL logs..."
	@docker-compose logs -f postgres

# Migration operations
migrate-up:
	@echo "Running migrations..."
	@migrate -path migrations -database "postgres://postgres:postgres@localhost:5444/aispace?sslmode=disable" up

migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path migrations -database "postgres://postgres:postgres@localhost:5444/aispace?sslmode=disable" down

migrate-create:
	@echo "Creating new migration: $(name)"
	@migrate create -ext sql -dir migrations $(name)

# Install development tools
tools:
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/a-h/templ/cmd/templ@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf tmp/

# Setup development environment
setup: deps tools templ swagger
	@echo "Development environment ready!"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@templ fmt .
