# Contributing

## Prerequisites

| Tool | Version source | Install |
| --- | --- | --- |
| Go | `go.mod` | <https://go.dev/doc/install> |
| Docker | any current release | <https://docs.docker.com/get-docker/> |
| Node.js | 22 or newer | <https://nodejs.org/> |
| golangci-lint | `.github/workflows/ci.yml` | <https://golangci-lint.run/welcome/install/> |
| gitleaks (optional) | `.github/workflows/ci.yml` | <https://github.com/gitleaks/gitleaks#installing> |

sqlc, govulncheck and lefthook come from `go.mod`; you do not install them.

## Set up

```bash
make setup
make up
```

`make setup` downloads modules and installs the git hooks. `make up` starts Postgres and prints `Healthy` when it is ready.

## Daily loop

```bash
make verify
```

This runs vet, build, lint, anti-slop rules, repo policy and Markdown lint. It finishes without errors and needs no database.

Before you open a pull request:

```bash
make check
```

This adds the sqlc drift check and the tests. It needs Postgres from `make up`.

## Add a feature

Follow the notes slice. Touch the files in this order:

1. `sql/schema.sql`: add the table or column.
2. `sql/queries.sql`: add named queries. Filter by `owner_id` on every read and write.
3. Run `make generate`. It updates `internal/db`.
4. `internal/handler/<feature>.go`: add handlers. Reuse `writeJSON`, `writeError` and `decodeJSON` from `internal/handler/respond.go`.
5. `cmd/server/main.go`: mount the routes in `SetupRouter`.
6. Tests: call the public endpoints through `SetupRouter`. Cover the success path, validation errors, 401 and cross-owner access (404).
7. Docs: update the code map in `ARCHITECTURE.md`, add a row to `docs/dependencies.md` for any new module, and add new config keys to `.env.example` and the README table.

A new top-level or `internal/` directory needs an ADR in `docs/decisions/` and a `layout.allow` change in `scripts/repo-policy.json` in the same pull request.

## Commits and pull requests

- Write Conventional Commits: `feat(notes): add note search`.
- Keep one logical change per commit.
- Do not add AI `Co-Authored-By` trailers. The `commit-msg` hook rejects them.
- Fill in the pull request template. Paste the quiet output of `make check`.

## Where docs go

See [docs/README.md](../docs/README.md).

## Maintainers

Resolve template markers first:

```bash
make template-debt
```

<!-- TEMPLATE-TODO(license): set the copyright holder in LICENSE for your organisation. -->

Import the branch ruleset once the repository exists:

1. Open Settings, then Rules, then Rulesets, then Import a ruleset.
2. Choose `.github/rulesets/main.json`.

Or with the GitHub CLI:

```bash
gh api repos/OWNER/REPO/rulesets --method POST --input .github/rulesets/main.json
```

Run the initialize workflow before you import the ruleset. It pushes directly to `main`, which the ruleset blocks.

Renovate reads `.github/renovate.json`. Versions in `go.mod`, in `# renovate:` annotated lines of the `Makefile` and in workflows are the only places tool versions live.
