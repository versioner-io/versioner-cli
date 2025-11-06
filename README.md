# Versioner CLI

> **Beta Release** - This CLI is currently in beta. We're actively seeking feedback from early users!

## 📌 What is Versioner?

**Versioner** is a deployment tracking and visibility system designed to help engineering teams:

- Detect drift, skipped steps, or misconfigurations across environments
- Maintain a shared view of what’s deployed where
- Reduce wasted time from unclear or missing deployment context
- Improve auditability, ownership, and approval processes in release pipelines

Versioner takes events such as "build completed" or "deployment started" in via the Events API and stores them in a database. It then provides a query endpoint to retrieve deployment history and a Slack app to query the database for deployment history in a chat-native interface.

### Full ecosystem architecture:

- **API**: Python-based REST API, accessed by the UI, Slack app and external events senders
- **Database**: Neon-hosted Postgres, with a **database-per-customer** design
- **Frontend**: React UI served via CloudFront
- **CLI (this project)**: Command-line interface for submitting events
- **GitHub Action**: GitHub Action for submitting events
- **Slack App**: Chat-native interface for querying deployment state and approvals
- **Events API**: External events (e.g. CI/CD notifications) flow through API Gateway → SQS → Lambda → DB

## This project

This project is the CLI for Versioner. It provides a simple way to submit build and deployment events to the Versioner API from any CI/CD system or deployment tool.

**Current Status**: Phase 1 (MVP) complete! The CLI is functional with core features including:
- ✅ Track build and deployment events
- ✅ Auto-detection for 7 CI/CD systems (GitHub Actions, GitLab CI, Jenkins, CircleCI, Bitbucket, Azure DevOps, Travis CI)
- ✅ Retry logic with exponential backoff
- ✅ Status value normalization
- ✅ Environment variable configuration

## Installation

### Beta Release - Direct Download

Download pre-built binaries from [GitHub Releases](https://github.com/versioner-io/versioner-cli/releases):

```bash
# Linux (amd64)
curl -L https://github.com/versioner-io/versioner-cli/releases/latest/download/versioner-linux-amd64 -o versioner
chmod +x versioner
sudo mv versioner /usr/local/bin/

# Linux (arm64)
curl -L https://github.com/versioner-io/versioner-cli/releases/latest/download/versioner-linux-arm64 -o versioner
chmod +x versioner
sudo mv versioner /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/versioner-io/versioner-cli/releases/latest/download/versioner-darwin-arm64 -o versioner
chmod +x versioner
sudo mv versioner /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/versioner-io/versioner-cli/releases/latest/download/versioner-darwin-amd64 -o versioner
chmod +x versioner
sudo mv versioner /usr/local/bin/

# Windows (amd64)
# Download versioner-windows-amd64.exe from releases and add to PATH
```

**Coming Soon:**
- Homebrew tap for easier installation on macOS/Linux
- Automated install script

## Quick Start

### Get Your API Key

1. Sign up at [app.versioner.io](https://app.versioner.io)
2. Navigate to Settings → API Keys
3. Create a new API key and save it securely

### Track a Build Event

```bash
versioner track build \
  --product=api-service \
  --version=1.2.3 \
  --status=completed
```

### Track a Deployment Event

```bash
versioner track deployment \
  --product=api-service \
  --environment=production \
  --version=1.2.3 \
  --status=success
```

### Configuration

Set your API key via environment variable:

```bash
export VERSIONER_API_KEY=your-api-key-here
export VERSIONER_API_URL=https://api.versioner.io  # Optional, defaults to production
```

Or use a config file at `~/.versioner/config.yaml`:

```yaml
api_key: your-api-key-here
api_url: https://api.versioner.io
```

## Usage Examples

### GitHub Actions

```yaml
- name: Track build
  run: |
    versioner track build \
      --product=api-service \
      --version=${{ github.sha }} \
      --status=completed
  env:
    VERSIONER_API_KEY: ${{ secrets.VERSIONER_API_KEY }}

- name: Track deployment
  run: |
    versioner track deployment \
      --product=api-service \
      --environment=production \
      --version=${{ github.sha }} \
      --status=success
  env:
    VERSIONER_API_KEY: ${{ secrets.VERSIONER_API_KEY }}
```

### GitLab CI

```yaml
build:
  script:
    - make build
  after_script:
    - versioner track build --product=api --version=$CI_COMMIT_SHA --status=completed

deploy:
  script:
    - make deploy
  after_script:
    - versioner track deployment --product=api --environment=$CI_ENVIRONMENT_NAME --version=$CI_COMMIT_SHA --status=success
```

### Jenkins

```groovy
stage('Build') {
  steps {
    sh 'make build'
    sh 'versioner track build --product=api --version=${BUILD_NUMBER} --status=completed'
  }
}

stage('Deploy') {
  steps {
    sh 'make deploy'
    sh 'versioner track deployment --product=api --environment=prod --version=${BUILD_NUMBER} --status=success'
  }
}
```

## Status Values

Both build and deployment events support these statuses:

- `pending` - Queued/scheduled
- `started` - Currently executing
- `completed` - Successfully finished
- `failed` - Failed with errors
- `aborted` - Cancelled or skipped

Aliases like `success`, `in_progress`, `cancelled`, etc. are automatically normalized.

## CI/CD Auto-Detection

The CLI automatically detects your CI/CD environment and extracts relevant metadata. Supported systems:

- **GitHub Actions** - Repository, commit SHA, run ID, actor
- **GitLab CI** - Project path, commit SHA, pipeline ID, user
- **Jenkins** - Repository, commit SHA, build number, user (with plugin)
- **CircleCI** - Repository, commit SHA, build number, workflow ID
- **Bitbucket Pipelines** - Repository, commit SHA, build number
- **Azure DevOps** - Repository, commit SHA, build number, user
- **Travis CI** - Repository, commit SHA, build number

When running in a supported CI/CD system, you can omit many flags:

```bash
# In GitHub Actions - auto-detects repository, SHA, build number, etc.
versioner track build --product=api-service --status=completed

# In GitLab CI - auto-detects project, SHA, pipeline info
versioner track deployment --environment=production --status=success
```

Use `--verbose` to see which values were auto-detected:

```bash
versioner track build --product=api --status=completed --verbose
```

## Feedback & Support

This is a beta release and we'd love your feedback!

- **Issues & Bug Reports:** [GitHub Issues](https://github.com/versioner-io/versioner-cli/issues)
- **Feature Requests:** [GitHub Discussions](https://github.com/versioner-io/versioner-cli/discussions)
- **Security Issues:** See [SECURITY.md](SECURITY.md)

## Documentation

For comprehensive documentation:

- **[API Reference](https://api.versioner.io/docs)** - Interactive OpenAPI documentation
- **[Web Dashboard](https://app.versioner.io)** - View your deployment history
- **[CLI Integration Guide](https://github.com/versioner-io/versioner-docs/blob/main/features/cli-integration.md)** - Complete feature documentation and roadmap

### Repository-Specific Docs

- [CONTRIBUTING.md](CONTRIBUTING.md) - Development workflow, setup, testing
- [SECURITY.md](SECURITY.md) - Security policy and vulnerability reporting
- [LICENSE](LICENSE) - MIT License
