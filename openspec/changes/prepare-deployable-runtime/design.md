## Context

MemoTree already has a local development runtime: Docker MySQL and MinIO, Go API, Go Worker, Vite Web app, schema migration SQL, storage initialization tooling, and GitHub Actions for Web and Go tests. The gap is that these pieces are not yet shaped as deployable runtime units for staging or production.

The next deployment step should prove that the application can run outside a developer terminal with persistent MySQL, S3-compatible object storage, health checks, logs, and repeatable verification. This is also the right time to correct drift in the existing checks, such as CI using Go 1.22 while the project targets Go 1.24+, and `tools/check.mjs` validating an archived OpenSpec change.

Stakeholders are the developer/operator running staging, family users whose media must not be lost, and future contributors who need a clear path to verify deployment changes.

## Goals / Non-Goals

**Goals:**

- Produce container images for API, Worker, and Web/static serving that can be built consistently in CI.
- Provide single-host staging-style deployment templates that wire API, Worker, Web, local Docker MySQL, and external S3-compatible storage together.
- Make runtime configuration explicit across `local`, `staging`, and `production`.
- Clarify schema migration and object bucket initialization responsibilities.
- Add CI coverage for specs, Go, Web, and Docker image builds.
- Document a first staging deployment runbook with smoke tests, logs, rollback, and data safety checks.
- Keep the MVP deployment photo-only so early operations do not depend on video transcoding capacity.

**Non-Goals:**

- No provider-specific production rollout to Alibaba OSS, Tencent COS, Cloudflare R2, or a specific MySQL hosting product.
- No full automatic CD pipeline that pushes images and deploys to a live host.
- No Kubernetes-only architecture; templates should stay usable for a small VPS or single-host staging environment.
- No frontend redesign or major UX migration.
- No full video-processing deployment path; video support is deferred to a later scaling-focused change.

## Decisions

### Decision: Use Three Runtime Images

Create separate runtime images for:

- `memotree-api`: Go API server and migration access.
- `memotree-worker`: Go Worker for photo rendition jobs; FFmpeg/FFprobe are not required for the MVP image.
- `memotree-web`: static Web build served by a small HTTP server such as Nginx or Caddy.

Rationale: API, Worker, and Web have different scaling, dependencies, and health characteristics. Early deployment only supports photos, so the Worker should stay lightweight and avoid FFmpeg as a hard dependency. API should expose `/healthz`; Worker should fail fast if required database or object storage dependencies are missing.

Alternatives considered:

- Single all-in-one image: simpler to run, but mixes static assets, API process, and Worker lifecycle.
- Shipping FFmpeg in the MVP Worker image: keeps video code easier to re-enable, but adds CPU-heavy dependency and startup requirements before the small VPS deployment can safely use it.
- API serving embedded Web assets: attractive later, but it would change API runtime packaging and frontend cache behavior in the same change.

### Decision: Deploy Photo-Only MVP

The staging and production-like runtime for this change should accept images only. The frontend should present an image-only picker, the API should reject video media types, and the Worker should process photo renditions without requiring FFmpeg.

Rationale: the initial server target is a small 2 vCPU / 2 GiB host running API, Worker, and MySQL together. Photo processing is predictable enough for that footprint; video transcoding would dominate CPU, memory, temporary disk, and operational debugging.

Alternatives considered:

- Keep video enabled but set Worker concurrency to 1: safer than higher concurrency, but still leaves long-running FFmpeg jobs as the first production bottleneck.
- Remove video code entirely: reduces surface now, but throws away useful scaffolding. Keeping the code path dormant is better until video support is revisited.

### Decision: Use Single-Host Staging First

The deployment template should target one VPS running Web, API, Worker, and MySQL through Docker Compose, while R2 or another S3-compatible service remains external.

Rationale: this matches the current cost and operations plan. A separate staging server is not required at this stage; the first real deployment can begin as `staging` and be promoted operationally to `production` once data, backups, and smoke tests are trusted.

