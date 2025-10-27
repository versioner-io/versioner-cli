# Development Guide

## 🏗️ Repository Structure

### High-Level Overview

This is a **Go CLI application** using the Cobra framework. If you're coming from Python, think of it like:
- **Go modules** = Python's `requirements.txt` or `pyproject.toml`
- **Packages** = Python modules/packages
- **`cmd/`** = Your `main.py` or entry point
- **`internal/`** = Your application code (not importable by other projects)

### Directory Structure

```
versioner-cli/
├── cmd/versioner/              # Application entry point (like main.py)
│   └── main.go                 # Calls the root command and exits
│
├── internal/                   # Private application code (cannot be imported by other projects)
│   ├── api/                    # API client for talking to Versioner backend
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
│   │   └── track_deployment.go # Track deployment subcommand
│   │
│   └── status/                 # Status value validation and normalization
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
├── go.mod                      # Go module definition (like requirements.txt)
├── go.sum                      # Dependency lock file (like poetry.lock)
├── justfile                    # Build commands (like Makefile)
└── README.md                   # User-facing documentation
```

## 🔑 Key Files Explained

### `cmd/versioner/main.go`
**Purpose**: Application entry point
**What it does**: Calls the root Cobra command and exits with appropriate code
**Python equivalent**: Your `if __name__ == "__main__":` block

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

**Key concepts**:
- **Cobra**: CLI framework (like Python's Click or argparse)
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
- `go.mod`: Declares module name and dependencies (like `requirements.txt`)
- `go.sum`: Cryptographic checksums for dependencies (like `poetry.lock`)

**Key dependencies**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

## 🛠️ Development Workflow

### 1. **Initial Setup**

```bash
# Clone the repository
git clone <repo-url>
cd versioner-cli

# Install dependencies (Go will download them automatically)
just setup_local_dev
```

This runs:
- `go mod download` - Downloads all dependencies
- `go mod tidy` - Cleans up unused dependencies

### 2. **Making Changes**

#### Adding a New Feature

1. **Create/modify files** in `internal/` directory
2. **Write tests** alongside your code (e.g., `myfile.go` → `myfile_test.go`)
3. **Run tests** to verify:
   ```bash
   just run_tests
   ```

#### Adding a New Command

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

### 3. **Building**

```bash
# Build for your current platform
just build

# This creates: bin/versioner
```

The binary is placed in `bin/` and can be run with:
```bash
./bin/versioner --help
```

### 4. **Testing**

```bash
# Run all tests
just run_tests

# Run only unit tests (faster)
just test_unit

# Run with coverage report
just test_coverage
# Opens coverage.html in browser
```

**Writing tests in Go**:
- Test files end with `_test.go`
- Test functions start with `Test` (e.g., `TestDetectGitHub`)
- Use `t.Errorf()` for failures

Example:
```go
func TestNormalize(t *testing.T) {
    result, _ := Normalize("success")
    if result != "completed" {
        t.Errorf("Expected 'completed', got '%s'", result)
    }
}
```

### 5. **Running Locally**

```bash
# Option 1: Build and run
just build
./bin/versioner track build --product=test --version=1.0.0

# Option 2: Run without building (slower, but convenient)
just run track build --product=test --version=1.0.0

# Option 3: Use go run directly
go run ./cmd/versioner track build --product=test --version=1.0.0
```

### 6. **Formatting & Linting**

```bash
# Format code (like black for Python)
just fmt

# Run linters (requires golangci-lint installed)
just lint
```

Go is very strict about formatting. `go fmt` is the standard formatter.

### 7. **Debugging**

#### Print Debugging
```go
fmt.Fprintf(os.Stderr, "Debug: value=%v\n", myVariable)
```

#### Verbose Mode
Run with `--verbose` to see detailed output:
```bash
./bin/versioner track build --product=test --version=1.0.0 --verbose
```

#### Debug Mode
Run with `--debug` to see HTTP requests/responses:
```bash
./bin/versioner track build --product=test --version=1.0.0 --debug
```

### 8. **Committing Changes**

```bash
# Stage changes
git add -A

# Commit with conventional commit format
git commit -m "feat: add new validation command"
git commit -m "fix: correct status normalization for edge case"
git commit -m "docs: update development guide"
git commit -m "test: add tests for new feature"
```

**Commit prefixes**:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test changes
- `refactor:` - Code refactoring
- `chore:` - Build/tooling changes

## 🧪 Testing Strategy

### Unit Tests
Test individual functions in isolation:
```bash
# Test a specific package
go test ./internal/status/...
go test ./internal/cicd/...

# Test with verbose output
go test -v ./internal/status/...
```

### Integration Tests
Test API client with mock server (TODO in Phase 2):
```bash
just test_integration
```

### Manual Testing
Test the full CLI flow:
```bash
# Set up test environment
export VERSIONER_API_KEY=test-key-123
export VERSIONER_API_URL=http://localhost:8000

# Test build tracking
./bin/versioner track build \
  --product=test-app \
  --version=1.0.0 \
  --status=completed \
  --verbose

# Test deployment tracking
./bin/versioner track deployment \
  --product=test-app \
  --environment=staging \
  --version=1.0.0 \
  --status=success \
  --verbose
```

## 📦 Building for Distribution

### Single Platform
```bash
just build
```

### All Platforms
```bash
just build_all
```

This creates binaries for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

All binaries are placed in `bin/` directory.

## 🐛 Common Issues & Solutions

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

## 🎓 Go Concepts for Python Developers

### Packages vs Modules
- **Go package** = Python module (a directory with `.go` files)
- **Go module** = Python project (defined by `go.mod`)

### Imports
```go
// Go
import "github.com/versioner-io/versioner-cli/internal/api"

# Python
from internal.api import client
```

### Structs vs Classes
```go
// Go struct (like a Python dataclass)
type BuildEvent struct {
    ProductName string
    Version     string
}

# Python equivalent
@dataclass
class BuildEvent:
    product_name: str
    version: str
```

### Error Handling
```go
// Go - explicit error handling
result, err := doSomething()
if err != nil {
    return err
}

# Python - exceptions
try:
    result = do_something()
except Exception as e:
    raise
```

### Nil vs None
- Go: `nil` (for pointers, interfaces, slices, maps, channels)
- Python: `None`

### Pointers
Go uses pointers explicitly:
```go
var x int = 5
var p *int = &x  // p is a pointer to x
fmt.Println(*p)  // Dereference: prints 5
```

Python hides this - everything is a reference.

## 🔗 Useful Commands Reference

```bash
# Development
just setup_local_dev    # Initial setup
just build              # Build binary
just run <args>         # Run without building
just clean              # Remove build artifacts

# Testing
just run_tests          # Run all tests
just test_unit          # Unit tests only
just test_integration   # Integration tests only
just test_coverage      # Generate coverage report

# Code Quality
just fmt                # Format code
just lint               # Run linters

# Building
just build_all          # Build for all platforms
```

## 📚 Additional Resources

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

## 🤝 Getting Help

1. **Check the docs** in `docs/` directory
2. **Read the code** - Go is designed to be readable
3. **Run with `--verbose`** to see what's happening
4. **Check test files** - They show how to use the code
5. **Ask questions** - Better to ask than to guess!

## 🎯 Quick Start Checklist

- [ ] Clone repository
- [ ] Run `just setup_local_dev`
- [ ] Run `just build`
- [ ] Run `just run_tests` to verify setup
- [ ] Try `./bin/versioner --help`
- [ ] Make a small change
- [ ] Run tests again
- [ ] Commit your change

Welcome to Go development! 🎉
