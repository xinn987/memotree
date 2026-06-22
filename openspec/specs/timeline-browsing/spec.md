# timeline-browsing Specification

## Purpose
TBD - created by archiving change family-shared-album-mvp. Update Purpose after archive.
## Requirements
### Requirement: Family Timeline

The system SHALL provide a family timeline that displays selected family photos and videos grouped by time.

#### Scenario: View recent timeline

- **WHEN** an active family member opens the album home screen
- **THEN** the system displays recent family media grouped by month and date

#### Scenario: Date grouping

- **WHEN** multiple media assets belong to the same calendar date
- **THEN** the system groups them under that date and identifies the uploading member where useful

#### Scenario: Render media assets, not original files

- **WHEN** the timeline renders family media
- **THEN** the system renders media asset records and their display assets rather than rendering original file records directly

#### Scenario: Sort by captured time first

- **WHEN** a media asset has a captured time
- **THEN** the system places it in the main timeline according to captured time

#### Scenario: Fall back to uploaded time

- **WHEN** a media asset does not have a captured time
- **THEN** the system places it in the main timeline according to uploaded time

#### Scenario: Preserve uploaded time separately

- **WHEN** a media asset appears in the main timeline
- **THEN** the system preserves uploaded time separately for detail metadata, audit, and future recently-added views

### Requirement: Fast Timeline Loading

The system SHALL optimize timeline loading for mobile viewing by using preview assets instead of original files.

#### Scenario: Load preview assets

- **WHEN** the timeline renders photo or video items
- **THEN** the system loads thumbnails, display images, or display videos instead of original media files

#### Scenario: Display live photo as one item

- **WHEN** the timeline renders a live photo media asset
- **THEN** the system displays it as one timeline item using its static image rendition by default

#### Scenario: Paginate timeline

- **WHEN** the family contains more media than the first timeline page can display efficiently
- **THEN** the system loads additional media through pagination or incremental loading

### Requirement: Media Detail View

The system SHALL allow active family members to open a media asset and view its detail.

#### Scenario: Open photo detail

- **WHEN** a member selects a photo from the timeline
- **THEN** the system displays a larger photo view with basic metadata

#### Scenario: Open video detail

- **WHEN** a member selects a video from the timeline
- **THEN** the system displays a playable video view with basic metadata

### Requirement: Minimal Browsing Filters

The system SHALL support lightweight browsing filters only when they help users find recent family media without turning the MVP into a media management system.

#### Scenario: Filter by media type

- **WHEN** a member chooses photos only or videos only
- **THEN** the timeline displays only matching media assets

#### Scenario: Filter by month

- **WHEN** a member selects a month
- **THEN** the timeline displays media assets from that month

### Requirement: Media Visibility After Deletion

The system SHALL allow administrators to hide mistaken or unwanted media from family browsing without deleting user accounts or membership history.

#### Scenario: Administrator soft deletes media

- **WHEN** an active administrator deletes an active media asset
- **THEN** the system marks the media asset as deleted and removes it from timeline and detail responses

#### Scenario: Member cannot delete media

- **WHEN** a non-administrator member attempts to delete a media asset
- **THEN** the system denies the action

#### Scenario: Removed member cannot delete media

- **WHEN** a removed member attempts to delete a media asset
- **THEN** the system denies the action

#### Scenario: Deleted media remains unavailable by direct detail request

- **WHEN** an active family member requests the detail endpoint for a deleted media asset
- **THEN** the system responds as if the media is not available for browsing
