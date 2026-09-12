# ShelfShare

ShelfShare application services.

## Services

- `apps/books-service`: Go API backed by PostgreSQL
- `apps/auth-service`: Bun/TypeScript API backed by MongoDB

## Development

Install dependencies and start the local databases:

```sh
bun install
docker compose up -d
```

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
