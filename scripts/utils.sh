#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

source "$SCRIPT_DIR/constants.sh"

wait_for_postgres() {
    echo "Waiting for Postgres to become healthy..."
    while true; do
        status=$(docker inspect --format='{{.State.Health.Status}}' "$POSTGRES_CONTAINER" 2>/dev/null || echo "starting")

        if [ "$status" = "healthy" ]; then
            echo "Postgres is healthy"
            break
        elif [ "$status" = "unhealthy" ]; then
            echo "Postgres is UNHEALTHY — check logs with 'make logs'"
            exit 1
        else
            echo "Current status: $status… waiting 1s"
            sleep 1
        fi
    done
}

wait_for_mongo() {
    echo "Waiting for Mongo to become healthy..."

    while true; do
        status=$(docker inspect --format='{{.State.Health.Status}}' "$MONGO_CONTAINER" 2>/dev/null || echo "starting")

        if [ "$status" = "healthy" ]; then
            echo "Mongo is healthy"
            break
        fi

        if [ "$status" = "unhealthy" ]; then
            echo "Mongo is UNHEALTHY — check logs with 'make logs'"
            exit 1
        fi

        if [ "$status" = "starting" ] || [ "$status" = "" ]; then
            if docker exec "$MONGO_CONTAINER" mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
                echo "Mongo is reachable"
                break
            fi
        fi

        echo "Current Mongo status: $status… waiting 1s"
        sleep 1
    done
}

cleanup_infra() {
    echo
    echo "Stopping infra..."
    cd "$PROJECT_ROOT"
    docker compose --env-file "$DEV_ENV_FILE" -f "$INFRA_COMPOSE_FILE" down
}
