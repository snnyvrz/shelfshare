#!/usr/bin/env sh
set -eu

# Resolve paths relative to this script, including when called from another directory.
ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$ROOT_DIR"

if [ ! -f .env.dev ]; then
    printf '%s\n' 'Missing .env.dev. Run: cp .env.dev.example .env.dev' >&2
    exit 1
fi

if ! command -v setsid >/dev/null 2>&1; then
    printf '%s\n' 'Missing setsid. Install util-linux to enable application process-group cleanup.' >&2
    exit 1
fi

child_pid=
child_group=
cleanup() {
    status=$?
    trap - EXIT
    # Let cleanup finish even if Ctrl+C is pressed again.
    trap '' INT TERM

    if [ -n "$child_pid" ]; then
        if [ -n "$child_group" ]; then
            kill -TERM "-$child_pid" 2>/dev/null || true
        else
            kill -TERM "$child_pid" 2>/dev/null || true
        fi
        wait "$child_pid" 2>/dev/null || true
    fi

    printf '%s\n' 'Stopping database containers (stored data is preserved)...'
    if ! bun run dev:stop; then
        if [ "$status" -eq 0 ]; then
            status=1
        fi
    fi
    exit "$status"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

printf '%s\n' 'Starting databases and waiting for readiness...'
docker compose --env-file .env.dev up -d --wait --wait-timeout 120 postgres mongo &
child_pid=$!
wait "$child_pid"
child_pid=

printf '%s\n' 'Starting APIs and frontend. Open http://localhost:5173 once ready.'
printf '%s\n' 'Press Ctrl+C to stop the applications and database containers.'
# Isolate the application process group so termination reaches all three services.
setsid bun x nx run-many -t serve -p books-service,auth-service,web --parallel=3 --outputStyle=stream &
child_pid=$!
child_group=1
wait "$child_pid"
