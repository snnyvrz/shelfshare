# ShelfShare

ShelfShare application services.

## Services

- `apps/books-service`: Go API backed by PostgreSQL
- `apps/auth-service`: Bun/TypeScript API backed by MongoDB

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
| --- | --- |
| `POSTGRES_USER` | `DB_USER` |
| `POSTGRES_PASSWORD` | `DB_PASS` |
| `POSTGRES_DB` | `DB_NAME` |

Compose uses the `POSTGRES_*` and `MONGO_*` variables when initializing the
database containers. `docker compose --env-file .env.dev up -d` is required
because Compose does not automatically load `.env.dev`.

The books service runs with `GIN_MODE=debug` during local development and
loads `.env.dev` from the repository root using `godotenv`. It then reads the
`DB_*` variables. Set `DB_HOST=localhost` when it runs on the host, or
`DB_HOST=postgres` when it runs inside the same Compose network.

The auth service reads MongoDB settings and `JWT_SECRET` directly from
`process.env`. Its Nx development target passes the repository-root `.env.dev`
to Bun explicitly. When running it directly from `apps/auth-service`, provide
that file explicitly as well:

```sh
bun --env-file=../../.env.dev --watch src/index.ts
```

The equivalent Nx command is:

```sh
bun x nx serve auth-service
```

Use `MONGO_HOST=localhost` for a host-run auth service, or `MONGO_HOST=mongo`
inside the Compose network. The `JWT_SECRET` and database passwords shown in
the example are only for local development and must be replaced in production.

Run a service through Nx:

```sh
bun x nx serve books-service
bun x nx serve auth-service
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
