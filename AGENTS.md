# Go Runtime Agent Guidelines (L1 Extension)

Inherits: Universal L0 Invariants (workflow/AGENTS.md).
Do not repeat global workflow rules here. Architecture and the code map live in [ARCHITECTURE.md](ARCHITECTURE.md).
This is the only agent instruction file. Claude Code, Codex, Cursor, Copilot and Gemini read it. Do not add CLAUDE.md, GEMINI.md or other per-tool copies: Claude Code ignores this file once a CLAUDE.md exists. `make policy` rejects them.

## 1. Project Technology Stack

- Runtime: Go (`go.mod`)
- Router: Chi (`github.com/go-chi/chi/v5`)
- Database: sqlc (`sqlc.yaml`, `sql/`) generating pgx/v5 code in `internal/db/`
- Linters: `golangci-lint` (`.golangci.yml`), `go vet`, `go test -race`
- Task runner: `Makefile`. Hooks, agent hooks and CI call `make` targets only.

## 2. Where things go

- New code goes where the code map in [ARCHITECTURE.md](ARCHITECTURE.md) says. Reuse the helpers listed there before writing new ones.
- Nest new code freely inside existing top-level directories. A new top-level directory needs a `layout.allow` entry in `scripts/repo-policy.json`; record the reason in `docs/decisions/` if it changes the architecture.
- Generic grab-bag packages (`utils`, `helpers`) are banned.

## 3. Go Idiomatic Invariants

- Configuration: `os.Getenv` and `os.LookupEnv` outside `internal/config` are banned [enforced by ast-grep].
- Tenant isolation: every query on `notes` filters by `owner_id` [review-only].
- Identity: `requireOwner` verifies the `Authorization` header against `AUTH_JWKS_URL` when it is set, else reads `x-user-id` only when trusted; it fails closed with 401 [enforced by tests].
- Responses: use `writeJSON`, `writeError` and `decodeJSON`; never `http.Error` with a JSON string [review-only].
- Errors: no blank error ignores; wrap with `%w` [enforced by errcheck, errorlint].
- Routing: rely on Chi registrations; no manual `r.Method` checks [review-only].
- Shutdown: `signal.NotifyContext` plus `server.Shutdown` with a drain timeout [review-only].

## 4. Verification Gates

Run `make check` before claiming done. Use `make verify` for a fast loop without a database.

1. Gate 1 (Typecheck): `go vet ./... && go build ./...`
2. Gate 2 (Lint and security): `golangci-lint run ./...` (includes gosec, bodyclose, errorlint, noctx, sqlclosecheck)
3. Gate 3 (Tests): `go test -race ./...`
4. Gate 4 (Dead code): `unused` in Gate 2
5. Gate 5 (Duplication): `dupl` in Gate 2, threshold 100
6. Gate 6 (Anti-slop): `make anti-slop` (ast-grep scan and test)

`make check` also runs `make policy` (layout, docs, dependency ledger, version drift), `sqlc diff` and the OpenAPI spec check. CI additionally runs `make licenses` and `make security`.
