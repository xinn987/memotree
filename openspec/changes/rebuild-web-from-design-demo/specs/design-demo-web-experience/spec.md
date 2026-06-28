## ADDED Requirements

### Requirement: Production Frontend Uses The Design Demo As Its Visual Contract
The production React frontend SHALL reproduce the visual system and page composition defined by `design/tokens.css`, `design/demo/assets/demo.css`, the seven demo pages, and their desktop/mobile reference screenshots.

#### Scenario: Shared visual system matches the demo
- **WHEN** any production page is rendered
- **THEN** it uses the demo's warm paper surfaces, warm ink colors, brick-red accent, Fraunces/Inter hierarchy, spacing, restrained radii, shadows, and component states

#### Scenario: Production pages do not retain the legacy dashboard style
- **WHEN** an authenticated or unauthenticated page is rendered
- **THEN** it does not use the old cold blue-gray palette, ghost-card stack, repeated dashboard panels, or legacy class structure

#### Scenario: Demo-only infrastructure is not shipped
- **WHEN** production styles and markup are built
- **THEN** demo navigation, inline demo data, screenshot tooling, and static-page-only behavior are excluded

### Requirement: Frontend Is Rewritten As A Formal Modular Application
The frontend SHALL be organized into application, API, shared component, feature, style, and utility boundaries without requiring compatibility with the old frontend modules.

#### Scenario: Product features are isolated
- **WHEN** a developer inspects the source tree
- **THEN** authentication, timeline, upload, invites, members, and media detail each have a clear feature boundary and do not share one monolithic page component

#### Scenario: Shared infrastructure has one responsibility
- **WHEN** multiple features need HTTP, routing, layout, controls, feedback, formatting, or copy behavior
- **THEN** the shared behavior is provided through a typed, focused module instead of duplicated feature code

#### Scenario: Legacy implementation is removed
- **WHEN** the rewritten frontend is complete
- **THEN** no active route depends on the old `FamilyHome`, `TimelinePanel`, `UploadPanel`, legacy page composition, or legacy stylesheet rules

### Requirement: Product Surfaces Have Canonical Routes
The frontend SHALL provide canonical routes for login, invitation joining, timeline, upload, invitation management, member management, and media detail.

#### Scenario: Authenticated family navigation
- **WHEN** an authenticated member navigates among timeline, upload, invites, members, and media detail
- **THEN** each surface has a stable family-scoped URL and browser back/forward navigation works

#### Scenario: Protected deep link restores after authentication
- **WHEN** an unauthenticated user opens a protected family route
- **THEN** the frontend sends the user through login and returns to an allowed destination after session recovery

#### Scenario: Invalid family route is rejected
- **WHEN** the route references a family outside the authenticated user's visible families
- **THEN** the frontend redirects to a visible family or an understandable empty/onboarding surface without issuing unauthorized feature requests

### Requirement: Authentication And Invitation Joining Match The Demo
The login, registration, and invitation-joining flows SHALL use the demo's emotional image panel and focused form layout on desktop and its retained image banner layout on mobile.

#### Scenario: Existing member logs in
- **WHEN** a user opens the login route
- **THEN** the page matches `auth.html`, submits through the existing login API, and exposes readable loading and error states

#### Scenario: New account registers without an invitation
- **WHEN** a user chooses the supported registration path
- **THEN** the frontend collects the existing API fields, creates the account, restores the session, and shows family onboarding when no family exists

#### Scenario: Invited family member joins
- **WHEN** a valid invitation token is present and the visitor needs an account
- **THEN** the frontend presents the `join.html` design and composes the existing register/login and join APIs without requiring a new backend endpoint

#### Scenario: Invitation join fails
- **WHEN** the invitation is invalid, expired, used, or the join request fails
- **THEN** the page keeps the demo layout, explains the failure in family-readable Chinese, and offers a valid recovery action

### Requirement: Timeline Is The Primary Family Experience
The authenticated family home SHALL lead with a photo-first monthly story timeline matching `timeline.html`.

#### Scenario: Timeline contains media
- **WHEN** timeline data is returned
- **THEN** media is grouped into month sections, image dimensions reserve their real aspect ratios, and desktop/mobile layouts follow the demo's editorial photo rhythm

