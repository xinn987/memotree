-- MemoTree MVP 初始账户与家庭模型。
-- 当前项目尚无线上数据，因此第一版迁移直接定义目标结构。

CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  display_name VARCHAR(120) NOT NULL,
  is_system_admin BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
);

CREATE TABLE IF NOT EXISTS user_credentials (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  login_name VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY uniq_user_credentials_login_name (login_name),
  CONSTRAINT fk_user_credentials_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS user_sessions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  token_hash CHAR(43) NOT NULL,
  expires_at DATETIME(6) NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY uniq_user_sessions_token_hash (token_hash),
  KEY idx_user_sessions_user_expires (user_id, expires_at),
  CONSTRAINT fk_user_sessions_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS families (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  display_name VARCHAR(120) NOT NULL,
  timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai',
  created_by BIGINT NOT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  CONSTRAINT fk_families_creator FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS family_members (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  display_name VARCHAR(120) NOT NULL,
  role VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  joined_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  removed_at DATETIME(6) NULL,
  UNIQUE KEY uniq_family_member (family_id, user_id),
  KEY idx_family_members_user_status (user_id, status),
  KEY idx_family_members_family_status (family_id, status),
  CONSTRAINT fk_family_members_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_family_members_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS family_invites (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  token_hash CHAR(43) NOT NULL,
  token_plaintext VARCHAR(128) NULL,
  created_by BIGINT NOT NULL,
  member_display_name VARCHAR(120) NULL,
  status VARCHAR(32) NOT NULL,
  expires_at DATETIME(6) NOT NULL,
  used_by BIGINT NULL,
  used_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY uniq_family_invites_token_hash (token_hash),
  KEY idx_family_invites_family_status (family_id, status),
  CONSTRAINT fk_family_invites_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_family_invites_creator FOREIGN KEY (created_by) REFERENCES users(id),
  CONSTRAINT fk_family_invites_used_by FOREIGN KEY (used_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS upload_batches (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  created_by BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL,
  active_slot TINYINT NULL DEFAULT 1,
  total_count INT NOT NULL DEFAULT 0,
  ready_count INT NOT NULL DEFAULT 0,
  failed_count INT NOT NULL DEFAULT 0,
  cancelled_count INT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  completed_at DATETIME(6) NULL,
  stopped_at DATETIME(6) NULL,
  UNIQUE KEY uniq_active_upload_batch (family_id, created_by, active_slot),
  KEY idx_upload_batches_family_created (family_id, created_at),
  KEY idx_upload_batches_creator_created (created_by, created_at),
  CONSTRAINT fk_upload_batches_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_upload_batches_creator FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS media_assets (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  family_id BIGINT NOT NULL,
  uploaded_by BIGINT NOT NULL,
  media_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL,
  rendition_status VARCHAR(32) NOT NULL,
  captured_at DATETIME(6) NULL,
  uploaded_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  deleted_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  KEY idx_media_timeline (family_id, status, rendition_status, captured_at, uploaded_at, id),
  KEY idx_media_uploaded_at (family_id, uploaded_at, id),
  KEY idx_media_uploader (uploaded_by),
  CONSTRAINT fk_media_assets_family FOREIGN KEY (family_id) REFERENCES families(id),
  CONSTRAINT fk_media_assets_uploader FOREIGN KEY (uploaded_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS media_originals (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  media_asset_id BIGINT NOT NULL,
  original_type VARCHAR(32) NOT NULL,
  object_key VARCHAR(512) NOT NULL,
  original_filename VARCHAR(255) NOT NULL,
  content_type VARCHAR(120) NOT NULL,
  byte_size BIGINT NOT NULL,
  checksum_sha256 CHAR(64) NULL,
  width INT NULL,
  height INT NULL,
  duration_millis BIGINT NULL,
  captured_at DATETIME(6) NULL,
  uploaded_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  UNIQUE KEY uniq_media_original_object_key (object_key),
  KEY idx_media_originals_asset (media_asset_id),
  CONSTRAINT fk_media_originals_asset FOREIGN KEY (media_asset_id) REFERENCES media_assets(id)
);

CREATE TABLE IF NOT EXISTS media_renditions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  media_asset_id BIGINT NOT NULL,
  rendition_type VARCHAR(32) NOT NULL,
  object_key VARCHAR(512) NOT NULL,
  content_type VARCHAR(120) NOT NULL,
  byte_size BIGINT NOT NULL,
  width INT NULL,
  height INT NULL,
  duration_millis BIGINT NULL,
  status VARCHAR(32) NOT NULL,
  error_message TEXT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  UNIQUE KEY uniq_media_rendition_type (media_asset_id, rendition_type),
  UNIQUE KEY uniq_media_rendition_object_key (object_key),
  KEY idx_media_renditions_asset_status (media_asset_id, status),
  CONSTRAINT fk_media_renditions_asset FOREIGN KEY (media_asset_id) REFERENCES media_assets(id)
);

CREATE TABLE IF NOT EXISTS upload_items (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  upload_batch_id BIGINT NOT NULL,
  media_asset_id BIGINT NULL,
  original_type VARCHAR(32) NOT NULL,
  original_filename VARCHAR(255) NOT NULL,
  content_type VARCHAR(120) NOT NULL,
  byte_size BIGINT NOT NULL,
  object_key VARCHAR(512) NOT NULL,
  status VARCHAR(32) NOT NULL,
  error_message TEXT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  completed_at DATETIME(6) NULL,
  UNIQUE KEY uniq_upload_item_object_key (object_key),
  KEY idx_upload_items_batch_status (upload_batch_id, status),
  KEY idx_upload_items_asset (media_asset_id),
  CONSTRAINT fk_upload_items_batch FOREIGN KEY (upload_batch_id) REFERENCES upload_batches(id),
  CONSTRAINT fk_upload_items_asset FOREIGN KEY (media_asset_id) REFERENCES media_assets(id)
);
