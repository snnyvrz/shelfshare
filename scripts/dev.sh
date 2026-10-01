#!/usr/bin/env sh
set -eu

# Resolve paths relative to this script, including when called from another directory.
ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

if [ ! -f .env.dev ]; then
    printf '%s\n' 'Missing .env.dev. Run: cp .env.dev.example .env.dev' >&2
    exit 1
fi

printf '%s\n' 'Starting databases and waiting for readiness...'
docker compose --env-file .env.dev up -d --wait --wait-timeout 120 postgres mongo

printf '%s\n' 'Starting APIs and frontend. Open http://localhost:5173 once ready.'
printf '%s\n' 'Press Ctrl+C to stop the applications. Use bun run dev:stop to stop the databases.'
exec bun x nx run-many -t serve -p books-service,auth-service,web --parallel=3 --outputStyle=stream
