SHELL       := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

# Phony targets
.PHONY: \
	help \
	set-env \
	build run dev \
	test test-watch test-bench test-coverage \
	profile-cpu profile-mem profile-trace \
	lint format vuln \
	deps-check deps-tidy deps-vendor deps-upgrade \
	migrate-create migrate-up migrate-down migrate-status migrate-drop \
	docker-up docker-down docker-logs docker-ps docker-stats docker-shell \
	install-lefthook \
	clean-bin clean-tmp clean-coverage clean-cache clean-vendor clean-migrations clean-docker clean-all

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
GO      := go
COMPOSE := docker compose

# Paths
PROJECT_ROOT   := $(shell pwd)
COVERAGE_DIR   := $(PROJECT_ROOT)/coverage
TMP_DIR        := $(PROJECT_ROOT)/tmp
BIN_DIR        := $(PROJECT_ROOT)/bin
TOOLS_DIR      := $(BIN_DIR)/tools
MIGRATIONS_DIR := $(PROJECT_ROOT)/migrations

# App
APP_NAME     := app
APP_BINARY   := $(BIN_DIR)/$(APP_NAME)

# Tool Versions
LINT_VERSION      := v2.7.2
VULN_VERSION      := v1.1.4
AIR_VERSION       := v1.63.4
LEFTHOOK_VERSION  := v2.0.13
GOTESTSUM_VERSION := v1.13.0
BENCHSTAT_VERSION := latest
PPROF_VERSION     := latest
MIGRATE_VERSION   := v4.19.1

# Tool Binaries
LINT      := $(TOOLS_DIR)/golangci-lint
VULN      := $(TOOLS_DIR)/govulncheck
AIR       := $(TOOLS_DIR)/air
LEFTHOOK  := $(TOOLS_DIR)/lefthook
GOTESTSUM := $(TOOLS_DIR)/gotestsum
BENCHSTAT := $(TOOLS_DIR)/benchstat
PPROF     := $(TOOLS_DIR)/pprof
MIGRATE   := $(TOOLS_DIR)/migrate

# Environment file
-include .env
export

# Database URL for migrations
DATABASE_URL ?= postgres://$(GO_AUTH_DATABASE_USER):$(GO_AUTH_DATABASE_PASSWORD)@$(GO_AUTH_DATABASE_HOST):$(GO_AUTH_DATABASE_PORT)/$(GO_AUTH_DATABASE_NAME)?sslmode=$(GO_AUTH_DATABASE_SSL_MODE)

# Test Variables
TEST_PKG   ?= ./...
TEST_RUN   ?= .
TEST_ARGS  ?=
TEST_BENCH ?= .
TEST_COUNT ?= 10

# Build Metadata
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date '+%Y-%m-%d-%H:%M:%S')
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -s -w -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)

# Default command
.DEFAULT_GOAL := help

# Directories
$(BIN_DIR) $(TOOLS_DIR) $(COVERAGE_DIR) $(TMP_DIR):
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

$(AIR): | $(TOOLS_DIR)
	@$(call print_header,"Installing air@$(AIR_VERSION)...")
	@GOBIN="$(TOOLS_DIR)" $(GO) install github.com/air-verse/air@$(AIR_VERSION)
	@$(call print_success,"air installed successfully!")

$(LEFTHOOK): | $(TOOLS_DIR)
	@$(call print_header,"Installing lefthook@$(LEFTHOOK_VERSION)...")
	@GOBIN="$(TOOLS_DIR)" $(GO) install github.com/evilmartians/lefthook/v2@$(LEFTHOOK_VERSION)
	@$(call print_success,"lefthook installed successfully!")

$(GOTESTSUM): | $(TOOLS_DIR)
	@$(call print_header,"Installing gotestsum@$(GOTESTSUM_VERSION)...")
	@GOBIN="$(TOOLS_DIR)" $(GO) install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)
	@$(call print_success,"gotestsum installed successfully!")

$(BENCHSTAT): | $(TOOLS_DIR)
	@$(call print_header,"Installing benchstat@$(BENCHSTAT_VERSION)...")
	@GOBIN="$(TOOLS_DIR)" $(GO) install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_VERSION)
	@$(call print_success,"benchstat installed successfully!")

$(PPROF): | $(TOOLS_DIR)
	@$(call print_header,"Installing pprof@$(PPROF_VERSION)...")
	@GOBIN="$(TOOLS_DIR)" $(GO) install github.com/google/pprof@$(PPROF_VERSION)
	@$(call print_success,"pprof installed successfully!")

