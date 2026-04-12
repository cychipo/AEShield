.PHONY: dev dev_be build clean install run deploy deploy-down deploy-build

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

deploy-build:
	cd deploy && docker compose --env-file .env build

deploy:
	cd deploy && docker compose --env-file .env up --build -d

deploy-down:
	cd deploy && docker compose --env-file .env down

clean:
	rm -rf frontend/dist
	rm -f backend/server

run:
	cd backend && ./server
