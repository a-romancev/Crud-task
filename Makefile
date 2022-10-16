.PHONY: start
start:
	docker-compose up

.PHONY: build
build:
	docker-compose build

.PHONY: lint
lint:
	docker run --rm -ti -w /app -v ${PWD}:/app golangci/golangci-lint:v1.45-alpine golangci-lint run

.PHONY: test
test:
	go test ./company/... -race -timeout 2m

.PHONY: test_e2e
test_e2e:
	go test ./cmd/... -race -timeout 2m

.PHONY: migrate
migrate:
	./mongo/init.sh
	./mongo/migrate.sh -source file://migrations/ -database "mongodb://tuser:tpass@mongo:27017/companies" up
	docker exec broker \
    kafka-topics --bootstrap-server broker:9092 \
                 --create \
                 --topic companies_changes