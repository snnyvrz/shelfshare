# ShelfShare Agent Guide

## Repository

- This is a Bun-managed Nx workspace with three applications: `apps/books-service` is a Go/Gin API backed by PostgreSQL; `apps/auth-service` is a Node/NestJS TypeScript API backed by MongoDB; `apps/web` is a SvelteKit/Svelte 5 frontend using adapter-node.
- Run commands from the repository root unless a command explicitly sets a working directory.
- `./configure` checks or installs the expected local tools and installs JavaScript dependencies. The expected versions are Go `1.25.4+`, Bun `1.3.3+`, Air `1.63.0+`, Docker `28.4.0+`, shfmt `3.12.0+`, and SOPS `3.11.0+`.
- Start local databases with `docker compose --env-file .env.dev up -d`; PostgreSQL defaults to port `5432` and MongoDB to port `27017`.

## Product And Domain Boundaries

- ShelfShare connects readers through public profiles, owned physical books, borrowing and real-time private messages. Catalog `Book`/`Author` records describe titles; `lending.Copy` represents one owner's physical copy. Never infer ownership of existing catalog records or use email as a foreign key: the auth-service MongoDB user ID (`JWT sub`) is the stable cross-service identity.
- Public profiles expose only ID, display name, biography and general location. Account email/password data stay private. Profiles live in auth-service; copies, requests, conversations, messages and blocks live in books-service/PostgreSQL.
- `apps/books-service/internal/lending` owns domain operations, HTTP endpoints, shared schema migration and WebSocket delivery. `Migrate` upgrades the catalog additively, records custom indexes in `schema_migrations` and takes a PostgreSQL advisory migration lock. Production and integration tests must use it. Preserve restrictive foreign keys and the partial unique indexes for active occupancy/open borrower requests.
- Borrowing states: `pending → accepted → borrowed → return_pending → returned`, plus `declined`/`cancelled`. Only owners approve/decline, confirm handover and confirm receipt; borrowers initiate return. Both participants can cancel an accepted reservation before handover. Approval declines competing pending requests. Borrower return alone must not release availability. Archive retains history and is forbidden while occupied.
- Physical-copy mutations derive the owner from verified identity. Private reads require `auth.Required`, not `auth.PublicReads`, which deliberately skips GET authentication. Public copy queries must exclude hidden/archived copies. Catalog deletion must not remove referenced copies or history.
- Use copy-row locks for lending transitions and pair advisory locks for blocking/direct-message delivery; keep lock order consistent. Messages persist before broadcast and use client-generated UUIDs for idempotent retries.
- Direct conversations are unique per user pair. Their initiator may send one message until the recipient accepts. Borrowing conversations are separate and created automatically. Blocking prevents direct messaging and presence but keeps active loan coordination usable. Closed loan threads are read-only.

## Development Commands

- Install dependencies with `bun install` (use `bun install --frozen-lockfile` to match CI).
- Run the services with `bun x nx serve books-service` and `bun x nx serve auth-service`.
- The books-service `serve` target runs Swagger generation before starting Air; use `bun x nx run books-service:swagger` when only regenerating `apps/books-service/internal/docs`.
- Run books-service unit tests and coverage with `bun x nx run books-service:test` or `bun x nx run books-service:coverage`.
- Run a focused Go test from `apps/books-service` with `go test -run 'TestName/optional_subtest' ./path/to/package`; run all unit tests with `go test ./...`.
- Run integration tests with `bun x nx run books-service:integration-test` or, from `apps/books-service`, `go test -tags=integration ./...`. They require a reachable PostgreSQL instance and `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, and `TZ` environment variables; tests auto-migrate and truncate the test tables.
- The auth service runs on Node, requires `JWT_SECRET` at startup, and reads MongoDB settings from `MONGO_INITDB_ROOT_USERNAME`, `MONGO_INITDB_ROOT_PASSWORD`, `MONGO_HOST`, `MONGO_PORT`, and `MONGO_INITDB_DATABASE`.

## Verification And Hooks

- CI generates Swagger, runs books-service coverage, starts PostgreSQL, then runs tagged integration tests. Preserve that order when reproducing CI locally.
- JavaScript/TypeScript linting and formatting use ESLint and Prettier; Go uses `gofmt`; shell uses `shfmt`. Pre-commit runs these through `lint-staged` on changed files.
- Pre-push runs `bun x nx affected --target=test`, and commit messages must use Conventional Commits with a non-empty scope of `books-service`, `auth-service`, or `root`.
- Prettier uses semicolons, ES5 trailing commas, and a 120-column print width. `.editorconfig` uses four-space indentation except YAML, which uses two spaces.

## Configuration And Generated Files

- Serve the frontend with `bun x nx serve web`; run `bun x nx run-many -t check,lint,test,build -p web` for frontend verification. `bun x nx run web:e2e` requires running APIs/databases and a Playwright Chromium installation (or `PLAYWRIGHT_CHANNEL=chrome`).
- The frontend development server reads `JWT_SECRET`, `BOOKS_API_URL`, `BOOKS_WS_URL`, and `AUTH_API_URL` from root `.env.dev`. All three applications share `JWT_SECRET`; books/author mutations require bearer authentication. Production frontend deployments must set `ORIGIN` and supply the server environment.
- Bun applies the checked-in `@nestjs/mongoose` patch to fix its Node ESM connection import; preserve it when installing dependencies or upgrading that package.

- The books service loads database settings from the repository-root `.env.dev` when `GIN_MODE=debug`; `.env.dev` is ignored and must not be committed.
- Swagger outputs under `apps/books-service/internal/docs` are generated by the Nx `swagger` target and should be regenerated when Swagger annotations or API routes change.
- Swagger discovery includes `internal/lending`; document new lending and messaging HTTP endpoints there. The WebSocket protocol and full endpoint inventory are documented in README.md.
- The web frontend includes `/shelves`, `/shelves/:userId`, `/my-shelf`, `/account/profile`, `/requests` and `/messages`. Public book detail pages list physical copies. The API helper forwards bearer tokens to both APIs; never expose session tokens in page data or browser code.
- The browser requests a 30-second, single-use origin-bound WebSocket ticket via SvelteKit `/api/realtime-ticket`, then connects directly to books-service `/api/ws`. Connections expire with the JWT. Set server-side `BOOKS_WS_URL` to a browser-reachable `wss://.../api/ws` URL in production; REST `BOOKS_API_URL` can remain internal. Preserve origin validation, membership checks, heartbeat/read deadlines, bounded queues and reconnect recovery.
- WebSocket events can have null data. Guard event payloads before accessing fields. Community page loads declare `shelfshare:community`; background refresh must not interrupt enhanced-form redirects or reset catalog editor drafts. Browser tests wait for the layout's hydration marker before filling SSR forms.
- Presence/tickets/live delivery are in-process: run one books-service instance until shared event distribution and ticket/presence storage are implemented. Configure the reverse proxy for WebSocket upgrades and idle timeouts longer than the 20-second heartbeat.
- The two-account lending E2E test retains archived copies, loan history and test accounts. Use disposable E2E databases. Tagged integration tests truncate all lending/catalog tables; never point them at a development database containing user data. New PostgreSQL concurrency tests live in the integration package so they share its sequential reset lifecycle.
- Deployment builds and publishes only the books-service image. Deployment manifests and Flux automation live in the separate `snnyvrz/shelfshare-infra` repository.
