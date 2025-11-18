.PHONY: help dev build test clean docker-up docker-down

# Colors for output
CYAN := \033[0;36m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(CYAN)MMO Development Commands:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

# Development
dev: ## Start development server (Go)
	@echo "$(CYAN)Starting development server...$(NC)"
	cd server && ENVIRONMENT=development go run ./cmd/server

dev-client: ## Open Godot client
	@echo "$(CYAN)Opening Godot client...$(NC)"
	cd client && godot project.godot

# Building
build: ## Build server binary
	@echo "$(CYAN)Building server...$(NC)"
	cd server && go build -o ../bin/mmo-server ./cmd/server
	@echo "$(GREEN)✓ Binary built: bin/mmo-server$(NC)"

build-demo: ## Build combat demo
	@echo "$(CYAN)Building combat demo...$(NC)"
	cd server && go build -o ../bin/combat_demo ./cmd/combat_demo
	@echo "$(GREEN)✓ Binary built: bin/combat_demo$(NC)"

build-all: build build-demo ## Build all binaries

# Testing
test: ## Run Go tests
	@echo "$(CYAN)Running tests...$(NC)"
	cd server && go test ./... -v

test-combat: ## Run combat demo
	@echo "$(CYAN)Running combat demo...$(NC)"
	cd server && go run ./cmd/combat_demo

test-websocket: ## Open WebSocket test client
	@echo "$(CYAN)Opening WebSocket test client...$(NC)"
	@echo "Navigate to: http://localhost:3000/test/websocket_test.html"
	@which open > /dev/null && open client/test/websocket_test.html || xdg-open client/test/websocket_test.html || echo "Please open client/test/websocket_test.html in your browser"

# Docker
docker-build: ## Build Docker image
	@echo "$(CYAN)Building Docker image...$(NC)"
	docker build -t mmo-server:latest .

docker-up: ## Start all services with Docker Compose
	@echo "$(CYAN)Starting Docker Compose...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Services started:$(NC)"
	@echo "  - Server: http://localhost:8080"
	@echo "  - Client: http://localhost:3000"
	@echo "  - Adminer: http://localhost:8081"

docker-down: ## Stop all Docker services
	@echo "$(CYAN)Stopping Docker Compose...$(NC)"
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f server

docker-restart: docker-down docker-up ## Restart Docker services

# Database
db-migrate: ## Run database migrations (future)
	@echo "$(YELLOW)Database migrations not yet implemented$(NC)"

db-seed: ## Seed database with test data (future)
	@echo "$(YELLOW)Database seeding not yet implemented$(NC)"

# Cleanup
clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -f server/mmo-server
	rm -f server/combat_demo
	@echo "$(GREEN)✓ Cleaned$(NC)"

clean-docker: ## Clean Docker volumes and images
	@echo "$(CYAN)Cleaning Docker resources...$(NC)"
	docker-compose down -v
	docker rmi mmo-server:latest 2>/dev/null || true
	@echo "$(GREEN)✓ Docker cleaned$(NC)"

# Dependencies
deps: ## Install/update Go dependencies
	@echo "$(CYAN)Installing dependencies...$(NC)"
	cd server && go mod download
	cd server && go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

# Environment setup
setup-dev: ## Setup development environment
	@echo "$(CYAN)Setting up development environment...$(NC)"
	@cp .env.development .env 2>/dev/null || echo "$(YELLOW).env already exists$(NC)"
	@echo "$(GREEN)✓ Development environment ready$(NC)"
	@echo ""
	@echo "$(CYAN)Next steps:$(NC)"
	@echo "  1. make deps        - Install dependencies"
	@echo "  2. make docker-up   - Start all services"
	@echo "  3. make dev         - Run server locally"

# Linting & Formatting
lint: ## Run linters
	@echo "$(CYAN)Running linters...$(NC)"
	cd server && go vet ./...
	cd server && gofmt -l .
	@echo "$(GREEN)✓ Linting complete$(NC)"

fmt: ## Format code
	@echo "$(CYAN)Formatting code...$(NC)"
	cd server && gofmt -w .
	@echo "$(GREEN)✓ Code formatted$(NC)"

# Performance
benchmark: ## Run benchmarks
	@echo "$(CYAN)Running benchmarks...$(NC)"
	cd server && go test -bench=. -benchmem ./...

# Quick start commands
quick-start: setup-dev deps docker-up ## Complete quick start setup
	@echo "$(GREEN)✓ Quick start complete!$(NC)"
	@echo ""
	@echo "$(CYAN)Services running:$(NC)"
	@echo "  - Server:  http://localhost:8080"
	@echo "  - Client:  http://localhost:3000"
	@echo "  - Test UI: http://localhost:3000/test/websocket_test.html"

# Production build
build-prod: ## Build production-optimized binary
	@echo "$(CYAN)Building production binary...$(NC)"
	cd server && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o ../bin/mmo-server-prod ./cmd/server
	@echo "$(GREEN)✓ Production binary built: bin/mmo-server-prod$(NC)"
