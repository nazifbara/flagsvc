include .env
export
export GOOSE_MIGRATION_DIR  ?= ./backend/sql/schema

.PHONY: gen dev-backend dev-frontend build-backend sqlc migrate-up migrate-down


gen:
	buf generate

sqlc:
	sqlc generate .

db-up:
	goose postgres $(DB_URL) up

db-down:
	goose postgres $(DB_URL) down

dev-backend:
	go run ./backend/cmd/server

dev-frontend:
	pnpm --filter frontend dev

build-backend:
	go build -o bin/server ./backend/cmd/server
