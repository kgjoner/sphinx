test-server:
	bash scripts/test.sh

test-e2e:
	go test -v --short ./test/e2e/...

test-unit:
	go test -v ./internal/...

test: test-server test-e2e test-unit

doc:
	bash scripts/swag.sh

dev:		
	bash scripts/dev.sh

integration: doc
	bash scripts/integration.sh

vela-ci: doc
	bash scripts/integration.sh --registry docker.kgjoner.com.br --platform linux/arm64

ed-ci: doc
	bash scripts/integration.sh --registry docker.elevendragons.com

publish: doc
	bash scripts/publish.sh

logs:
	bash scripts/logs.sh