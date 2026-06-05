## 1. Project Foundation

- [x] 1.1 Create React + Vite + TypeScript + Tailwind PWA project structure
- [x] 1.2 Create Go API project structure with configuration loading and health check
- [x] 1.3 Configure MySQL migrations and choose sqlc or GORM for data access
- [x] 1.4 Define environment configuration for R2 buckets, signed URL expiration, upload limits, and media processing
- [x] 1.5 Add baseline automated checks for frontend linting/type checking and backend tests

## 2. Family Access

- [ ] 2.1 Implement user authentication entry suitable for family members
- [ ] 2.2 Implement family creation with creator administrator membership
- [ ] 2.3 Implement invitation creation, revocation, expiration, and join flow
- [ ] 2.4 Enforce active family membership on all family media APIs

## 3. Media Upload

- [ ] 3.1 Implement media asset, original file, rendition, upload status, preview status, and soft deletion metadata model
- [ ] 3.2 Implement authorized upload flow to Cloudflare R2 private object storage
- [ ] 3.3 Implement batch upload UI with progress, partial failure handling, and retry
- [ ] 3.4 Implement Go Worker flow for photo thumbnail and display image generation after upload
- [ ] 3.5 Implement Go Worker flow for video thumbnail and display video generation with FFmpeg after upload
- [ ] 3.6 Show pending or processing state while preview assets are unavailable

## 4. Timeline Browsing

- [ ] 4.1 Implement timeline query grouped by month and date
- [ ] 4.2 Implement mobile-first timeline UI using preview assets
- [ ] 4.3 Implement incremental loading or pagination for large timelines
- [ ] 4.4 Implement media detail view for photos and videos
- [ ] 4.5 Implement lightweight filters for media type and month

## 5. Original Download

- [ ] 5.1 Implement permission-gated original media download authorization
- [ ] 5.2 Generate short-lived access URLs or equivalent authorized download responses
- [ ] 5.3 Add download actions to media detail views
- [ ] 5.4 Verify removed or non-member users cannot download original files

## 6. Verification

- [ ] 6.1 Add tests for family membership authorization and invitation edge cases
- [ ] 6.2 Add tests for upload success, unsupported files, partial batch failure, and retry behavior
- [ ] 6.3 Add tests for timeline grouping, pagination, and preview-only loading
- [ ] 6.4 Add tests for authorized and unauthorized original downloads
- [ ] 6.5 Perform mobile browser smoke testing for upload, timeline browsing, video detail, and download
