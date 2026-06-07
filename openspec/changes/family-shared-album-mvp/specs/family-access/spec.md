## ADDED Requirements

### Requirement: Account Authentication

The system SHALL provide persistent account authentication for MVP users.

#### Scenario: Register with invitation

- **WHEN** a person opens a valid family invitation and submits a display name, login name, and password
- **THEN** the system creates a persistent user account, login credential, authenticated session, and family membership for that family

#### Scenario: Login with account password

- **WHEN** an existing user submits a valid login name and password
- **THEN** the system creates an authenticated session for that user

#### Scenario: Use session for subsequent requests

- **WHEN** a user with a valid authenticated session requests family data
- **THEN** the system resolves the session to the persistent user account before checking family membership

### Requirement: Family Space

The system SHALL allow a user to create a private family space for sharing selected baby and family photos and videos.

#### Scenario: Create family space

- **WHEN** an authenticated user creates a family with a display name
- **THEN** the system creates a private family space and adds the creator as an administrator member

#### Scenario: Access only joined families

- **WHEN** a user requests family data
- **THEN** the system returns data only for families where the user is an active member

#### Scenario: User can belong to multiple families

- **WHEN** a user is an active member of more than one family
- **THEN** the system preserves separate family memberships and permissions for each family

#### Scenario: Family does not require child assignment

- **WHEN** a member uploads media to a family
- **THEN** the system associates the media with the family without requiring a child, baby, album, or folder selection

### Requirement: Family Invitation

The system SHALL allow an administrator member to invite other family members through an invitation code or invitation link.

#### Scenario: Join with valid invitation

- **WHEN** an authenticated user submits a valid unused or active invitation for a family
- **THEN** the system adds the user as a member of that family

#### Scenario: Administrator lists invitations

- **WHEN** an administrator member opens invitation management for a family
- **THEN** the system returns invitations for that family including status, intended member display name, expiration time, and a reusable token only when the invitation is still pending

#### Scenario: Administrator revokes pending invitation

- **WHEN** an administrator member revokes a pending invitation
- **THEN** the system marks the invitation as revoked and prevents that invitation token from joining the family

#### Scenario: Reject invalid invitation

- **WHEN** a user submits an expired, revoked, or unknown invitation
- **THEN** the system rejects the join request and does not create a family membership

### Requirement: Basic Member Roles

The system SHALL support a minimal role model for MVP access control.

#### Scenario: Member can use core album features

- **WHEN** an active family member uploads or views media in the family
- **THEN** the system permits the action after verifying membership

#### Scenario: Only administrator can manage invitations

- **WHEN** a non-administrator member attempts to create or revoke an invitation
- **THEN** the system denies the action

#### Scenario: Active administrator can manage family members

- **WHEN** an active administrator updates member display names or removes a member
- **THEN** the system permits the action after verifying administrator membership

#### Scenario: Prevent family without administrator

- **WHEN** an administrator removal or role change would leave the family without any active administrator
- **THEN** the system rejects the change

### Requirement: Membership Removal

The system SHALL prevent removed members from accessing family media after removal.

#### Scenario: Removed member loses access

- **WHEN** a removed member requests timeline data, media detail, or upload authorization
- **THEN** the system denies the request

#### Scenario: Removed member content remains in family

- **WHEN** a member is removed from a family
- **THEN** the system keeps previously uploaded family media available to active family members
