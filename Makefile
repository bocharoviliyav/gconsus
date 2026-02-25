.PHONY: help build test build-images publish-images deploy clean dev up down logs migrate-up migrate-down

# Variables
DOCKER_REGISTRY ?=
IMAGE_TAG ?= latest
BACKEND_IMAGE = $(DOCKER_REGISTRY)gconsus-backend
FRONTEND_IMAGE = $(DOCKER_REGISTRY)gconsus-frontend

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Git Analytics Platform - Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Development
dev: ## Start development environment
	@echo "$(BLUE)Starting development environment...$(NC)"
	docker-compose up -d --build
	@echo "$(GREEN)Waiting for services to be ready...$(NC)"
	@sleep 5
	@echo "$(GREEN)Development environment is ready!$(NC)"
	@echo "$(YELLOW)PostgreSQL:$(NC) localhost:5432"
	@echo "$(YELLOW)Frontend:$(NC) http://localhost"
	@echo "$(YELLOW)Backend:$(NC) http://localhost:8000"
	@echo "$(YELLOW)Keycloak:$(NC) http://localhost:8090"

up: ## Start all services
	@echo "$(BLUE)Starting all services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)All services started!$(NC)"
	@echo "$(YELLOW)Frontend:$(NC) http://localhost"
	@echo "$(YELLOW)Backend:$(NC) http://localhost:8000"
	@echo "$(YELLOW)Keycloak:$(NC) http://localhost:8090"

down: ## Stop all services
	@echo "$(BLUE)Stopping all services...$(NC)"
	docker-compose down
	@echo "$(GREEN)All services stopped$(NC)"

logs: ## Show logs from all services
	docker-compose logs -f

logs-backend: ## Show backend logs
	docker-compose logs -f backend

logs-frontend: ## Show frontend logs
	docker-compose logs -f frontend

# Build
build: build-backend build-frontend ## Build backend and frontend

build-backend: ## Build backend binary
	@echo "$(BLUE)Building backend...$(NC)"
	go mod download
	go build -o bin/git-analytics .
	@echo "$(GREEN)Backend built successfully$(NC)"

build-frontend: ## Build frontend
	@echo "$(BLUE)Building frontend...$(NC)"
	cd frontend && bun install && bun run build
	@echo "$(GREEN)Frontend built successfully$(NC)"

# Test
test: test-backend test-frontend ## Run all tests

test-backend: ## Run backend tests
	@echo "$(BLUE)Running backend tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Backend tests completed$(NC)"

test-backend-short: ## Run backend tests (short)
	@echo "$(BLUE)Running backend tests (short)...$(NC)"
	go test -short -v ./...

test-frontend: ## Run frontend tests
	@echo "$(BLUE)Running frontend tests...$(NC)"
	cd frontend && bun test
	@echo "$(GREEN)Frontend tests completed$(NC)"

# Lint
lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Run backend linter
	@echo "$(BLUE)Running backend linter...$(NC)"
	golangci-lint run
	@echo "$(GREEN)Backend linting completed$(NC)"

lint-frontend: ## Run frontend linter
	@echo "$(BLUE)Running frontend linter...$(NC)"
	cd frontend && bun run lint
	@echo "$(GREEN)Frontend linting completed$(NC)"

# Format
fmt: ## Format code
	@echo "$(BLUE)Formatting backend code...$(NC)"
	go fmt ./...
	@echo "$(BLUE)Formatting frontend code...$(NC)"
	cd frontend && bun run format
	@echo "$(GREEN)Code formatted$(NC)"

# Docker Images
build-images: build-image-backend build-image-frontend ## Build Docker images

build-image-backend: ## Build backend Docker image
	@echo "$(BLUE)Building backend Docker image...$(NC)"
	docker build -t $(BACKEND_IMAGE):$(IMAGE_TAG) \
		-t $(BACKEND_IMAGE):latest \
		--target production .
	@echo "$(GREEN)Backend image built: $(BACKEND_IMAGE):$(IMAGE_TAG)$(NC)"

build-image-frontend: ## Build frontend Docker image
	@echo "$(BLUE)Building frontend Docker image...$(NC)"
	docker build -t $(FRONTEND_IMAGE):$(IMAGE_TAG) \
		-t $(FRONTEND_IMAGE):latest \
		--build-arg VITE_API_URL=${VITE_API_URL} \
		--build-arg VITE_KEYCLOAK_URL=${VITE_KEYCLOAK_URL} \
		--build-arg VITE_KEYCLOAK_REALM=${VITE_KEYCLOAK_REALM} \
		--build-arg VITE_KEYCLOAK_CLIENT_ID=${VITE_KEYCLOAK_CLIENT_ID} \
		frontend/
	@echo "$(GREEN)Frontend image built: $(FRONTEND_IMAGE):$(IMAGE_TAG)$(NC)"

