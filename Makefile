.PHONY: start
start:
	docker-compose build
	docker-compose up

.PHONY: build
build:
	docker-compose build
