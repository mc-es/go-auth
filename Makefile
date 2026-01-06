SHELL := /bin/bash

.PHONY: \
	help \
	build run \
	test test-race test-coverage \
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

define print_error
	echo "$(RED)✘ $(1)$(RESET)"
endef

# Commands
GO := go

# Paths
APP_NAME     := app
PROJECT_ROOT := $(shell pwd)
COVERAGE_DIR := $(PROJECT_ROOT)/coverage
BIN_DIR      := $(PROJECT_ROOT)/bin
APP_BINARY   := $(BIN_DIR)/$(APP_NAME)

# Build Metadata
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY  := $(shell git diff --quiet || echo "+CHANGES")
BUILD_DATE := $(shell date '+%Y-%m-%d-%H:%M:%S')
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -s -w -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)

# Default command
.DEFAULT_GOAL := help

# Directories
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

$(COVERAGE_DIR):
	@mkdir -p $(COVERAGE_DIR)


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
	@if $(GO) build -o "$(APP_BINARY)" -ldflags '$(LDFLAGS)' $(PROJECT_ROOT)/cmd/app; then \
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


# --- Testing & Coverage ---
test: ## Run tests (fast, no race)
	@$(call print_header,"Running tests...")
	@$(GO) test -v $(PROJECT_ROOT)/...

test-race: ## Run tests with race detector
	@$(call print_header,"Running tests with race detector...")
	@if $(GO) test -v -race $(PROJECT_ROOT)/...; then \
		$(call print_success,"Race tests passed!"); \
	else \
		$(call print_error,"Tests failed!"); \
		exit 1; \
	fi

test-coverage: | $(COVERAGE_DIR) ## Run tests with coverage
	@$(call print_header,"Generating coverage report...")
	@if $(GO) test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out $(PROJECT_ROOT)/...; then \
		$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html; \
		$(call print_success,"Report saved: $(COVERAGE_DIR)/coverage.html"); \
	else \
		$(call print_error,"Tests failed!"); \
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


# --- Cleanup ---
clean: ## Clean directories (bin, coverage)
	@$(call print_header,"Cleaning directories...")
	@rm -rf "$(BIN_DIR)" "$(COVERAGE_DIR)"
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
