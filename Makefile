.PHONY: help up down test test-backend test-portal lint validate build deploy

help:
	@echo "make up            Levanta DynamoDB Local, la API, el portal, MySQL y Mailpit"
	@echo "make down          Detiene el entorno local y borra sus volúmenes"
	@echo "make test          Pruebas de la API (con DynamoDB Local) y del portal"
	@echo "make lint          golangci-lint, Larastan y Pint"
	@echo "make validate      Valida la plantilla SAM con cfn-lint"
	@echo "make build         Compila las funciones Lambda (arm64) con SAM"
	@echo "make deploy        Despliega el stack con SAM (pide los parámetros la primera vez)"

up:
	docker compose up --build -d

down:
	docker compose down -v

test: test-backend test-portal

test-backend:
	docker compose up -d dynamodb
	cd backend && DYNAMODB_ENDPOINT=http://localhost:8000 go test -race -count=1 ./...

test-portal:
	docker build --target test -t events-portal-test portal
	docker run --rm events-portal-test

lint:
	cd backend && golangci-lint run ./...
	docker build --target test -t events-portal-test portal
	docker run --rm events-portal-test sh -c "vendor/bin/phpstan analyse --memory-limit=1G --no-progress && vendor/bin/pint --test"

validate:
	sam validate --lint

build:
	sam build

deploy: build
	sam deploy --guided