$(MIGRATE): | $(TOOLS_DIR)
	@$(call print_header,"Installing migrate@$(MIGRATE_VERSION)...")
	@GOBIN="$(TOOLS_DIR)" $(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	@$(call print_success,"migrate installed successfully!")

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


# --- Set Environment ---
set-env: ## Create .env file from example
	@if [ -f .env ]; then \
		$(call print_warning,.env already exists, skipping); \
	else \
		cp .env.example .env; \
		$(call print_success,.env file created from .env.example); \
	fi


# --- Build & Run ---
build: | $(BIN_DIR) ## Build the binary
	@$(call print_header,"Building binary...")
	@$(GO) build -o "$(APP_BINARY)" -ldflags '$(LDFLAGS)' ./cmd/app
	@$(call print_success,"Build complete: $(APP_BINARY)")

run: build ## Run the application
	@$(call print_header,"Starting application...")
	@"$(APP_BINARY)"

dev: $(AIR) ## Run with live reload (Air)
	@$(call print_header,"Starting Air...")
	@"$(AIR)"


# --- Testing & Coverage ---
test: $(GOTESTSUM) ## Run tests (Use TEST_PKG=./path, TEST_RUN="Func/CaseName", TEST_ARGS="flags...")
	@$(call print_header,"Running tests on $(TEST_PKG)")
	@"$(GOTESTSUM)" --format-hide-empty-pkg --format testdox --format-icons hivis \
		-- \
		-run "$(TEST_RUN)" \
		$(TEST_PKG) $(TEST_ARGS)
	@$(call print_success,"Tests passed!")

test-watch: $(GOTESTSUM) ## Run tests with watch (Same opts as 'test')
	@$(call print_header,"Running tests with watch on $(TEST_PKG)")
	@"$(GOTESTSUM)" --watch --format-hide-empty-pkg --format testdox --format-icons hivis \
		-- \
		-run "$(TEST_RUN)" \
		$(TEST_PKG) $(TEST_ARGS)

test-bench: $(BENCHSTAT) | $(TMP_DIR) ## Run benchmarks (Use TEST_PKG=./path, TEST_BENCH="Func/CaseName", TEST_COUNT=10, TEST_ARGS="flags...")
	@$(call print_header,"Running benchmark tests on $(TEST_PKG)")
	@$(GO) test \
		-run=^$$ \
		-benchmem \
		-bench="$(TEST_BENCH)" \
		-count=$(TEST_COUNT) \
		$(TEST_PKG) \
		$(TEST_ARGS) \
		| tee $(TMP_DIR)/bench.txt
	@echo ""
	@$(call print_header,"Calculating statistics...")
	@"$(BENCHSTAT)" $(TMP_DIR)/bench.txt
	@$(call print_success,"Statistics calculated: $(TMP_DIR)/bench.txt")

test-coverage: | $(COVERAGE_DIR) ## Generate coverage report (Same opts as 'test')
	@$(call print_header,"Generating coverage for $(TEST_PKG)")
	@$(GO) test \
		-coverprofile=$(COVERAGE_DIR)/coverage.out \
		-run "$(TEST_RUN)" \
		$(TEST_PKG) $(TEST_ARGS)
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@$(call print_success,"Report generated: $(COVERAGE_DIR)/coverage.html")


# --- Profiling ---
profile-cpu: $(PPROF) | $(TMP_DIR) ## Generate CPU profile (Use TEST_PKG=./path, TEST_BENCH="Func/CaseName", TEST_COUNT=10)
	@if [ "$(TEST_PKG)" = "./..." ]; then \
		$(call print_error,"Please specify a single package. Example: TEST_PKG=./path/to/pkg"); \
		exit 1; \
	fi
	@$(call print_header,"Generating CPU profile...")
	@$(GO) test -run=^$$ -bench="$(TEST_BENCH)" \
		-count=$(TEST_COUNT) \
		-o $(TMP_DIR)/cpu.test \
		-cpuprofile=$(TMP_DIR)/cpu.pprof \
		$(TEST_PKG)
	@$(call print_success,"CPU profile generated: $(TMP_DIR)/cpu.pprof")
	@echo "Press Ctrl+C to stop server"
	@$(PPROF) -http=:3000 $(TMP_DIR)/cpu.pprof

profile-mem: $(PPROF) | $(TMP_DIR) ## Generate Memory profile (Same opts as 'profile-cpu')
	@if [ "$(TEST_PKG)" = "./..." ]; then \
		$(call print_error,"Please specify a single package. Example: TEST_PKG=./path/to/pkg"); \
		exit 1; \
	fi
	@$(call print_header,"Generating Memory profile...")
	@$(GO) test -run=^$$ -bench="$(TEST_BENCH)" \
		-count=$(TEST_COUNT) \
		-o $(TMP_DIR)/mem.test \
		-memprofile=$(TMP_DIR)/mem.pprof \
		$(TEST_PKG)
	@$(call print_success,"Memory profile generated: $(TMP_DIR)/mem.pprof")
	@echo "Press Ctrl+C to stop server"
	@$(PPROF) -http=:3001 $(TMP_DIR)/mem.pprof

profile-trace: | $(TMP_DIR) ## Generate execution trace (Same opts as 'profile-cpu')
	@if [ "$(TEST_PKG)" = "./..." ]; then \
		$(call print_error,"Please specify a single package. Example: TEST_PKG=./path/to/pkg"); \
		exit 1; \
	fi
	@$(call print_header,"Generating execution trace...")
	@$(GO) test -run=^$$ -bench="$(TEST_BENCH)" \
		-count=$(TEST_COUNT) \
		-o $(TMP_DIR)/trace.test \
		-trace=$(TMP_DIR)/trace.out \
		$(TEST_PKG)
	@$(call print_success,"Trace generated: $(TMP_DIR)/trace.out")
	@echo "Ctrl+C to stop server"
	@$(GO) tool trace $(TMP_DIR)/trace.out


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


# --- Migrations ---
migrate-create: $(MIGRATE) ## Create a new migration (Use NAME=my_migration)
	@test -n "$(NAME)" || ($(call print_error,"Usage: make migrate-create NAME=description") && exit 1)
	@$(call print_header,"Creating migration: $(NAME)")
	@"$(MIGRATE)" create -ext sql -dir "$(MIGRATIONS_DIR)" -seq "$(NAME)"
	@$(call print_success,"Migration files created in $(MIGRATIONS_DIR)!")

migrate-up: $(MIGRATE) ## Apply all pending migrations
	@$(call print_header,"Running migrations up...")
	@"$(MIGRATE)" -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" up
	@$(call print_success,"Migrations applied!")

migrate-down: $(MIGRATE) ## Rollback migrations (Use N=1 for number of steps, default 1)
	@$(call print_header,"Running migrations down - N=$(or $(N),1)")
	@"$(MIGRATE)" -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" down $(or $(N),1)
	@$(call print_success,"Migrations rolled back!")

migrate-status: $(MIGRATE) ## Show migration version
	@$(call print_header,"Migration status")
	@"$(MIGRATE)" -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" version

migrate-drop: $(MIGRATE) ## Drop all tables (use FORCE=1 to confirm)
	@$(call print_header,"Drop all tables")
	@if [ "$(FORCE)" != "1" ]; then \
		$(call print_warning,This will drop all tables. Set FORCE=1 to proceed.); \
		exit 1; \
	fi
	@"$(MIGRATE)" -path "$(MIGRATIONS_DIR)" -database "$(DATABASE_URL)" drop
	@$(call print_success,"Database schema dropped!")


# --- Docker ---
docker-up: ## Start Docker services (Postgres) in background
	@$(call print_header,"Starting Docker services...")
	@$(COMPOSE) up -d --wait
	@$(call print_success,"Docker services started!")

docker-down: ## Stop and remove Docker containers, networks
	@$(call print_header,"Stopping Docker services...")
	@$(COMPOSE) down --remove-orphans
	@$(call print_success,"Docker services stopped!")

docker-logs: ## Tail logs from Docker services (Use SERVICE=postgres for specific service)
	@$(COMPOSE) logs -f $(if $(SERVICE),$(SERVICE),)

docker-ps: ## List running Docker Compose services
	@$(COMPOSE) ps

docker-stats: ## Show Docker resource usage statistics
	@$(COMPOSE) stats

docker-shell: ## Open shell in Docker container (Use SERVICE=postgres for specific service)
	@if [ -z "$(SERVICE)" ]; then \
		$(call print_warning,Usage: make docker-shell SERVICE=<service>); \
		exit 1; \
	fi; \
	$(COMPOSE) exec $(SERVICE) /bin/sh


# --- Lefthook ---
install-lefthook: $(LEFTHOOK) ## Install lefthook and configure them (pre-commit, commit-msg, pre-push)
	@if [ -d ".git" ]; then \
		"$(LEFTHOOK)" install; \
		$(call print_success,Lefthook installed and configured!); \
	else \
		$(call print_warning,Not a git repo, skipping hook configuration!); \
	fi


# --- Cleanup ---
clean-bin: ## Remove binary files
	@$(call print_header,"Cleaning binary files...")
	@if [ -f "$(LEFTHOOK)" ]; then "$(LEFTHOOK)" uninstall; fi
	@rm -rf "$(BIN_DIR)"
	@$(call print_success,"Binary files cleaned!")

clean-tmp: ## Remove temporary files
	@$(call print_header,"Cleaning temporary files...")
	@rm -rf "$(TMP_DIR)"
	@$(call print_success,"Temporary files cleaned!")

clean-coverage: ## Remove coverage files
	@$(call print_header,"Cleaning coverage files...")
	@rm -rf "$(COVERAGE_DIR)"
	@$(call print_success,"Coverage files cleaned!")

clean-cache: ## Remove Go cache (cache, testcache)
	@$(call print_header,"Cleaning Go cache...")
	@$(GO) clean -cache -testcache
	@$(call print_success,"Go cache cleaned!")

clean-vendor: ## Remove vendor directory
	@$(call print_header,"Cleaning vendor directory...")
	@rm -rf vendor
	@$(call print_success,"Vendor directory cleaned!")

clean-migrations: ## Remove migrations directory
	@$(call print_header,"Cleaning migrations directory...")
	@rm -rf "$(MIGRATIONS_DIR)"
	@$(call print_success,"Migrations directory cleaned!")

clean-docker: ## Remove Docker containers, networks, volumes, and images
	@$(call print_header,"Cleaning Docker...")
	@$(COMPOSE) down --volumes --remove-orphans --rmi all
	@$(call print_success,"Docker cleaned!")

clean-all: clean-bin clean-tmp clean-coverage clean-cache clean-vendor clean-migrations clean-docker ## Deep clean everything
