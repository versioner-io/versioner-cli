# Development Guide

Complete guide for developing the Versioner CLI.

## Quick Start

### Initial Setup

```bash
# Clone and setup
git clone <repo-url>
cd versioner-cli
just setup_local_dev
```

### Development Workflow

1. **Make your changes** - Edit files in `internal/` or `cmd/` directories
2. **Run tests** - `just run_tests`
3. **Build** - `just build`
4. **Test locally** - `./bin/versioner track build --product=test --version=1.0.0`
5. **Commit** - `git commit -m "feat: description"`

## Repository Structure

```
versioner-cli/
├── cmd/versioner/              # Application entry point (main.go)
│   └── main.go                 # Calls the root command and exits
│
├── internal/                   # Private application code
│   ├── api/                    # API client for Versioner backend
│   │   ├── client.go           # HTTP client with retry logic
│   │   ├── build.go            # Build event types and API calls
│   │   └── deployment.go       # Deployment event types and API calls
│   │
│   ├── cicd/                   # CI/CD system auto-detection
│   │   ├── detector.go         # Detects CI system and extracts metadata
│   │   └── detector_test.go    # Tests for detection logic
│   │
│   ├── cmd/                    # Cobra command definitions
│   │   ├── root.go             # Root command (versioner)
│   │   ├── version.go          # Version command
│   │   ├── track.go            # Track parent command
│   │   ├── track_build.go      # Track build subcommand
│   │   ├── track_deployment.go # Track deployment subcommand
│   │   ├── metadata.go         # Extra metadata parsing
│   │   └── metadata_test.go    # Metadata tests
│   │
│   └── status/                 # Status value validation
│       ├── validator.go        # Status normalization logic
│       └── validator_test.go   # Tests for status validation
│
├── docs/                       # Documentation
│   ├── api-contract.md         # API specification
│   ├── cicd-env-vars.md        # CI/CD environment variable reference
│   ├── development-plan.md     # Project roadmap
│   └── development-guide.md    # This file!
│
├── bin/                        # Compiled binaries (gitignored)
│   └── versioner               # Built executable
│
├── go.mod                      # Go module definition
├── go.sum                      # Dependency lock file
├── justfile                    # Build commands
└── README.md                   # User-facing documentation
```

## Key Files Explained

### `cmd/versioner/main.go`
**Purpose**: Application entry point  
**What it does**: Calls the root Cobra command and exits with appropriate code

```go
func main() {
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### `internal/cmd/root.go`
**Purpose**: Root command definition and global configuration  
**What it does**:
- Defines the base `versioner` command
- Sets up global flags (`--verbose`, `--api-key`, etc.)
- Initializes Viper for configuration management
- Handles config file loading

**Key frameworks**:
- **Cobra**: CLI framework (like Click or argparse)
- **Viper**: Configuration management (handles env vars, config files, flags)

### `internal/cmd/track_build.go` & `track_deployment.go`
**Purpose**: Subcommand implementations  
**What they do**:
- Define command flags and help text
- Validate required fields
- Auto-detect CI/CD environment
- Build API request payload
- Send request to API
- Handle errors and exit codes

**Flow**:
1. Parse flags and env vars (via Viper)
2. Auto-detect CI/CD values (if available)
3. Validate required fields
4. Create API client
5. Build event payload
6. Send to API
7. Print result

### `internal/api/client.go`
**Purpose**: HTTP client with retry logic  
**What it does**:
- Makes HTTP requests to Versioner API
- Implements retry logic (3 attempts, exponential backoff)
- Handles authentication (Bearer token)
- Parses error responses

**Key features**:
- 30-second timeout
- Retries on 5xx errors and network failures
- No retry on 4xx errors (except 429)

### `internal/cicd/detector.go`
**Purpose**: Auto-detect CI/CD environment  
**What it does**:
- Checks environment variables to identify CI system
- Extracts metadata (repo, SHA, build number, etc.)
- Provides fallback values for CLI flags

**Supported systems**: GitHub Actions, GitLab CI, Jenkins, CircleCI, Bitbucket, Azure DevOps, Travis CI

### `go.mod` & `go.sum`
**Purpose**: Dependency management  
**What they do**:
- `go.mod`: Declares module name and dependencies
- `go.sum`: Cryptographic checksums for dependencies

**Key dependencies**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

## Common Commands

```bash
# Setup/Install dependencies
just setup_local_dev

# Build the CLI
just build

# Run all tests
just run_tests

# Run tests with coverage
just test_coverage

# Format code
just fmt

# Lint code (requires golangci-lint)
just lint

# Clean build artifacts
just clean

# Build for all platforms
just build_all

# Run without building (slower)
just run track build --product=test --version=1.0.0
```

## Testing Locally

### Basic Testing

```bash
# Track a build event
./bin/versioner track build \
  --api-url https://development-api.versioner.io \
  --product=test-app \
  --version=1.0.0 \
  --status=completed

# Track a deployment event
./bin/versioner track deployment \
  --api-url https://development-api.versioner.io \
  --product=test-app \
  --environment=development \
  --version=1.0.0 \
  --status=completed

# With extra metadata
./bin/versioner track build \
  --api-url https://development-api.versioner.io \
  --product=test-app \
  --version=1.0.0 \
  --extra-metadata '{"docker_image": "myorg/app:1.0.0"}'

