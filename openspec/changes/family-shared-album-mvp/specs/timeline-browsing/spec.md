## ADDED Requirements

### Requirement: Family Timeline

The system SHALL provide a family timeline that displays uploaded photos and videos grouped by time.

#### Scenario: View recent timeline

- **WHEN** an active family member opens the album home screen
- **THEN** the system displays recent family media grouped by month and date

#### Scenario: Date grouping

- **WHEN** multiple media assets belong to the same calendar date
- **THEN** the system groups them under that date and identifies the uploading member where useful

### Requirement: Fast Timeline Loading

The system SHALL optimize timeline loading for mobile viewing by using preview assets instead of original files.

#### Scenario: Load preview assets

- **WHEN** the timeline renders photo or video items
- **THEN** the system loads thumbnails or video cover images instead of original media files

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
