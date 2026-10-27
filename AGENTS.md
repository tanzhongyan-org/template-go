# Go Runtime Agent Guidelines (L1 Extension)

Inherits: Universal L0 Invariants (workflow/AGENTS.md).
Do not repeat global workflow rules here.

## 1. Project Technology Stack
- Runtime: Go 1.23+ (`go.mod`)
- Project Layout: Standard Go layout (`cmd/server/main.go`, `internal/`)
- HTTP Router: Chi (`github.com/go-chi/chi/v5`) with standard middleware (`RequestID`, `RealIP`, `Logger`, `Recoverer`)
- Database: SQLC (`sqlc.yaml`, `sql/schema.sql`, `sql/queries.sql`) generating type-safe Go in `internal/db/` using `pgx/v5`
- Linters: `golangci-lint` (`.golangci.yml`), `go vet`, `go test -race`

## 2. Package Architecture & Layout
- Shallow Packages: Keep package hierarchy flat (`internal/config`, `internal/handler`, `internal/db`). Avoid deep Java-like hierarchies (`core/`, `api/handlers/`).
- Entry Point: `cmd/server/main.go` parses configuration, initializes dependencies, binds Chi router, and manages lifecycle via `signal.NotifyContext`.
- Dependency Injection: Pass initialized dependencies (config, database pool) as explicit parameters or struct fields. Never use global package singletons.
- Transport Handlers: Colocate route handlers and response contracts in `internal/handler/`. Do not create artificial "domain" packages for infrastructural probes.
- Database Layer: Schema migrations and queries stay in `sql/`. Code is generated into `internal/db/` via `make generate` (`sqlc generate`).

## 3. Go Idiomatic Invariants
- Configuration Boundary: Direct calls to `os.Getenv` outside `internal/config` are banned.
- Error Handling: Blank error ignores (`_ = err`) are forbidden. Wrap contextual errors with `%w` (`fmt.Errorf("...: %w", err)`).
- HTTP Method Routing: Rely on Chi route registrations (`r.Get`, `r.Post`); do not write defensive manual `r.Method != ...` checks inside handlers.
- Safe Graceful Shutdown: Handle process interruption using `signal.NotifyContext` and `server.Shutdown(ctx)` with an explicit drain timeout.
