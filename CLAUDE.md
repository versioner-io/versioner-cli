# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Versioner CLI - a Go command-line tool for tracking build and deployment events in CI/CD pipelines. Go 1.24, Cobra, Viper.

## Cross-Repo Context

This repo is part of the Versioner ecosystem. Before starting work:
- Use the `/kanban-markdown` skill or read `versioner-dev-docs/.devtool/features/` for current status and priorities
- Read relevant feature docs from `/Users/phil.austin/projects/versioner-workspace/versioner-dev-docs/specs/`
- Read relevant architecture docs from `/Users/phil.austin/projects/versioner-workspace/versioner-dev-docs/architecture/`
- Update feature file status in `versioner-dev-docs/.devtool/features/` as you complete tasks, add entries to `changelog.md`

## Build & Test Commands

All commands use `just` (run `just` to list all):

```bash
just setup_local_dev    # go mod download + install golangci-lint
just build              # Build binary to bin/versioner
just build_all          # Cross-compile for Linux, macOS, Windows (amd64 + arm64)
just run_tests          # go test -v ./...
just test_unit          # Unit tests only (-short flag)
just test_integration   # Integration tests only
just test_coverage      # Generate HTML coverage report
just lint               # golangci-lint
just fmt                # go fmt + gofmt
just ci                 # fmt -> tests -> lint -> build (full check)
```

## Architecture

### Entry Point & Commands
- `cmd/versioner/main.go` - Entry point, calls `cmd.Execute()`
- `internal/cmd/root.go` - Root command with global flags (--config, --verbose, --debug, --api-url, --api-key)
- `internal/cmd/track_build.go` - `versioner track build` subcommand
- `internal/cmd/track_deployment.go` - `versioner track deployment` subcommand

### Key Modules
- `internal/api/` - HTTP client with retry logic (3 retries, exponential backoff)
- `internal/cicd/detector.go` - Auto-detects 8 CI/CD systems (GitHub Actions, GitLab, Jenkins, CircleCI, Bitbucket, Azure DevOps, Travis CI, Rundeck) via environment variables
- `internal/status/validator.go` - Status value validation and normalization
- `internal/github/annotations.go` - GitHub workflow annotations
- `internal/version/version.go` - Version info injected via ldflags at build time

### Configuration
Config loaded from (in order): `--config` flag, `$HOME/.versioner/config.yaml`, `./config.yaml`. Environment variable prefix: `VERSIONER_` (e.g., `VERSIONER_API_KEY`).

## Conventions

- Follow existing code patterns and style
- Blank lines must be completely empty (no whitespace-only lines)
- Add tests for new features; ensure tests pass before committing
- Prefer editing existing files over creating new ones
