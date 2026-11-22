.PHONY: all test build build-backend clean help

all: test build

help:
	@echo "Serverless Events Platform - Monorepo Commands:"
	@echo "  make test          Run unit test suites across all services"
	@echo "  make build         Build backend microservice binaries"
	@echo "  make clean         Remove build artifacts"

test:
	@echo "==> Running Go backend unit tests..."
	cd backend && go test -v -race ./pkg/...

build: build-backend

build-backend:
	@echo "==> Compiling Go AWS Lambda microservices..."
	cd backend && make build

clean:
	@echo "==> Cleaning build artifacts..."
	cd backend && make clean
