# CI/CD Environment Variables Reference

This document lists environment variables from popular CI/CD systems that the Versioner CLI can auto-detect and map to build and deployment event fields.

## Build vs Deployment Context

The CLI needs to determine whether to track a **build event** or a **deployment event**:

- **Build events** (`track build`) - Track CI/CD build lifecycle (compiling, testing, creating artifacts)
- **Deployment events** (`track deployment`) - Track deployment lifecycle (deploying to environments)

### Auto-Detection Strategy

When the CLI auto-detects CI/CD variables:

1. **Explicit command** - User specifies `track build` or `track deployment`
2. **Environment hint** - If `--environment` flag is provided or `VERSIONER_ENVIRONMENT` is set → deployment
3. **CI context** - Some CI systems have deployment-specific variables (e.g., `CI_ENVIRONMENT_NAME` in GitLab)
4. **Default** - If ambiguous, default to build event (safer assumption)

### Recommendation

For clarity, always explicitly use `track build` or `track deployment` rather than relying on auto-detection.

## GitHub Actions

| Environment Variable | Maps To | Description |
|---------------------|---------|-------------|
| `GITHUB_REPOSITORY` | `scm_repository` | Repository name (owner/repo) |
| `GITHUB_SHA` | `scm_sha` | Commit SHA that triggered the workflow |
| `GITHUB_REF_NAME` | `version` (fallback) | Branch or tag name |
| `GITHUB_RUN_ID` | `invoke_id` | Unique workflow run ID |
| `GITHUB_RUN_NUMBER` | `build_number` | Workflow run number |
| `GITHUB_ACTOR` | `deployed_by` | Username that triggered the workflow |
| `GITHUB_SERVER_URL` | - | GitHub server URL (for building links) |
| `GITHUB_WORKFLOW` | `source_system` | Always "github" |
| `GITHUB_ACTION` | - | Name of the action currently running |

**Build URL Construction**: `$GITHUB_SERVER_URL/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID`

## GitLab CI/CD

| Environment Variable | Maps To | Description |
|---------------------|---------|-------------|
| `CI_PROJECT_PATH` | `scm_repository` | Project path (namespace/project) |
| `CI_COMMIT_SHA` | `scm_sha` | Commit SHA |
| `CI_COMMIT_REF_NAME` | `version` (fallback) | Branch or tag name |
| `CI_PIPELINE_ID` | `invoke_id` | Pipeline ID |
| `CI_PIPELINE_IID` | `build_number` | Pipeline IID (internal ID) |
| `CI_JOB_ID` | - | Job ID |
| `GITLAB_USER_LOGIN` | `deployed_by` | Username that triggered the pipeline |
| `GITLAB_USER_EMAIL` | `deployed_by_email` | Email of user who triggered pipeline |
| `GITLAB_USER_NAME` | `deployed_by_name` | Display name of user |
| `CI_PROJECT_URL` | - | Project URL |
| `CI_PIPELINE_URL` | `build_url` | Direct link to pipeline |
| `CI_SERVER` | `source_system` | Always "gitlab" |

## Jenkins

| Environment Variable | Maps To | Description |
|---------------------|---------|-------------|
| `GIT_URL` | `scm_repository` | Git repository URL |
| `GIT_COMMIT` | `scm_sha` | Git commit SHA |
| `GIT_BRANCH` | `version` (fallback) | Git branch name |
| `BUILD_NUMBER` | `build_number` | Jenkins build number |
| `BUILD_ID` | `invoke_id` | Build ID (timestamp format) |
| `BUILD_URL` | `build_url` | Direct link to build |
| `JOB_NAME` | - | Jenkins job name |
| `BUILD_USER` | `deployed_by` | User who started the build (requires plugin) |
| `BUILD_USER_EMAIL` | `deployed_by_email` | User email (requires plugin) |
| `JENKINS_URL` | - | Jenkins server URL |

