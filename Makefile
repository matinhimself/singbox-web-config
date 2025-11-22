.PHONY: help build build-server build-generator dev clean test

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: build-server build-generator ## Build all binaries

build-server: ## Build the web server binary
	@echo "Building server..."
	@go build -v -o singbox-web-config ./cmd/server
	@echo "✓ Server built: ./singbox-web-config"

build-generator: ## Build the type generator binary
	@echo "Building generator..."
	@go build -v -o singbox-generator ./cmd/generator
	@echo "✓ Generator built: ./singbox-generator"

dev: ## Run the server in development mode
	@echo "Starting development server..."
	@go run ./cmd/server

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f singbox-web-config singbox-generator
	@echo "✓ Clean complete"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

.DEFAULT_GOAL := help
