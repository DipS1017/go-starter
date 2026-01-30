# Simple Makefile for a Go project

# Build the application
all: build test
-include .env

build:
	@echo "Building..."
	
	
	@go build -o main main.go

db-up:
	@echo "Starting database..."
	@docker compose --profile db up --detach
	
db-down:
	@echo "Starting database..."
	@docker compose --profile db down


# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@if [ -d ./internal/database ]; then \
		go test ./internal/database -v; \
	else \
		echo "Skipping integration tests: ./internal/database not found"; \
	fi

# End-to-End Tests
e2e:
	@echo "Running E2E tests..."
	@if [ -d ./tests/e2e/suites ]; then \
		go test ./tests/e2e/suites/... -v -timeout 10m; \
	else \
		echo "Skipping E2E tests: ./tests/e2e/suites not found"; \
	fi

# Run specific E2E test suite
e2e-suite:
	@echo "Running E2E test suite: $(SUITE)"
	@if [ -d ./tests/e2e/suites ]; then \
		go test ./tests/e2e/suites -v -timeout 10m -run $(SUITE); \
	else \
		echo "Skipping E2E suite: ./tests/e2e/suites not found"; \
	fi

# Run E2E tests with coverage
e2e-coverage:
	@echo "Running E2E tests with coverage..."
	@if [ -d ./tests/e2e/suites ]; then \
		go test ./tests/e2e/suites/... -v -timeout 10m -coverprofile=e2e-coverage.out; \
		go tool cover -html=e2e-coverage.out -o e2e-coverage.html; \
		echo "Coverage report generated: e2e-coverage.html"; \
	else \
		echo "Skipping E2E coverage: ./tests/e2e/suites not found"; \
	fi

# Run all tests (unit, integration, and E2E)
test-all: test itest e2e
	@echo "All tests completed!"

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch docker-run docker-down itest e2e e2e-suite e2e-coverage test-all
	 
rollback:
	POSTGRESQL_URL="$(POSTGRESQL_URL)" tern migrate --migrations internal/db/migrations/ --destination ${dest}
