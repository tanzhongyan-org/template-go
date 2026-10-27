# Go Runtime Template

Universal Go application boilerplate with standard project layout, Chi router, production middleware, SQLC data layer, agentic quality guardrails, and verification gates.

## Features

- **Runtime:** Go 1.27+ with standard project layout (`cmd/`, `internal/`)
- **Transport:** Chi router (`github.com/go-chi/chi/v5`) with production middleware stack (`RequestID`, `Logger`, `Recoverer`)
- **Database:** Type-safe SQL generation using SQLC (`sql/schema.sql`, `sql/queries.sql`, `sqlc.yaml` -> `internal/db`)
- **Build & Task Automation:** Universal `Makefile` with `setup`, `dev`, `test`, `lint`, `fix`, `typecheck`, and `check` targets
- **Code Quality:** `.golangci.yml` standard linter configuration (`errcheck`, `govet`, `staticcheck`, `unused`)
- **Config & Validation:** Validated environment configuration module
- **Agent Guardrails:**
  - `lefthook.yml`: Sub-second local Git pre-commit hooks
  - `sgconfig.yml` + `.ast-grep/`: Automatic detection of placeholder stubs and unfulfilled code
  - `.github/workflows/ci.yml`: Automated CI running the 6 Universal Verification Gates via `make check`
  - `AGENTS.md`: Runtime-specific agent invariants

## Getting Started

### 1. Create a New Repository
Click **"Use this template"** on GitHub or use the GitHub CLI:
```bash
gh repo create my-go-service --template tanzhongyan-org/template-go --private --clone
cd my-go-service
```
*(On first push to `main`, GitHub Actions automatically renames all template placeholders to your new repository name and self-deletes the initialization workflow).*

### 2. Local Setup
```bash
# Download dependencies
make setup
# or: go mod download

# Run server with Chi router and graceful shutdown
make dev
# or: go run ./cmd/server/main.go
```

### 3. Database & SQLC
Database schema and SQL queries are located under `sql/`:
- `sql/schema.sql`: Table definitions
- `sql/queries.sql`: Named queries for SQLC type generation
- `sqlc.yaml`: Configuration mapping queries and schema to `internal/db`

Generate type-safe Go query code:
```bash
make generate
# or: sqlc generate
```

### 4. Run Verification Gates
```bash
# Run universal check
make check
```
