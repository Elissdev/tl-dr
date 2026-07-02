.PHONY: build test lint clean

BINARY_NAME ?= tldr
BUILD_DIR ?= build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BINARY_NAME) .

test:
	@echo "Running tests..."
	go test ./... -v -count=1

lint:
	@echo "Running linter..."
	golangci-lint run ./... 2>/dev/null || echo "golangci-lint não encontrado — instale com: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Done."
