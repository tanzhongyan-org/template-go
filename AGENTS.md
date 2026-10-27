# Go Runtime Agent Guidelines (L1 Extension)

Inherits: Universal L0 Invariants (workflow/AGENTS.md).
Do not repeat global workflow rules here.

## 1. Project Technology Stack
- Runtime: Go 1.27+ (`go.mod`)
- Project Layout: Standard Go layout (`cmd/server/main.go`, `internal/`)
- HTTP Router: Chi (`github.com/go-chi/chi/v5`) with standard middleware (`RequestID`, `Logger`, `Recoverer`)
- Database: SQLC (`sqlc.yaml`, `sql/schema.sql`, `sql/queries.sql`) generating type-safe Go in `internal/db/` using `pgx/v5`
- Linters: `golangci-lint` (`.golangci.yml`), `go vet`, `go test -race`
- Task Runner: Universal Makefile (`setup`, `dev`, `test`, `lint`, `fix`, `typecheck`, `check`, `audit`)

## 2. Package Architecture & Layout
- Shallow Packages: Keep package hierarchy flat (`internal/config`, `internal/handler`, `internal/db`). Avoid deep Java-like hierarchies (`core/`, `api/handlers/`).
- Entry Point: `cmd/server/main.go` parses configuration, initializes dependencies, binds Chi router, and manages lifecycle via `signal.NotifyContext`.
- Dependency Injection: Pass initialized dependencies (config, database pool) as explicit parameters or struct fields. Never use global package singletons.
- Transport Handlers: Colocate route handlers and response contracts in `internal/handler/`. Do not create artificial "domain" packages for infrastructural probes.
- Database Layer: Schema migrations and queries stay in `sql/`. Code is generated into `internal/db/` via `make generate` (`sqlc generate`).

## 3. Go Idiomatic Invariants
- Configuration Boundary: Direct calls to `os.Getenv` or `os.LookupEnv` outside `internal/config` are banned [enforced by ast-grep].
- Tenant Isolation: Every read and write query on `notes` must filter by `owner_id` [review-only].
- Identity Boundary: Extract `owner_id` from the `x-user-id` header at the HTTP boundary [review-only].
- Client IP Handling: Client IP comes from `r.RemoteAddr`; trust proxy headers only behind a known reverse proxy [review-only].
- Error Handling: Blank error ignores (`_ = err`) are forbidden. Wrap contextual errors with `%w` (`fmt.Errorf("...: %w", err)`) [enforced by errcheck].
- HTTP Method Routing: Rely on Chi route registrations (`r.Get`, `r.Post`); do not write defensive manual `r.Method != ...` checks inside handlers [review-only].
- Safe Graceful Shutdown: Handle process interruption using `signal.NotifyContext` and `server.Shutdown(ctx)` with an explicit drain timeout [review-only].

## 4. Verification Gates
Run `make check` to verify task completion:
1. Gate 1 (Compile & Vet): `go vet ./... && go build ./...`
2. Gate 2 (Lint & Static Analysis): `golangci-lint run ./...`
3. Gate 3 (Tests with Race Detector): `go test -v -race ./...`
4. Gate 4 (Dead Code): Covered by `unused` in Gate 2 (golangci-lint)
5. Gate 5 (Duplication): N/A (TypeScript/Web specific)
6. Gate 6 (Anti-Slop): `npx --yes --package @ast-grep/cli@0.51.0 ast-grep scan && ast-grep test`