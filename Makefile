# Variables
BINARY_NAME=nullmail
MAIN_PATH=cmd/nullmail/main.go
BUILD_DIR=bin
GO_VERSION=1.21.3

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: run
run: ## Run the SMTP server in development mode
	go run $(MAIN_PATH)

.PHONY: run-client
run-client: ## Run the Next.js client in development mode
	cd client && pnpm run dev

.PHONY: build
build: ## Build the binary for current platform
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)


.PHONY: test
test: ## Run tests
	go test -v ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	go test -race -v ./...

.PHONY: test-cover
test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: bench
bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

.PHONY: fmt
fmt: ## Format Go code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint to be installed)
	golangci-lint run

.PHONY: check
check: fmt vet ## Run formatting and vetting

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: install
install: build ## Install binary to $GOPATH/bin
	go install $(MAIN_PATH)

# Docker (optional)
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME) .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run -p 2525:2525 $(BINARY_NAME)

.PHONY: test-smtp
test-smtp: ## Test SMTP server with swaks (requires swaks to be installed)
	@echo "Testing SMTP server (make sure server is running on port 2525)..."
	swaks --to user@nullmail.local --from sender@example.com --server localhost:2525 --body "Test email from Makefile"

.PHONY: redis-logs
redis-logs: ## View Redis logs
	cd infra && docker-compose logs -f redis

.PHONY: redis-init
redis-init: ## Initialize Redis with sample data
	cd infra && docker-compose exec redis sh /usr/local/etc/redis/init.sh

.PHONY: redis-reset
redis-reset: ## Reset Redis data
	cd infra && docker-compose exec redis redis-cli -a dev123 FLUSHALL

.PHONY: docker-up
docker-up: ## Start all development services
	cd infra && docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop all development services
	cd infra && docker-compose down