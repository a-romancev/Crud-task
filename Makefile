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

.PHONY: create_topic
create_topic:
	docker exec broker \
    kafka-topics --bootstrap-server broker:9092 \
                 --create \
                 --topic companies_changes
