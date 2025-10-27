# Versioner CLI Tool - Implementation Plan

## Project Overview
Building a Go-based CLI tool to send deployment events to an existing Versioner deployment tracking API. This CLI will provide broad compatibility across multiple CI/CD systems and deployment tools.

## Context
- **Existing**: Custom GitHub Action already built and working
- **API**: Versioner deployment tracking API already exists and operational (see `docs/api-contract.md`)
- **Goal**: Enable build and deployment event tracking from any CI/CD system or deployment tool
- **Target Endpoints**:
  - `POST /build-events/` - Track CI/CD build lifecycle events (auto-creates products and versions)
  - `POST /deployment-events/` - Track deployment lifecycle events (auto-creates products, versions, and environments)

## Why Go?
- Compiles to single static binary (no runtime dependencies)
- Easy cross-platform compilation (Linux, macOS, Windows)
- Simple distribution model
- Gentle learning curve from Python
- Industry standard for CLI tools in DevOps space

## Target Use Cases
1. **CI/CD Systems**: Jenkins, GitLab CI, CircleCI, Bitbucket Pipelines, Azure DevOps
2. **Infrastructure as Code**: Terraform/Terragrunt runs
3. **Deployment Tools**: Rundeck, custom scripts
4. **Manual Deployments**: Ad-hoc shell execution

## CLI Design Requirements

### Basic Command Structure
```bash
# Track a build event
versioner track build \
  --product=api-service \
  --version=1.2.3 \
  --status=completed

# Track a deployment event
versioner track deployment \
  --product=api-service \
  --environment=production \
  --version=1.2.3 \
  --status=success
```

### Key Features
- **Simple, intuitive command structure**
- **Environment variable support** for API keys/config (e.g., `VERSIONER_API_KEY`)
- **Multiple input methods**: CLI flags, env vars, config file
- **Dry-run mode** for testing
- **Good error messages** with actionable guidance
- **Exit codes** that follow Unix conventions (0=success, non-zero=failure)
- **Optional verbose/debug output**

### Required Parameters

**For `track build`:**
- Product/application name (`--product`)
- Version string (`--version`) - e.g., 1.2.3, build number, or commit SHA
- Status (`--status`, optional) - defaults to `completed`

**For `track deployment`:**
- Product/application name (`--product`)
- Environment name (`--environment`) - e.g., production, staging, dev
- Version string (`--version`) - e.g., 1.2.3, build number, or commit SHA
- Status (`--status`, optional) - defaults to `success`

### Status Values

Both build and deployment events support the same **5 canonical statuses**:

**Canonical values** (recommended):
- `pending` - Queued/scheduled but not started
- `started` - Currently executing
- `completed` - Successfully finished
- `failed` - Failed with errors
- `aborted` - Cancelled or skipped

**Accepted aliases** (normalized by API):
- For `pending`: `queued`, `scheduled`
- For `started`: `in_progress`, `init`, `building`/`deploying`
- For `completed`: `success`, `complete`, `finished`, `built`/`deployed`
- For `failed`: `fail`, `failure`, `error`
- For `aborted`: `abort`, `cancelled`, `cancel`, `skipped`

**Note**: The API automatically normalizes status values. In verbose mode, the CLI will show which canonical status your input maps to.

### Optional Parameters

**Common to both commands:**
- Versioner API endpoint (`--api-url` or `VERSIONER_API_URL`)
- Versioner API key (`--api-key` or `VERSIONER_API_KEY`) - **prefer env var for security**
- Source system (`--source-system`) - e.g., github, jenkins, gitlab, circleci
- Build number (`--build-number`)
- SCM commit SHA (`--scm-sha`) - full 40-character Git hash
- SCM repository (`--scm-repository`) - e.g., owner/repo
- Build URL (`--build-url`) - link to CI/CD build
- Invoke/run ID (`--invoke-id`) - CI system run identifier
- Additional metadata (`--metadata`) - JSON string or key=value pairs

**For `track build` only:**
- SCM branch (`--scm-branch`) - Git branch name
- Built by (`--built-by`, `--built-by-email`, `--built-by-name`)
- Started at (`--started-at`) - ISO 8601 format
- Completed at (`--completed-at`) - ISO 8601 format

