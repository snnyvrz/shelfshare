# ShelfShare

Share the physical books on your shelf, borrow from other readers, and arrange
exchanges through real-time private messages.

## How it works

1. Register with a public display name and complete your profile.
2. Add your physical copies in **My shelf**. Choose an existing catalog title,
   or add missing book/author details first. Record condition and edition notes,
   then choose whether the copy is public and offered for lending.
3. Browse **Readers' shelves** or the available copies on a catalog book page.
   Request a particular copy and introduce yourself to its owner.
4. The owner approves or declines, confirms handover, and optionally records an
   agreed due date. The borrower initiates return; the owner confirms receipt.
5. Use **Messages** to arrange the exchange. You can also start a direct
   conversation from a public profile without requesting a book.

Profiles and visible shelves are public; requesting, lending and messaging
require an account. Email stays private. Public profile fields include display
name, biography and an optional general location.

Catalog titles describe a work; **physical copies belong to a user**. Multiple
readers can own the same title, and one reader can list several copies. Existing
catalog records remain usable and are not automatically assigned an owner.

### Borrowing and messaging rules

Requests follow `pending → accepted → borrowed → return_pending → returned`.
Owners can decline pending requests. Borrowers can cancel pending requests;
either participant can cancel an accepted reservation before handover.
Approval reserves one copy and declines competing pending requests atomically.
A borrower reporting return does not release the copy until its owner confirms
receipt. Overdue labels are derived from the due date.

Direct conversations start as message requests: the initiator may send one
introduction, and the recipient accepts or declines before further messages.
Borrowing requests automatically create a separate exchange conversation.
The unified inbox includes message previews, unread counts, read receipts,
participant-only online presence and typing indicators. Blocking stops direct
messages and presence sharing; active loan conversations remain usable for
return arrangements. Closed loan conversations remain readable but cannot
receive new messages.

Copies can be hidden or archived when they are not reserved/on loan. Archiving
retains borrowing and message history. Shared catalog metadata is still editable
by members, but referenced titles/authors cannot be deleted.

## Services

- `apps/books-service`: Go/Gin API, PostgreSQL catalog, owned copies, borrowing,
  conversations/messages and WebSocket delivery
- `apps/auth-service`: NestJS/TypeScript API, MongoDB accounts/public profiles,
  password authentication and JWT identity
- `apps/web`: SvelteKit/Svelte 5 frontend with a Node server, server-side session
  handling and a browser WebSocket client

## Development

First-time setup (use `./configure` to check/install the required local tools):

```sh
bun install
cp .env.dev.example .env.dev
```

Start the full development stack in one terminal:

```sh
bun run dev
```

This starts PostgreSQL and MongoDB, waits for both databases to be healthy, and
runs books-service, auth-service, and web together through Nx with labeled logs
and hot reload. Open **http://localhost:5173** once the applications are ready.
The books-service Swagger generation still runs before its API starts.

Press **Ctrl+C** to stop the applications. The databases stay running for quick
restarts. To stop the database containers as well, retaining their stored data:

```sh
bun run dev:stop
```

The launcher is in `scripts/dev.sh`. To start only the databases:

```sh
docker compose --env-file .env.dev up -d --wait
```

### Behavior-driven development (BDD)

Executable Gherkin specifications for requests, reservations, handover, returns,
cancellation and archiving live in
`apps/books-service/internal/integration/features/`. Godog runs them through
`TestBDD` in the existing PostgreSQL integration package. Lending steps call the
production service; archive and shelf steps use in-process Gin routes with real
JWT authentication. No running APIs, MongoDB or browser are required.

Use a **disposable PostgreSQL database**: each scenario truncates lending and
catalog tables. Do not point this suite at a database containing development data.
For example, start an isolated container (no persistent volume):

```sh
docker run --detach --rm --name shelfshare-bdd-postgres \
    -e POSTGRES_USER=shelfshare -e POSTGRES_PASSWORD=shelfshare \
    -e POSTGRES_DB=shelfshare_bdd -p 127.0.0.1:55432:5432 postgres:18
```

Once PostgreSQL is ready, run from the repository root:

```sh
POSTGRES_HOST=localhost POSTGRES_PORT=55432 POSTGRES_USER=shelfshare \
    POSTGRES_PASSWORD=shelfshare POSTGRES_DB=shelfshare_bdd TZ=UTC \
    bun x nx run books-service:bdd
```

Use the same environment with `bun x nx run books-service:integration-test` to
run all integration tests, including BDD. Both targets bypass caching and the
scenarios run sequentially with fresh state. The suite uses `lending.Migrate`;
undefined or pending steps fail the run. CI executes BDD once as part of its
PostgreSQL integration stage, after Swagger generation and unit coverage.
Stop the disposable database with `docker stop shelfshare-bdd-postgres`.

