.PHONY: setup up down dev build generate generate-check test lint fix typecheck anti-slop policy docs-lint audit verify check template-debt

# renovate: datasource=npm depName=@ast-grep/cli
AST_GREP_VERSION := 0.51.0
# renovate: datasource=npm depName=markdownlint-cli2
MARKDOWNLINT_VERSION := 0.24.0

AST_GREP := npx --yes --package @ast-grep/cli@$(AST_GREP_VERSION) ast-grep
MARKDOWNLINT := npx --yes markdownlint-cli2@$(MARKDOWNLINT_VERSION)

setup:
	go mod download
	go tool lefthook install

up:
	docker compose up -d --wait

down:
	docker compose down

dev: up
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

generate:
	go tool sqlc generate

generate-check:
	go tool sqlc diff

test:
	go test -race ./...

lint:
	golangci-lint run ./...

fix:
	golangci-lint run --fix ./...

typecheck:
	go vet ./...
	go build ./...

anti-slop:
	$(AST_GREP) scan
	$(AST_GREP) test --skip-snapshot-tests

policy:
	node scripts/repo-policy.mjs all

docs-lint:
	$(MARKDOWNLINT)

audit:
	go tool govulncheck ./...

template-debt:
	@git grep -n "TEMPLATE-TODO(" -- . ':!Makefile' || true

# Fast gates for agent hooks and pre-push: no database.
verify: typecheck lint anti-slop policy docs-lint

# Full gate set, run by CI.
check: verify generate-check test
