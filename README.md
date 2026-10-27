# Go Runtime Template

Universal Go application boilerplate with standard project layout, agentic quality guardrails, and verification gates.

## Features

- **Runtime:** Go 1.23+ with standard project layout (`cmd/`, `internal/`)
- **Transport:** Standard library `net/http` router with structured JSON encoding
- **Config & Validation:** Validated environment configuration module
- **Agent Guardrails:**
  - `lefthook.yml`: Sub-second local Git pre-commit hooks
  - `fallow.jsonc`: AST-aware code duplication and cognitive complexity limits
  - `sgconfig.yml` + `.ast-grep/`: Automatic detection of placeholder stubs and unfulfilled code
  - `.github/workflows/ci.yml`: Automated CI running the 6 Universal Verification Gates
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
# Run server
go run ./cmd/server/main.go
```

### 3. Run Verification Gates
```bash
# Gate 1: Vet & Build
go vet ./...
go build ./...

# Gate 3: Test
go test -v ./...

# Gate 5: Duplication Audit
npx --yes fallow@3.33.1 dupes --quiet --no-fragments

# Gate 6: Anti-Slop Scan
npx --yes --package @ast-grep/cli ast-grep scan
```
