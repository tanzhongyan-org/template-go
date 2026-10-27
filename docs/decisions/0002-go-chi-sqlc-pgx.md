# 0002. Go with Chi, sqlc and pgx

## Context

The template needs a small HTTP service with typed database access, fast builds and few moving parts. The code must stay easy for humans and agents to read and check mechanically.

## Decision

We use Go, the Chi router, pgx v5 for Postgres, and sqlc to generate query code from `sql/schema.sql` and `sql/queries.sql`. Packages stay shallow: `cmd/server` and `internal/{config,db,handler}`. Handlers share one JSON response and error envelope. Tool binaries (sqlc, govulncheck, lefthook) are `tool` directives in `go.mod`.

## Consequences

- SQL is written by hand and checked at generation time; `sqlc diff` in CI catches drift.
- Chi is a small dependency on top of the standard library handlers, so handlers stay standard.
- Schema changes need a migration tool once real data exists; the server currently applies `CREATE ... IF NOT EXISTS` at start.
- Tool versions live in `go.mod`, where Renovate updates them.
