# Variables
APP_NAME := biolynq
SERVER_ENTRY := cmd/server/main.go
WORKER_ENTRY := cmd/worker/main.go

.PHONY: help run-server run-worker build test docker-up docker-down migration-gen migration-apply migration-status

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Development Targets:"
	@echo "  run-server       Run the API server locally"
	@echo "  run-worker       Run the background task worker locally"
	@echo "  build            Build server and worker binaries"
	@echo "  test             Run all Go tests"
	@echo ""
	@echo "Docker Targets:"
	@echo "  docker-up        Start all Docker compose services"
	@echo "  docker-down      Stop all Docker compose services"
	@echo ""
	@echo "Migration Targets (requires Atlas CLI):"
	@echo "  migration-gen    Generate a new migration. Usage: make migration-gen name=<migration_name>"
	@echo "  migration-apply  Apply pending migrations to database"
	@echo "  migration-status Check migration status"

run-server:
	go run $(SERVER_ENTRY)

run-worker:
	go run $(WORKER_ENTRY)

build:
	go build -o bin/server $(SERVER_ENTRY)
	go build -o bin/worker $(WORKER_ENTRY)

test:
	go test -v ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down

migration-gen:
	@if [ -z "$(name)" ]; then \
		echo "Error: 'name' variable is required. Example: make migration-gen name=your_migration_name"; \
		exit 1; \
	fi
	atlas migrate diff $(name) --env local

migration-apply:
	atlas migrate apply --env local

migration-status:
	atlas migrate status --env local
