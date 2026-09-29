.PHONY: run dev migrate test check build fmt
run:
	go run ./cmd/api
dev:
	air
migrate:
	go run ./cmd/migrate
test:
	go test -race ./...
check:
	go vet ./...
	go test ./...
build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate
fmt:
	gofmt -w cmd internal migrations
