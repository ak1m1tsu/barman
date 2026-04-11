.PHONY: lint test mock build docker-build up down

lint:
	golangci-lint run

test:
	go test ./...

mock:
	go tool mockery

build:
	go build -o bin/bot ./cmd/bot/

docker-build:
	docker build -t barman .

up:
	docker compose up -d

down:
	docker compose down