#### Scenario: Timeline filtering
- **WHEN** a user changes media type or month filters
- **THEN** the existing timeline query contract is used and the desktop segmented toolbar or mobile sticky filter bar reflects the active filter

#### Scenario: Timeline pagination
- **WHEN** more timeline media is available
- **THEN** the user can request earlier items without losing the current groups or disrupting the photo layout

#### Scenario: Timeline is empty or fails
- **WHEN** the family has no media or the timeline request fails
- **THEN** the page shows a demo-consistent teaching empty state or inline retry state instead of a blank grid or backend terminology

#### Scenario: Upload remains secondary
- **WHEN** a member browses the timeline
- **THEN** a demo-matched upload FAB remains reachable without replacing or visually dominating the timeline

### Requirement: Upload Surface Preserves Existing Upload Behavior
The upload route SHALL match `upload.html` while preserving the current upload-intent, direct upload, completion/failure reporting, processing polling, retry, stop, and recent-task behavior.

#### Scenario: Desktop member selects files
- **WHEN** a user selects supported files from the desktop drop zone
- **THEN** the frontend creates upload intents, uploads originals, reports progress, and renders the active task in the demo task layout

#### Scenario: Mobile member selects files
- **WHEN** a user opens the upload page on a mobile viewport
- **THEN** the large drag target is replaced by the demo's mobile file action while task status remains readable and unobstructed

#### Scenario: Upload item changes state
- **WHEN** an item is waiting, uploading, uploaded, processing, ready, upload-failed, processing-failed, or cancelled
- **THEN** its chip, progress bar, copy, and available recovery action use the demo's family-readable state vocabulary

#### Scenario: Unsupported file is selected
- **WHEN** a file type outside the backend-supported MVP set is selected
- **THEN** the frontend refuses it before creating an upload intent and explains the supported format without breaking the task layout

#### Scenario: Browser upload is interrupted
- **WHEN** a user attempts to leave while original files are still uploading from the browser
- **THEN** the frontend gives an explicit interruption warning while background processing-only work may continue

### Requirement: Invitation Management Matches The Demo
Administrators SHALL manage family invitations through a dedicated route matching `invite.html`.

#### Scenario: Administrator creates an invitation
- **WHEN** an administrator submits a family member name
- **THEN** the existing invitation API creates the invitation and the generated link expands inline in the same action block with a copy action

#### Scenario: Administrator reviews invitations
- **WHEN** invitation records are loaded
- **THEN** pending invitations remain actionable, resolved invitations are visually de-emphasized, and mobile can collapse resolved records as shown in the demo

#### Scenario: Administrator revokes an invitation
- **WHEN** an administrator chooses to invalidate a pending invitation
- **THEN** the existing revoke API is called and the row updates without claiming success before the response

#### Scenario: Non-administrator reaches invitation management
- **WHEN** a member without invitation-management permission opens the route
- **THEN** administrator controls are not exposed as working operations and the user receives a clear route-level permission state

### Requirement: Member Management Matches The Demo
Administrators SHALL manage family members through a dedicated route matching `members.html`, using real existing API behavior where available.

#### Scenario: Administrator views members
- **WHEN** member data is loaded
- **THEN** the page shows the current user, names, roles, statuses, desktop actions, and mobile overflow actions in the demo's layout

#### Scenario: Administrator changes a member name
- **WHEN** an administrator submits a valid changed display name
- **THEN** the existing member update API persists it and the current-user family label is refreshed when applicable

#### Scenario: Administrator removes a member
- **WHEN** an administrator confirms removal of another active member
- **THEN** the existing remove API is called through a demo-matched confirmation dialog and success is reflected only after the response

#### Scenario: Impossible destructive action is prevented
- **WHEN** an action targets the current user, an already removed member, or another backend-forbidden state
- **THEN** the frontend does not send the request and exposes a clear disabled or explanatory state

### Requirement: Media Detail Matches The Demo
The media detail route SHALL provide the warm immersive viewing surface defined by `media-detail.html`.

#### Scenario: Photo or video loads
- **WHEN** the existing media detail API returns a photo, video, or live-photo preview
- **THEN** the viewer centers the display rendition on the warm dark surface and renders the available metadata in the demo layout

