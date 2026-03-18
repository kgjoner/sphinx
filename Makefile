CI_RELEASE_GOALS := release

ifneq ($(filter $(CI_RELEASE_GOALS),$(MAKECMDGOALS)),)
RELEASE ?= canary
VALID_RELEASE_KINDS := stable rc canary nightly

ifeq ($(filter $(RELEASE),$(VALID_RELEASE_KINDS)),)
$(error Invalid RELEASE '$(RELEASE)'. Allowed values: $(VALID_RELEASE_KINDS))
endif

RELEASE_FLAG := --$(RELEASE)
endif

doc:
	bash build/scripts/swag.sh

dev: doc
	bash build/scripts/dev.sh

dev-down:
	docker compose -f docker-compose.yaml down -v

test-env:
	bash build/scripts/test-env.sh

test-env-down:
	docker compose -f docker-compose.test.yaml down -v

test-e2e: test-env
	@echo -e "\n🧪 Running E2E tests with real database and auxiliary services...\n"
	go test -v --short ./test/e2e/... || (make test-env-down && exit 1)
	@make test-env-down

test-unit:
	go test -v ./internal/...
	go test -v ./pkg/...

test: test-e2e test-unit

release:
	bash build/scripts/tag.sh $(RELEASE_FLAG)
	bash build/scripts/integration.sh $(RELEASE_FLAG) --platform=${PLATFORM}

# Deploy logic is incomplete. Use with caution.
# It deploys to $ENV namespace using helm/${ENV}-values.yaml file. Certify that the file exists 
# and is correct before running this command.
deploy:
	@echo "Deploying to $(ENV)..."
	@if [ "$(ENV)" = "prod" ]; then \
		read -p "Are you SURE you want to deploy to PRODUCTION? [y/N] " ans && [ $${ans:-N} = y ]; \
	fi

	bash build/scripts/deploy.sh --env=$(ENV)
