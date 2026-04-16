.PHONY: lint test mock build docker-build up down hooks

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run

test:
	go test ./...

mock:
	go tool mockery

build:
	go build -o bin/bot ./cmd/bot/

run: build
	./bin/bot --config ./configs/config.yaml

docker-build:
	docker build -t barman .

up:
	docker compose up -d

down:
	docker compose down

hooks:
	git config core.hooksPath .githooks
