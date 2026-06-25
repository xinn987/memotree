## 1. Frontend Structure

- [ ] 1.1 Split shared API request helpers, route parsing/navigation helpers, formatting helpers, and user-facing copy maps out of the monolithic `App.tsx`.
- [ ] 1.2 Create feature-level React modules for auth/onboarding, family home, timeline, upload, family operations, and media detail while preserving current API behavior.
- [ ] 1.3 Create small shared UI primitives for buttons, icon buttons, fields, chips, empty states, inline errors, loading/skeleton states, and app shell layout.

## 2. Visual System

- [ ] 2.1 Port the relevant `design/tokens.css` Warm Editorial tokens into production `web/src/styles.css` or an equivalent imported CSS module.
- [ ] 2.2 Replace cold blue-gray MVP styling with warm paper surfaces, warm ink text, restrained brick-red accents, conservative radii, and non-ghost-card panel treatment.
- [ ] 2.3 Add complete interactive states for shared controls: hover, focus-visible, active, disabled, loading, error, and reduced-motion behavior.

## 3. Core Product Surfaces

- [ ] 3.1 Rebuild the authenticated app shell so brand, family context, account/logout, and family switching are clear without dominating the first viewport.
- [ ] 3.2 Rebuild auth and onboarding/join screens using the Warm Editorial emotional-panel/form direction and readable Chinese copy.
- [ ] 3.3 Rebuild the family home so timeline browsing is primary and upload/family operations are secondary but reachable.

## 4. Timeline And Media Detail

- [ ] 4.1 Rebuild timeline groups as mobile-first monthly story sections with photo-first layout, lightweight filters, empty state, refresh, pagination, and accessible media buttons.
- [ ] 4.2 Rebuild media detail as a warm immersive viewing surface with metadata, back navigation, admin media removal, loading/error states, and mobile-safe layout.
- [ ] 4.3 Decide and implement or explicitly defer previous/next media navigation based on available client-side timeline context.

## 5. Upload And Family Operations

- [ ] 5.1 Rebuild upload selection, active task summary, per-item progress, retry, retry-processing, stop, and recent tasks using family-readable Chinese copy.
- [ ] 5.2 Rebuild invite creation/recent invites and family member management with clear admin/member affordances, disabled states, and no impossible destructive action for non-admin users.
- [ ] 5.3 Rewrite touched user-facing copy to UTF-8 Chinese and replace backend terms such as task status, processing failure, and media asset with family-readable language.

## 6. Verification

- [ ] 6.1 Run `node tools/check-web.mjs` and fix TypeScript/build issues.
- [ ] 6.2 Run OpenSpec validation for `refactor-web-warm-editorial`.
- [ ] 6.3 Verify representative browser states on desktop and mobile: logged out, empty timeline, populated timeline, active upload, failed upload, media detail, admin family operations, and member-only family operations.
- [ ] 6.4 Check for visual regressions against the design direction: text overflow, overlapping controls, weak contrast, ghost-card patterns, repeated card-grid dominance, and keyboard focus visibility.
