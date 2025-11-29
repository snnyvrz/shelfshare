.ONESHELL:
SHELL := /bin/bash

.DEFAULT_GOAL := check_configured

CONFIG_STAMP := .configured

.PHONY: check_configured
check_configured:
	@if [ ! -f "$(CONFIG_STAMP)" ]; then \
		echo "Please run the configure script before using any make commands:"; \
		echo ""; \
		echo "    chmod +x ./configure"; \
		echo "    ./configure"; \
		echo ""; \
		exit 1; \
	fi
	echo "Environment ready. Use a specific make command"

ENV_DEV  := .env
ENV_TEST := .env.test
ENV_LOCALPROD := .env.localprod

INFRA_COMPOSE    := -f docker-compose.infra.yml
POSTGRES_SERVICE := postgres
POSTGRES_CONTAINER := shelfshare-postgres

DC = docker compose --env-file $(ENV_FILE) $(INFRA_COMPOSE)

.PHONY: \
	books-dev books-test books-coverage books-swagger books-integration-test \
	books-infra-up infra-down logs books-localprod

define wait_for_postgres
	@echo "Waiting for Postgres to become healthy..."
	@while true; do \
		status=$$(docker inspect --format='{{.State.Health.Status}}' $(POSTGRES_CONTAINER) 2>/dev/null || echo "starting"); \
		if [ "$$status" = "healthy" ]; then \
			echo "Postgres is healthy"; \
			break; \
		elif [ "$$status" = "unhealthy" ]; then \
			echo "Postgres is UNHEALTHY — check logs with 'make logs'"; \
			exit 1; \
		else \
			echo "Current status: $$status… waiting 1s"; \
			sleep 1; \
		fi; \
	done
endef

books-dev: ENV_FILE=$(ENV_DEV)
books-dev:
	@set -e
	echo "Starting infra..."
	$(DC) up -d $(POSTGRES_SERVICE)

	cleanup() {
		echo ""
		echo "Stopping infra..."
		$(DC) down
	}
	trap cleanup INT TERM EXIT

	$(call wait_for_postgres)

	echo "Starting books-service via Nx (Ctrl+C to stop everything)..."
	bunx nx serve books-service || true

books-test:
	bunx nx test books-service

books-coverage:
	bunx nx coverage books-service
	nohup xdg-open apps/books-service/coverage/coverage.html >/dev/null 2>&1 & echo "" || true

books-swagger:
	bunx nx swagger books-service

books-integration-test: ENV_FILE=$(ENV_TEST)
books-integration-test:
	@set -e
	echo "Starting infra..."
	$(DC) up -d $(POSTGRES_SERVICE)

	$(call wait_for_postgres)

	echo "Running books-service integration tests..."
	set -a; . $(ENV_TEST); set +a;

	docker exec $(POSTGRES_CONTAINER) psql -U $$POSTGRES_USER -d postgres \
		-tc "SELECT 1 FROM pg_database WHERE datname = '$$POSTGRES_DB'" | grep -q 1 || \
	docker exec $(POSTGRES_CONTAINER) psql -U $$POSTGRES_USER -d postgres \
		-c "CREATE DATABASE $$POSTGRES_DB"

	bunx nx integration-test books-service

	echo "books-service integration tests completed."
	echo "Stopping infra..."
	$(DC) down

books-infra-up: ENV_FILE=$(ENV_DEV)
books-infra-up:
	$(DC) up -d $(POSTGRES_SERVICE)

infra-down: ENV_FILE=$(ENV_DEV)
infra-down:
	$(DC) down

logs: ENV_FILE=$(ENV_DEV)
logs:
	$(DC) logs -f

books-localprod:
	@set -e
	echo "Starting localprod stack (Postgres + books-api)..."

	cleanup() {
		echo ""
		echo "Stopping localprod stack..."
		docker compose --env-file $(ENV_LOCALPROD) \
			-f docker-compose.localprod.yml \
			down
	}
	trap cleanup INT TERM EXIT

	docker compose --env-file $(ENV_LOCALPROD) \
		-f docker-compose.localprod.yml \
		up --build
