## 1. Release Env Generation

- [x] 1.1 Update ACR image publishing tooling so it writes `deploy/releases/staging-current.env` with image tags and release metadata after a successful push.
- [x] 1.2 Add validation that the generated release env contains no long-lived staging secrets.

## 2. Server Deployment Scripts

- [x] 2.1 Add a staging deploy shell script that requires `deploy/.env.staging` and a release env file, validates Compose config, pulls API/Worker/Web images, restarts business services, and runs health checks.
- [x] 2.2 Add a staging log shell script for recent and follow-mode API/Worker/Web/MySQL logs.
- [x] 2.3 Ensure routine deploy does not automatically run `schema-init` or `init-storage`.

## 3. Documentation

- [x] 3.1 Update the deployment runbook with the routine release env handoff, server deploy command, log helper, and rollback flow.
- [x] 3.2 Clarify that first-time infrastructure setup remains manual and separate from routine releases.

## 4. Verification

- [x] 4.1 Validate OpenSpec artifacts for `automate-staging-release-flow`.
- [x] 4.2 Run script syntax/help checks and Compose config validation.
