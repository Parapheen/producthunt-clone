.DEFAULT_GOAL := help

# Variables
APP_NAME := ph-clone
DB_PATH := data.db
MIGRATIONS_DIR := migrations
CSS_INPUT := assets/css/input.css
CSS_OUTPUT := assets/css/output.css
GO := go
TAILWINDCSS := ./tailwindcss

.PHONY: help migrate-up migrate-status migrate-down migrate-reset migrate-create run up build install deps clean lint fmt vet setup dev test coverage css-build css-watch css-dev

# Database commands (goose)
migrate-up:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up

migrate-status:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) status

migrate-down:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) down

migrate-reset:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) reset

# create migration: supports `make migrate-create NAME=add_users` or interactive prompt
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		read -p "Enter migration name: " name; \
	else \
		name="$(NAME)"; \
	fi; \
	if [ -z "$$name" ]; then \
		echo "Migration name required"; exit 1; \
	fi; \
	goose -dir $(MIGRATIONS_DIR) create $$name sql

# Application commands
run:
	$(GO) run cmd/main.go

# up: run migrations then app
up: migrate-up run

build:
	$(GO) build -o $(APP_NAME) cmd/main.go

deps:
	$(GO) mod tidy

dev: deps
	# запуск в режиме разработки (пример) — можно добавить переменные окружения
	APP_ENV=development $(GO) run cmd/main.go

# Utility commands
clean:
	$(GO) clean
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint: fmt vet


# CSS Build/Watch commands
css-build:
	$(TAILWINDCSS) -i $(CSS_INPUT) -o $(CSS_OUTPUT)

css-watch:
	$(TAILWINDCSS) -i $(CSS_INPUT) -o $(CSS_OUTPUT) --watch

css-dev: css-build css-watch
# Setup
# setup: install deps and run migrations (non-interactive)
setup: deps migrate-up
	@echo "Project setup complete!"
	@echo "Run 'make dev' to start development server"

help:
	@echo "Available commands:"
	@echo ""
	@echo "Database:"
	@echo "  migrate-up        - Run database migrations"
	@echo "  migrate-status    - Show migration status"
	@echo "  migrate-down      - Rollback one migration"
	@echo "  migrate-reset     - Rollback all migrations"
	@echo "  migrate-create    - Create new migration (NAME=... or interactively)"
	@echo ""
	@echo "Application:"
	@echo "  run               - Run the application"
	@echo "  up                - Run migrations and start app"
	@echo "  build             - Build the application"
	@echo ""
	@echo "Development:"
	@echo "  deps              - Download and tidy dependencies"
	@echo "  dev               - Run in development mode"
	@echo "  lint              - Run linter"
	@echo "  fmt               - Format code"
	@echo "  vet               - Run go vet"
	@echo "  test              - Run unit tests"
	@echo "  coverage          - Generate coverage report"
	@echo ""
	@echo "Utility:"
	@echo "  clean             - Clean build artifacts"
	@echo "  setup             - Initial project setup"
	@echo "  help              - Show this help message"
	@echo "Tailwind:"
	@echo "  css-dev           - Autoreload styles"

