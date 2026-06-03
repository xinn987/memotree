-- MVP 初始数据模型，先覆盖账号、家庭、邀请和媒体资产的核心关系。
CREATE TABLE users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  display_name VARCHAR(120) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE login_identities (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  identity_type VARCHAR(32) NOT NULL,
  identity_value VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_identity (identity_type, identity_value),
  CONSTRAINT fk_login_identities_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE families (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  display_name VARCHAR(120) NOT NULL,
  created_by BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_families_creator FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE family_members (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  removed_at TIMESTAMP NULL,
  UNIQUE KEY uniq_family_member (family_id, user_id),
  KEY idx_family_members_user_status (user_id, status),
  CONSTRAINT fk_family_members_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_family_members_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE invitations (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  token_hash VARBINARY(64) NOT NULL,
  created_by BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  used_by BIGINT NULL,
  used_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uniq_invitation_token_hash (token_hash),
  KEY idx_invitations_family_status (family_id, status),
  CONSTRAINT fk_invitations_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_invitations_creator FOREIGN KEY (created_by) REFERENCES users(id),
  CONSTRAINT fk_invitations_used_by FOREIGN KEY (used_by) REFERENCES users(id)
);

CREATE TABLE media_assets (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  uploaded_by BIGINT NOT NULL,
  media_type VARCHAR(32) NOT NULL,
  original_object_key VARCHAR(512) NOT NULL,
  preview_object_key VARCHAR(512) NULL,
  original_filename VARCHAR(255) NOT NULL,
  content_type VARCHAR(120) NOT NULL,
  byte_size BIGINT NOT NULL,
  captured_at TIMESTAMP NULL,
  uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  upload_status VARCHAR(32) NOT NULL,
  preview_status VARCHAR(32) NOT NULL,
  KEY idx_media_timeline (family_id, captured_at, uploaded_at, id),
  KEY idx_media_uploader (uploaded_by),
  CONSTRAINT fk_media_assets_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_media_assets_uploader FOREIGN KEY (uploaded_by) REFERENCES users(id)
);
