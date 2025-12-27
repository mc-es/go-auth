SHELL := /bin/bash

.PHONY: \
	help \
	build run dev \
	test test-race test-coverage \
	lint format vuln \
	deps-check deps-tidy deps-vendor deps-upgrade \
	setup-golangci-lint setup-govulncheck setup-air setup-all \
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

define print_error
	echo "$(RED)✘ $(1)$(RESET)"
endef

# Commands
GO := go

# Paths
APP_NAME     := app
PROJECT_ROOT := $(shell pwd)
COVERAGE_DIR := $(PROJECT_ROOT)/coverage
TMP_DIR      := $(PROJECT_ROOT)/tmp
BIN_DIR      := $(PROJECT_ROOT)/bin
TOOLS_DIR    := $(BIN_DIR)/tools

# App Binary
APP_BINARY   := $(BIN_DIR)/$(APP_NAME)

# Tool Versions
LINT_VERSION     := v2.7.2
VULN_VERSION     := v1.1.4
AIR_VERSION      := v1.63.4

# Tool Paths
LINT     := $(TOOLS_DIR)/golangci-lint
VULN     := $(TOOLS_DIR)/govulncheck
AIR      := $(TOOLS_DIR)/air

# Build Metadata
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY  := $(shell git diff --quiet && git diff --cached --quiet || echo "+CHANGES")
BUILD_DATE := $(shell date '+%Y-%m-%d-%H:%M:%S')
LDFLAGS    := -w -s -X main.Commit=$(GIT_COMMIT)$(GIT_DIRTY) -X main.BuildDate=$(BUILD_DATE)

# Default command
.DEFAULT_GOAL := help

# Directories
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

$(COVERAGE_DIR):
	@mkdir -p $(COVERAGE_DIR)

$(TOOLS_DIR):
	@mkdir -p $(TOOLS_DIR)

$(TMP_DIR):
	@mkdir -p $(TMP_DIR)


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
build: | $(BIN_DIR) ## Build the application
	@$(call print_header,"Building binary...")
	@if $(GO) build -o "$(APP_BINARY)" -ldflags '$(LDFLAGS)' ./cmd/app; then \
		$(call print_success,"Build complete: $(APP_BINARY)"); \
		echo "  Commit: $(GIT_COMMIT)"; \
		echo "  Date:   $(BUILD_DATE)"; \
	else \
		$(call print_error,"Build failed!"); \
		exit 1; \
	fi

run: build ## Run the application
	@$(call print_header,"Running application...")
	@"$(APP_BINARY)"

dev: | $(TOOLS_DIR) $(TMP_DIR) ## Run the application with air
	@if [ ! -f "$(AIR)" ]; then \
		$(call print_warning,"Air not found. Auto-installing..."); \
		$(MAKE) -s setup-air; \
	fi
	@$(call print_header,"Running application with air...")
	@"$(AIR)"


# --- Testing & Coverage ---
test: ## Run tests (fast, no race)
	@$(call print_header,"Running tests...")
	@if $(GO) test -v ./...; then \
		$(call print_success,"Tests passed!"); \
	else \
		$(call print_error,"Tests failed!"); \
		exit 1; \
	fi

test-race: ## Run tests with race detector
	@$(call print_header,"Running tests with race detector...")
	@if $(GO) test -v -race ./...; then \
		$(call print_success,"Race tests passed!"); \
	else \
		$(call print_error,"Tests failed!"); \
		exit 1; \
	fi

test-coverage: | $(COVERAGE_DIR) ## Run tests with coverage
	@$(call print_header,"Generating coverage report...")
	@if $(GO) test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...; then \
		$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html; \
		$(call print_success,"Report saved: $(COVERAGE_DIR)/coverage.html"); \
	else \
		$(call print_error,"Tests failed!"); \
		exit 1; \
	fi


