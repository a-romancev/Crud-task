.PHONY: start
start:
	docker-compose build
	docker-compose up

.PHONY: build
build:
	docker-compose build

.PHONY: lint
lint:
	docker run --rm -ti -w /app -v ${PWD}:/app golangci/golangci-lint:v1.45-alpine golangci-lint run
