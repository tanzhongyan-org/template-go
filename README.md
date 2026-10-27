# template-go

[![CI](https://github.com/tanzhongyan-org/template-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tanzhongyan-org/template-go/actions/workflows/ci.yml)
[![CodeQL](https://github.com/tanzhongyan-org/template-go/actions/workflows/codeql.yml/badge.svg)](https://github.com/tanzhongyan-org/template-go/actions/workflows/codeql.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A starting point for a Go HTTP service backed by Postgres. Use it when you want a Chi router, sqlc queries and a verification gate set that agents and humans run the same way.

## Quick start

You need Go (version in `go.mod`), Docker, Node 22+ and `make`.

```bash
gh repo create my-service --template tanzhongyan-org/template-go --private --clone
cd my-service
cp .env.example .env
make setup
make dev
```

In a second terminal:

```bash
curl -s localhost:8080/health
```

You should see `{"status":"healthy","appName":"template-go",...}`.

## What you get

- Chi router with request ID, logging, panic recovery and a 1 MiB request body limit.
- A notes example slice (`/notes`, `/api/v1/notes`) with per-owner isolation.
- Postgres access through sqlc-generated code in `internal/db`.
- One error envelope for every endpoint: `{"error":{"message":"..."}}`.
- Gates, security scans and agent hooks, all driven by `make`.

## After creating your repo

1. Run the **Initialize Template Repository** workflow from the Actions tab. It renames the module and opens a "Template follow-ups" issue.
2. List what you still have to change:

   ```bash
   make template-debt
   ```

   You should see one line per marker, such as the auth stub and the `CODEOWNERS` owner.

## Configuration

Copy `.env.example` to `.env`. The server reads these variables in `internal/config`.

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `APP_NAME` | no | `template-go` | Name shown in `/` and `/health`. |
| `APP_ENV` | no | `development` | Environment label shown in `/health`. |
| `PORT` | no | `8080` | HTTP listen port. |
| `DATABASE_URL` | no | local Postgres URL | Postgres connection string. |
| `TRUST_IDENTITY_HEADER` | no | `false` | When `true`, the `x-user-id` header identifies the caller. When `false`, `/notes` answers 401. |

> [!WARNING]
> `TRUST_IDENTITY_HEADER=true` lets any caller claim any identity. Use it for local work and tests only, then replace it with real authentication.

## Commands

| Command | What it does |
| --- | --- |
| `make setup` | Download modules and install the git hooks. |
| `make up` / `make down` | Start or stop Postgres in Docker. |
| `make dev` | Start Postgres and run the server. |
| `make build` | Build `bin/server`. |
| `make generate` | Regenerate `internal/db` from `sql/`. |
| `make test` | Run tests with the race detector (needs Postgres). |
| `make lint` / `make fix` | Run golangci-lint, optionally with fixes. |
| `make verify` | Fast gates: vet, build, lint, anti-slop, policy, Markdown lint. |
| `make check` | Full gates: `make verify`, sqlc drift check and tests. |
| `make audit` | Scan dependencies with govulncheck. |
| `make template-debt` | List template markers a derived repo must resolve. |

## Project layout

| Path | Contents |
| --- | --- |
| `cmd/server` | Entry point and router. |
| `internal` | Config, handlers and generated database code. |
| `sql` | Schema and queries for sqlc. |
| `docs` | Decisions, specs, plans and the dependency ledger. |
| `scripts` | Agent hook and repo policy checker. |

See [ARCHITECTURE.md](ARCHITECTURE.md) for the code map.

## More

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [LICENSE](LICENSE)
