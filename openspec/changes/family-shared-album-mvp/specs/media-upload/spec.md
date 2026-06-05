## ADDED Requirements

### Requirement: Photo And Video Upload

The system SHALL allow active family members to upload selected photo, video, and live photo files to a family space.

#### Scenario: Upload supported media

- **WHEN** an active family member selects supported photo or video files for upload
- **THEN** the system accepts the files and associates the resulting media assets with the selected family

#### Scenario: Create one media asset per timeline item

- **WHEN** a member uploads multiple independent photos or videos in one user flow
- **THEN** the system creates one media asset per photo or video rather than creating a post, album, or upload group for timeline browsing

#### Scenario: Reject unsupported file

- **WHEN** a user attempts to upload an unsupported file type
- **THEN** the system rejects that file and does not create a media asset for it

### Requirement: Batch Upload

The system SHALL allow active family members to upload multiple photos and videos in one user flow.

#### Scenario: Upload multiple files

- **WHEN** a member selects multiple supported media files
- **THEN** the system uploads each file independently and reports progress for each file or the batch

#### Scenario: Partial upload failure

- **WHEN** one file in a batch fails while other files succeed
- **THEN** the system preserves the successful uploads and marks the failed file as retryable

### Requirement: Private Original Storage

The system SHALL store original photos and videos as private objects that are not publicly accessible through permanent URLs.

#### Scenario: Store original media privately

- **WHEN** a media file upload completes
- **THEN** the original file is stored in private object storage and linked to media metadata without exposing a public permanent URL

#### Scenario: Preserve original file formats

- **WHEN** a member uploads HEIC, JPG, PNG, MOV, MP4, or another supported original media format
- **THEN** the system stores the original file without requiring the browser to render that original format directly

### Requirement: Media Metadata And Preview Generation

The system SHALL record basic media metadata and generate Web-compatible display assets for browsing and detail viewing.

#### Scenario: Process photo upload

- **WHEN** a photo upload completes
- **THEN** the system records file metadata and creates a thumbnail and display image suitable for browser rendering

#### Scenario: Process video upload

- **WHEN** a video upload completes
- **THEN** the system records file metadata and creates a thumbnail and display video suitable for browser rendering

#### Scenario: Process live photo upload

- **WHEN** the system identifies a still image and matching video as one live photo
- **THEN** the system creates one live photo media asset, stores both original files, and displays the static image by default

#### Scenario: Do not duplicate live photo video in timeline

- **WHEN** a live photo includes a video original
- **THEN** the system does not show that video original as a separate timeline media asset

#### Scenario: Processing pending state

- **WHEN** preview generation has not completed yet
- **THEN** the system shows the media asset in a pending or processing state rather than blocking the whole timeline

### Requirement: Media Deletion

The system SHALL allow administrators to remove mistaken or unwanted media from the family timeline without deleting the global user or member record.

#### Scenario: Administrator soft deletes media

- **WHEN** an active administrator deletes an active media asset
- **THEN** the system marks the media asset as deleted and removes it from timeline, detail, and download responses

#### Scenario: Member cannot delete media

- **WHEN** a non-administrator member attempts to delete a media asset
- **THEN** the system denies the action
