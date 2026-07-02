.PHONY: build test test-verbose test-race test-integration lint lint-ci clean

BINARY_NAME ?= tldr
BUILD_DIR ?= build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	@echo "Running tests..."
	go test ./... -count=1

test-verbose:
	@echo "Running tests (verbose)..."
	go test ./... -v -count=1

test-race:
	@echo "Running tests with race detector..."
	go test ./... -race -count=1

test-integration:
	@echo "Running integration tests (requires TLDR_API_KEY)..."
	go test -tags=integration -v -count=1 ./internal/integration/

lint:
	@echo "Running linter..."
	golangci-lint run ./... 2>/dev/null || echo "golangci-lint não encontrado — instale com: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

lint-ci:
	@echo "Running linter (strict)..."
	golangci-lint run ./... --timeout=5m

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Done."
