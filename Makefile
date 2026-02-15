SHELL := /bin/bash

MIGRATIONS_DIR := migrations
DOCKER_MIGRATE_DSN := postgres://postgres:postgres@postgres:5432/aiden?sslmode=disable

.PHONY: run test lint migrate-up migrate-down migrate-seed dev-deps-up dev-migrate-up-docker dev-migrate-down-docker dev-bootstrap

run:
	go run ./cmd/aiden

test:
	go test ./...

lint:
	go vet ./...

migrate-up:
	@if [[ -z "$$DB_DSN" ]]; then \
		echo "DB_DSN is required"; \
		exit 1; \
	fi
	@files=$$(find $(MIGRATIONS_DIR) -maxdepth 1 -type f -name '*.up.sql' | sort); \
	if [[ -z "$$files" ]]; then \
		echo "No up migrations found in $(MIGRATIONS_DIR)"; \
		exit 0; \
	fi; \
	for file in $$files; do \
		echo "Applying $$file"; \
		psql "$$DB_DSN" -v ON_ERROR_STOP=1 -f "$$file"; \
	done

migrate-down:
	@if [[ -z "$$DB_DSN" ]]; then \
		echo "DB_DSN is required"; \
		exit 1; \
	fi
	@files=$$(find $(MIGRATIONS_DIR) -maxdepth 1 -type f -name '*.down.sql' | sort -r); \
	if [[ -z "$$files" ]]; then \
		echo "No down migrations found in $(MIGRATIONS_DIR)"; \
		exit 0; \
	fi; \
	for file in $$files; do \
		echo "Rolling back $$file"; \
		psql "$$DB_DSN" -v ON_ERROR_STOP=1 -f "$$file"; \
	done

migrate-seed:
	@if [[ -z "$$DB_DSN" ]]; then \
		echo "DB_DSN is required"; \
		exit 1; \
	fi
	@echo "Seeding migrations/seed_m1.sql"
	@psql "$$DB_DSN" -v ON_ERROR_STOP=1 -f "$(MIGRATIONS_DIR)/seed_m1.sql"

dev-deps-up:
	docker compose up -d postgres

dev-migrate-up-docker:
	docker compose run --rm --no-deps app bash -lc 'set -euo pipefail; \
		for i in {1..30}; do \
			if pg_isready -h postgres -U postgres -d aiden >/dev/null 2>&1; then \
				break; \
			fi; \
			sleep 1; \
		done; \
		files=$$(find $(MIGRATIONS_DIR) -maxdepth 1 -type f -name "*.up.sql" | sort); \
		if [[ -z "$$files" ]]; then \
			echo "No up migrations found in $(MIGRATIONS_DIR)"; \
			exit 0; \
		fi; \
		for file in $$files; do \
			echo "Applying $$file"; \
			psql "$(DOCKER_MIGRATE_DSN)" -v ON_ERROR_STOP=1 -f "$$file"; \
		done'

dev-migrate-down-docker:
	docker compose run --rm --no-deps app bash -lc 'set -euo pipefail; \
		for i in {1..30}; do \
			if pg_isready -h postgres -U postgres -d aiden >/dev/null 2>&1; then \
				break; \
			fi; \
			sleep 1; \
		done; \
		files=$$(find $(MIGRATIONS_DIR) -maxdepth 1 -type f -name "*.down.sql" | sort -r); \
		if [[ -z "$$files" ]]; then \
			echo "No down migrations found in $(MIGRATIONS_DIR)"; \
			exit 0; \
		fi; \
		for file in $$files; do \
			echo "Rolling back $$file"; \
			psql "$(DOCKER_MIGRATE_DSN)" -v ON_ERROR_STOP=1 -f "$$file"; \
		done'

dev-bootstrap: dev-deps-up dev-migrate-up-docker
	@echo "Dependencies are up and migrations are applied."
