## Why

MemoTree MVP now has the core album, upload, worker, and family operation flows, but it still runs like a local development application rather than a deployable service. Before real family data enters the system, the API, Worker, Web app, database, object storage, CI, and runbooks need a coherent staging/production runtime boundary.

This is the right moment to harden deployment because storage, soft delete, retry, and worker processing are now meaningful enough to test as a complete service, but the frontend redesign has not yet created more moving UI surface to deploy.

## What Changes

- Add production-oriented container definitions for API, Worker, and Web/static serving.
- Add deployment templates for single-host staging-style runtime wiring, including API, Worker, Web, local Docker MySQL, external S3-compatible object storage, service environment variables, health checks, and startup order.
- Clarify configuration profiles for `local`, `staging`, and `production`, including required secrets and safe defaults.
- Establish migration and bucket initialization strategy for deployable environments.
- Update CI so it verifies Go, Web, OpenSpec specs, and Docker image builds with versions aligned to the repository requirements.
- Update tooling and docs so `node tools/check.mjs` validates current specs rather than archived changes.
- Add a deployment runbook covering first staging deploy, smoke tests, logs, rollback, and data safety checks.
- Make the MVP deployment photo-only: upload paths reject videos, Worker deployment does not require FFmpeg, and video processing remains a later expansion path.
- Do not implement a provider-specific production deployment or full CD pipeline in this change.

## Capabilities

### New Capabilities

- `deployment-readiness`: Defines deployable runtime requirements for service images, environment configuration, migrations, object storage initialization, CI checks, and staging runbooks.

### Modified Capabilities

- None.

## Impact

- `deploy/`: API/Web/Worker Dockerfiles, single-host deployment templates, environment examples, and production-oriented compose or staging manifests.
- `.github/workflows/ci.yml`: Go/Web/OpenSpec checks and Docker image build validation.
- `tools/`: check script updates, possible image-build or deploy-smoke helper scripts.
- `docs/wiki/`: deployment, release, CI, environment, and operations documentation.
- `server/api/`, `server/worker/`, `web/`: minimal changes only where needed to support container runtime, static serving, health/readiness behavior, photo-only MVP upload policy, or configuration clarity.
