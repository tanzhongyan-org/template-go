# Go Runtime Agent Guidelines (L1 Extension)

Inherits: Universal L0 Invariants (workflow/AGENTS.md).
Do not repeat global workflow rules here.

## 1. Project Technology Stack
- Runtime: Go 1.23+
- Project Layout: Standard Go project layout (cmd/ and internal/)
- Standard Router: net/http with structured JSON encoding
- Lint & Verification: go vet, go test

## 2. Physical File Placement Rules
- Entry Point: Thin server initialization in cmd/server/main.go
- Transport Handlers: Place HTTP route binders in internal/api/handlers/
- Domain Logic: Place features and models in internal/domain/<feature>/
- Configuration: Singleton environment loader in internal/core/config/
- Test Suites: Sister-file colocated tests (e.g. internal/domain/<feature>/<feature>_test.go)

## 3. Data Access & Boundary Invariants
- Direct calls to os.Getenv in domain packages are strictly banned. Load config via internal/core/config.
- Always handle and wrap returned errors. Blank error ignores (_ = err) in domain paths are forbidden.

## 4. The 6 Verification Gates
Run these commands to prove task completion:
1. Gate 1 (Vet & Build): go vet ./... && go build ./...
2. Gate 2 (Lint): golangci-lint run (or go vet ./...)
3. Gate 3 (Tests): go test -v ./...
4. Gate 4 (Dead Code): Zero unused exports or orphaned packages
5. Gate 5 (Duplication): npx --yes fallow@3.33.1 dupes --quiet --no-fragments
6. Gate 6 (Anti-Slop): npx --yes --package @ast-grep/cli ast-grep scan
