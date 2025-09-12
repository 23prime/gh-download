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
- `task go:test` - Run tests
- `task go:lint` - Run golangci-lint
- `task go:lint:fix` - Auto-fix linting issues
- `task go:fmt` - Format code (via golangci-lint fmt)
- `task go:fix` - Run formatting and auto-fixes
- `task go:check` - Complete check (fix, lint, test, build)

### Project Management

- `task list` (or `task l`) - List all available tasks
- `task check` (or `task c`) - Run all checks (YAML, JSON, Markdown, GitHub Actions, Go)
- `task tag:X.Y.Z` - Create and push a git tag for release

## Architecture

- **Single binary**: The main application is in `main.go`
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

- **Go**: Uses golangci-lint with govet, staticcheck, and errcheck enabled
- **Formatters**: gofmt and goimports are configured
- **Multi-language**: Separate linting pipelines for YAML, JSON, Markdown, and GitHub Actions

## CI/CD

GitHub Actions workflows automatically:

- Run `task go:fix` and commit changes on Go file changes
- Run comprehensive checks on pushes and PRs
- Validate multiple file types (Go, YAML, JSON, Markdown, GitHub Actions)
