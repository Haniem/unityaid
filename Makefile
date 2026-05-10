COMPOSE ?= docker compose
BACKEND_SERVICE ?= backend
FRONTEND_SERVICE ?= frontend
DB_SERVICE ?= db

.PHONY: help up down restart logs ps build migrate seed backend-shell frontend-shell db-shell

help:
	@echo "UnityAid infrastructure commands"
	@echo "  make up              Start development stack"
	@echo "  make down            Stop development stack"
	@echo "  make restart         Restart all services"
	@echo "  make logs            Follow service logs"
	@echo "  make ps              Show service status"
	@echo "  make build           Build development images"
	@echo "  make migrate         Apply backend migrations"
	@echo "  make seed            Apply backend seeds"
	@echo "  make backend-shell   Open backend shell"
	@echo "  make frontend-shell  Open frontend shell"
	@echo "  make db-shell        Open PostgreSQL shell"

up:
	$(COMPOSE) up --build

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) restart

logs:
	$(COMPOSE) logs -f

ps:
	$(COMPOSE) ps

build:
	$(COMPOSE) build

migrate:
	$(COMPOSE) run --rm $(BACKEND_SERVICE) go run ./cmd/migrate

seed:
	$(COMPOSE) run --rm $(BACKEND_SERVICE) go run ./cmd/seed

backend-shell:
	$(COMPOSE) exec $(BACKEND_SERVICE) sh

frontend-shell:
	$(COMPOSE) exec $(FRONTEND_SERVICE) sh

db-shell:
	$(COMPOSE) exec $(DB_SERVICE) psql -U $${POSTGRES_USER:-unityaid} -d $${POSTGRES_DB:-unityaid}