# --- Code Quality ---
lint: | $(TOOLS_DIR) ## Run linter
	@$(call print_header,"Running golangci-lint...")
	@if [ ! -f "$(LINT)" ]; then \
		$(call print_warning,"Golangci-lint not found. Auto-installing..."); \
		$(MAKE) -s setup-golangci-lint; \
	fi
	@if "$(LINT)" run ./...; then \
		$(call print_success,"Lint checks passed!"); \
	else \
		$(call print_error,"Lint checks failed!"); \
		exit 1; \
	fi

format: ## Run go fmt
	@$(call print_header,"Running code formatter...")
	@$(GO) fmt ./...
	@$(call print_success,"Code formatted!")

vuln: | $(TOOLS_DIR) ## Run security scan
	@$(call print_header,"Checking vulnerabilities...")
	@if [ ! -f "$(VULN)" ]; then \
		$(call print_warning,"Govulncheck not found. Auto-installing..."); \
		$(MAKE) -s setup-govulncheck; \
	fi
	@if "$(VULN)" ./...; then \
		$(call print_success,"Vulnerability checks passed!"); \
	else \
		$(call print_error,"Vulnerability checks failed!"); \
		exit 1; \
	fi


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


# --- Tooling ---
setup-golangci-lint: $(TOOLS_DIR) ## Install GolangCI-Lint
	@if [ -f "$(LINT)" ]; then \
		$(call print_warning,"golangci-lint already installed at $(LINT)"); \
	else \
		$(call print_header,"Installing golangci-lint@$(LINT_VERSION)..."); \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(TOOLS_DIR)" $(LINT_VERSION) 2>/dev/null; \
		if [ ! -f "$(LINT)" ]; then \
			$(call print_error,"Failed to install golangci-lint!"); \
			exit 1; \
		fi; \
		$(call print_success,"Golangci-lint installed successfully!"); \
	fi

setup-govulncheck: $(TOOLS_DIR) ## Install Govulncheck
	@if [ -f "$(VULN)" ]; then \
		$(call print_warning,"govulncheck already installed at $(VULN)"); \
	else \
		$(call print_header,"Installing govulncheck@$(VULN_VERSION)..."); \
		GOBIN="$(TOOLS_DIR)" $(GO) install golang.org/x/vuln/cmd/govulncheck@$(VULN_VERSION); \
		if [ ! -f "$(VULN)" ]; then \
			$(call print_error,"Failed to install govulncheck!"); \
			exit 1; \
		fi; \
		$(call print_success,"Govulncheck installed successfully!"); \
	fi

setup-air: $(TOOLS_DIR) ## Install Air
	@if [ -f "$(AIR)" ]; then \
		$(call print_warning,"air already installed at $(AIR)"); \
	else \
		$(call print_header,"Installing air@$(AIR_VERSION)..."); \
		GOBIN="$(TOOLS_DIR)" $(GO) install github.com/air-verse/air@$(AIR_VERSION); \
		if [ ! -f "$(AIR)" ]; then \
			$(call print_error,"Failed to install air!"); \
			exit 1; \
		fi; \
		$(call print_success,"Air installed successfully!"); \
	fi

setup-all: setup-golangci-lint setup-govulncheck setup-air ## Install all tools


# --- Cleanup ---
clean: ## Clean directories (bin, coverage, tmp)
	@$(call print_header,"Cleaning directories...")
	@rm -rf "$(BIN_DIR)" "$(COVERAGE_DIR)" "$(TMP_DIR)"
	@$(call print_success,"Directories cleaned")

clean-cache: ## Clean Go cache (cache, testcache)
	@$(call print_header,"Cleaning Go cache...")
	@$(GO) clean -cache -testcache
	@$(call print_success,"Go cache cleaned")

clean-vendor: ## Clean vendor directory
	@$(call print_header,"Cleaning vendor directory...")
	@rm -rf vendor
	@$(call print_success,"Vendor directory cleaned")

clean-all: clean clean-cache clean-vendor ## Deep clean everything
