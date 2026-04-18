APP_NAME=go-etl-studio

.PHONY: run-server run-worker tidy fmt test build

run-server:
	go run ./cmd/server

run-worker:
	go run ./cmd/worker

tidy:
	go mod tidy

fmt:
	gofmt -w ./cmd ./internal ./api ./pkg

test:
	go test ./...

build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
