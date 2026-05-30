## ADDED Requirements

### Requirement: Photo And Video Upload

The system SHALL allow active family members to upload photo and video files to a family space.

#### Scenario: Upload supported media

- **WHEN** an active family member selects supported photo or video files for upload
- **THEN** the system accepts the files and associates the resulting media assets with the selected family

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

### Requirement: Media Metadata And Preview Generation

The system SHALL record basic media metadata and generate lightweight preview assets for browsing.

#### Scenario: Process photo upload

- **WHEN** a photo upload completes
- **THEN** the system records file metadata and creates at least one thumbnail suitable for timeline browsing

#### Scenario: Process video upload

- **WHEN** a video upload completes
- **THEN** the system records file metadata and creates a video cover image suitable for timeline browsing

#### Scenario: Processing pending state

- **WHEN** preview generation has not completed yet
- **THEN** the system shows the media asset in a pending or processing state rather than blocking the whole timeline
