# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a GitHub CLI extension for downloading files from releases. It's written in Go and uses the `github.com/cli/go-gh/v2` library to interact with GitHub's API.

## Development Setup

This project uses Taskfile and mise for tooling management:

1. **Setup project**: `task setup` (installs all tools via mise)
2. **Run development**: `task go:run`
3. **Check everything**: `task check` (runs all linting, testing, and building)

## Key Commands

### Go Development

- `task go:run` - Run the application
- `task go:build` - Build the binary (outputs `gh-download`)
- `task go:test` or `task go:test:unit` - Run unit tests (internal/ packages only)
- `task go:test:integration` - Run integration tests (tests/ directory, uses real GitHub API)
- `task go:lint` - Run golangci-lint
- `task go:lint:fix` - Auto-fix linting issues
- `task go:fmt` - Format code (via golangci-lint fmt)
- `task go:fix` - Run formatting and auto-fixes
- `task go:check` - Complete check (fix, lint, unit test, build)

### Project Management

- `task list` (or `task l`) - List all available tasks
- `task check` (or `task c`) - Run all checks (YAML, JSON, Markdown, GitHub Actions, Go)
- `task tag:X.Y.Z` - Create and push a git tag for release

## Architecture

### Code Structure

- **Main application**: `main.go` - Entry point with minimal logic
- **Internal packages**: Organized under `internal/` following Go best practices
  - `internal/config/` - CLI argument parsing and configuration
  - `internal/github/` - GitHub API operations with HTTPClient interface abstraction
  - `internal/download/` - Download functionality for assets and archives

### Testing Strategy

- **Unit tests**: Mock-based tests in `internal/` packages with 100% coverage for core logic
- **Integration tests**: End-to-end tests in `tests/` directory that execute the actual CLI with real GitHub API
- **HTTPClient interface**: Enables comprehensive testing without external dependencies

### Development Tools

- **GitHub CLI extension**: Uses `github.com/cli/go-gh/v2` for GitHub API integration
- **Task-based workflow**: All development tasks are managed through Taskfile with modular task files in `tasks/`
- **Multi-language linting**: Separate linting for YAML, JSON, Markdown, GitHub Actions, and Go

## Tool Requirements

The project uses mise (mise.toml) to manage these tools:

- Go 1.24
- golangci-lint v2
- yamllint, markdownlint-cli, biome, actionlint, shellcheck
- fzf for interactive task selection

## Linting Configuration

- **Go**: Uses golangci-lint with govet, staticcheck, errcheck, unparam, unused, and ineffassign enabled
- **Formatters**: gofmt and goimports are configured
- **Multi-language**: Separate linting pipelines for YAML, JSON, Markdown, and GitHub Actions

## Testing Guidelines

### Unit Tests

- Located in `internal/` packages alongside source code
- Use MockHTTPClient for testing GitHub API interactions
- Achieve 100% statement coverage for critical functions
- Focus on pure functions and business logic

### Integration Tests  

- Located in `tests/integration_test.go`
- Execute actual CLI commands via `os/exec`
- Use real GitHub API with stable public repositories (cli/cli)
- Include file download verification and error handling
- Run with `go test ./tests/` or `task go:test:integration`
- Slower execution (~13 seconds) due to network operations

### Test Execution

- **Fast feedback**: `task go:test` runs only unit tests
- **Full validation**: `task go:test:integration` runs end-to-end tests  
- **CI/CD**: Only unit tests are included in `task check` for speed

## CI/CD

GitHub Actions workflows automatically:

- Run `task go:fix` and commit changes on Go file changes
- Run comprehensive checks on pushes and PRs
- Validate multiple file types (Go, YAML, JSON, Markdown, GitHub Actions)