publish-images: ## Publish Docker images to registry
	@echo "$(BLUE)Publishing images...$(NC)"
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "$(YELLOW)Warning: DOCKER_REGISTRY is not set. Set it to publish images.$(NC)"; \
		exit 1; \
	fi
	docker push $(BACKEND_IMAGE):$(IMAGE_TAG)
	docker push $(BACKEND_IMAGE):latest
	docker push $(FRONTEND_IMAGE):$(IMAGE_TAG)
	docker push $(FRONTEND_IMAGE):latest
	@echo "$(GREEN)Images published successfully$(NC)"

# Database
migrate-up: ## Run database migrations up
	@echo "$(BLUE)Running migrations...$(NC)"
	docker-compose exec postgres psql -U git_analytics -d git_analytics -f /docker-entrypoint-initdb.d/000001_init_schema.up.sql
	@echo "$(GREEN)Migrations completed$(NC)"

migrate-down: ## Run database migrations down
	@echo "$(BLUE)Rolling back migrations...$(NC)"
	docker-compose exec postgres psql -U git_analytics -d git_analytics -f /docker-entrypoint-initdb.d/000001_init_schema.down.sql
	@echo "$(GREEN)Rollback completed$(NC)"

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "$(YELLOW)Usage: make migrate-create NAME=migration_name$(NC)"; \
		exit 1; \
	fi
	@TIMESTAMP=$$(date +%s); \
	echo "-- Migration: $(NAME)" > migrations/$${TIMESTAMP}_$(NAME).up.sql; \
	echo "-- Rollback: $(NAME)" > migrations/$${TIMESTAMP}_$(NAME).down.sql; \
	echo "$(GREEN)Created migrations/$${TIMESTAMP}_$(NAME).{up,down}.sql$(NC)"

db-shell: ## Open PostgreSQL shell
	docker-compose exec postgres psql -U git_analytics -d git_analytics

# Clean
clean: ## Clean build artifacts and containers
	@echo "$(BLUE)Cleaning...$(NC)"
	rm -rf bin/
	rm -rf frontend/dist/
	rm -f coverage.out coverage.html
	docker-compose down -v
	@echo "$(GREEN)Cleaned$(NC)"

clean-images: ## Remove Docker images
	@echo "$(BLUE)Removing Docker images...$(NC)"
	docker rmi $(BACKEND_IMAGE):$(IMAGE_TAG) $(BACKEND_IMAGE):latest || true
	docker rmi $(FRONTEND_IMAGE):$(IMAGE_TAG) $(FRONTEND_IMAGE):latest || true
	@echo "$(GREEN)Images removed$(NC)"

# Deploy
deploy: ## Deploy to production (customize as needed)
	@echo "$(BLUE)Deploying to production...$(NC)"
	@echo "$(YELLOW)Implement your deployment strategy here$(NC)"
	@echo "$(YELLOW)Examples:$(NC)"
	@echo "  - kubectl apply -f k8s/"
	@echo "  - docker stack deploy -c docker-stack.yml git-analytics"
	@echo "  - ansible-playbook deploy.yml"

# CI/CD Pipeline
ci: lint test build-images ## Run CI pipeline
	@echo "$(GREEN)CI pipeline completed successfully$(NC)"

# Install development tools
install-tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Development tools installed$(NC)"

# Quick start
quickstart: ## Quick start guide
	@echo "$(BLUE)=== Git Analytics Platform - Quick Start ===$(NC)"
	@echo ""
	@echo "$(YELLOW)1. Copy .env.example to .env and configure:$(NC)"
	@echo "   cp .env.example .env"
	@echo ""
	@echo "$(YELLOW)2. Start the development environment:$(NC)"
	@echo "   make dev"
	@echo ""
	@echo "$(YELLOW)3. In another terminal, run the backend:$(NC)"
	@echo "   go run main.go"
	@echo ""
	@echo "$(YELLOW)4. In another terminal, run the frontend:$(NC)"
	@echo "   cd frontend && bun dev"
	@echo ""
	@echo "$(YELLOW)5. Or start everything with Docker:$(NC)"
	@echo "   make up"
	@echo ""
	@echo "$(GREEN)For more commands, run: make help$(NC)"
