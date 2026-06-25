## Why

The current `web/` MVP works functionally, but it still reads like a cold admin tool: panel-heavy layout, blue-gray controls, repeated cards, and developer-facing upload/status copy. MemoTree's next useful step is to move the real React app toward the already-approved `design/` Warm Editorial direction so family members experience it as a private family album rather than a backend dashboard.

## What Changes

- Rebuild the `web/` presentation layer around the `design/` visual system: warm paper surfaces, restrained accent color, serif editorial headings, photo-first timeline rhythm, and consistent product controls.
- Refactor the large single-file React UI into clearer page/feature/component boundaries while keeping existing API contracts and route semantics.
- Convert the home timeline into a mobile-first monthly story flow with mixed photo presentation, lightweight filters, and upload access that remains visible but secondary.
- Redesign auth, onboarding/join, upload, invite/member management, media detail, loading, empty, and error states with family-readable Chinese copy.
- Preserve MVP scope: no public sharing, no AI organization, no full album/NAS file-management model, and no backend API rewrite in this change.

## Capabilities

### New Capabilities

- `web-warm-editorial-experience`: Defines the target frontend experience, visual migration boundaries, component expectations, responsive behavior, copy voice, and quality checks for the MemoTree React web app.

### Modified Capabilities

- None.

## Impact

- `web/src/app/`: React component structure, route-level screens, UI state composition, and user-facing copy.
- `web/src/styles.css`: design tokens, component styles, responsive layout, focus states, loading/empty/error states, and reduced-motion behavior.
- `design/`: remains the visual source of truth; no expected changes unless implementation reveals a small mismatch.
- API contracts: no intentional backend endpoint, request, or response changes.
- Dependencies: no new heavy frontend framework; optional lightweight additions must be justified before use.
