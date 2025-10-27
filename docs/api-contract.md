# Versioner API Contract

## Overview
The Versioner CLI interacts with the Versioner API to submit build and deployment events. This document defines the exact API contracts for both endpoints.

## Endpoints

- **POST** `/build-events/` - Track CI/CD build lifecycle events
- **POST** `/deployment-events/` - Track deployment lifecycle events

## Authentication

API requests require authentication via the `Authorization` header:

```
Authorization: Bearer <API_KEY>
```

**Security Note**: The CLI should warn users if they pass API keys via command-line flags (visible in process lists). Prefer environment variables or config files.

## Request Schemas

### BuildEventCreate (POST /build-events/)

**Content-Type**: `application/json`

#### Required Fields

| Field | Type | Max Length | Description |
|-------|------|------------|-------------|
| `product_name` | string | 255 | Product/application name (e.g., 'api-service') |
| `version` | string | 100 | Version string (e.g., '1.2.3' or build number) |
| `status` | string | 50 | Build status (see Status Values section) |

#### Optional Fields

| Field | Type | Max Length | Description |
|-------|------|------------|-------------|
| `source_system` | string | 50 | System that triggered build (github, jenkins, gitlab, circleci, etc.) |
| `build_number` | string | 100 | Build number from CI system |
| `scm_sha` | string | 40 | Git commit SHA (full 40-character hash) |
| `scm_branch` | string | 100 | Git branch name |
| `scm_repository` | string | 500 | Source control repository (e.g., 'myorg/my-api') |
| `build_url` | string | 500 | Link to CI/CD build run |
| `invoke_id` | string | 255 | Invocation/run ID from CI system |
| `built_by` | string | 255 | User identifier (username, email, or other unique ID) |
| `built_by_email` | string | 255 | User email (improves matching) |
| `built_by_name` | string | 255 | User display name |
| `started_at` | string (ISO 8601) | - | Build start timestamp |
| `completed_at` | string (ISO 8601) | - | Build completion timestamp |
| `extra_metadata` | object | - | Additional metadata as JSON object |

### DeploymentEventCreate (POST /deployment-events/)

**Content-Type**: `application/json`

#### Required Fields

| Field | Type | Max Length | Description |
|-------|------|------------|-------------|
| `product_name` | string | 255 | Product/application name (e.g., 'api-service') |
| `version` | string | 100 | Version string (e.g., '1.2.3' or build number) |
| `environment_name` | string | 100 | Environment name (e.g., 'production', 'staging') |
| `status` | string | 50 | Deployment status (see Status Values section) |

#### Optional Fields

| Field | Type | Max Length | Description |
|-------|------|------------|-------------|
| `source_system` | string | 50 | System that triggered deployment (github, jenkins, gitlab, circleci, etc.) |
| `build_number` | string | 100 | Build number from CI system |
| `scm_sha` | string | 40 | Git commit SHA (full 40-character hash) |
| `scm_repository` | string | 500 | Source control repository (e.g., 'myorg/my-api') |
| `build_url` | string | 500 | Link to CI/CD build run |
| `invoke_id` | string | 255 | Invocation/run ID from CI system |
| `deployed_by` | string | 255 | User identifier (username, email, or other unique ID) |
| `deployed_by_email` | string | 255 | User email (improves matching) |
| `deployed_by_name` | string | 255 | User display name |
| `completed_at` | string (ISO 8601) | - | Deployment completion timestamp |
| `extra_metadata` | object | - | Additional metadata as JSON object |

## Example Requests

### Build Event Example

```json
POST /build-events/
{
  "product_name": "api-service",
  "version": "1.2.3",
  "status": "completed",
  "source_system": "github",
  "build_number": "456",
  "scm_sha": "abc123def456789012345678901234567890abcd",
  "scm_branch": "main",
  "scm_repository": "myorg/api-service",
  "build_url": "https://github.com/myorg/api-service/actions/runs/456",
  "invoke_id": "456",
  "built_by": "github-actions",
  "built_by_email": "ci@myorg.com",
  "built_by_name": "GitHub Actions",
  "started_at": "2025-10-23T10:00:00Z",
  "completed_at": "2025-10-23T10:05:00Z",
  "extra_metadata": {
    "docker_image": "myorg/api-service:1.2.3",
    "artifacts": ["binary", "docker-image"]
  }
}
```

### Deployment Event Example

