# List available commands
default:
    just --list


setup_local_dev:
    # Set up local development environment
    @echo "Setting up local development environment..."
    go mod download
    go mod tidy
    @echo "Development environment ready!"

build:
    # Build the CLI binary
    @echo "Building versioner..."
    go build -o bin/versioner ./cmd/versioner

build_all:
    # Build for all platforms
    @echo "Building for all platforms..."
    GOOS=linux GOARCH=amd64 go build -o bin/versioner-linux-amd64 ./cmd/versioner
    GOOS=linux GOARCH=arm64 go build -o bin/versioner-linux-arm64 ./cmd/versioner
    GOOS=darwin GOARCH=amd64 go build -o bin/versioner-darwin-amd64 ./cmd/versioner
    GOOS=darwin GOARCH=arm64 go build -o bin/versioner-darwin-arm64 ./cmd/versioner
    GOOS=windows GOARCH=amd64 go build -o bin/versioner-windows-amd64.exe ./cmd/versioner
    @echo "All platform binaries built in bin/"

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
