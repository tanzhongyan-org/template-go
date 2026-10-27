.PHONY: setup dev test lint fix typecheck check up down build generate audit

setup:
	go mod download
	npx --yes lefthook@2.2.2 install

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

audit:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

check:
	go vet ./...
	golangci-lint run ./...
	go test -v -race ./...
	go build ./...
	npx --yes --package @ast-grep/cli@0.51.0 ast-grep scan
	npx --yes --package @ast-grep/cli@0.51.0 ast-grep test --skip-snapshot-tests
