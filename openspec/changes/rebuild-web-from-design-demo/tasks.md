## 1. Frontend Foundation

- [ ] 1.1 Replace the legacy frontend entry structure with `app`, `api`, `components`, `features`, `styles`, and `utils` module boundaries, including Chinese intent comments at file and critical-code level.
- [ ] 1.2 Add and configure the approved routing, production font, and frontend test dependencies without changing backend or deployment contracts.
- [ ] 1.3 Define typed API contracts for session, families, timeline, media, uploads, invitations, and members from the existing server responses.
- [ ] 1.4 Implement the shared HTTP client with API base path, credentials, JSON handling, abort support, and normalized user-facing errors.
- [ ] 1.5 Add the centralized frontend capability registry and placeholder feedback model for demo capabilities without backend support.
- [ ] 1.6 Configure unit/component test setup and add baseline tests for the HTTP client, capability registry, formatting, and status-copy utilities.

## 2. Visual System And Shared Components

- [ ] 2.1 Port the production design tokens from `design/tokens.css`, including exact colors, typography, spacing, radii, shadows, motion, z-index, and layout widths.
- [ ] 2.2 Load Fraunces and Inter with the demo-required weights and define global reset, Chinese fallbacks, focus, selection, image, and reduced-motion behavior.
- [ ] 2.3 Implement shared buttons, icon buttons, fields, selects, chips, progress bars, skeletons, empty states, inline errors, toasts, dialogs, and accessible disabled/loading/placeholder states.
- [ ] 2.4 Implement the shared desktop/mobile app bar, family context, avatar, page containers, auth split shell, and demo-consistent navigation entry points.
- [ ] 2.5 Add shared component tests covering keyboard focus, loading/disabled behavior, placeholder feedback, and reduced-motion-safe class/state output.

## 3. Application Routing And Session

- [ ] 3.1 Implement canonical React routes for login, invitation join, timeline, upload, invitations, members, and media detail.
- [ ] 3.2 Implement session provider, protected-route guard, family-access validation, root redirects, logout, and deep-link return behavior using existing auth/session APIs.
- [ ] 3.3 Implement family switching and no-family onboarding without reusing the legacy topbar, account bar, or split dashboard layout.
- [ ] 3.4 Add route/session tests for authenticated, unauthenticated, no-family, invalid-family, and deep-link states.

## 4. Authentication And Invitation Join

- [ ] 4.1 Rebuild login and registration to match `design/demo/auth.html` on desktop and mobile, including emotional image panel, typography, form, copy, loading, and error states.
- [ ] 4.2 Rebuild invitation joining to match `design/demo/join.html`, composing the existing register/login and invitation-join APIs without adding a backend endpoint.
- [ ] 4.3 Implement invalid, expired, used, and failed invitation states without changing the auth/join page composition.
- [ ] 4.4 Add tests for login, registration, join composition, session recovery, validation, and API failure recovery.

## 5. Timeline Experience

- [ ] 5.1 Rebuild the timeline page to match `design/demo/timeline.html`, including desktop hero/toolbar, mobile compact app bar/sticky filters, monthly headings, notes placeholder policy, and upload FAB.
- [ ] 5.2 Implement rendition-ratio-aware editorial photo layout, video/live-photo marks, hover/focus metadata, image loading/failure states, and media-detail navigation.
- [ ] 5.3 Reconnect existing media-type/month filters, refresh, pagination, loading skeleton, empty state, and retry behavior.
- [ ] 5.4 Add timeline tests for populated, empty, failed, filtered, paginated, long-copy, and keyboard-navigation states.

## 6. Upload Experience

- [ ] 6.1 Rebuild the upload route to match `design/demo/upload.html`, including desktop drop zone, mobile file action, active task, file rows, summary, controls, and recent uploads.
- [ ] 6.2 Reconnect upload intents, direct upload progress, completion/failure reporting, processing polling, retry-upload, retry-processing, stop, and timeline refresh behavior.
- [ ] 6.3 Implement supported-file validation, local-file retry limitations, browser-leave warning, family-readable status/error copy, and all progress/chip visual states.
- [ ] 6.4 Add upload tests for waiting, uploading, processing, ready, both failure types, cancellation, unsupported files, retry, stop, and interrupted navigation.

## 7. Invitation And Member Management

- [ ] 7.1 Rebuild invitation management to match `design/demo/invite.html`, including inline generated result, pending/resolved rows, mobile resolved-record collapse, and compact member summary.
- [ ] 7.2 Reconnect create/list/copy/revoke invitation behavior and administrator authorization without adding backend capabilities.
- [ ] 7.3 Rebuild member management to match `design/demo/members.html`, including desktop actions, mobile overflow menu, current-user treatment, warning copy, and confirmation dialog.
- [ ] 7.4 Reconnect member listing, display-name updates, and removals; implement administrator transfer and any other unsupported demo controls through the centralized honest-placeholder policy.
- [ ] 7.5 Add invitation/member tests for administrator, ordinary member, empty, loading, failed, long-name, resolved-invite, rename, remove, and placeholder-action states.

## 8. Media Detail

- [ ] 8.1 Rebuild media detail to match `design/demo/media-detail.html`, including warm dark viewer, desktop metadata rail, mobile metadata disclosure, top navigation, image/video handling, and admin removal.
- [ ] 8.2 Implement previous/next navigation from client-known timeline context and demo-consistent unavailable placeholders for direct deep links without adjacent context.
- [ ] 8.3 Implement placeholder controls for original download and capture-time editing until corresponding backend APIs are approved.
- [ ] 8.4 Add detail tests for photo, video, live-photo fallback, loading, not-found/error, admin/member permissions, removal, adjacent navigation, and placeholder actions.

## 9. Cleanup And Verification

- [ ] 9.1 Remove unused legacy components, route parsing, icons, styles, and dead dependencies after all new routes are connected.
- [ ] 9.2 Run frontend type checks, unit/component tests, production build, and the repository web check; fix all failures and warnings introduced by the rewrite.
- [ ] 9.3 Validate the OpenSpec change and confirm every requirement has an implemented or explicitly approved placeholder path.
- [ ] 9.4 Verify every page against its desktop and mobile demo screenshot, correcting material differences in layout, font, color, spacing, image ratio, fixed controls, and motion.
- [ ] 9.5 Verify keyboard focus, reduced motion, touch targets, text overflow, empty/error/loading states, and representative administrator/member permissions in the browser.
- [ ] 9.6 Document any remaining backend capability gaps for user confirmation without modifying the server in this change.
