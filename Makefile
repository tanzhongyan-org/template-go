.PHONY: setup up down dev build generate generate-check test lint fix typecheck anti-slop policy docs-lint audit licenses security openapi-client spec-check verify check template-debt

# renovate: datasource=npm depName=@ast-grep/cli
AST_GREP_VERSION := 0.51.0
# renovate: datasource=npm depName=markdownlint-cli2
MARKDOWNLINT_VERSION := 0.24.0
# renovate: datasource=npm depName=@hey-api/openapi-ts
HEY_API_VERSION := 0.99.0
# TypeScript stays below 7: hey-api 0.99 imports the TypeScript 5 default
# export, which TypeScript 7 removed. Drop the second package once hey-api
# supports TypeScript 7.
# renovate: datasource=npm depName=typescript
TYPESCRIPT_VERSION := 5.9.3
# renovate: datasource=github-releases depName=gitleaks/gitleaks extractVersion=^v(?<version>.*)$
GITLEAKS_VERSION := 8.30.1

AST_GREP := npx --yes --package @ast-grep/cli@$(AST_GREP_VERSION) ast-grep
MARKDOWNLINT := npx --yes markdownlint-cli2@$(MARKDOWNLINT_VERSION)
HEY_API := npx --yes --package @hey-api/openapi-ts@$(HEY_API_VERSION) --package typescript@$(TYPESCRIPT_VERSION) openapi-ts

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

licenses:
	go tool go-licenses report --include_tests ./... > NOTICE
	node scripts/licenses.mjs

security: audit
	node scripts/gitleaks.mjs $(GITLEAKS_VERSION)
	node scripts/new-deps.mjs

openapi-client:
	$(HEY_API) -i ./api/openapi.yaml -o ./clients/ts -c @hey-api/client-fetch

spec-check:
	node scripts/spec-check.mjs $(HEY_API_VERSION) $(TYPESCRIPT_VERSION)

template-debt:
	@git grep -n "TEMPLATE-TODO(" -- . ':!Makefile' || true

# Fast gates for agent hooks and pre-push: no database.
verify: typecheck lint anti-slop policy docs-lint spec-check

# Full gate set, run by CI.
check: verify generate-check test