```json
POST /deployment-events/
{
  "product_name": "api-service",
  "version": "1.2.3",
  "environment_name": "production",
  "status": "success",
  "source_system": "github",
  "build_number": "456",
  "scm_sha": "abc123def456789012345678901234567890abcd",
  "scm_repository": "myorg/api-service",
  "build_url": "https://github.com/myorg/api-service/actions/runs/456",
  "invoke_id": "456",
  "deployed_by": "github-actions",
  "deployed_by_email": "deploy@myorg.com",
  "deployed_by_name": "GitHub Actions",
  "completed_at": "2025-10-23T10:10:00Z",
  "extra_metadata": {
    "deployment_duration_seconds": 120,
    "rollback_enabled": true
  }
}
```

## Responses

### Success Response (200 OK)

**Build Event** returns a `BuildResponse` object:

```json
{
  "id": "uuid-here",
  "product_id": "uuid-here",
  "version_id": "uuid-here",
  "status": "completed",
  "started_at": "2025-10-23T10:00:00Z",
  "completed_at": "2025-10-23T10:05:00Z",
  ...
}
```

**Deployment Event** returns a `DeploymentResponse` object:

```json
{
  "id": "uuid-here",
  "product_id": "uuid-here",
  "version_id": "uuid-here",
  "environment_id": "uuid-here",
  "status": "completed",
  "deployed_at": "2025-10-23T10:10:00Z",
  ...
}
```

### Error Responses

#### 401 Unauthorized
Invalid or missing API key.

```json
{
  "detail": "Invalid authentication credentials"
}
```

#### 422 Validation Error
Invalid request payload.

```json
{
  "detail": [
    {
      "loc": ["body", "product_name"],
      "msg": "field required",
      "type": "value_error.missing"
    }
  ]
}
```

#### 500 Internal Server Error
Server-side error.

```json
{
  "detail": "Internal server error"
}
```

## Behavior Notes

### Build Events
1. **Auto-creation**: The API automatically creates products and versions if they don't exist (using natural keys).
2. **Version Immutability**: Versions are immutable artifacts. Multiple builds can be associated with the same version.
3. **Notifications**: The API sends notifications to configured channels based on build status changes.

### Deployment Events
1. **Auto-creation**: The API automatically creates products, versions, and environments if they don't exist (using natural keys).
2. **Multiple Deployments**: Multiple deployments of the same version to the same environment are tracked separately.
3. **Notifications**: The API sends notifications to configured channels based on deployment status changes.

## Status Values

Both build and deployment events support the same **5 canonical statuses**. The API automatically normalizes input values to these canonical forms.

### Canonical Status Values

| Status | Description | Triggers Notification |
|--------|-------------|----------------------|
| `pending` | Queued/scheduled but not started | Yes (`.pending` event) |
| `started` | Currently executing | Yes (`.started` event) |
| `completed` | Successfully finished | Yes (`.completed` event) |
| `failed` | Failed with errors | Yes (`.failed` event) |
| `aborted` | Cancelled or skipped | Yes (`.aborted` event) |

### Accepted Aliases

The API accepts these aliases and normalizes them to canonical values:

**For `pending`:**
- `queued`
- `scheduled`

**For `started`:**
- `in_progress`
- `init`
- `building` (builds)
- `deploying` (deployments)

**For `completed`:**
- `success`
- `complete`
- `finished`
- `built` (builds)
- `deployed` (deployments)

**For `failed`:**
- `fail`
- `failure`
- `error`

**For `aborted`:**
- `abort`
- `cancelled`
- `cancel`
- `skipped`

### CLI Behavior

The CLI accepts any of the canonical values or aliases. In verbose mode, it will display which canonical status the input maps to:

```bash
$ versioner track build --status=success --verbose
ℹ Status 'success' will be normalized to 'completed' by the API
✓ Build event tracked successfully
```

## Rate Limiting

**TBD**: Document rate limits once confirmed with API team.

## Retry Strategy

The CLI implements retry logic for transient failures:

- **Retry on**: 5xx errors, network timeouts, connection errors
- **Do not retry on**: 4xx errors (except 429 Too Many Requests)
- **Configuration**: 3 retries with exponential backoff (1s, 2s, 4s)
- **Timeout**: 30 seconds per request

## Notification Events

Both endpoints trigger notification events based on status:

### Build Events
- `build.pending` - Build queued
- `build.started` - Build started
- `build.completed` - Build succeeded
- `build.failed` - Build failed
- `build.aborted` - Build cancelled/skipped

### Deployment Events
- `deployment.pending` - Deployment queued
- `deployment.started` - Deployment started
- `deployment.completed` - Deployment succeeded
- `deployment.failed` - Deployment failed
- `deployment.aborted` - Deployment cancelled/skipped
