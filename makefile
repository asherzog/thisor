.PHONY: build

build:
	go build -o ./bin/thisor ./cmd/.

.PHONY: run-local

run-local: build
	./bin/thisor

.PHONY: docker-build

docker-build:
	docker-compose build

.PHONY: run

run: docker-build
	docker-compose up