**For `track deployment` only:**
- Deployed by (`--deployed-by`, `--deployed-by-email`, `--deployed-by-name`)
- Completed at (`--completed-at`) - ISO 8601 format

### Configuration
Support multiple configuration sources (in priority order):
1. CLI flags (highest priority)
2. Environment variables (`VERSIONER_*` prefix)
3. CI/CD auto-detection (see `docs/cicd-env-vars.md`)
4. Config file (`~/.versioner/config.yaml` or `.versioner.yaml`)
5. Defaults

### Exit Codes
Follow Unix conventions for CI/CD integration:
- `0` - Success (event recorded)
- `1` - General error (invalid arguments, config issues)
- `2` - API error (network failure, authentication, server error)
- `3` - Validation error (invalid status value, missing required fields)

### Security Considerations
- **API Key Storage**: Config file should have `0600` permissions
- **CLI Flag Warning**: Warn if `--api-key` is used (visible in process list)
- **Recommended**: Use `VERSIONER_API_KEY` environment variable or config file
- **User-Agent**: Include CLI version in requests: `versioner/1.0.0 (github)`

## Distribution Strategy

### Phase 1: Initial Release
1. **GitHub Releases** with pre-built binaries
   - linux-amd64, linux-arm64
   - darwin-amd64 (Intel Mac)
   - darwin-arm64 (Apple Silicon)
   - windows-amd64
2. **Simple install script**: `curl -L https://install.versioner.io | sh`

### Phase 2: Package Managers
1. **Homebrew tap** (priority - very common for Go CLIs)
   - Create `homebrew-versioner` repository
   - Users: `brew install versioner-io/versioner`
2. **Direct downloads** from GitHub Releases

### Phase 3: Linux Package Managers (if demand exists)
1. APT repository (Debian/Ubuntu)
2. YUM/DNF repository (RHEL/Fedora/CentOS)

## Recommended Tooling

### GoReleaser
Use GoReleaser to automate:
- Cross-platform binary compilation
- GitHub Release creation
- Homebrew formula generation
- Debian/RPM package creation (later)
- Changelog generation

Benefits: Single config file handles entire release process

### Go Libraries to Consider
- **cobra**: CLI framework (used by kubectl, hugo, etc.)
- **viper**: Configuration management (flags, env vars, config files)
- **go-pretty**: Table output for status commands
- Standard library `net/http` for API calls (no need for heavy dependencies)

## Implementation Phases

### Phase 1: Core CLI (MVP)
- [x] Basic CLI structure with cobra
- [x] Core `track` command with `build` and `deployment` subcommands
  - [x] `track build` - sends to `/build-events/`
  - [x] `track deployment` - sends to `/deployment-events/`
  - [x] Subcommand-specific required field validation
- [x] HTTP client to send events to API
  - [x] Bearer token authentication
  - [x] Retry logic: 3 attempts with exponential backoff (1s, 2s, 4s)
  - [x] 30-second timeout per request
  - [x] Retry on 5xx errors and network failures
  - [x] Do not retry on 4xx errors (except 429)
- [x] Environment variable configuration (`VERSIONER_*` prefix)
- [x] CI/CD auto-detection (GitHub Actions, GitLab CI, Jenkins, CircleCI, Bitbucket, Azure DevOps, Travis CI)
  - [x] Auto-detect CI system
  - [x] Map CI variables to appropriate fields
- [x] Error handling with proper exit codes (0, 1, 2)
- [x] Security: Warn if API key passed via CLI flag
- [x] Versioning: `versioner version` command
- [x] Unit tests for CI detection
- [x] Status value validation and normalization warnings (verbose mode)
- [ ] Integration tests with mock API server (deferred to Phase 2)
- [ ] README with usage examples (in progress)

### Phase 2: Enhanced UX
- [ ] Config file support (YAML)
  - [ ] `~/.versioner/config.yaml` and `.versioner.yaml`
  - [ ] Set file permissions to 0600 on creation
- [ ] Dry-run mode (`--dry-run`) - show what would be sent without sending
- [ ] Verbose/debug output (`--verbose`, `--debug`)
  - [ ] Show auto-detected values
  - [ ] Show which config source was used for each field
  - [ ] HTTP request/response logging
