.PHONY: dev build lint test test-unit test-e2e clean

dev:
	go run github.com/air-verse/air@latest -c .air.toml

build:
	go build -o server cmd/main.go

lint:
	golangci-lint run

test:
	go test ./... -v

test-unit:
	go test ./internal/... -v

test-e2e:
	go test ./e2e/... -v

clean:
	rm -f server
