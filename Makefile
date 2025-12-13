.PHONY: help build run dev test test-coverage lint vuln vet deps-check deps-upgrade deps-tidy deps-vendor install-tools clean clean-cache clean-deps clean-all

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

# Configuration & Variables
GO           := go
APP_NAME     := app
PROJECT_ROOT := $(shell pwd)
LOCAL_BIN    := $(PROJECT_ROOT)/bin
TOOLS_BIN    := $(LOCAL_BIN)/tools
APP_BINARY   := $(LOCAL_BIN)/$(APP_NAME)

# Tool Versions
AIR_VERSION      := v1.63.4
LINT_VERSION     := v2.7.2
VULN_VERSION     := v1.1.4
LEFTHOOK_VERSION := v2.0.11

# Tool Paths
AIR      := $(TOOLS_BIN)/air
LINT     := $(TOOLS_BIN)/golangci-lint
VULN     := $(TOOLS_BIN)/govulncheck
LEFTHOOK := $(TOOLS_BIN)/lefthook

# Build Metadata
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_DIRTY  := $(shell git diff --quiet || echo "+CHANGES")
BUILD_DATE := $(shell date '+%Y-%m-%d-%H:%M:%S')
LDFLAGS    := -w -s -X main.Commit=$(GIT_COMMIT)$(GIT_DIRTY) -X main.BuildDate=$(BUILD_DATE)

# --- Help ---
help: ## Show this help message
	@echo "$(BOLD)Available Commands:$(RESET)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# --- Build & Run ---
build: ## Build the application
	@$(call print_header,"Building binary...")
	@mkdir -p "$(LOCAL_BIN)"
	@if $(GO) build -o "$(APP_BINARY)" -ldflags '$(LDFLAGS)' ./cmd/app; then \
		$(call print_success,"Build complete: $(APP_BINARY)"); \
		echo "  Commit: $(GIT_COMMIT)"; \
		echo "  Date:   $(BUILD_DATE)"; \
	else \
		$(call print_error,"Build failed!"); \
		exit 1; \
	fi

run: ## Run the application
	@$(call print_header,"Running application...")
	@if [ ! -f "$(APP_BINARY)" ]; then \
		$(call print_warning,"Application binary not found. Building..."); \
		$(MAKE) -s build; \
	fi
	@"$(APP_BINARY)"

dev: ## Start development server (Air)
	@$(call print_header,"Starting Air server...")
	@if [ ! -f "$(AIR)" ]; then \
		$(call print_warning,"Air not found. Auto-installing..."); \
		$(MAKE) -s install-tools; \
	fi
	@"$(AIR)"

# --- Testing & Coverage ---
test: ## Run tests
	@$(call print_header,"Running tests...")
	@if $(GO) test -v -race -parallel 4 ./...; then \
		$(call print_success,"Tests passed!"); \
	else \
		$(call print_error,"Tests failed!"); \
		exit 1; \
	fi

test-coverage: ## Run tests with coverage
	@$(call print_header,"Generating coverage report...")
	@mkdir -p coverage
	@if $(GO) test -v -race -parallel 4 -coverprofile=coverage/coverage.out ./...; then \
		$(GO) tool cover -html=coverage/coverage.out -o coverage/coverage.html; \
		$(call print_success,"Report saved: coverage/coverage.html"); \
	else \
		$(call print_error,"Tests failed!"); \
		exit 1; \
	fi

# --- Quality Assurance ---
lint: ## Run linter (Check only)
	@$(call print_header,"Running golangci-lint...")
	@if [ ! -f "$(LINT)" ]; then \
		$(call print_warning,"Golangci-lint not found. Auto-installing..."); \
		$(MAKE) -s install-tools; \
	fi
	@if "$(LINT)" run ./...; then \
		$(call print_success,"Lint checks passed!"); \
	else \
		$(call print_error,"Lint checks failed!"); \
		exit 1; \
	fi

vet: ## Run go vet
	@$(call print_header,"Running go vet...")
	@if $(GO) vet ./...; then \
		$(call print_success,"Vet passed!"); \
	else \
		$(call print_error,"Vet checks failed!"); \
		exit 1; \
	fi

vuln: ## Run security scan
	@$(call print_header,"Checking vulnerabilities...")
	@if [ ! -f "$(VULN)" ]; then \
		$(call print_warning,"Govulncheck not found. Auto-installing..."); \
		$(MAKE) -s install-tools; \
	fi
	@if "$(VULN)" ./...; then \
		$(call print_success,"Vulnerability checks passed!"); \
	else \
		$(call print_error,"Vulnerability checks failed!"); \
		exit 1; \
	fi

