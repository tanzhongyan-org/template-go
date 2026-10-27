.PHONY: install dev build generate test lint lint-fix check

install:
	go mod download

dev:
	go run ./cmd/server/main.go

build:
	go build -o bin/server ./cmd/server

generate:
	sqlc generate

test:
	go test -v -race ./...

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