For a new behavior, agree on the rule and concrete examples first, write a small
scenario in reader-facing language, then implement its steps and behavior.
Review scenarios alongside code changes. Use Scenario Outlines for permission
and state variations; keep fixtures and HTTP details in step helpers. Fast unit
tests, PostgreSQL concurrency tests and Playwright user journeys remain useful
at their respective layers; feature files describe the agreed acceptance rules.

### Configuration

`.env.dev.example` is a committed template containing safe development-only
values. `.env.dev` is ignored and must not contain production credentials.
Production supplies its own database endpoints and secrets.

The same file contains variables for Docker Compose and the applications, but
the PostgreSQL names are intentionally different:

| Database container setting | books-service application setting |
| -------------------------- | --------------------------------- |
| `POSTGRES_USER`            | `DB_USER`                         |
| `POSTGRES_PASSWORD`        | `DB_PASS`                         |
| `POSTGRES_DB`              | `DB_NAME`                         |

Compose uses the `POSTGRES_*` and `MONGO_*` variables when initializing the
database containers. The launcher passes `--env-file .env.dev` to Compose
because Compose does not automatically load `.env.dev`. Pass the same flag
when running Compose manually.

The books service runs with `GIN_MODE=debug` during local development and
loads `.env.dev` from the repository root using `godotenv`. It then reads the
`DB_*` variables. Set `DB_HOST=localhost` when it runs on the host, or
`DB_HOST=postgres` when it runs inside the same Compose network.

The auth service reads MongoDB settings and `JWT_SECRET` from its environment.
Its Nx development target loads the repository-root `.env.dev` for Node. When
running it directly from `apps/auth-service`, provide that file explicitly:

```sh
node --env-file=../../.env.dev --import tsx --watch src/main.ts
```

The equivalent Nx command is:

```sh
bun x nx serve auth-service
bun x nx serve web
```

### Frontend

Open **http://localhost:5173** after starting the databases and both APIs.
The frontend provides public reader shelves (`/shelves`, `/shelves/:userId`),
the searchable book catalog and author directory, personal copy management
(`/my-shelf`), profile settings (`/account/profile`), requests/loans (`/requests`)
and a real-time inbox (`/messages`).

The web development server loads `JWT_SECRET`, `BOOKS_API_URL`, `BOOKS_WS_URL`, and `AUTH_API_URL`
from the root `.env.dev`. API URLs default to `http://localhost:8080/api` and
`http://localhost:3030/api/auth`. REST calls go through the SvelteKit server.
The browser connects directly to the books-service WebSocket endpoint; set
`BOOKS_WS_URL` to its browser-reachable URL when it differs from `BOOKS_API_URL`.

**All three services must use the same `JWT_SECRET`.** The books API now requires
a valid auth-service bearer token for mutations and all private reads. Catalog,
public profile and visible-copy reads remain public. SvelteKit verifies sessions and keeps tokens in HttpOnly,
SameSite cookies; tokens expire with the auth service's `JWT_EXPIRES_IN` setting.

Run frontend checks:

```sh
bun x nx run-many -t check,lint,test,build -p web
```

With both APIs and databases running, exercise the browser journey:

```sh
bun x playwright install chromium
bun x nx run web:e2e
```

The browser tests start the frontend automatically and cover registration,
login/logout, catalog CRUD/filtering/pagination, mobile layouts, unauthenticated
API rejection, and a two-account real-time messaging and complete lending
journey. Run against disposable databases: the catalog test removes its records,
while lending history and uniquely named test accounts are retained. Set `PLAYWRIGHT_CHANNEL=chrome` to use
an existing Chrome installation instead of downloading Chromium.

The production build is written to `apps/web/build`. Provide the API URLs and
shared secret in the server environment, and set `ORIGIN` to the frontend's
public origin:

```sh
ORIGIN=https://shelfshare.example PORT=3000 node apps/web/build/index.js
```

The frontend uses responsive typographic book covers because the books API
does not yet include images. Fonts use Google Fonts with local serif/sans-serif
fallbacks.

`patches/@nestjs%2Fmongoose@12.0.0.patch` fixes the dependency's ESM Mongoose
import so it can open a MongoDB connection on Node. Bun applies it automatically
during installation.

Use `MONGO_HOST=localhost` for a host-run auth service, or `MONGO_HOST=mongo`
inside the Compose network. The `JWT_SECRET` and database passwords shown in
the example are only for local development and must be replaced in production.
Existing Bun-generated Argon2id password hashes remain usable after migration.
The Nest app binds to port `3030` by default; set `PORT` to override it.

