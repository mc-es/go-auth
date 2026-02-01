# Agents Guide

This repository is a Go codebase for the `go-auth` project. Use this document
as the operating guide for automated changes.

## Project Layout

- `cmd/app/`: application entrypoint.
- `cmd/commitlint/`: commit lint tool.
- `internal/`: application internals (config, response, app errors, bootstrap).
- `pkg/`: reusable packages (logger adapters, registry, options).

## Go Version

- Go version is defined in `go.mod` (`go 1.25.5`).

## Common Commands

Run from repository root:

- Build: `make build`
- Run: `make run`
- Dev (live reload): `make dev`
- Tests: `make test`
- Coverage: `make test-coverage`
- Benchmarks: `make test-bench`
- Lint: `make lint`
- Format: `make format`
- Vulnerability scan: `make vuln`
- Tidy modules: `make deps-tidy`

Most commands have optional variables (see `Makefile`), for example:
`make test TEST_PKG=./internal/... TEST_RUN="TestName"`.

## Testing Guidance

- Prefer table-driven tests.
- Keep unit tests fast and focused.
- Use `testing.T` with `t.Helper()` for helpers.
- When changing behavior, add tests in the same package.

## Code Style

- Keep functions short and single-purpose.
- Always check and wrap errors with context (`fmt.Errorf("context: %w", err)`).
- Avoid global state; prefer dependency injection via constructors.
- Run `make format` and `make lint` before pushing.

## Dependencies

- Use `go get` to add new dependencies, then run `make deps-tidy`.
- Prefer standard library packages where possible.

## Notes

- `make dev` starts a long-running process (Air). Avoid running it in CI or
  automation unless explicitly needed.
