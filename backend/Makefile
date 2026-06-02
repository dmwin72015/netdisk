# ----- Variables -----
SHELL := /bin/bash
GO    := go
PG_DSN ?= postgres://postgres:root1234@localhost:15432/netdisk?sslmode=disable

MIGRATE_IMAGE := migrate/migrate:latest
SQLC_IMAGE    := sqlc/sqlc:v1.27.0

# Prefer locally installed binaries; fall back to docker if absent.
SQLC_BIN := $(shell command -v sqlc 2>/dev/null)
ifeq ($(SQLC_BIN),)
SQLC_BIN := $(shell ls $$HOME/go/bin/sqlc 2>/dev/null)
endif
MIGRATE_BIN := $(shell command -v migrate 2>/dev/null)
ifeq ($(MIGRATE_BIN),)
MIGRATE_BIN := $(shell ls $$HOME/go/bin/migrate 2>/dev/null)
endif

# ----- Targets -----
.PHONY: help dev-infra dev-down migrate-up migrate-down sqlc-gen run build lint test clean tidy

help:
	@echo "Targets:"
	@echo "  dev-infra   Start PostgreSQL + Redis via docker compose"
	@echo "  dev-down    Stop docker compose services"
	@echo "  migrate-up  Apply database migrations"
	@echo "  migrate-down Rollback the last migration"
	@echo "  sqlc-gen    Generate sqlc code"
	@echo "  run         Run the backend server"
	@echo "  build       Build the backend binary"
	@echo "  lint        Run golangci-lint"
	@echo "  test        Run go tests"
	@echo "  tidy        go mod tidy"

dev-infra:
	cd .. && docker compose up -d

dev-down:
	cd .. && docker compose down

migrate-up:
ifeq ($(MIGRATE_BIN),)
	docker run --rm -v $(PWD)/internal/db/migrations:/migrations \
		--network host $(MIGRATE_IMAGE) \
		-path=/migrations -database "$(PG_DSN)" up
else
	$(MIGRATE_BIN) -path=internal/db/migrations -database "$(PG_DSN)" up
endif

migrate-down:
ifeq ($(MIGRATE_BIN),)
	docker run --rm -v $(PWD)/internal/db/migrations:/migrations \
		--network host $(MIGRATE_IMAGE) \
		-path=/migrations -database "$(PG_DSN)" down 1
else
	$(MIGRATE_BIN) -path=internal/db/migrations -database "$(PG_DSN)" down 1
endif

# Build frontend with static adapter and copy to backend for embedding
.PHONY: build-frontend-static
build-frontend-static:
	cd ../frontend && cp svelte.config.js svelte.config.js.bak && cp svelte.config.static.js svelte.config.js && pnpm build && mv svelte.config.js.bak svelte.config.js
	rm -rf internal/web/build && mkdir -p internal/web/build
	cp -r ../frontend/build/* internal/web/build/
	find internal/web/build -name '_*' -depth -exec sh -c 'for f; do d=$$(dirname "$$f"); b=$$(basename "$$f"); mv "$$f" "$$d/$${b#_}"; done' _ {} +
	find internal/web/build -type f \( -name '*.html' -o -name '*.js' -o -name '*.css' -o -name '*.json' \) -exec sed -i 's|/_app/|/app/|g' {} +

# Build single container image (frontend + backend)
.PHONY: build-single
build-single:
	docker compose -f ../docker-compose.single.yml build

sqlc-gen:
ifeq ($(SQLC_BIN),)
	docker run --rm -v $(PWD):/src -w /src $(SQLC_IMAGE) generate
else
	$(SQLC_BIN) generate
endif

run:
	$(GO) run ./cmd/server

build:
	$(GO) build -o bin/server ./cmd/server

lint:
	@command -v golangci-lint >/dev/null || (echo "golangci-lint not installed; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

test:
	$(GO) test -race -count=1 ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf bin/ tmp/
