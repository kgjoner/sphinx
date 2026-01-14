test-server:
	bash build/scripts/test.sh

test-e2e:
	go test -v --short ./test/e2e/...

test-unit:
	go test -v ./internal/...

test: test-server test-e2e test-unit

doc:
	bash build/scripts/swag.sh

artifact: doc
	bash build/scripts/tag.sh --canary
	bash build/scripts/integration.sh --canary

dev:
	swag init -g server.go --dir internal/server,internal/domains/auth/gateway --parseDependency --parseInternal
	bash build/scripts/dev.sh

arm-ci:
	bash build/scripts/tag.sh
	bash build/scripts/integration.sh --registry=${DOCKER_REGISTRY} --platform linux/arm64

amd-ci:
	bash build/scripts/tag.sh
	bash build/scripts/integration.sh --registry=${DOCKER_REGISTRY}

ci:
	bash build/scripts/tag.sh
	bash build/scripts/integration.sh --registry=${DOCKER_REGISTRY} --platform linux/amd64,linux/arm64

cd: doc
	bash build/scripts/deploy.sh
