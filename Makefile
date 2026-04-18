.PHONY: server worker ui build docker-up docker-down test tidy

## --- Dev ---
server:
	go run ./cmd/server

worker:
	go run ./cmd/worker

ui:
	cd web/ui && npm run dev

## --- Build ---
build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker

## --- Docker ---
docker-up:
	docker compose -f deploy/docker/docker-compose.yml up --build -d

docker-down:
	docker compose -f deploy/docker/docker-compose.yml down

## --- Go ---
test:
	go test ./... -race -cover

tidy:
	go mod tidy

## --- UI ---
ui-build:
	cd web/ui && npm run build

ui-install:
	cd web/ui && npm install
