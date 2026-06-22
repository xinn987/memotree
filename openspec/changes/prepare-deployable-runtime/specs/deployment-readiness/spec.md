## ADDED Requirements

### Requirement: Service Images

The system SHALL provide buildable runtime images for the API, Worker, and Web app.

#### Scenario: API image builds

- **WHEN** CI or an operator builds the API runtime image from the repository
- **THEN** the build produces an image that starts the API server and exposes the existing health endpoint

#### Scenario: Worker image builds for photo processing

- **WHEN** CI or an operator builds the Worker runtime image from the repository
- **THEN** the build produces an image that contains the Worker binary and can process photo rendition jobs without requiring FFmpeg/FFprobe

#### Scenario: Web image builds static assets

- **WHEN** CI or an operator builds the Web runtime image from the repository
- **THEN** the build produces an image that serves the production Web bundle and routes API requests to the API service boundary

### Requirement: Runtime Environment Configuration

The system SHALL document and validate the required runtime configuration for local, staging, and production-like environments.

#### Scenario: Required deployment variables are documented

- **WHEN** an operator reads the deployment environment example
- **THEN** it identifies required values for API address, session cookie settings, MySQL DSN, object storage endpoint, object storage credentials, and bucket names

#### Scenario: Local defaults remain usable

- **WHEN** a developer starts the existing local development environment
- **THEN** the system continues to use Docker MySQL and MinIO defaults without requiring production secrets

#### Scenario: Missing required production dependency fails fast

- **WHEN** an API or Worker process starts in a production-like environment without required database or object storage configuration
- **THEN** the process fails clearly rather than silently falling back to in-memory or disabled storage behavior

### Requirement: Photo-Only MVP Deployment

The system SHALL make the deployable MVP image-only until video processing is revisited in a later change.

#### Scenario: Video upload is rejected before processing

- **WHEN** a user or client attempts to create an upload intent for a video media type in the deployable MVP
- **THEN** the API rejects the request with a clear unsupported media type response before any Worker job is created

#### Scenario: Web upload entry point is image-only

- **WHEN** a user chooses files from the MVP upload UI
- **THEN** the UI only offers image file types and communicates unsupported selections clearly

#### Scenario: Worker deployment does not require video tools

- **WHEN** the Worker starts in the MVP deployment profile
- **THEN** it does not fail solely because FFmpeg or FFprobe are unavailable

### Requirement: Deployment Initialization

The system SHALL provide explicit deployment initialization steps for database schema and object storage buckets.

#### Scenario: Schema initialization can be run explicitly

- **WHEN** an operator prepares a staging environment
- **THEN** they can run a documented command or process that applies the current database schema before starting long-running services

#### Scenario: Object storage buckets can be initialized explicitly

- **WHEN** an operator prepares a staging environment
- **THEN** they can run a documented command or process that ensures the originals and previews buckets exist

#### Scenario: Initialization is separated from long-running service permissions

- **WHEN** deployment docs describe staging or production credentials
- **THEN** they distinguish one-time initialization permissions from long-running API and Worker runtime permissions

### Requirement: CI Deployment Checks

The system SHALL verify deployment readiness in CI without automatically deploying to an environment.

#### Scenario: CI validates current specs

- **WHEN** CI runs OpenSpec validation
- **THEN** it validates the current repository specs or active changes rather than an archived change name

#### Scenario: CI uses the repository-supported runtime versions

- **WHEN** CI runs Web and Go checks
- **THEN** it uses Node and Go versions aligned with the repository development requirements

#### Scenario: CI builds service images

- **WHEN** CI runs on pull requests or main
- **THEN** it builds API, Worker, and Web runtime images without pushing or deploying them

### Requirement: Staging Deployment Runbook

The system SHALL provide a staging deployment runbook that an operator can follow to verify the deployed service.

#### Scenario: First staging deployment is documented

- **WHEN** an operator follows the staging runbook
- **THEN** it covers configuring environment variables, initializing schema and buckets, starting services, and checking health

#### Scenario: Smoke test covers image media pipeline

- **WHEN** an operator completes a staging deployment
- **THEN** the runbook includes smoke tests for account creation, family creation, image upload authorization, object upload, Worker processing, timeline visibility, media soft delete, and processing retry

#### Scenario: Rollback and logs are documented

- **WHEN** a staging deployment fails or behaves unexpectedly
- **THEN** the runbook identifies service logs, dependency logs, and rollback steps for returning to the previous image or configuration
