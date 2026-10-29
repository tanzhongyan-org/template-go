# 0003 Product-lane auth and contracts

Status: accepted

## Context

The product lane asks backend templates for product JWT checks, an
OpenAPI-first example API with a generated TypeScript client and a CI
spec-diff check, plus `make licenses` and `make security` gates.

## Decision

- JWTs are verified against the product JWKS (`internal/auth`) with
  `github.com/golang-jwt/jwt/v5` for parsing and a hand-written JWKS
  client (RS256, ES256) with refresh-on-unknown-key-ID. The `x-user-id`
  header stays for local work only, and the server refuses to boot with
  it on outside development.
- `api/openapi.yaml` is the API contract. `clients/ts` is generated
  with hey-api, a route-coverage test fails when a route is missing
  from the spec, and `make spec-check` fails on client drift.
- `make licenses` checks the go-licenses report against the licence
  allowlist and regenerates `NOTICE`. `make security` runs govulncheck
  (which queries osv.dev, so no separate osv-scanner runs for Go), a
  pinned gitleaks binary, and a new-dependency age and ledger check.

## Consequences

- Derived services point `AUTH_JWKS_URL`, `AUTH_ISSUER` and
  `AUTH_AUDIENCE` at their product auth and delete the header path when
  they no longer need local stubbing.
- Every route change updates `api/openapi.yaml` and regenerates
  `clients/ts` before CI passes.
- New dependencies must be 14 days old and carry a ledger row, or land
  through a recorded exception.
