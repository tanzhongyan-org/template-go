# Go Runtime Agent Guidelines (L0)

## 1. Universal Architectural Invariants
- **Standard Go Project Layout:**
  - `cmd/<app-name>/main.go`: Thin entry point, flag parsing, dependency wiring.
  - `internal/<feature>/`: Cohesive domain boundary protected by Go internal package visibility.
  - `internal/core/config/`: Validated configuration. Direct calls to `os.Getenv` in domain code are strictly banned.
  - `pkg/`: Public library code intended for external consumption (only if required).
- **Search-Before-Author:** Search existing packages before authoring helpers or math utilities. Parameterize existing functions; duplicate code is strictly prohibited.
- **Error Handling:** Always handle and wrap errors. Blank error ignores (`_ = err`) in critical paths are forbidden.

## 2. Directory Layout
```
cmd/
└── server/
    └── main.go                  # Main entry point & dependency injection
internal/
├── api/
│   └── handlers/                # Transport: HTTP handlers & route binders
├── domain/                      # Cohesive Domain Boundaries (Colocated)
│   └── <feature>/               # Domain logic & models
└── core/
    └── config/                  # Validated environment configuration
```

## 3. The 6 Verification Gates
Before claiming task completion or creating PRs, all gates must pass:
1. **Typecheck & Compile:** `go vet ./...` & `go build ./...`
2. **Lint:** `golangci-lint run`
3. **Tests:** `go test -v ./...`
4. **Dead Code:** Zero unused exports or orphaned packages
5. **Duplication Audit:** `npx --yes fallow@3.33.1 dupes --quiet --no-fragments`
6. **Anti-Slop Scan:** `npx --yes --package @ast-grep/cli ast-grep scan`
