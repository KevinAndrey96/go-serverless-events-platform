.PHONY: all test build build-backend clean docker-up docker-down help

all: test build

help:
	@echo "Serverless Events Platform - Monorepo Commands:"
	@echo "  make test          Run unit test suites across all services"
	@echo "  make build         Build backend microservice binaries"
	@echo "  make docker-up     Start local DynamoDB, MySQL, and Mailpit services"
	@echo "  make docker-down   Stop and clean up local Docker containers"
	@echo "  make clean         Remove build artifacts"

test:
	@echo "==> Running Go backend unit tests..."
	cd backend && go test -v -race ./pkg/...

build: build-backend

build-backend:
	@echo "==> Compiling Go AWS Lambda microservices..."
	cd backend && make build

docker-up:
	@echo "==> Booting local DynamoDB, Admin UI, MySQL & Mailpit..."
	docker compose up -d

docker-down:
	@echo "==> Stopping local Docker containers..."
	docker compose down

clean:
	@echo "==> Cleaning build artifacts..."
	cd backend && make clean
