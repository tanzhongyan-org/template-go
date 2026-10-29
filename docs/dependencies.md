# Dependency ledger

Every direct module and tool in `go.mod` has a row. `make policy` fails when a row is missing or stale.

| Dependency | Purpose | Why the default was not enough | What it replaced |
| --- | --- | --- | --- |
| `github.com/go-chi/chi/v5` | HTTP router and middleware. | The standard `ServeMux` lacks middleware chaining and route mounting in the shape the handlers use. | Hand-written routing and middleware wrappers. |
| `github.com/google/uuid` | Note identifiers. | The standard library has no UUID generator. | Nothing. |
| `github.com/jackc/pgx/v5` | Postgres driver and connection pool. | database/sql has no Postgres-native types or pool tuning; sqlc generates pgx code. | lib/pq with database/sql. |
| `github.com/golang-jwt/jwt/v5` | Parses and verifies product JWTs in `internal/auth`. | The standard library has no JWT support; the JWKS key handling above it is hand-written. | Nothing. |
| `gopkg.in/yaml.v3` | Reads `api/openapi.yaml` in the route-coverage test. | The standard library has no YAML parser. | Nothing. |
| `github.com/google/go-licenses/v2` | Reports dependency licenses for `make licenses` (go tool). | `go list` does not classify licenses. | Nothing. |
| `github.com/evilmartians/lefthook/v2` | Git hooks (go tool). | Git has no hook installer shared across clones. | `npx lefthook` in the Makefile. |
| `github.com/sqlc-dev/sqlc/cmd/sqlc` | Generates `internal/db` from SQL (go tool). | Hand-written scanning code drifts from the schema. | `sqlc@latest` installs. |
| `golang.org/x/vuln/cmd/govulncheck` | Known-vulnerability scan (go tool). | `go vet` does not check advisories. | `govulncheck@latest` installs. |
