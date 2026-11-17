SQLC_IMAGE    := sqlc/sqlc:1.30.0
MIGRATE_IMAGE := migrate/migrate:v4.19.1
OWNER_DB_URL  := postgres://owner:owner@localhost:5432/portal?sslmode=disable
BACKEND       := backend

.PHONY: help env up down reset logs migrate-up migrate-down sqlc sqlc-check run web test test-integration test-web vet

help: ## List targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*## "}{printf "  %-18s %s\n",$$1,$$2}'

env: ## Create .env with a random JWT secret (if missing)
	@test -f .env || (sed "s/^JWT_SECRET=$$/JWT_SECRET=$$(openssl rand -hex 32)/" .env.example > .env && echo "created .env")

up: env ## Start postgres, run migrations, the API and the web app (http://localhost:5173)
	docker compose up -d --build

down: ## Stop containers (keeps data)
	docker compose down

reset: ## Stop containers and delete all data
	docker compose down -v

logs: ## Follow API logs
	docker compose logs -f api

migrate-up: ## Apply migrations to the running postgres
	docker run --rm --network host -v $(PWD)/$(BACKEND)/pkg/database/migrations:/migrations $(MIGRATE_IMAGE) \
		-path=/migrations -database="$(OWNER_DB_URL)" up

migrate-down: ## Roll back the last migration
	docker run --rm --network host -v $(PWD)/$(BACKEND)/pkg/database/migrations:/migrations $(MIGRATE_IMAGE) \
		-path=/migrations -database="$(OWNER_DB_URL)" down 1

sqlc: ## Generate Go code from SQL
	docker run --rm -v $(PWD)/$(BACKEND):/src -w /src $(SQLC_IMAGE) generate

sqlc-check: ## Fail if generated code is out of date (CI)
	docker run --rm -v $(PWD)/$(BACKEND):/src -w /src $(SQLC_IMAGE) diff

run: env ## Run the API locally against the compose postgres
	cd $(BACKEND) && set -a && . ../.env && set +a && UPLOAD_DIR=../uploads go run ./app

web: ## Run the frontend dev server (proxies /api to localhost:8080)
	cd frontend && npm run dev

test-web: ## Frontend typecheck + unit tests
	cd frontend && npm run typecheck && npm test

vet: ## go vet (including integration tests)
	cd $(BACKEND) && go vet -tags integration ./...

test: ## Unit tests (no database)
	cd $(BACKEND) && go test -race ./...

test-integration: ## Integration tests against PostgreSQL 17 in Docker
	cd $(BACKEND) && go test -race -tags integration -count=1 ./tests/integration/
