.PHONY: run dev test lint build tidy

run:
	go run ./cmd/api

dev:
	air

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api

tidy:
	go mod tidy
