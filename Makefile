.PHONY: install dev test lint lint-fix check

install:
	go mod download

dev:
	go run ./cmd/server/main.go

test:
	go test -v ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

check:
	go vet ./...
	go test -v ./...
	go build ./...
	npx --yes fallow@3.33.1 dupes --quiet --no-fragments
	npx --yes --package @ast-grep/cli ast-grep scan
