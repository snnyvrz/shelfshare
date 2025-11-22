APP_NAME := go-book-crud-gin
MAIN := ./cmd/server/main.go

DOCKER_IMAGE_DEV := $(APP_NAME)-dev
DOCKER_IMAGE_PROD := $(APP_NAME)
DOCKER_PORT := 8080

.PHONY: dev dev-down dev-logs run build test tidy fmt lint clean docker-clean

dev:
	trap 'docker compose down' INT TERM HUP; docker compose up --build

dev-down:
	docker compose down

dev-logs:
	docker compose logs -f

run:
	docker build --target prod -t $(DOCKER_IMAGE_PROD) .
	docker run --rm -p $(DOCKER_PORT):$(DOCKER_PORT) $(DOCKER_IMAGE_PROD)

build:
	docker build --target prod -t $(DOCKER_IMAGE_PROD) .

test:
	docker run --rm \
		-v "$$PWD":/app \
		-w /app \
		golang:1.25.4-bookworm \
		go test ./... -v

fmt:
	go fmt ./...

lint:
	go vet ./...

clean:
	rm -rf bin/

tidy:
	go mod tidy

docker-clean:
	- docker rmi $(DOCKER_IMAGE_DEV) $(DOCKER_IMAGE_PROD) 2>/dev/null || true
