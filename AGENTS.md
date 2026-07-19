# AGENTS.md — versioner-cli

Go CLI for tracking build and deployment events in CI/CD. Go 1.24, Cobra, Viper.

## Cross-repo context

This repo is one of the Versioner app repos.

1. Resolve `versioner-workspace` (nested parent **or** sibling — see that repo's `docs/agent/ECOSYSTEM.md` layout discovery).
2. Read:
   - `docs/agent/ECOSYSTEM.md` (routing / which repo)
   - `docs/agent/CONVENTIONS.md` (branches, PRs, commits)
3. Then follow **this file** for repo-local setup, test, and architecture.

Do **not** use `versioner-dev-docs`, kanban-markdown feature trees, or machine-specific paths like `/Users/phil...`.


## Build & test

```bash
just setup_local_dev    # go mod download + install golangci-lint
just build              # bin/versioner
just build_all          # cross-compile linux/mac/windows amd64+arm64
just run_tests          # go test -v ./...
just test_unit          # -short
just test_integration   # integration only
just test_coverage      # HTML coverage
just lint               # golangci-lint
just fmt                # go fmt + gofmt
just ci                 # fmt -> tests -> lint -> build
```

## Architecture

### Entry & commands

- `cmd/versioner/main.go` → `cmd.Execute()`
- `internal/cmd/root.go` — global flags (`--config`, `--verbose`, `--debug`, `--api-url`, `--api-key`)
- `internal/cmd/track_build.go` — `versioner track build`
- `internal/cmd/track_deployment.go` — `versioner track deployment`

### Key modules

- `internal/api/` — HTTP client, 3 retries, exponential backoff
- `internal/cicd/detector.go` — detects 8 CI systems via env
- `internal/status/validator.go` — status validation/normalization
- `internal/github/annotations.go` — GHA annotations
- `internal/version/version.go` — ldflags version

### Config

Order: `--config` flag → `$HOME/.versioner/config.yaml` → `./config.yaml`.  
Env prefix: `VERSIONER_` (e.g. `VERSIONER_API_KEY`).

## Pre-commit / pre-PR gate (required)

Mirrors GitHub Actions `test.yml` on PRs to `main`/`develop`: **test** (`go test -race`), **lint** (golangci-lint), **matrix build**.

| When | Command | Notes |
|------|---------|--------|
| **Final handoff (required)** | `just ci` | `fmt` → `run_tests` → `lint` → `build` |
| Closest to CI tests | `go test -v -race ./...` | CI adds `-race`; `just run_tests` is plain `go test -v` |
| Lint only | `just lint` | CI runs golangci-lint-action |

**Rules for agents**
1. Run **`just ci`** before commit/push (or at least tests + lint + build if `ci` skips lint when golangci-lint missing — install linter rather than skip).
2. Prefer `go test -v -race ./...` once before handoff if `just ci` does not pass `-race`.
3. Note commands + results on the card/PR.

## Local conventions

- Follow existing patterns
- Blank lines must be completely empty
- Add tests for new behavior; pass before PR
- Prefer editing existing files over creating new ones
