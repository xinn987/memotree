## ADDED Requirements

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
