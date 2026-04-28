SHELL := /bin/bash

ifneq (,$(wildcard ./.env))
	include .env
	DATABASE_URL = postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable
	REDIS_URL = redis://localhost:$(REDIS_PORT)/0
	export
endif

ifneq ($(filter release,$(MAKECMDGOALS)),)
KIND ?= canary
VALID_RELEASE_KINDS := stable rc canary nightly

ifeq ($(filter $(KIND),$(VALID_RELEASE_KINDS)),)
$(error Invalid KIND '$(KIND)'. Allowed values: $(VALID_RELEASE_KINDS))
endif

RELEASE_FLAG := --$(KIND)
endif

doc:
	@./build/scripts/swag.sh

dev: doc
	@./build/scripts/dev.sh

dev-down:
	@docker compose -f docker-compose.yaml down -v

test-env:
	@./build/scripts/test-env.sh

test-env-down:
	@docker compose -f docker-compose.test.yaml down -v

test-e2e: test-env
	@echo -e "\n🧪 Running E2E tests with real database and auxiliary services...\n"
	@set -o pipefail; \
	go test -v --short --coverpkg=./internal/... --coverprofile=docs/coverage_e2e.out ./test/e2e/... \
	| grep -E '^(ok|---)' || (make test-env-down && echo -e "\n❌ E2E tests failed!\n" && exit 1)
	@make test-env-down	
	@echo -e "\n======================================================"
	@echo "✅ All E2E tests passed!"
	@echo "📊 Coverage:"
	@go tool cover -func=docs/coverage_e2e.out | awk '/total/ {print "	" $$3}'
	@echo -e "======================================================\n"

test-unit:
	@go test ./internal/...
# 	@go test --cover ./pkg/...

test: test-unit test-e2e 

release:
	@echo "Starting release process with KIND=$(RELEASE_FLAG) and PLATFORM=$(PLATFORM)..."
	@./build/scripts/tag.sh $(RELEASE_FLAG)
	@./build/scripts/integration.sh $(RELEASE_FLAG) --platform=$(PLATFORM)

# Deploy logic is incomplete. Use with caution.
# It deploys to $ENV namespace using helm/${ENV}-values.yaml file. Certify that the file exists 
# and is correct before running this command.
deploy:
	@echo "Deploying to $(ENV)..."
	@if [ "$(ENV)" = "prod" ]; then \
		read -p "Are you SURE you want to deploy to PRODUCTION? [y/N] " ans && [ $${ans:-N} = y ]; \
	fi

	@./build/scripts/deploy.sh --env=$(ENV)
