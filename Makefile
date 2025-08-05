all: generate tidy-all lint-all docker-test-server test

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: tidy-test-server
tidy-test-server:
	cd internal/tests/test_server && go mod tidy

.PHONY: tidy-all
tidy-all: tidy tidy-test-server

.PHONY: docker-test-server
docker-test-server:
	docker build -t agrirouter-test-server -f internal/tests/test_server/Dockerfile .

.PHONY: vet
vet:
	go vet ./...

.PHONY: vet-test-server
vet-test-server:
	cd internal/tests/test_server && go vet ./...

.PHONY: vet-all
vet-all: vet vet-test-server

.PHONY: lint
lint: vet
lint:
	tools/golang_ci_lint.sh

.PHONY: lint-all
lint-all: vet-all
lint-all:
	tools/golang_ci_lint.sh all

.PHONY: generate
generate:
	tools/oapi/generate.sh