**Note**: `BUILD_USER` variables require the [Build User Vars Plugin](https://plugins.jenkins.io/build-user-vars-plugin/).

**Source System**: "jenkins"

## CircleCI

| Environment Variable | Maps To | Description |
|---------------------|---------|-------------|
| `CIRCLE_PROJECT_USERNAME` | - | GitHub/Bitbucket username |
| `CIRCLE_PROJECT_REPONAME` | - | Repository name |
| `CIRCLE_REPOSITORY_URL` | `scm_repository` | Full repository URL |
| `CIRCLE_SHA1` | `scm_sha` | Git commit SHA |
| `CIRCLE_BRANCH` | `version` (fallback) | Git branch name |
| `CIRCLE_TAG` | `version` (fallback) | Git tag (if building a tag) |
| `CIRCLE_BUILD_NUM` | `build_number` | CircleCI build number |
| `CIRCLE_WORKFLOW_ID` | `invoke_id` | Workflow ID |
| `CIRCLE_BUILD_URL` | `build_url` | Direct link to build |
| `CIRCLE_USERNAME` | `deployed_by` | GitHub/Bitbucket username of user |

**Source System**: "circleci"

**Repository Construction**: `$CIRCLE_PROJECT_USERNAME/$CIRCLE_PROJECT_REPONAME`

## Bitbucket Pipelines

| Environment Variable | Maps To | Description |
|---------------------|---------|-------------|
| `BITBUCKET_REPO_FULL_NAME` | `scm_repository` | Full repository name (workspace/repo) |
| `BITBUCKET_COMMIT` | `scm_sha` | Commit hash |
| `BITBUCKET_BRANCH` | `version` (fallback) | Branch name |
| `BITBUCKET_TAG` | `version` (fallback) | Tag name (if building a tag) |
| `BITBUCKET_BUILD_NUMBER` | `build_number` | Build number |
| `BITBUCKET_PIPELINE_UUID` | `invoke_id` | Pipeline UUID |
| `BITBUCKET_GIT_HTTP_ORIGIN` | - | Repository HTTP URL |
| `BITBUCKET_WORKSPACE` | - | Workspace name |
| `BITBUCKET_REPO_SLUG` | - | Repository slug |

**Source System**: "bitbucket"

**Build URL Construction**: `https://bitbucket.org/$BITBUCKET_REPO_FULL_NAME/pipelines/results/$BITBUCKET_BUILD_NUMBER`

## Azure DevOps

| Environment Variable | Maps To | Description |
|---------------------|---------|-------------|
| `BUILD_REPOSITORY_NAME` | `scm_repository` | Repository name |
| `BUILD_SOURCEVERSION` | `scm_sha` | Commit SHA |
| `BUILD_SOURCEBRANCHNAME` | `version` (fallback) | Branch name |
| `BUILD_BUILDNUMBER` | `build_number` | Build number |
| `BUILD_BUILDID` | `invoke_id` | Build ID |
| `BUILD_BUILDURI` | `build_url` | Build URI |
| `BUILD_REQUESTEDFOR` | `deployed_by` | User who requested the build |
| `BUILD_REQUESTEDFOREMAIL` | `deployed_by_email` | User email |
| `SYSTEM_TEAMPROJECT` | - | Project name |
| `SYSTEM_COLLECTIONURI` | - | Azure DevOps collection URI |

**Source System**: "azure-devops"

## Travis CI

| Environment Variable | Maps To | Description |
|---------------------|---------|-------------|
| `TRAVIS_REPO_SLUG` | `scm_repository` | Repository slug (owner/repo) |
| `TRAVIS_COMMIT` | `scm_sha` | Commit SHA |
| `TRAVIS_BRANCH` | `version` (fallback) | Branch name |
| `TRAVIS_TAG` | `version` (fallback) | Tag name (if building a tag) |
| `TRAVIS_BUILD_NUMBER` | `build_number` | Build number |
| `TRAVIS_BUILD_ID` | `invoke_id` | Build ID |
| `TRAVIS_BUILD_WEB_URL` | `build_url` | Build URL |

**Source System**: "travis"

## Argo CD

Argo CD doesn't provide standard environment variables during sync operations. For Argo CD integrations, consider:

1. **Resource Hooks**: Use PreSync/PostSync hooks to run the CLI
2. **Manual Variables**: Pass variables explicitly via hook annotations
3. **Argo CD Notifications**: Integrate with Argo CD's notification system

**Source System**: "argocd"

## Generic/Manual Events

For systems not listed above or manual tracking, all fields should be provided explicitly via CLI flags or environment variables prefixed with `VERSIONER_`:

### Common Fields

| CLI Flag | Environment Variable | Description |
|----------|---------------------|-------------|
| `--product` | `VERSIONER_PRODUCT` | Product name |
| `--version` | `VERSIONER_VERSION` | Version string |
| `--status` | `VERSIONER_STATUS` | Event status |
| `--source-system` | `VERSIONER_SOURCE_SYSTEM` | Source system identifier |
| `--build-number` | `VERSIONER_BUILD_NUMBER` | Build number |
| `--scm-sha` | `VERSIONER_SCM_SHA` | Git commit SHA |
| `--scm-repository` | `VERSIONER_SCM_REPOSITORY` | Repository name |
| `--build-url` | `VERSIONER_BUILD_URL` | Build/deployment URL |

### Build-Specific Fields

| CLI Flag | Environment Variable | Description |
|----------|---------------------|-------------|
| `--scm-branch` | `VERSIONER_SCM_BRANCH` | Git branch name |
| `--built-by` | `VERSIONER_BUILT_BY` | Builder username |
| `--built-by-email` | `VERSIONER_BUILT_BY_EMAIL` | Builder email |
| `--built-by-name` | `VERSIONER_BUILT_BY_NAME` | Builder display name |
| `--started-at` | `VERSIONER_STARTED_AT` | Build start timestamp |
| `--completed-at` | `VERSIONER_COMPLETED_AT` | Build completion timestamp |

### Deployment-Specific Fields

| CLI Flag | Environment Variable | Description |
|----------|---------------------|-------------|
| `--environment` | `VERSIONER_ENVIRONMENT` | Environment name (required for deployments) |
| `--deployed-by` | `VERSIONER_DEPLOYED_BY` | Deployer username |
| `--deployed-by-email` | `VERSIONER_DEPLOYED_BY_EMAIL` | Deployer email |
| `--deployed-by-name` | `VERSIONER_DEPLOYED_BY_NAME` | Deployer display name |
| `--completed-at` | `VERSIONER_COMPLETED_AT` | Deployment completion timestamp |

## Auto-Detection Priority

When auto-detection is enabled, the CLI should:

1. **Detect CI system**: Check for system-specific environment variables (e.g., `GITHUB_ACTIONS`, `GITLAB_CI`, `JENKINS_URL`, `CIRCLECI`, `BITBUCKET_BUILD_NUMBER`)
2. **Determine event type**: Build or deployment based on command and context
3. **Extract values**: Map system-specific variables to appropriate event fields
4. **Allow overrides**: CLI flags and `VERSIONER_*` env vars take precedence over auto-detected values
5. **Provide feedback**: In verbose mode, show which values were auto-detected vs. explicitly provided

### Example: GitHub Actions Build

```bash
# Auto-detected from GitHub Actions environment
$ versioner track build --verbose

ℹ Auto-detected CI system: github
ℹ Auto-detected values:
  --product: api-service (from GITHUB_REPOSITORY)
  --version: abc123def (from GITHUB_SHA)
  --scm-repository: myorg/api-service (from GITHUB_REPOSITORY)
  --scm-sha: abc123def456 (from GITHUB_SHA)
  --build-number: 42 (from GITHUB_RUN_NUMBER)
  --source-system: github
✓ Build event tracked successfully
```

### Example: GitLab Deployment

```bash
# Auto-detected from GitLab CI environment
$ versioner track deployment --environment=production --verbose

ℹ Auto-detected CI system: gitlab
ℹ Auto-detected values:
  --product: api-service (from CI_PROJECT_PATH)
  --version: abc123def (from CI_COMMIT_SHA)
  --environment: production (from --environment flag)
  --scm-repository: myorg/api-service (from CI_PROJECT_PATH)
  --scm-sha: abc123def456 (from CI_COMMIT_SHA)
  --build-number: 123 (from CI_PIPELINE_IID)
  --source-system: gitlab
✓ Deployment event tracked successfully
```

## Implementation Notes

- **Normalization**: Strip `https://` and `.git` suffixes from repository URLs
- **Validation**: Ensure SHA values are valid hex strings (40 characters for Git)
- **Fallbacks**: If version is not provided, fall back to branch/tag name, then SHA (truncated to 8 chars)
- **User-Agent**: Include CLI version and detected CI system in API requests (e.g., `versioner/1.0.0 (github)`)
- **Event Type Detection**: Use command (`build` vs `deployment`) and presence of `--environment` flag to determine endpoint
- **Status Defaults**: Build events default to `completed`, deployment events default to `success` if not specified
