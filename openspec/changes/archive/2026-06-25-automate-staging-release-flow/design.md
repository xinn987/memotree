## Context

The first real staging deployment proved the MemoTree runtime works on a small Alibaba Cloud server with API, Worker, Web, Docker MySQL, ACR images, and R2 storage. The remaining friction is routine deployment: every code update currently requires manual image publishing, copying image tags, editing staging env values, pulling images, restarting containers, and checking health.

The main stakeholder is the operator doing frequent MVP iteration. The deployment scripts should reduce repeatable mistakes without hiding first-time setup responsibilities such as R2 CORS, ACR authentication, server security groups, and staging secrets.

## Goals / Non-Goals

**Goals:**

- Produce a per-release env file containing only release metadata and API/Worker/Web image tags.
- Keep long-lived secrets and runtime settings in `deploy/.env.staging`.
- Provide server-side scripts for routine staging updates: validate config, pull business images, restart API/Worker/Web, and run health checks.
- Provide a log helper that makes failure triage predictable.
- Document the operator handoff from local image publish to server deploy.

**Non-Goals:**

- No fully automatic SSH deployment in the first iteration.
- No automatic mutation of R2 CORS, ACR permissions, Alibaba Cloud security groups, or domain settings.
- No automatic database migration framework beyond the existing explicit schema initialization flow.
- No automatic rollback of MySQL data.
- No GitHub Actions CD pipeline yet.

## Decisions

### Decision: Use A Release Env File As The Handoff

The local release helper will write `deploy/releases/staging-current.env` with `API_IMAGE`, `WORKER_IMAGE`, `WEB_IMAGE`, `RELEASE_COMMIT`, and `RELEASE_CREATED_AT`. The server deploy script will accept that file as an argument and use it together with `deploy/.env.staging`.

Rationale: a release env file is easy to inspect, copy with `scp`, archive, and roll back. It avoids repeatedly editing the secret-bearing `.env.staging` file.

Alternatives considered:

- Directly edit `deploy/.env.staging`: convenient but mixes release tags with secrets and makes rollback harder.
- Fully automatic SSH from local script: attractive later, but it introduces SSH key, remote path, and remote failure-handling complexity before the basic deployment interface is stable.

### Decision: Keep Server Scripts Shell-Based

Server scripts will be POSIX shell scripts under `deploy/`, because the Alibaba Cloud server already has Docker Compose and shell available while Node.js was intentionally avoided during deployment.

Rationale: routine server deployment should not require installing Node on the server.

Alternatives considered:

- Node-based server deploy script: easier to share code with existing tools, but adds a runtime dependency to the server.
- Makefile: concise, but less friendly for explicit health and log handling on a minimal server.

### Decision: Restart Only Long-Running Business Services

Routine deploy will pull and restart `api`, `worker`, and `web`. It will not automatically run `schema-init` or `init-storage`.

Rationale: schema and bucket initialization are privileged, infrequent operations. Keeping them explicit prevents a routine application deploy from accidentally touching data or storage configuration.

Alternatives considered:

- Always run `schema-init`: simple, but weakens the separation between one-time initialization and daily deploy.
- Run `init-storage` on every deploy: unnecessary after buckets exist and requires broader credentials than daily runtime may need.

### Decision: Health Checks Gate Success

The server deploy script will run Docker Compose config validation, pull images, restart services, print service status, and verify `http://127.0.0.1/healthz` and `http://127.0.0.1/api/healthz`.

Rationale: a deploy is not complete just because containers started. The Web and API health endpoints are the fastest signal that the release is reachable on the host.

Alternatives considered:

- Trust Docker Compose exit codes only: insufficient because containers can start but fail health checks.
- Add a Worker HTTP health endpoint now: useful later, but out of scope for this small deployment automation change.

## Risks / Trade-offs

- Release env file can drift from server `.env.staging` if copied incorrectly -> The deploy script will require the file and print the release metadata it is using.
- Compose support for multiple env files may vary by version -> The script will source the release env into the process environment and use the existing `--env-file deploy/.env.staging` for long-lived settings.
- Restart may pass health checks while Worker later fails processing jobs -> The log helper keeps Worker logs easy to inspect; deeper Worker readiness can be a later observability change.
- Manual `scp` is still a step -> This is intentional for the first iteration and can be automated later once the release env contract is stable.

## Migration Plan

1. Add local release env generation to the image publishing flow.
2. Add server deployment and log scripts.
3. Update documentation with the routine update workflow.
4. Validate scripts locally where possible without requiring live server credentials.

Rollback for this change is operational: keep prior release env files and re-run the server deploy script with an older file after confirming the older images still exist in ACR.
