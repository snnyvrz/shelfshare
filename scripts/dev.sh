#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/constants.sh"

SERVICE="${1:-}"

case "$SERVICE" in
    books)
        echo "🚀 Starting books-service..."
        echo "Starting infra..."

        docker compose --env-file "$ENV_FILE" -f "$INFRA_COMPOSE_FILE" up -d "$POSTGRES_SERVICE"

        trap cleanup_infra INT TERM EXIT

        wait_for_postgres

        echo "Starting books-service via Nx (Ctrl+C to stop everything)..."
        cd apps/books-service
        bunx nx serve books-service || true
        ;;

    auth)
        echo "🔐 Starting auth-service..."
        echo "Starting infra..."

        docker compose --env-file "$ENV_FILE" -f "$INFRA_COMPOSE_FILE" up -d "$MONGO_SERVICE"

        trap cleanup_infra INT TERM EXIT

        wait_for_mongo

        echo "Starting auth-service via Nx (Ctrl+C to stop everything)..."
        cd apps/auth-service
        bun --watch src/index.ts || true
        ;;

    "")
        echo "Argument required to specify which service to run."
        echo "Usage: make dev [books|auth]"
        ;;

    *)
        echo "❌ Unknown service: $SERVICE"
        echo "Usage: make dev [books|auth]"
        exit 1
        ;;
esac