# --- Dependencies ---
deps-check: ## Verify dependencies
	@$(call print_header,"Verifying dependencies...")
	@if $(GO) mod verify; then \
		$(call print_success,"Dependencies verified!"); \
	else \
		$(call print_error,"Dependencies verification failed!"); \
		exit 1; \
	fi

deps-upgrade: ## Upgrade direct dependencies
	@$(call print_header,"Upgrading dependencies...")
	@if $(GO) get -u ./... && $(GO) mod tidy; then \
		$(call print_success,"Dependencies upgraded!"); \
	else \
		$(call print_error,"Dependencies upgrade failed!"); \
		exit 1; \
	fi

deps-tidy: ## Tidy go.mod
	@$(call print_header,"Tidying dependencies...")
	@if $(GO) mod tidy; then \
		$(call print_success,"Dependencies tidied!"); \
	else \
		$(call print_error,"Dependencies tidying failed!"); \
		exit 1; \
	fi

deps-vendor: ## Create vendor directory
	@$(call print_header,"Vendoring dependencies...")
	@if $(GO) mod vendor; then \
		$(call print_success,"Dependencies vendored!"); \
	else \
		$(call print_error,"Dependencies vendoring failed!"); \
		exit 1; \
	fi

# --- Tooling ---
install-tools: ## Install dev tools locally to ./bin
	@$(call print_header,"Installing Tools to $(LOCAL_BIN)...")
	@mkdir -p "$(TOOLS_BIN)"

	@$(call print_header,"Installing air@$(AIR_VERSION)...")
	@GOBIN="$(TOOLS_BIN)" $(GO) install github.com/air-verse/air@$(AIR_VERSION)

	@$(call print_header,"Installing govulncheck@$(VULN_VERSION)...")
	@GOBIN="$(TOOLS_BIN)" $(GO) install golang.org/x/vuln/cmd/govulncheck@$(VULN_VERSION)

	@$(call print_header,"Installing lefthook@$(LEFTHOOK_VERSION)...")
	@GOBIN="$(TOOLS_BIN)" $(GO) install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)

	@$(call print_header,"Installing golangci-lint@$(LINT_VERSION)...")
	@if [ -n "$$(which curl)" ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(TOOLS_BIN)" $(LINT_VERSION) 2>/dev/null; \
	elif [ -n "$$(which wget)" ]; then \
		wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(TOOLS_BIN)" $(LINT_VERSION) 2>/dev/null; \
	else \
		$(call print_error,"curl or wget not found. Please install curl or wget to install Golangci-lint."); \
		exit 1; \
	fi

	# add path to git hooks for local tooling
	@if [ -d ".git" ]; then \
		"$(LEFTHOOK)" install; \
		$(call print_header,"Patching Git hooks for local tooling..."); \
		for hook in pre-commit pre-push commit-msg prepare-commit-msg; do \
			if [ -f ".git/hooks/$$hook" ]; then \
				echo '#!/bin/sh' > ".git/hooks/$$hook.tmp"; \
				echo 'export PATH="$(TOOLS_BIN):$$PATH"' >> ".git/hooks/$$hook.tmp"; \
				tail -n +2 ".git/hooks/$$hook" >> ".git/hooks/$$hook.tmp"; \
				mv ".git/hooks/$$hook.tmp" ".git/hooks/$$hook"; \
				chmod +x ".git/hooks/$$hook"; \
				$(call print_success,"$$hook patched successfully."); \
			else \
				$(call print_error,"$$hook not found."); \
			fi \
		done; \
		$(call print_success,"Git hooks patched successfully."); \
	fi

	@$(call print_success,"All tools installed locally! Ready to rock.")

# --- Cleanup ---
clean: ## Clean build artifacts and local tools (bin, coverage, tmp)
	@$(call print_header,"Cleaning build artifacts...")
	@rm -rf bin coverage tmp
	@$(call print_success,"Build artifacts and local tools cleaned")

clean-cache: ## Clean Go cache (cache, testcache)
	@$(call print_header,"Cleaning Go cache...")
	@$(GO) clean -cache -testcache
	@$(call print_success,"Go cache cleaned")

clean-deps: ## Remove vendor directory
	@$(call print_header,"Removing vendor directory...")
	@rm -rf vendor
	@$(call print_success,"Vendor directory removed")

clean-all: clean clean-cache clean-deps ## Deep clean everything
	@$(call print_success,"All clean operations completed")
