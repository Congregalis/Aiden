SHELL := /bin/bash

MIGRATIONS_DIR := migrations

.PHONY: run test lint migrate-up migrate-down migrate-seed

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
