test-server:
	bash scripts/test.sh

test-e2e:
	go test -v --short ./test/e2e/...

test-unit:
	go test -v ./internal/...

test: test-server test-e2e test-unit

doc:
	bash scripts/swag.sh

artifact: doc
	bash scripts/tag.sh --canary
	bash scripts/integration.sh --canary

dev: build
	bash scripts/dev.sh

arm-ci: doc
	bash scripts/tag.sh
	bash scripts/integration.sh --registry ${DOCKER_REGISTRY:-docker.kgjoner.com.br} --platform linux/arm64

amd-ci: doc
	bash scripts/tag.sh
	bash scripts/integration.sh --registry=${DOCKER_REGISTRY}

ci: doc
	bash scripts/tag.sh
	bash scripts/integration.sh --registry=${DOCKER_REGISTRY} --platform linux/amd64,linux/arm64

cd: doc
	bash scripts/deploy.sh
