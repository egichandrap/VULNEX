SHELL := /bin/bash
.DEFAULT_GOAL := help

BIN_DIR := bin
BACKEND_MODULE := ./backend
INSTALLER_MODULE := ./installer

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: ## Format Go source code and verify formatting
	@echo "Formatting Go code..."
	@cd $(BACKEND_MODULE) && go fmt ./...
	@cd $(INSTALLER_MODULE) && go fmt ./...

.PHONY: vet
vet: ## Run go vet across all modules
	@echo "Running go vet on backend..."
	@cd $(BACKEND_MODULE) && go vet ./...
	@echo "Running go vet on installer..."
	@cd $(INSTALLER_MODULE) && go vet ./...

.PHONY: test
test: ## Run unit tests across all modules
	@echo "Running unit tests on backend..."
	@cd $(BACKEND_MODULE) && go test -v -cover ./...
	@echo "Running unit tests on installer..."
	@cd $(INSTALLER_MODULE) && go test -v -cover ./...

.PHONY: test-race
test-race: ## Run unit tests with race detection enabled
	@echo "Running race tests on backend..."
	@cd $(BACKEND_MODULE) && go test -race -v ./...
	@echo "Running race tests on installer..."
	@cd $(INSTALLER_MODULE) && go test -race -v ./...

.PHONY: test-security
test-security: ## Run dedicated security test suites (SSRF, DNS rebinding, scope invariants)
	@echo "Running security tests..."
	@cd $(BACKEND_MODULE) && go test -v -tags=security ./...

.PHONY: test-integration
test-integration: ## Run adapter integration tests
	@echo "Running integration tests..."
	@cd $(BACKEND_MODULE) && go test -v -tags=integration ./...
	@cd $(INSTALLER_MODULE) && go test -v -tags=integration ./...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@cd tests && go test -v ./...

.PHONY: test-arch
test-arch: ## Run Clean Architecture and package boundary validation tests
	@echo "Running Architecture boundary tests..."
	@cd $(BACKEND_MODULE) && go test -v ./internal/archtest/...

.PHONY: build
build: build-api build-worker build-cli ## Build all executables

.PHONY: build-api
build-api: ## Build backend API binary
	@mkdir -p $(BIN_DIR)
	@echo "Building vulnex-api..."
	@cd $(BACKEND_MODULE) && go build -o ../$(BIN_DIR)/vulnex-api ./cmd/api

.PHONY: build-worker
build-worker: ## Build backend worker binary
	@mkdir -p $(BIN_DIR)
	@echo "Building vulnex-worker..."
	@cd $(BACKEND_MODULE) && go build -o ../$(BIN_DIR)/vulnex-worker ./cmd/worker

.PHONY: build-cli
build-cli: ## Build installer CLI binary
	@mkdir -p $(BIN_DIR)
	@echo "Building vulnex CLI..."
	@cd $(INSTALLER_MODULE) && go build -o ../$(BIN_DIR)/vulnex ./cmd/vulnex

.PHONY: clean
clean: ## Remove compiled binaries and test outputs
	@rm -rf $(BIN_DIR) coverage.out coverage.html
