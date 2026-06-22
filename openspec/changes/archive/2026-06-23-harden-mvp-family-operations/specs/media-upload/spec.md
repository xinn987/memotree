## MODIFIED Requirements

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
- **THEN** the system keeps the media asset out of the main timeline while exposing processing status through the upload batch or upload item result

#### Scenario: Retry upload failure

- **WHEN** an upload item fails before the original file is stored
- **THEN** the system allows retrying that item by uploading the original file again

#### Scenario: Retry processing failure

- **WHEN** an upload item fails after the original file is stored but before display assets are ready
- **THEN** the system allows retrying media processing without requiring the original file to be uploaded again

#### Scenario: Upload creator can retry processing failure

- **WHEN** the upload creator retries a processing-failed upload item
- **THEN** the system resets that item and its media asset for background processing without creating a new upload item

#### Scenario: Administrator can retry family processing failure

- **WHEN** an active administrator retries a processing-failed upload item created by another family member
- **THEN** the system resets that item and its media asset for background processing

#### Scenario: Non-owner member cannot retry another member processing failure

- **WHEN** a non-administrator member retries another member's processing-failed upload item
- **THEN** the system denies the action
