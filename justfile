# List available commands
default:
    just --list

ci:
    # Run all CI checks locally (tests, lint, build)
    @echo "Running CI checks..."
    @echo "\n=== Formatting ==="
    just fmt
    @echo "\n=== Tests ==="
    just run_tests
    @echo "\n=== Linting ==="
    -just lint || echo "⚠️  golangci-lint not installed (skipping lint check)"
    @echo "\n=== Build Check ==="
    just build
    @echo "\n✅ All CI checks passed!"

setup_local_dev:
    # Set up local development environment
    @echo "Setting up local development environment..."
    go mod download
    go mod tidy
    just install_linter
    @echo "✅ Development environment ready!"

install_linter:
    # Install golangci-lint
    @echo "Installing golangci-lint..."
    @if command -v golangci-lint >/dev/null 2>&1; then \
        echo "✅ golangci-lint already installed ($(golangci-lint --version))"; \
    else \
        echo "Installing golangci-lint..."; \
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
        if [ -f "$$HOME/go/bin/golangci-lint" ]; then \
            echo "✅ golangci-lint installed to $$HOME/go/bin/golangci-lint"; \
            echo ""; \
            echo "⚠️  Add to your PATH by adding this to ~/.zshrc or ~/.bashrc:"; \
            echo "   export PATH=\"\$$HOME/go/bin:\$$PATH\""; \
            echo ""; \
            echo "Then run: source ~/.zshrc (or restart terminal)"; \
        else \
            echo "✅ golangci-lint installed"; \
        fi; \
    fi

build:
    # Build the CLI binary (static for maximum compatibility)
    @echo "Building versioner..."
    @mkdir -p bin
    CGO_ENABLED=0 go build -ldflags="-X 'github.com/versioner-io/versioner-cli/internal/version.Version=dev' \
        -X 'github.com/versioner-io/versioner-cli/internal/version.Commit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)' \
        -X 'github.com/versioner-io/versioner-cli/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)'" \
        -o bin/versioner ./cmd/versioner

build_all VERSION="dev":
    # Build for all platforms with version injection (static binaries)
    @echo "Building for all platforms (version: {{VERSION}})..."
    @mkdir -p bin
    #!/usr/bin/env bash
    set -euo pipefail
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    LDFLAGS="-X 'github.com/versioner-io/versioner-cli/internal/version.Version={{VERSION}}' -X 'github.com/versioner-io/versioner-cli/internal/version.Commit=$COMMIT' -X 'github.com/versioner-io/versioner-cli/internal/version.BuildDate=$BUILD_DATE'"
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o bin/versioner-linux-amd64 ./cmd/versioner
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$LDFLAGS" -o bin/versioner-linux-arm64 ./cmd/versioner
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$LDFLAGS" -o bin/versioner-darwin-amd64 ./cmd/versioner
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$LDFLAGS" -o bin/versioner-darwin-arm64 ./cmd/versioner
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$LDFLAGS" -o bin/versioner-windows-amd64.exe ./cmd/versioner
    echo "✅ All platform binaries built in bin/ with version {{VERSION}} (static)"

run *ARGS:
    # Run the CLI (pass arguments after --)
    @echo "Running versioner..."
    go run ./cmd/versioner {{ARGS}}

clean:
    # Clean build artifacts
    @echo "Cleaning build artifacts..."
    rm -rf bin/
    rm -f coverage.out coverage.html

lint:
    # Run linters
    @echo "Running linters..."
    golangci-lint run ./...

fmt:
    # Format code
    @echo "Formatting code..."
    go fmt ./...
    gofmt -s -w .

run_tests:
    # Run all tests
    @echo "Running all tests..."
    go test -v ./...

test_unit:
    # Run unit tests only
    @echo "Running unit tests..."
    go test -v -short ./...

test_integration:
    # Run integration tests only
    @echo "Running integration tests..."
    go test -v -run Integration ./...

test_coverage:
    # Run tests with coverage report
    @echo "Running tests with coverage..."
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

down_tests:
    # Clean up test artifacts
    @echo "Cleaning up test artifacts..."
    rm -f coverage.out coverage.html
