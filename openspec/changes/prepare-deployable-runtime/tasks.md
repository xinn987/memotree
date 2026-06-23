## 1. Runtime Images

- [x] 1.1 Add an API Dockerfile that builds the Go API binary and starts the API server with production-compatible defaults.
- [x] 1.2 Review and adjust the Worker Dockerfile so it builds under CI for photo-only MVP processing and documents required runtime environment variables.
- [x] 1.3 Add a Web/static Dockerfile and server configuration that serves the Vite production bundle and routes `/api/*` to the API service boundary.
- [x] 1.4 Add a local image-build helper or documented commands for building API, Worker, and Web images from the repository root.

## 2. Deployment Configuration

- [x] 2.1 Add staging-style deployment templates that wire API, Worker, Web, MySQL, and S3-compatible object storage together.
- [x] 2.2 Update environment examples to distinguish local defaults from staging/production required secrets.
- [x] 2.3 Ensure production-like API and Worker startup fails clearly when required MySQL or object storage configuration is missing.
- [x] 2.4 Keep existing local development scripts working with Docker MySQL and MinIO defaults.
- [x] 2.5 Enforce the MVP photo-only upload policy in Web/API configuration or code paths so videos are rejected clearly before Worker processing.

## 3. Initialization And Operations

- [x] 3.1 Document or add an explicit schema initialization command suitable for staging before long-running services start.
- [x] 3.2 Document or add an explicit object storage bucket initialization command suitable for staging.
- [x] 3.3 Clarify one-time initialization credentials versus long-running API/Worker credentials in deployment docs.
- [x] 3.4 Document service logs, dependency logs, and rollback expectations for staging.

## 4. CI Readiness

- [x] 4.1 Align GitHub Actions Go version with the repository-supported Go version.
- [x] 4.2 Update CI and `tools/check.mjs` so OpenSpec validation targets current specs or all relevant active planning artifacts rather than archived changes.
- [x] 4.3 Add CI image-build checks for API, Worker, and Web images without pushing or deploying them.
- [x] 4.4 Keep Web type check/build and Go test coverage in CI.

## 5. Staging Runbook And Verification

- [x] 5.1 Update release/deployment documentation with a first staging deployment runbook.
- [x] 5.2 Add smoke-test steps for account creation, family creation, image upload authorization, object upload, Worker processing, timeline visibility, media soft delete, and processing retry.
- [x] 5.3 Run or document local compose-based deployment verification for the new templates.
- [x] 5.4 Run Go tests, frontend check/build, Docker image builds, and OpenSpec strict validation.
