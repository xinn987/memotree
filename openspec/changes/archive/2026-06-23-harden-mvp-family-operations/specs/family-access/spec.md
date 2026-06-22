## MODIFIED Requirements

### Requirement: Basic Member Roles

The system SHALL support a minimal role model for MVP access control.

#### Scenario: Member can use core album features

- **WHEN** an active family member uploads or views media in the family
- **THEN** the system permits the action after verifying membership

#### Scenario: Only administrator can manage invitations

- **WHEN** a non-administrator member attempts to create or revoke an invitation
- **THEN** the system denies the action

#### Scenario: Active administrator can list family members

- **WHEN** an active administrator requests the member list for a family
- **THEN** the system returns active and removed members for that family with display name, role, status, and joined time

#### Scenario: Active administrator can update member display names

- **WHEN** an active administrator updates a family member display name
- **THEN** the system stores the new family-specific display name without changing the global user account display name

#### Scenario: Active administrator can remove family members

- **WHEN** an active administrator removes a member from the family
- **THEN** the system marks that family membership as removed

#### Scenario: Administrator cannot remove themself

- **WHEN** an active administrator attempts to remove their own family membership through member management
- **THEN** the system rejects the action

#### Scenario: Non-administrator cannot manage members

- **WHEN** a non-administrator member attempts to list, update, or remove family members
- **THEN** the system denies the action

#### Scenario: Prevent family without administrator

- **WHEN** an administrator removal or role change would leave the family without any active administrator
- **THEN** the system rejects the change

### Requirement: Membership Removal

The system SHALL prevent removed members from accessing family media and family operations after removal.

#### Scenario: Removed member loses access to browsing

- **WHEN** a removed member requests timeline data or media detail
- **THEN** the system denies the request

#### Scenario: Removed member loses access to uploading

- **WHEN** a removed member requests upload authorization or upload task data
- **THEN** the system denies the request

#### Scenario: Removed member loses access to family management

- **WHEN** a removed member requests invitation or member management data
- **THEN** the system denies the request

#### Scenario: Removed member content remains in family

- **WHEN** a member is removed from a family
- **THEN** the system keeps previously uploaded family media available to active family members
