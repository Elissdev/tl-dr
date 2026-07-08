.PHONY: build test test-verbose test-race test-integration lint lint-ci clean commit-push release

BINARY_NAME ?= tldr
BUILD_DIR ?= build
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags="-X github.com/Elissdev/tl-dr/cmd.version=$(VERSION)"

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

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

# Commit + Push: cria um commit com a mensagem e faz push para a branch atual
# Uso: make commit-push m="mensagem do commit"
commit-push:
	@BRANCH=$$(git branch --show-current); \
	git add -A && git commit -m "$(m)" && git push origin $$BRANCH

# Release: cria uma tag SemVer, faz push (dispara o job release no CI)
# Uso: make release v=v0.1.0
release:
	@if [ -z "$(v)" ]; then \
		echo "Uso: make release v=v0.1.0"; \
		exit 1; \
	fi
	@echo "Criando release $(v)..."
	git tag -a "$(v)" -m "Release $(v)"
	git push origin "$(v)"
	@echo "✅ Release $(v) criada e enviada!"
	@echo "O CI criará o release no GitHub automaticamente."
