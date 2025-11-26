.ONESHELL:
SHELL := /bin/bash

.PHONY: books-dev books-test books-coverage books-integration-test books-infra-up infra-down logs

books-dev:
	@set -e

	echo "Starting infra..."
	docker compose --env-file .env -f docker-compose.infra.yml up -d postgres

	cleanup() {
		echo ""
		echo "Stopping infra..."
		docker compose --env-file .env -f docker-compose.infra.yml down
	}
	
	trap cleanup INT TERM EXIT

	echo "Waiting for infra to become healthy..."
	while true; do
		status=$$(docker inspect --format='{{.State.Health.Status}}' shelfshare-postgres 2>/dev/null || echo "starting")
		if [ "$$status" = "healthy" ]; then
			echo "Postgres is healthy"
			break
		elif [ "$$status" = "unhealthy" ]; then
			echo "Postgres is UNHEALTHY — check logs with 'make logs'"
			exit 1
		else
			echo "Current status: $$status… waiting 1s"
			sleep 1
		fi
	done

	echo "Starting books-api via Nx (Ctrl+C to stop everything)..."

	bunx nx serve books-api || true

books-test:
	bunx nx test books-api

books-coverage:
	bunx nx coverage books-api
	nohup xdg-open apps/books-api/coverage/coverage.html >/dev/null 2>&1 & echo "" || true

books-integration-test:
	@set -e

	echo "Starting infra..."
	docker compose --env-file .env.test -f docker-compose.infra.yml up -d postgres

	echo "Waiting for infra to become healthy..."
	while true; do
		status=$$(docker inspect --format='{{.State.Health.Status}}' shelfshare-postgres 2>/dev/null || echo "starting")
		if [ "$$status" = "healthy" ]; then
			echo "Postgres is healthy"
			break
		elif [ "$$status" = "unhealthy" ]; then
			echo "Postgres is UNHEALTHY — check logs with 'make logs'"
			exit 1
		else
			echo "Current status: $$status… waiting 1s"
			sleep 1
		fi
	done

	echo "Running books-api integration tests..."
	set -a; . .env.test; set +a;
	docker exec shelfshare-postgres psql -U $$DB_USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$$DB_NAME'" | grep -q 1 || docker exec shelfshare-postgres psql -U $$DB_USER -d postgres -c "CREATE DATABASE $$DB_NAME"
	cd apps/books-api && go test -tags=integration ./...
	echo "Books-api integration tests completed."
	echo "Stopping infra..."
	cd ../.. && docker compose --env-file .env.test -f docker-compose.infra.yml down

books-infra-up:
	docker compose --env-file .env -f docker-compose.infra.yml up -d postgres

infra-down:
	docker compose --env-file .env -f docker-compose.infra.yml down

logs:
	docker compose --env-file .env -f docker-compose.infra.yml logs -f

books-local:
	@set -e

	echo "Starting infra and local books-api..."

	cleanup() {
		echo ""
		echo "Stopping books-api and infra..."
		docker compose --env-file .env.prod \
			-f docker-compose.infra.yml \
			-f docker-compose.local.yml \
			down
	}
	trap cleanup INT TERM EXIT

	# Start both services (postgres + books-api) in one compose project
	docker compose --env-file .env.prod \
		-f docker-compose.infra.yml \
		-f docker-compose.local.yml \
		up --build
