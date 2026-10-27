.PHONY: setup dev test lint fix typecheck check up down build generate

setup:
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

fix:
	golangci-lint run --fix ./...

typecheck:
	go vet ./...

check:
	go vet ./...
	golangci-lint run ./...
	go test -v -race ./...
	go build ./...
	npx --yes --package @ast-grep/cli ast-grep scan
