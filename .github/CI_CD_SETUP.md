# GitHub Actions CI/CD

## Workflows

- `CI`: runs `gofmt`, `go vet`, `go test`, and a Docker build check on every push and pull request.
- `CD`: on pushes to `main`, builds and publishes `ghcr.io/<owner>/<repo>` and deploys the production stack over SSH.

## Required GitHub Secrets

- `DEPLOY_HOST`: server hostname or IP
- `DEPLOY_USER`: SSH user for the server
- `DEPLOY_PATH`: absolute path on the server where the compose files should live
- `DEPLOY_SSH_KEY`: private SSH key for the deploy user
- `GHCR_USERNAME`: GitHub username that can pull the package from GHCR
- `GHCR_TOKEN`: GitHub personal access token with package read access

## First-Time Server Setup

1. Install Docker and the Docker Compose plugin on the server.
2. Ensure ports `80` and `443` are open.
3. Let the workflow create `.env` from `.env.example`, then update the server-side `.env` with the real `APP_DOMAIN` and `LETSENCRYPT_EMAIL`.
