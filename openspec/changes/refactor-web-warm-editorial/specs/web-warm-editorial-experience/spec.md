## ADDED Requirements

### Requirement: Warm Editorial Visual System
The React web app SHALL use the MemoTree Warm Editorial visual direction as its production frontend baseline.

#### Scenario: Production app uses warm album tokens
- **WHEN** a user opens the React web app
- **THEN** the UI uses warm paper surfaces, warm ink text, restrained brick-red accent actions, and the MemoTree type hierarchy derived from `design/` rather than the old cold blue-gray dashboard palette

#### Scenario: Controls remain product-familiar
- **WHEN** a user interacts with buttons, inputs, selects, chips, and icon actions
- **THEN** those controls use consistent product UI affordances with visible hover, focus, disabled, loading, and error states

#### Scenario: Decorative dashboard patterns are removed
- **WHEN** the app renders primary surfaces
- **THEN** it avoids ghost-card stacking, repeated generic card grids, decorative gradients, and section-level eyebrow labels that make the album read like an admin dashboard

### Requirement: Photo-First Family Home
The React web app SHALL make family timeline browsing the primary family-home experience.

#### Scenario: Family home prioritizes recent memories
- **WHEN** an authenticated family member opens a family space
- **THEN** the first browsing surface emphasizes recent family photos grouped by time before upload tasks, invite controls, or member administration

#### Scenario: Timeline uses monthly story rhythm
- **WHEN** timeline media is available
- **THEN** the app presents media in time-based story groups with album-like headings and photo-first layout rather than a dense management list

#### Scenario: Empty timeline teaches next action
- **WHEN** a family has no visible timeline media
- **THEN** the app shows a warm, understandable empty state that explains how the family can add selected photos without exposing backend terms

### Requirement: Secondary Operations Remain Reachable
The React web app SHALL keep upload, invite, member, and admin actions available without letting them dominate ordinary browsing.

#### Scenario: Upload entry is visible but secondary
- **WHEN** a family member is browsing the family home
- **THEN** the app provides a clear way to add selected photos while keeping the timeline as the main visual priority

#### Scenario: Admin tools are scoped to administrators
- **WHEN** an active administrator views family operations
- **THEN** invite, member management, and media removal controls are available with clear labels and disabled states where actions are not allowed

#### Scenario: Member tools avoid impossible actions
- **WHEN** a non-administrator member views family operations
- **THEN** the app does not present administrator-only destructive actions as available controls

### Requirement: Family-Readable Copy
The React web app SHALL translate technical state into Chinese copy that family members can understand.

#### Scenario: Upload status copy is human
- **WHEN** an upload item or task is waiting, uploading, processing, ready, failed, cancelled, or stopped
- **THEN** the app uses family-readable Chinese status labels such as waiting to upload, uploading, organizing, ready, not uploaded, retry, and stop for now rather than backend workflow terms

#### Scenario: Errors explain next action
- **WHEN** an operation fails in auth, upload, timeline, invite, member, or media detail surfaces
- **THEN** the app shows a concise Chinese error message and a visible recovery action when recovery is possible

### Requirement: Responsive And Accessible Album UI
The React web app SHALL remain usable on mobile and desktop with accessible interaction states.

#### Scenario: Mobile layout preserves core tasks
- **WHEN** the app is viewed on a mobile-width viewport
- **THEN** browsing, upload entry, auth, media detail, and family operations remain readable without text overflow or overlapping controls

#### Scenario: Desktop layout enhances browsing
- **WHEN** the app is viewed on a desktop-width viewport
- **THEN** the layout uses additional width to improve photo browsing and operation placement without turning into a dense admin dashboard

#### Scenario: Keyboard and reduced motion are supported
- **WHEN** a user navigates with keyboard or has reduced motion enabled
- **THEN** focus states remain visible and motion-dependent effects degrade without hiding content

### Requirement: Frontend Structure Supports Iteration
The React web app SHALL separate product surfaces and shared UI helpers enough to support continued frontend work.

#### Scenario: Surfaces are split from monolithic app file
- **WHEN** the refactor is complete
- **THEN** auth/onboarding, family home, timeline, upload, family operations, media detail, shared API helpers, shared formatting, and shared UI primitives are no longer all implemented as one monolithic component file

#### Scenario: API behavior remains compatible
- **WHEN** the refactored frontend performs existing auth, family, timeline, upload, invite, member, and media requests
- **THEN** it sends compatible requests to the existing API endpoints and handles the existing response shapes

#### Scenario: Frontend checks pass
- **WHEN** the frontend refactor is ready for review
- **THEN** TypeScript checks and production web build complete successfully
