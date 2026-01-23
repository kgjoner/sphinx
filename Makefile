doc:
	bash build/scripts/swag.sh

dev: doc
	bash build/scripts/dev.sh

dev-down:
	docker-compose -f docker-compose.yaml down -v

test-env:
	bash build/scripts/test-env.sh

test-env-down:
	docker-compose -f docker-compose.test.yaml down -v

test-e2e: test-env
	@echo "🧪 Running E2E tests with real database..."
	@DATABASE_URL=postgres://postgres:postgres@localhost:5433/sphinx_test?sslmode=disable \
	REDIS_URL=redis://localhost:6380/0 \
	HERMES__BASE_URL=http://localhost:8082/v1 \
	go test -v --short ./test/e2e/... || (make test-env-down && exit 1)
	@make test-env-down

test-unit:
	go test -v ./internal/...

test: test-e2e test-unit

artifact:
	bash build/scripts/tag.sh --canary
	bash build/scripts/integration.sh --canary

arm-ci:
	bash build/scripts/tag.sh
	bash build/scripts/integration.sh --registry=${DOCKER_REGISTRY} --platform linux/arm64

amd-ci:
	bash build/scripts/tag.sh
	bash build/scripts/integration.sh --registry=${DOCKER_REGISTRY}

ci:
	bash build/scripts/tag.sh
	bash build/scripts/integration.sh --registry=${DOCKER_REGISTRY} --platform linux/amd64,linux/arm64

cd:
	bash build/scripts/deploy.sh
