.PHONY: run build tidy test fmt lint

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go vet ./...
	go test ./...

fmt:
	gofmt -w .
	goimports -w .

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
