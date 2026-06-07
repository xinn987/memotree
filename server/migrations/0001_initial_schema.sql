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

ALTER TABLE family_invites
  ADD COLUMN token_plaintext VARCHAR(128) NULL AFTER token_hash;
