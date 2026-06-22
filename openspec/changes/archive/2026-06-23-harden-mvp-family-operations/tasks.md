## 1. Store And Data Boundaries

- [x] 1.1 Audit existing MySQL schema and store models for member status, media status, upload item status, and rendition status reuse.
- [x] 1.2 Implement store operations for listing members, updating member display names, removing members, and protecting the last active admin.
- [x] 1.3 Implement store operations for soft deleting media and retrying processing-failed upload items.
- [x] 1.4 Keep MemoryStore and MySQLStore behavior consistent for all new operations.

## 2. API And Worker Behavior

- [x] 2.1 Add family member management routes with active admin authorization.
- [x] 2.2 Add media soft delete route with active admin authorization.
- [x] 2.3 Add processing retry route allowing upload creator or active admin to retry processing failures.
- [x] 2.4 Ensure removed members are denied for timeline, detail, upload authorization, upload task, invite, member, delete, and retry routes.
- [x] 2.5 Ensure Worker picks up retried processing items and can regenerate previews from stored originals.

## 3. Frontend Minimum Controls

- [x] 3.1 Add a minimal admin member management panel for listing members, editing display names, and removing members.
- [x] 3.2 Add a minimal admin media delete action on media detail.
- [x] 3.3 Add processing-failed retry action in upload task UI when the current user is allowed to retry.
- [x] 3.4 Keep member-facing copy family-friendly and avoid large visual redesign.

## 4. Verification

- [x] 4.1 Add store and API tests for member management, last admin protection, and removed member denial.
- [x] 4.2 Add store and API tests for media soft delete visibility and authorization.
- [x] 4.3 Add store/API/worker tests for processing failure retry.
- [x] 4.4 Run Go tests, frontend type check/build, and OpenSpec strict validation.
