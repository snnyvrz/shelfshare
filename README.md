# ShelfShare

ShelfShare application services.

## Services

- `apps/books-service`: Go API backed by PostgreSQL
- `apps/auth-service`: NestJS/TypeScript API backed by MongoDB
- `apps/web`: SvelteKit/Svelte 5 frontend with a Node server

## Development

Install dependencies and start the local databases:

```sh
bun install
cp .env.dev.example .env.dev
docker compose --env-file .env.dev up -d
```

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
database containers. `docker compose --env-file .env.dev up -d` is required
because Compose does not automatically load `.env.dev`.

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
The frontend provides a public book catalog with search, sorting, author/date
filters and pagination; an author directory; detail pages; and authenticated
create, edit and delete forms. Register an account to contribute. Any registered
user can manage the shared collection.

The web development server loads `JWT_SECRET`, `BOOKS_API_URL`, and `AUTH_API_URL`
from the root `.env.dev`. API URLs default to `http://localhost:8080/api` and
`http://localhost:3030/auth`. These are server-side settings, so the browser
does not need direct API access or CORS configuration.

**All three services must use the same `JWT_SECRET`.** The books API now requires
a valid auth-service bearer token for POST, PATCH and DELETE requests. Browsing
remains public. SvelteKit verifies sessions and keeps tokens in HttpOnly,
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

The browser test starts the frontend automatically and covers registration,
login/logout, book and author CRUD, filters, pagination, mobile layouts, and
unauthenticated API rejection. It removes its catalog records afterward and
creates a uniquely named test account. Set `PLAYWRIGHT_CHANNEL=chrome` to use
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

## Images

The `deploy.yml` workflow publishes the books service image to:

```text
ghcr.io/snnyvrz/shelfshare-books-service
```

Deployment manifests and Flux image automation live in the separate
`snnyvrz/shelfshare-infra` repository.
