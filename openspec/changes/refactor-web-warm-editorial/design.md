## Context

`web/src/app/App.tsx` currently contains routing, session handling, upload orchestration, timeline rendering, invite/member management, media detail, and most UI markup in one large file. `web/src/styles.css` is a functional MVP stylesheet, but its cold blue-gray palette, card-heavy panels, ghost-card shadows, repeated small eyebrow labels, and backend-like copy conflict with MemoTree's product direction.

The desired visual system already exists in `design/` and is documented by `PRODUCT.md`, `DESIGN.md`, `design/tokens.css`, and the static demo pages. This change migrates the real React app toward that system while preserving current backend API contracts and the existing deployment/runtime shape.

Primary stakeholders are family viewers, especially older relatives, and uploaders/admins who need simple private-album workflows without feeling like they are operating a storage dashboard.

## Goals / Non-Goals

**Goals:**

- Make the production React app feel like a warm, private family album rather than an admin console.
- Establish reusable frontend structure: app shell, UI primitives, feature pages, feature hooks/helpers, and shared formatting/copy helpers.
- Port the visual language from `design/`: Warm Editorial colors, restrained shadows, serif editorial headings where appropriate, photo-first timeline, family-readable Chinese copy, and mobile-first layouts.
- Keep current API calls, auth/session behavior, upload task behavior, invite/member operations, and media deletion semantics intact.
- Improve loading, empty, error, disabled, focus, and mobile states enough that the app is usable during real MVP iteration.

**Non-Goals:**

- No backend API redesign, data model changes, or new storage behavior.
- No public sharing, comment/like social features, AI organization, face recognition, or NAS/file-manager concepts.
- No full design-system package extraction; component structure can live inside `web/src/` for now.
- No heavy frontend dependency adoption unless implementation proves a narrow need.
- No production font hosting work in this change unless it is needed to avoid layout or network regressions.

## Decisions

### Decision: Keep API/Data Contracts Stable

The refactor will preserve existing endpoint usage and response types. Data fetching and mutation helpers may move into smaller files, but request payloads and response handling remain behaviorally compatible.

Rationale: the backend was just made deployable and family operations were hardened. Changing API contracts now would mix product polish with backend risk.

Alternatives considered:

- Rewrite the client around a query library: useful later, but adds dependency and migration surface before the UI structure is stable.
- Redesign endpoints around new views: out of scope for a frontend migration.

### Decision: Split By Product Surface, Not By Generic Widget First

The app will be split into surfaces such as auth/onboarding, family home, timeline, upload, family operations, and media detail. Shared primitives such as buttons, fields, chips, empty states, and icon buttons should be extracted only when two or more surfaces need the same interaction pattern.

Rationale: the current problem is product-surface entanglement. Starting with too many generic abstractions would slow iteration and hide real workflow differences.

Alternatives considered:

- Build a complete component library first: too much upfront work for MVP.
- Keep a single file and only rewrite CSS: faster at first, but keeps the app hard to reason about and makes future polish brittle.

### Decision: Use `design/` Tokens As The Visual Source, Adapted Into `web/src/styles.css`

The implementation should copy or map the relevant OKLCH tokens from `design/tokens.css` into `web/src/styles.css`, then define production class names around React surfaces. `design/` remains reference material; the production app does not need to import demo CSS directly.

Rationale: demo CSS contains demo-only navigation, CDN font imports, and static-page assumptions. Production CSS should be deliberate and scoped to the real app.

Alternatives considered:

- Import `design/demo/assets/demo.css` directly: tempting, but it would leak demo-only rules and make production behavior harder to audit.
- Keep the old palette and only tweak layout: misses the main product experience goal.

### Decision: Make Timeline The First-Class Home Surface

The family home should lead with the timeline and photos. Upload, invites, member management, and admin actions remain available but should not dominate the first viewport for ordinary family viewers.

Rationale: MemoTree succeeds when family members quickly see recent moments. Admin controls are necessary, but they are not the product's emotional center.

Alternatives considered:

- Keep the current two-column dashboard: efficient for development, but visually reads as a management console.
- Move all operations to separate routes now: cleaner eventually, but routing breadth can wait until the first visual migration lands.

### Decision: Use Photo-First Responsive Layouts With Conservative Motion

Timeline media should use preview renditions, preserve meaningful aspect ratios, and avoid same-sized card repetition where possible. Motion is limited to short state transitions, focus/hover feedback, progress updates, and media detail transitions, with reduced-motion fallback.

Rationale: photos are the product. Motion and layout should help browsing, not perform.

Alternatives considered:

- Masonry library: avoid for now; CSS grid/columns and simple span rules are enough for MVP.
- Fully uniform grid: simpler, but it preserves the current file-browser feeling.

## Risks / Trade-offs

- Large React split may introduce regressions → Keep tasks staged, run `node tools/check-web.mjs`, and preserve existing API helpers until surfaces are stable.
- Warm typography can hurt readability if overused → Use serif display only for brand, month, page titles, and album notes; keep controls and dense UI in sans.
- Mixed photo layout can crop important content → Use existing rendition aspect ratios and conservative object-fit behavior; detail view remains the full inspection surface.
- Moving upload/admin controls down in hierarchy may make them harder for admins → Keep visible entry points and admin sections, but reduce visual dominance on normal browsing.
- Existing Chinese copy appears garbled in some source/tool outputs → Rewrite touched user-facing copy as UTF-8 Chinese and verify in browser rather than trusting terminal rendering.

## Migration Plan

1. Establish production tokens, base styles, and shared UI primitives in `web/src/`.
2. Move API types/helpers and formatting/copy helpers out of the monolithic app file without changing behavior.
3. Rebuild app shell/auth/onboarding around Warm Editorial layout and accessible form states.
4. Rebuild family home so timeline is primary, with upload and family operations as secondary but reachable surfaces.
5. Rebuild timeline, upload task, family members/invites, and media detail components incrementally.
6. Validate TypeScript/build, run browser checks on desktop/mobile, and inspect representative states: logged out, empty family, populated timeline, active upload, failed upload, media detail, admin family operations.

Rollback is code-level: revert the frontend refactor commit(s). No database or backend migration is introduced.

## Open Questions

- Whether to self-host Fraunces/Inter immediately or use the existing font stack until production font hosting is worth the setup.
- Whether upload should remain inline on the family home for this iteration or become a route-like overlay/page after the first migration.
- Whether media detail previous/next navigation should be implemented from currently loaded timeline items only, or wait for backend-supported adjacent-media queries.
