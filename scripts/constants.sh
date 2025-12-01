#!/usr/bin/env bash

POSTGRES_CONTAINER="shelfshare-postgres"
POSTGRES_SERVICE="postgres"

MONGO_CONTAINER="shelfshare-mongo"
MONGO_SERVICE="mongo"

INFRA_COMPOSE_FILE="docker-compose.infra.yml"

DEV_ENV_FILE=".env.dev"
