## Why

MemoTree can now run on the Alibaba Cloud staging server, but every update still requires copying image tags, editing environment files, pulling images, restarting containers, and checking health by hand. This is manageable for the first deployment, but it is too error-prone for routine iteration before the frontend rebuild.

## What Changes

- Add a release helper that publishes API, Worker, and Web images to ACR and writes a small release env file containing the image tags for that release.
- Add server-side staging deployment scripts that consume the release env file, pull business images, restart long-running services, and run health checks.
- Add an operational logs helper so API, Worker, Web, and MySQL logs can be inspected consistently during a failed deploy.
- Keep first-time setup manual: DNS, R2 CORS, ACR login, MySQL initialization, bucket initialization, and secret editing remain explicit operator steps.
- Document the routine update flow and the boundary between long-lived staging secrets and per-release image tags.

## Capabilities

### New Capabilities

- `staging-release-automation`: Defines routine staging release behavior, release env handoff, server deployment checks, and log inspection helpers.

### Modified Capabilities

- None.

## Impact

- `tools/`: release helper updates or a new script for producing release env files.
- `deploy/`: staging deployment, log, and possibly health-check shell scripts.
- `docs/wiki/`: deployment runbook updates for the routine release flow.
- OpenSpec: new staging release automation capability and tasks.
