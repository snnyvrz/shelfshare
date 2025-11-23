.ONESHELL:
SHELL := /bin/bash

.PHONY: dev infra-up infra-down logs

dev:
	set -e

	echo "Starting infra..."
	docker compose --env-file .env -f infra/docker-compose.infra.yml up -d

	cleanup() {
		echo ""
		echo "Stopping infra..."
		docker compose --env-file .env -f infra/docker-compose.infra.yml down
	}
	
	trap cleanup INT TERM EXIT

	echo "Waiting for infra to become healthy..."
	while true; do
		status=$$(docker inspect --format='{{.State.Health.Status}}' shelfshare-postgres 2>/dev/null || echo "starting")
		if [ "$$status" = "healthy" ]; then
			echo "Postgres is healthy ✅"
			break
		elif [ "$$status" = "unhealthy" ]; then
			echo "Postgres is UNHEALTHY ❌ — check logs with 'make logs'"
			exit 1
		else
			echo "Current status: $$status… waiting 1s"
			sleep 1
		fi
	done

	echo "Starting books-api via Nx (Ctrl+C to stop everything)..."

	bunx nx serve books-api || true

infra-up:
	docker compose --env-file .env -f infra/docker-compose.infra.yml up -d

infra-down:
	docker compose --env-file .env -f infra/docker-compose.infra.yml down

logs:
	docker compose --env-file .env -f infra/docker-compose.infra.yml logs -f