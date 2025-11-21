APP_NAME := go-book-crud-gin
MAIN := ./cmd/server/main.go

.PHONY: run build test tidy fmt lint clean

## Run the app in dev mode
dev:
	@echo "Running $(APP_NAME) in dev mode..."
	air

## Run the app
run:
	@echo "Running $(APP_NAME)..."
	go run $(MAIN)

## Build the binary
build:
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) $(MAIN)

## Run tests
test:
	@echo "Running tests..."
	go test ./... -v

## Format code
fmt:
	@echo "Formatting..."
	go fmt ./...

## Run 'go vet' (lightweight static analysis)
lint:
	@echo "Linting..."
	go vet ./...

## Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

## Keep go.mod tidy
tidy:
	@echo "Tidying..."
	go mod tidy
