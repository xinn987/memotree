## ADDED Requirements

### Requirement: Download Original Media

The system SHALL allow active family members to download original photo and video files.

#### Scenario: Download original photo

- **WHEN** an active family member requests to download an original photo
- **THEN** the system authorizes the request and provides access to the original photo file

#### Scenario: Download original video

- **WHEN** an active family member requests to download an original video
- **THEN** the system authorizes the request and provides access to the original video file

### Requirement: Permission-Gated Download Links

The system SHALL gate original media downloads through membership authorization and short-lived access.

#### Scenario: Generate short-lived download link

- **WHEN** an active family member requests an original media download
- **THEN** the system verifies family membership and returns a short-lived download URL or equivalent authorized response

#### Scenario: Deny unauthorized download

- **WHEN** a non-member, removed member, or unauthenticated user requests an original media download
- **THEN** the system denies the request and does not reveal a usable file URL
