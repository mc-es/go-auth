SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: \
	help \
	build run \
	test test-race test-coverage \
	lint format vuln \
	deps-check deps-tidy deps-vendor deps-upgrade \
	clean clean-cache clean-vendor clean-all

# Colors & Formatting
RED    := $(shell tput setaf 1 2>/dev/null || echo "")
GREEN  := $(shell tput setaf 2 2>/dev/null || echo "")
YELLOW := $(shell tput setaf 3 2>/dev/null || echo "")
BLUE   := $(shell tput setaf 4 2>/dev/null || echo "")
CYAN   := $(shell tput setaf 6 2>/dev/null || echo "")
BOLD   := $(shell tput bold 2>/dev/null || echo "")
RESET  := $(shell tput sgr0 2>/dev/null || echo "")

# Helper functions
define print_header
	echo "$(BOLD)$(BLUE)>>> $(1)$(RESET)"
endef

define print_success
	echo "$(GREEN)✔ $(1)$(RESET)"
endef

define print_warning
	echo "$(YELLOW)⚠ $(1)$(RESET)"
endef

# Commands
GO := go

# Paths
PROJECT_ROOT := $(shell pwd)
COVERAGE_DIR := $(PROJECT_ROOT)/coverage
BIN_DIR      := $(PROJECT_ROOT)/bin
TOOLS_DIR    := $(BIN_DIR)/tools

# App
APP_NAME     := app
APP_BINARY   := $(BIN_DIR)/$(APP_NAME)

# Tool Versions
LINT_VERSION     := v2.7.2
VULN_VERSION     := v1.1.4

# Tool Binaries
LINT     := $(TOOLS_DIR)/golangci-lint
VULN     := $(TOOLS_DIR)/govulncheck

# Build Metadata
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date '+%Y-%m-%d-%H:%M:%S')
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -s -w -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)

# Default command
.DEFAULT_GOAL := help

# Directories
$(BIN_DIR) $(TOOLS_DIR) $(COVERAGE_DIR):
	@mkdir -p $@

# Tools Installation
$(LINT): | $(TOOLS_DIR)
	@$(call print_header,"Installing golangci-lint@$(LINT_VERSION)...")
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(TOOLS_DIR)" $(LINT_VERSION)
	@$(call print_success,"golangci-lint installed successfully!")

$(VULN): | $(TOOLS_DIR)
	@$(call print_header,"Installing govulncheck@$(VULN_VERSION)...")
	@GOBIN="$(TOOLS_DIR)" $(GO) install golang.org/x/vuln/cmd/govulncheck@$(VULN_VERSION)
	@$(call print_success,"govulncheck installed successfully!")


# --- Help ---
help: ## Show this help message
	@echo "$(BOLD)Available Commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} \
		/^# ---/ { \
			print ""; \
			gsub(/^# /, "", $$0); \
			printf " $(YELLOW)%s$(RESET)\n", $$0; \
		} \
		/^[a-zA-Z0-9_-]+:.*?## / { \
			printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 \
		}' $(MAKEFILE_LIST)


# --- Build & Run ---
build: | $(BIN_DIR) ## Build the binary
	@$(call print_header,"Building binary...")
	@$(GO) build -o "$(APP_BINARY)" -ldflags '$(LDFLAGS)' ./cmd/app
	@$(call print_success,"Build complete: $(APP_BINARY)")

run: build ## Run the application
	@$(call print_header,"Starting application...")
	@"$(APP_BINARY)"


# --- Testing & Coverage ---
test: ## Run unit tests
	@$(call print_header,"Running tests...")
	@$(GO) test -v ./...
	@$(call print_success,"All tests passed!")

test-race: ## Run tests with race detector
	@$(call print_header,"Running race tests...")
	@$(GO) test -v -race ./...
	@$(call print_success,"Race tests passed!")

test-coverage: | $(COVERAGE_DIR) ## Generate coverage report
	@$(call print_header,"Generating coverage...")
	@$(GO) test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(call print_success,"Report generated: $(COVERAGE_DIR)/coverage.html")


# --- Code Quality ---
lint: $(LINT) ## Run linter (golangci-lint)
	@$(call print_header,"Running golangci-lint...")
	@"$(LINT)" run ./...
	@$(call print_success,"Lint checks passed!")

format: ## Format code (gofmt)
	@$(call print_header,"Formatting code...")
	@$(GO) fmt ./...
	@$(call print_success,"Code formatted!")

vuln: $(VULN) ## Check vulnerabilities (govulncheck)
	@$(call print_header,"Scanning for vulnerabilities...")
	@"$(VULN)" ./...
	@$(call print_success,"No vulnerabilities found!")


# --- Dependencies ---
deps-check: ## Verify dependencies
	@$(GO) mod verify
	@$(call print_success,"Dependencies verified!")

deps-tidy: ## Tidy go.mod
	@$(GO) mod tidy
	@$(call print_success,"Dependencies tidied!")

deps-vendor: ## Create vendor directory
	@$(GO) mod vendor
	@$(call print_success,"Dependencies vendored!")

deps-upgrade: ## Upgrade direct dependencies
	@$(GO) get -u ./... && $(GO) mod tidy
	@$(call print_success,"Dependencies upgraded!")


# --- Cleanup ---
clean: ## Remove artifacts (bin, coverage)
	@$(call print_header,"Cleaning artifacts...")
	@rm -rf "$(BIN_DIR)" "$(COVERAGE_DIR)"
	@$(call print_success,"Artifacts cleaned!")

clean-cache: ## Remove Go cache (cache, testcache)
	@$(call print_header,"Cleaning Go cache...")
	@$(GO) clean -cache -testcache
	@$(call print_success,"Go cache cleaned!")

clean-vendor: ## Remove vendor directory
	@$(call print_header,"Cleaning vendor directory...")
	@rm -rf vendor
	@$(call print_success,"Vendor directory cleaned!")

clean-all: clean clean-cache clean-vendor ## Deep clean everything