Run a service through Nx:

```sh
bun x nx serve books-service
bun x nx serve auth-service
```

Build and test the auth service:

```sh
bun x nx run auth-service:build
bun x nx run auth-service:test
```

Run the books service tests:

```sh
bun x nx run books-service:test
bun x nx run books-service:integration-test
```

### Schema upgrades and integration databases

Books-service startup calls `internal/lending.Migrate`, also used by integration
tests. It upgrades the existing catalog additively using GORM, with a versioned
`schema_migrations` ledger for custom SQL indexes. Schema upgrades are atomic
and serialized with a PostgreSQL advisory lock. Partial unique indexes prevent
two active loans/reservations for one copy and duplicate open borrower requests.
Foreign keys preserve copy and loan history. No ownership backfill is performed.

Follow CI order: generate Swagger, run coverage, start PostgreSQL, then run
tagged integration tests. Supply `POSTGRES_HOST`, `POSTGRES_PORT`,
`POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, and `TZ`. Use a dedicated
test database: the integration suite migrates and truncates its domain tables.
It includes simultaneous approval checks against real PostgreSQL.

### API overview

Auth-service routes are under `/api/auth`:

- `POST /register`, `POST /login`
- `GET /profiles?q=...&page=...`, `GET /profiles/:userId` — public reader
  discovery by display name/general location, deliberately limited projection
- `GET /me`, `PATCH /me` — authenticated profile access

Books-service routes are under `/api` (full generated docs at `/swagger/index.html`):

- `/books`, `/authors` — shared catalog
- `GET /copies?ownerId=...&bookId=...&page=...`, `GET /me/copies`
- `POST /copies`, `PATCH /copies/:id`, `DELETE /copies/:id` (archive)
- `POST /copies/:id/requests`, `GET /requests`
- `POST /requests/:id/:action` — `accept`, `decline`, `cancel`, `handover`,
  `return`, `confirm-return`; optional RFC3339 `dueAt` on approval/handover
- `GET/POST /conversations`, `GET /conversations/:id`,
  `GET /request-conversations/:requestId`
- `GET/POST /conversations/:id/messages`, `POST /conversations/:id/:action`
  (`accept`, `decline`, `read`)
- `GET /blocks`, `POST/DELETE /blocks/:userId`
- `POST /ws-ticket`, `GET /ws?ticket=...` (WebSocket upgrade)

Shelf, request and conversation lists use 50-item pages. Message history returns
the newest 50 messages; use `before` (RFC3339 timestamp) and `beforeId` as a
compound cursor for older messages. All private resources are participant- or
owner-scoped; clients cannot supply ownership or borrowing status directly.

### WebSocket configuration and protocol

The browser requests a connection ticket from SvelteKit's
`POST /api/realtime-ticket`. SvelteKit forwards its server-held JWT to
books-service with the browser origin. Tickets expire after 30 seconds, are
single-use, and are bound to that exact origin. A connection expires with the
JWT; the session JWT is never exposed to browser JavaScript.

Frames are JSON `{ "event": "...", "data": ... }` from the server. Clients send:

```json
{ "event": "message", "conversationId": "uuid", "id": "client-generated-uuid", "body": "Hello!" }
```

Clients can also send `typing` or `presence` with a `conversationId`. The server
checks membership, acceptance and blocking rules. Events include `message`,
`ack`, `error`, `presence`, `typing`, `read`, `accept`, `decline`, `blocked`,
`conversations` and `requests`; invalidation events can have null data.
Messages commit before broadcast; a repeated UUID/body is acknowledged without
creating a duplicate. The client reconnects with backoff, fetches missed history,
and retains an unacknowledged message for an explicit safe retry. Heartbeats,
bounded payloads/queues and per-connection event limits bound resource use.

For production, set `BOOKS_WS_URL=wss://books.example/api/ws`. Route WebSocket
upgrade requests to books-service, preserve `Origin`, support HTTP/1.1 upgrade
headers and use a proxy idle timeout longer than the 20-second heartbeat.
The web server still needs its own `ORIGIN`. Local development defaults to
`ws://localhost:8080/api/ws`.

Live delivery, presence and connection tickets currently use an in-process hub:
deploy **one books-service instance**. Horizontal scaling requires shared
ticket/presence storage and a cross-instance event bus; persistent message and
borrowing data already live in PostgreSQL. API and frontend deployments need to
roll out together; this repository's image automation currently publishes only
books-service.

## Images

The `deploy.yml` workflow publishes the books service image to:

```text
ghcr.io/snnyvrz/shelfshare-books-service
```

Deployment manifests and Flux image automation live in the separate
`snnyvrz/shelfshare-infra` repository.