Alternatives considered:

- Managed MySQL from day one: operationally cleaner, but cost is high for the MVP.
- Separate staging and production hosts now: cleaner isolation, but not necessary before there is real user traffic.

### Decision: Keep Deployment Templates Vendor-Neutral

Add example deployment files that use environment variables and S3-compatible storage settings rather than provider-specific SDK or infrastructure assumptions.

Rationale: the repository already treats R2/OSS/COS/MinIO through the S3-compatible adapter. The deployment layer should preserve that boundary until a real provider is selected.

Alternatives considered:

- Pick one cloud provider now: faster for one environment, but it would prematurely encode cost, region, ICP/备案, and network assumptions.

### Decision: Prefer Explicit One-Time Initialization Commands

For staging, schema migration and bucket initialization should be explicit deploy steps or one-shot commands, not hidden inside every service process. The current API startup migration can remain for local development, but the runbook should make production expectations clear.

Rationale: implicit startup migrations are convenient locally but risky in multi-instance or production-like deploys. Bucket creation also needs clear permissions that may not belong to the long-running API or Worker credentials.

Alternatives considered:

- Keep API startup migration as the only path: easy but creates ambiguity about who owns schema changes in production.
- Build a full migration framework now: useful later, but too large for this deployment-readiness change.

### Decision: CI Builds Images But Does Not Deploy

CI should run Web checks, Go tests, OpenSpec spec validation, and Docker image build checks for API, Worker, and Web. It should not push images or deploy automatically yet.

Rationale: image build coverage catches missing files, wrong Go versions, Worker packaging issues, and static build assumptions without committing to registry or environment secrets.

Alternatives considered:

- Add CD now: valuable later, but it forces registry, credentials, environment, and rollback decisions before staging has been proven manually.
- Push to a container registry in this change: useful soon, but the first iteration can build locally on the server or in CI without choosing GHCR, Docker Hub, or a cloud registry yet.

### Decision: Staging Runbook Is Part Of The Deliverable

The change should update docs with a copy-pasteable staging flow:

1. Build or pull images.
2. Configure secrets and environment.
3. Run migration and bucket initialization.
4. Start services.
5. Run smoke tests for registration, family creation, image upload, Worker processing, timeline, soft delete, and retry.
6. Inspect logs and rollback if needed.

Rationale: deployment readiness is only real if someone can repeat it.

## Risks / Trade-offs

- Provider-neutral templates may be less turnkey than a provider-specific guide -> Keep the template concrete enough for Docker Compose staging, and document where provider values go.
- Startup migration behavior can diverge between local and staging -> Document the intended difference and keep local convenience scripts intact.
- Docker image build checks may slow CI -> Start with build-only checks, use layer caching only if CI time becomes painful.
- Web static serving may need API base URL decisions -> Use same-origin `/api` routing in the example server config where possible.
- Worker readiness is harder than API readiness because it is a background process -> Require fail-fast dependency checks and log-based heartbeat guidance rather than a fake HTTP endpoint unless a real endpoint is introduced.
- Photo-only policy can surprise users who select videos -> Make the frontend and API error explicit, and document video as a deferred capability rather than a broken upload type.

## Migration Plan

1. Add and validate runtime Dockerfiles and static Web serving config.
2. Add deployment environment examples and staging compose/template files.
3. Update initialization scripts and docs so schema and buckets can be prepared explicitly.
4. Update CI to validate specs and image builds.
5. Update deployment docs and run a local compose-based smoke test.

Rollback strategy for this change is simple because it does not change production data shape: revert deployment files, CI changes, and docs. If API startup behavior is changed, preserve local development compatibility and document the fallback command.

## Open Questions

- Should staging initially use a single-host Docker Compose deployment, or should the repository also include a second template for a managed platform?
- Should Web be served by Nginx or Caddy in the example image?
- Should API startup migration remain enabled in staging by default, or should staging require a separate migration command from the start?