- [ ] JSON output mode (`--output=json`) for parsing
- [ ] Better error messages with actionable guidance
- [ ] Additional CI/CD system support (Azure DevOps, Travis CI, Argo CD)

### Phase 3: Distribution
- [ ] GoReleaser configuration
- [ ] GitHub Actions workflow for releases
- [ ] Homebrew tap setup
- [ ] Installation documentation

### Phase 4: Additional Commands (Nice-to-have)
- [ ] `versioner init` - Initialize config file
- [ ] `versioner validate` - Validate configuration/connectivity
- [ ] `versioner status` - Query recent deployments


## Integration Examples to Document

### GitHub Actions - Build
```yaml
- name: Track build
  run: |
    versioner track build \
      --product=api-service \
      --version=${{ github.sha }} \
      --status=completed
```

### GitHub Actions - Deployment
```yaml
- name: Track deployment
  run: |
    versioner track deployment \
      --product=api-service \
      --environment=production \
      --version=${{ github.sha }} \
      --status=success
```

### Jenkins Pipeline - Build
```groovy
sh 'versioner track build --product=api --version=${BUILD_NUMBER} --status=completed'
```

### Jenkins Pipeline - Deployment
```groovy
sh 'versioner track deployment --product=api --environment=prod --version=${BUILD_NUMBER} --status=success'
```

### GitLab CI - Build
```yaml
build:
  script:
    - make build
  after_script:
    - versioner track build --product=api --version=$CI_COMMIT_SHA --status=completed
```

### GitLab CI - Deployment
```yaml
deploy:
  script:
    - make deploy
  after_script:
    - versioner track deployment --product=api --environment=$CI_ENVIRONMENT_NAME --version=$CI_COMMIT_SHA --status=success
```

### Terraform/Terragrunt - Deployment
```hcl
resource "null_resource" "deployment_tracking" {
  provisioner "local-exec" {
    command = "versioner track deployment --product=infra --environment=${var.environment} --version=${var.version} --status=success"
  }
}
```

## Testing Strategy

### Unit Tests
- Configuration parsing and priority
- CI/CD environment variable detection and mapping
- Input validation (status values, field lengths)
- Exit code logic

### Integration Tests
- Mock API server for testing HTTP client
- Retry logic with simulated failures
- Authentication handling
- Error response parsing

### E2E Tests (CI)
- Real API calls in GitHub Actions workflow
- Cross-platform binary testing (Linux, macOS, Windows)

### Test Commands (justfile)
```bash
just run_tests          # Run all tests
just test_unit          # Unit tests only
just test_integration   # Integration tests only
```

## Questions to Resolve During Implementation
1. ✅ API endpoint structure - **RESOLVED**: See `docs/api-contract.md`
2. Rate limiting considerations - **TBD**: Confirm with API team
3. ✅ Retry logic - **RESOLVED**: 3 retries, exponential backoff, included in Phase 1
4. Should CLI support batch operations? - **DEFERRED**: Not in MVP
5. Telemetry/analytics for CLI usage? - **DEFERRED**: Skip for MVP, consider later if needed
   - If added: Must be opt-in, easy to disable, minimal data collection
   - Potential use: Track CLI version adoption, most-used CI systems, error rates
   - Privacy: No sensitive data (API keys, repo names, etc.)

## Success Metrics
- Installation friction: Can users install in under 1 minute?
- Usage friction: Can users send their first event in under 2 minutes?
- Cross-platform: Does it work seamlessly on Linux, macOS, Windows?
- Documentation: Can users figure it out without asking questions?

## Reference Tools for Inspiration
- `gh` (GitHub CLI) - excellent UX and distribution
- `kubectl` - standard cobra/viper usage
- `terraform` - good cross-platform binary distribution
- `stripe` CLI - good examples of API-focused CLI tools

## Documentation Files
- `README.md` - Project overview, installation, quick start
- `docs/development-plan.md` - This file (canonical project plan)
- `docs/api-contract.md` - Complete API specification for `/deployment-events/`
- `docs/cicd-env-vars.md` - CI/CD environment variable reference and mapping
