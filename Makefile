.PHONY: install up down dev build generate test lint lint-fix check

install:
	go mod download

up:
	docker compose up -d --wait

down:
	docker compose down

dev: up
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
	npx --yes fallow@3.33.1 dupes --root . --quiet --no-fragments
	npx --yes --package @ast-grep/cli ast-grep scan