#### Scenario: Administrator removes media
- **WHEN** an administrator confirms the supported remove action
- **THEN** the existing delete API is called and the user returns to the timeline only after success

#### Scenario: Detail is opened from timeline context
- **WHEN** the user opens media from a currently loaded timeline sequence
- **THEN** previous and next controls navigate within that client-known sequence without requiring a new backend endpoint

#### Scenario: Detail lacks adjacent context
- **WHEN** the detail route is opened directly and no adjacent-media API exists
- **THEN** previous and next controls preserve the demo layout but expose an honest unavailable placeholder state

### Requirement: Unsupported Demo Capabilities Use Honest Placeholders
Capabilities visible in the demo but unsupported by the current backend SHALL preserve their designed position without pretending that data was persisted.

#### Scenario: User activates a placeholder action
- **WHEN** a user activates an unsupported but visible action such as administrator transfer, original download, or capture-time editing
- **THEN** no backend request or persistent local mutation occurs and the frontend displays a consistent “暂未开放” response

#### Scenario: Placeholder is intentionally disabled
- **WHEN** an unsupported action must not be interactive
- **THEN** it remains visually consistent with the demo, is marked disabled or `aria-disabled`, and has an accessible explanation

#### Scenario: Unsupported content field is displayed
- **WHEN** the demo reserves space for a family note, title, or other field absent from API data
- **THEN** the frontend uses neutral placeholder copy and never invents a specific family memory, author, or persisted value

#### Scenario: Backend support is added later
- **WHEN** a future change supplies a real endpoint or field
- **THEN** the placeholder can be replaced through the feature capability and API adapter without restructuring the page

### Requirement: Existing Backend Contracts Remain Stable
The frontend rewrite SHALL use the existing backend endpoints, request payloads, response shapes, authentication model, and authorization behavior unless a separate approved change modifies them.

#### Scenario: Supported operation is executed
- **WHEN** the rewritten frontend performs an existing auth, family, timeline, upload, invitation, member, or media operation
- **THEN** it sends a request compatible with the current API contract and handles the current response shape

#### Scenario: Required backend capability is missing
- **WHEN** implementation discovers that a demo behavior cannot be delivered honestly without new server behavior
- **THEN** the real integration pauses for user confirmation and the current change keeps only the approved placeholder

### Requirement: Responsive, Accessible And Motion-Safe Behavior Is Complete
Every production page SHALL remain usable at demo desktop and mobile sizes and SHALL provide accessible interaction states.

#### Scenario: Mobile layout
- **WHEN** the viewport is 720 pixels wide or narrower
- **THEN** each page follows its demo mobile composition without horizontal overflow, covered actions, clipped text, or desktop-only controls

#### Scenario: Keyboard navigation
- **WHEN** a user navigates with a keyboard
- **THEN** interactive elements have logical order, visible focus, meaningful labels, and reachable equivalents for hover-only information

#### Scenario: Reduced motion
- **WHEN** `prefers-reduced-motion: reduce` is active
- **THEN** nonessential transforms and animated transitions are removed without hiding content or breaking feedback

#### Scenario: Dynamic content extremes
- **WHEN** names, filenames, errors, dates, or counts are empty, unusually short, or unusually long
- **THEN** the interface wraps, truncates, or falls back predictably while preserving the demo hierarchy

### Requirement: Frontend Quality Is Verifiable
The rewritten frontend SHALL pass automated structural and behavioral checks and SHALL be manually verified against the demo references.

#### Scenario: Automated checks
- **WHEN** the change is ready for review
- **THEN** TypeScript checks, unit/component tests, and the production web build complete successfully

#### Scenario: Browser visual verification
- **WHEN** representative logged-out, joining, populated timeline, empty timeline, active upload, failed upload, invitation, member, and detail states are rendered
- **THEN** desktop and mobile screenshots have been compared with the corresponding demo references and material discrepancies have been corrected

#### Scenario: Interaction-state verification
- **WHEN** controls are inspected
- **THEN** default, hover, focus-visible, active, disabled, loading, error, placeholder, and reduced-motion states are present where applicable
