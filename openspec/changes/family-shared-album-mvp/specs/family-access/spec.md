## ADDED Requirements

### Requirement: Family Space

The system SHALL allow a user to create a private family space for sharing baby photos and videos.

#### Scenario: Create family space

- **WHEN** an authenticated user creates a family with a display name
- **THEN** the system creates a private family space and adds the creator as an administrator member

#### Scenario: Access only joined families

- **WHEN** a user requests family data
- **THEN** the system returns data only for families where the user is an active member

### Requirement: Family Invitation

The system SHALL allow an administrator member to invite other family members through an invitation code or invitation link.

#### Scenario: Join with valid invitation

- **WHEN** an authenticated user submits a valid unused or active invitation for a family
- **THEN** the system adds the user as a member of that family

#### Scenario: Reject invalid invitation

- **WHEN** a user submits an expired, revoked, or unknown invitation
- **THEN** the system rejects the join request and does not create a family membership

### Requirement: Basic Member Roles

The system SHALL support a minimal role model for MVP access control.

#### Scenario: Member can use core album features

- **WHEN** an active family member uploads, views, or downloads media in the family
- **THEN** the system permits the action after verifying membership

#### Scenario: Only administrator can manage invitations

- **WHEN** a non-administrator member attempts to create or revoke an invitation
- **THEN** the system denies the action

### Requirement: Membership Removal

The system SHALL prevent removed members from accessing family media after removal.

#### Scenario: Removed member loses access

- **WHEN** a removed member requests timeline data, media detail, upload authorization, or download authorization
- **THEN** the system denies the request
