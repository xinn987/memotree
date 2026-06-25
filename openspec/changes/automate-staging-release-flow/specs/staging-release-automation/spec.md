## ADDED Requirements

### Requirement: Release Env Handoff
The system SHALL provide a release handoff file for routine staging deployments that contains release image tags and metadata without long-lived secrets.

#### Scenario: Release env is generated after image publish
- **WHEN** an operator publishes API, Worker, and Web images for a staging release
- **THEN** the tooling writes a release env file containing `API_IMAGE`, `WORKER_IMAGE`, `WEB_IMAGE`, `RELEASE_COMMIT`, and `RELEASE_CREATED_AT`

#### Scenario: Release env excludes secrets
- **WHEN** the release env file is generated
- **THEN** it MUST NOT include MySQL passwords, storage access keys, session secrets, or bucket credentials

### Requirement: Routine Staging Deploy
The system SHALL provide a server-side routine deploy command that consumes long-lived staging configuration and a per-release env file.

#### Scenario: Deploy validates configuration before restart
- **WHEN** an operator runs the staging deploy command with a release env file
- **THEN** the command validates required files and Docker Compose configuration before pulling or restarting services

#### Scenario: Deploy pulls and restarts business services
- **WHEN** staging deploy validation succeeds
- **THEN** the command pulls the API, Worker, and Web images from the release env and restarts the API, Worker, and Web services

#### Scenario: Deploy does not run privileged initialization
- **WHEN** the routine staging deploy command runs
- **THEN** it MUST NOT automatically run schema initialization or object storage bucket initialization

### Requirement: Deploy Health Checks
The system SHALL verify that a routine staging deploy is reachable after services restart.

#### Scenario: Health checks pass
- **WHEN** API, Worker, and Web services are restarted
- **THEN** the command checks Docker Compose service status and verifies the Web and API health endpoints from the server host

#### Scenario: Health checks fail
- **WHEN** a health check fails after restart
- **THEN** the command exits non-zero and prints guidance for inspecting service logs

### Requirement: Staging Log Inspection
The system SHALL provide a consistent command for inspecting staging service logs.

#### Scenario: Operator inspects logs
- **WHEN** an operator runs the staging log helper
- **THEN** the helper shows recent or follow-mode logs for API, Worker, Web, and MySQL using the staging Compose file and env configuration

### Requirement: Routine Release Documentation
The system SHALL document the routine staging release flow separately from first-time infrastructure setup.

#### Scenario: Operator follows routine release docs
- **WHEN** an operator reads the deployment runbook after first-time setup is complete
- **THEN** it explains how to publish images, copy the release env file to the server, run the deploy command, inspect logs, and roll back to a prior release env file