# With verbose output
./bin/versioner track build \
  --api-url https://development-api.versioner.io \
  --product=test-app \
  --version=1.0.0 \
  --verbose
```

### Using Environment Variables

```bash
# Set API credentials
export VERSIONER_API_URL=https://development-api.versioner.io
export VERSIONER_API_KEY=your-key-here

# Now you can omit --api-url and --api-key flags
./bin/versioner track build --product=test --version=1.0.0
```

## Development Workflow Details

### Adding a New Feature

1. **Add code** in `internal/` package
2. **Add tests** in `*_test.go` file
3. **Run tests**: `just run_tests`
4. **Build**: `just build`
5. **Test manually**: `./bin/versioner ...`
6. **Commit**: `git commit -m "feat: description"`

### Adding a New Command

1. Create new file in `internal/cmd/` (e.g., `validate.go`)
2. Define Cobra command structure
3. Register command in `init()` function
4. Implement command logic in `RunE` function

Example structure:
```go
var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Validate configuration",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Your logic here
        return nil
    },
}

func init() {
    rootCmd.AddCommand(validateCmd)
}
```

### Writing Tests

Test files end with `_test.go` and test functions start with `Test`:

```go
func TestNormalize(t *testing.T) {
    result, _ := Normalize("success")
    if result != "completed" {
        t.Errorf("Expected 'completed', got '%s'", result)
    }
}
```

Run tests:
```bash
# All tests
just run_tests

# Specific package
go test ./internal/status/...

# Specific test
go test ./internal/cmd/... -v -run TestParseExtraMetadata

# With coverage
just test_coverage
```

## Debugging

### Verbose and Debug Modes

```bash
# Verbose output
./bin/versioner track build --product=test --version=1.0.0 --verbose

# Debug mode (shows HTTP requests/responses)
./bin/versioner track build --product=test --version=1.0.0 --debug
```

### Print Debugging

```go
fmt.Fprintf(os.Stderr, "Debug: value=%v\n", myVariable)
```

### Running Specific Tests

```bash
# Run specific test
go test ./internal/cmd/... -v -run TestParseExtraMetadata

# Run with verbose output
go test -v ./internal/status/...

# Check what's in bin/
ls -lh bin/
```

## Code Quality

### Formatting

Go is strict about formatting. Always run:

```bash
just fmt
```

This runs `go fmt` and `gofmt -s -w .`

### Linting

```bash
just lint
```

Requires `golangci-lint` to be installed.

## Building for Distribution

### Single Platform

```bash
just build
```

Creates `bin/versioner` for your current platform.

### All Platforms

```bash
just build_all
```

Creates binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

All binaries are placed in `bin/` directory.

## Common Issues

### "command not found: versioner"
**Problem**: Binary not in PATH  
**Solution**: Use `./bin/versioner` or add `bin/` to your PATH

### "package X is not in GOROOT"
**Problem**: Missing dependencies  
**Solution**: Run `go mod download` or `just setup_local_dev`

### Tests failing after changes
**Problem**: Broke existing functionality  
**Solution**:
1. Read test output carefully
2. Fix the issue or update tests if behavior changed intentionally
3. Run `just run_tests` to verify

### "cannot find package"
**Problem**: Import path incorrect  
**Solution**: Use full import path: `github.com/versioner-io/versioner-cli/internal/...`

### Build errors
**Problem**: Dependency issues  
**Solution**: Run `go mod tidy` to clean up dependencies

## Go Concepts

### Packages and Modules
- **Package**: A directory with `.go` files
- **Module**: Defined by `go.mod`, contains one or more packages
- **Import path**: Full path like `github.com/versioner-io/versioner-cli/internal/api`

### Structs
Go's equivalent of classes:

```go
type BuildEvent struct {
    ProductName string
    Version     string
}
```

### Error Handling
Go uses explicit error returns:

```go
result, err := doSomething()
if err != nil {
    return err
}
```

### Nil
Go's equivalent of `null`:
- Used for pointers, interfaces, slices, maps, channels
- Check with `if x == nil`

### Pointers
Go uses explicit pointers:

```go
var x int = 5
var p *int = &x  // p is a pointer to x
fmt.Println(*p)  // Dereference: prints 5
```

### Capitalization
- **Uppercase** = Public/exported (e.g., `ProductName`)
- **Lowercase** = Private/unexported (e.g., `productName`)

## Commit Message Format

Use conventional commit format:

```bash
git commit -m "feat: add new validation command"
git commit -m "fix: correct status normalization for edge case"
git commit -m "docs: update development guide"
git commit -m "test: add tests for new feature"
git commit -m "refactor: simplify error handling"
git commit -m "chore: update dependencies"
```

**Prefixes**:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test changes
- `refactor:` - Code refactoring
- `chore:` - Build/tooling changes

## Additional Resources

### Go Learning
- [Tour of Go](https://go.dev/tour/) - Interactive tutorial
- [Effective Go](https://go.dev/doc/effective_go) - Best practices
- [Go by Example](https://gobyexample.com/) - Code examples

### Frameworks Used
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management

### Project Documentation
- [Development Plan](development-plan.md) - Project roadmap
- [API Contract](api-contract.md) - API specification
- [CI/CD Env Vars](cicd-env-vars.md) - Auto-detection reference

## Getting Help

1. Check the docs in `docs/` directory
2. Read the code - Go is designed to be readable
3. Run with `--verbose` to see what's happening
4. Check test files - they show how to use the code
5. Ask questions!
