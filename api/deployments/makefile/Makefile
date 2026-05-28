SHELL := /bin/bash

COMPOSE_FILE      := deployments/docker/docker-compose.yml
ENV_FILE          := configs/.env

COMPOSE_FILE_TEST := deployments/docker/docker-compose.test.yml
ENV_FILE_TEST     := configs/.env.test

ifneq (,$(wildcard $(ENV_FILE)))
POSTGRES_USER     := $(shell grep ^POSTGRES_USER     $(ENV_FILE) | cut -d= -f2-)
POSTGRES_PASSWORD := $(shell grep ^POSTGRES_PASSWORD $(ENV_FILE) | cut -d= -f2-)
POSTGRES_HOST     := $(shell grep ^POSTGRES_HOST     $(ENV_FILE) | cut -d= -f2-)
POSTGRES_PORT     := $(shell grep ^POSTGRES_PORT     $(ENV_FILE) | cut -d= -f2-)
POSTGRES_DB       := $(shell grep ^POSTGRES_DB       $(ENV_FILE) | cut -d= -f2-)
endif

ifneq (,$(wildcard $(ENV_FILE_TEST)))
TEST_POSTGRES_USER     := $(shell grep ^POSTGRES_USER     $(ENV_FILE_TEST) | cut -d= -f2-)
TEST_POSTGRES_PASSWORD := $(shell grep ^POSTGRES_PASSWORD $(ENV_FILE_TEST) | cut -d= -f2-)
TEST_POSTGRES_HOST     := $(shell grep ^POSTGRES_HOST     $(ENV_FILE_TEST) | cut -d= -f2-)
TEST_POSTGRES_PORT     := $(shell grep ^POSTGRES_PORT     $(ENV_FILE_TEST) | cut -d= -f2-)
TEST_POSTGRES_DB       := $(shell grep ^POSTGRES_DB       $(ENV_FILE_TEST) | cut -d= -f2-)
endif

POSTGRES_HOST := $(or $(POSTGRES_HOST),localhost)

DC ?= docker compose
DC_CMD      := $(DC) -p atom-maintenance      --env-file $(ENV_FILE)      -f $(COMPOSE_FILE)
DC_TEST_CMD := $(DC) -p atom-maintenance-test --env-file $(ENV_FILE_TEST) -f $(COMPOSE_FILE_TEST)

GO      ?= go
API_PKG := ./cmd/api
API_BIN := bin/api

DB_URL      := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
DB_URL_TEST := postgres://$(TEST_POSTGRES_USER):$(TEST_POSTGRES_PASSWORD)@$(TEST_POSTGRES_HOST):$(TEST_POSTGRES_PORT)/$(TEST_POSTGRES_DB)?sslmode=disable

MIGRATIONS_DIR_CORE := migrations/core
MIGRATE_RUN := docker run --rm -v $(PWD):/work -w /work --network host migrate/migrate

.PHONY: up down restart logs ps api-run api-build api-clean dev tidy fmt vet \
        migrate-up migrate-down migrate-version migrate-new \
        up-test down-test restart-test logs-test ps-test \
        migrate-test-up migrate-test-down migrate-test-version \
        test-integration test-integration-no-up test-coverage \
        simulator-install simulator-run simulator-rebuild seed seed-equipment

up:
	@test -f $(ENV_FILE) || (echo "ERROR: $(ENV_FILE) not found"; exit 1)
	$(DC_CMD) up -d

down:
	$(DC_CMD) down

restart:
	$(DC_CMD) restart

logs:
	$(DC_CMD) logs -f

ps:
	$(DC_CMD) ps


api-run:
	$(GO) run $(API_PKG)

api-build:
	@mkdir -p bin
	$(GO) build -o $(API_BIN) $(API_PKG)
	@echo "Built $(API_BIN)"

api-clean:
	@rm -f $(API_BIN)

dev:
	$(MAKE) -f deployments/makefile/Makefile up
	$(MAKE) -f deployments/makefile/Makefile api-run

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

migrate-up:
	$(MIGRATE_RUN) -path $(MIGRATIONS_DIR_CORE) -database "$(DB_URL)" -verbose up

migrate-down:
	$(MIGRATE_RUN) -path $(MIGRATIONS_DIR_CORE) -database "$(DB_URL)" -verbose down

migrate-version:
	$(MIGRATE_RUN) -path $(MIGRATIONS_DIR_CORE) -database "$(DB_URL)" version

migrate-new:
	@test -n "$(name)" || (echo "usage: make migrate-new name=<snake_case>"; exit 1)
	$(MIGRATE_RUN) create -dir $(MIGRATIONS_DIR_CORE) -ext sql -seq $(name)

up-test:
	@test -f $(ENV_FILE_TEST) || (echo "ERROR: $(ENV_FILE_TEST) not found"; exit 1)
	$(DC_TEST_CMD) up -d

down-test:
	$(DC_TEST_CMD) down -v

restart-test:
	$(DC_TEST_CMD) restart

logs-test:
	$(DC_TEST_CMD) logs -f

ps-test:
	$(DC_TEST_CMD) ps

migrate-test-up: up-test
	$(MIGRATE_RUN) -path $(MIGRATIONS_DIR_CORE) -database "$(DB_URL_TEST)" -verbose up

migrate-test-down:
	$(MIGRATE_RUN) -path $(MIGRATIONS_DIR_CORE) -database "$(DB_URL_TEST)" -verbose down

migrate-test-version:
	$(MIGRATE_RUN) -path $(MIGRATIONS_DIR_CORE) -database "$(DB_URL_TEST)" version

test-integration: migrate-test-up
	CONFIG_PATH=$(PWD)/configs/config.test.yml \
	$(GO) test -v -count=1 ./test/integration/...

test-integration-no-up:
	CONFIG_PATH=$(PWD)/configs/config.test.yml \
	$(GO) test -v -count=1 -p=1 ./test/integration/...

test-coverage: migrate-test-up
	CONFIG_PATH=$(PWD)/configs/config.test.yml \
	$(GO) test -count=1 -p=1 ./test/integration/... \
	-coverpkg=atom-maintenance/internal/... \
	-coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	xdg-open coverage.html

simulator-install:
	cd tools/simulator && pip install -r requirements.txt

simulator-run:
	cd tools/simulator && uvicorn main:app --host 0.0.0.0 --port 8090 --reload

simulator-rebuild:
	@test -f $(ENV_FILE) || (echo "ERROR: $(ENV_FILE) not found"; exit 1)
	$(DC_CMD) up -d --build simulator

seed:
	$(GO) run ./cmd/seed

seed-equipment:
	$(GO) run ./cmd/seed-equipment
