.PHONY: dev dev_be build clean install

install:
	cd frontend && yarn install
	cd backend && go mod download

dev_be:
	cd backend && go run cmd/main.go

dev:
	cd backend && go run cmd/main.go

build:
	cd frontend && yarn build
	cd backend && go build -o server ./cmd/main.go

clean:
	rm -rf frontend/dist
	rm -f backend/server

run:
	cd backend && ./server
