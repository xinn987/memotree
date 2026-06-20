// MySQL store 实现。
//
// 这个文件只负责把 Store 接口翻译成 SQL 操作，不处理 HTTP 入参和响应。
// 需要保持业务语义和 MemoryStore 一致，避免测试和真实运行出现两套规则。
package store

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// MySQLStore 是 Store 的 MySQL 持久化实现。
type MySQLStore struct {
	db *sql.DB
}

// NewMySQLStore 使用已有 sql.DB 创建 MySQL store，便于测试或 main 统一管理连接生命周期。
func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

// OpenMySQL 打开并验证 MySQL 连接。
// 调用方负责在进程退出时关闭返回的 *sql.DB。
func OpenMySQL(ctx context.Context, dsn string) (*sql.DB, *MySQLStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	return db, NewMySQLStore(db), nil
}

// ApplySchema 应用本地开发用 schema。
// 当前迁移文件使用 CREATE TABLE IF NOT EXISTS，因此 API 启动时重复执行是安全的。
func ApplySchema(ctx context.Context, db *sql.DB, schemaSQL string) error {
	cleaned := stripSQLLineComments(schemaSQL)
	for _, statement := range strings.Split(cleaned, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, statement); err != nil {
			if isDuplicateColumn(err) {
				continue
			}
			return err
		}
	}
	return nil
}

// isDuplicateColumn 兼容本地开发 schema 的重复执行。
// 新库会先通过 CREATE TABLE 建出列，旧库则需要后面的 ALTER TABLE 补列；重复列错误可以安全跳过。
func isDuplicateColumn(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1060
}

// stripSQLLineComments 移除行注释，避免简单按分号切分 SQL 时把注释带入执行语句。
func stripSQLLineComments(schemaSQL string) string {
	var builder strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(schemaSQL))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "--") {
			continue
		}
		builder.WriteString(scanner.Text())
		builder.WriteByte('\n')
	}
	return builder.String()
}

// CreateUser 创建用户和登录凭证。
// 事务内先统计用户数，确保第一位注册用户被标记为系统初始管理员。
func (s *MySQLStore) CreateUser(ctx context.Context, loginName string, passwordHash string, displayName string) (User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer rollbackUnlessCommitted(tx)

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return User{}, err
	}
	isSystemAdmin := count == 0

	result, err := tx.ExecContext(ctx, `INSERT INTO users (display_name, is_system_admin) VALUES (?, ?)`, displayName, isSystemAdmin)
	if err != nil {
		return User{}, err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_credentials (user_id, login_name, password_hash) VALUES (?, ?, ?)`, userID, loginName, passwordHash); err != nil {
		if isDuplicateKey(err) {
			return User{}, ErrAlreadyExists
		}
		return User{}, err
	}
	if err := tx.Commit(); err != nil {
		return User{}, err
	}

	return User{ID: userID, LoginName: loginName, DisplayName: displayName, IsSystemAdmin: isSystemAdmin}, nil
}

// FindUserByLoginName 按登录名读取用户和密码哈希，用于登录校验。
func (s *MySQLStore) FindUserByLoginName(ctx context.Context, loginName string) (User, string, error) {
	var found User
	var passwordHash string
	err := s.db.QueryRowContext(ctx, `
SELECT u.id, c.login_name, u.display_name, u.is_system_admin, c.password_hash
FROM user_credentials c
JOIN users u ON u.id = c.user_id
WHERE c.login_name = ?
`, loginName).Scan(&found.ID, &found.LoginName, &found.DisplayName, &found.IsSystemAdmin, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, "", ErrNotFound
	}
	if err != nil {
		return User{}, "", err
	}
	return found, passwordHash, nil
}

func (s *MySQLStore) FindUserByID(ctx context.Context, id int64) (User, error) {
	var found User
	err := s.db.QueryRowContext(ctx, `
SELECT u.id, c.login_name, u.display_name, u.is_system_admin
FROM users u
JOIN user_credentials c ON c.user_id = u.id
WHERE u.id = ?
`, id).Scan(&found.ID, &found.LoginName, &found.DisplayName, &found.IsSystemAdmin)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return found, nil
}

// CreateSession 保存会话 token hash；token 原文只存在于浏览器 cookie 中。
func (s *MySQLStore) CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO user_sessions (user_id, token_hash, expires_at) VALUES (?, ?, ?)`, userID, tokenHash, expiresAt.UTC())
	return err
}

// FindSession 只返回未过期 session，过期 session 视为不存在。
func (s *MySQLStore) FindSession(ctx context.Context, tokenHash string, now time.Time) (Session, bool, error) {
	var found Session
	err := s.db.QueryRowContext(ctx, `
SELECT user_id, token_hash, expires_at
FROM user_sessions
WHERE token_hash = ? AND expires_at > ?
`, tokenHash, now.UTC()).Scan(&found.UserID, &found.TokenHash, &found.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}
	return found, true, nil
}

func (s *MySQLStore) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM user_sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// CreateFamily 创建家庭，并在同一事务内把创建者写成 active admin。
// 这样不会出现“家庭已创建但没有管理员”的半完成状态。
func (s *MySQLStore) CreateFamily(ctx context.Context, displayName string, timezone string, creatorID int64, creatorDisplayName string) (FamilySummary, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return FamilySummary{}, err
	}
	defer rollbackUnlessCommitted(tx)

	result, err := tx.ExecContext(ctx, `INSERT INTO families (display_name, timezone, created_by) VALUES (?, ?, ?)`, displayName, timezone, creatorID)
	if err != nil {
		return FamilySummary{}, err
	}
	familyID, err := result.LastInsertId()
	if err != nil {
		return FamilySummary{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO family_members (family_id, user_id, display_name, role, status)
VALUES (?, ?, ?, ?, ?)
`, familyID, creatorID, creatorDisplayName, MemberRoleAdmin, MemberStatusActive); err != nil {
		return FamilySummary{}, err
	}
	if err := tx.Commit(); err != nil {
		return FamilySummary{}, err
	}
	return FamilySummary{ID: familyID, DisplayName: displayName, Timezone: timezone, Role: MemberRoleAdmin, MemberDisplayName: creatorDisplayName}, nil
}

func (s *MySQLStore) ListFamiliesForUser(ctx context.Context, userID int64) ([]FamilySummary, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT f.id, f.display_name, f.timezone, m.role, m.display_name
FROM family_members m
JOIN families f ON f.id = m.family_id
WHERE m.user_id = ? AND m.status = ?
ORDER BY f.id ASC
`, userID, MemberStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 明确初始化为空切片，保证没有家庭时 JSON 编码结果是 [] 而不是 null。
	result := []FamilySummary{}
	for rows.Next() {
		var item FamilySummary
		if err := rows.Scan(&item.ID, &item.DisplayName, &item.Timezone, &item.Role, &item.MemberDisplayName); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (s *MySQLStore) IsActiveAdmin(ctx context.Context, familyID int64, userID int64) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM family_members
WHERE family_id = ? AND user_id = ? AND status = ? AND role = ?
`, familyID, userID, MemberStatusActive, MemberRoleAdmin).Scan(&count)
	return count > 0, err
}

func (s *MySQLStore) IsActiveMember(ctx context.Context, familyID int64, userID int64) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM family_members
WHERE family_id = ? AND user_id = ? AND status = ?
`, familyID, userID, MemberStatusActive).Scan(&count)
	return count > 0, err
}

// CreateInvite 保存邀请 token hash、token 原文和预填成员显示名。
// MVP 阶段保存 token 原文，方便管理员在邀请管理里重新复制邀请链接。
func (s *MySQLStore) CreateInvite(ctx context.Context, familyID int64, tokenHash string, tokenPlaintext string, createdBy int64, memberDisplayName string, expiresAt time.Time) (FamilyInvite, error) {
	result, err := s.db.ExecContext(ctx, `
INSERT INTO family_invites (family_id, token_hash, token_plaintext, created_by, member_display_name, status, expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
`, familyID, tokenHash, tokenPlaintext, createdBy, nullableString(memberDisplayName), InviteStatusPending, expiresAt.UTC())
	if err != nil {
		return FamilyInvite{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return FamilyInvite{}, err
	}
	return FamilyInvite{
		ID:                id,
		FamilyID:          familyID,
		TokenHash:         tokenHash,
		TokenPlaintext:    tokenPlaintext,
		CreatedBy:         createdBy,
		MemberDisplayName: memberDisplayName,
		Status:            InviteStatusPending,
		ExpiresAt:         expiresAt,
	}, nil
}

func (s *MySQLStore) ListInvitesForFamily(ctx context.Context, familyID int64) ([]FamilyInvite, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, family_id, token_hash, token_plaintext, created_by, member_display_name, status, expires_at, used_by, used_at
FROM family_invites
WHERE family_id = ?
ORDER BY id DESC
`, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []FamilyInvite{}
	for rows.Next() {
		var item FamilyInvite
		var tokenPlaintext sql.NullString
		var memberDisplayName sql.NullString
		var usedBy sql.NullInt64
		var usedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.FamilyID, &item.TokenHash, &tokenPlaintext, &item.CreatedBy, &memberDisplayName, &item.Status, &item.ExpiresAt, &usedBy, &usedAt); err != nil {
			return nil, err
		}
		if tokenPlaintext.Valid {
			item.TokenPlaintext = tokenPlaintext.String
		}
		if memberDisplayName.Valid {
			item.MemberDisplayName = memberDisplayName.String
		}
		if usedBy.Valid {
			item.UsedBy = usedBy.Int64
		}
		if usedAt.Valid {
			item.UsedAt = usedAt.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// RevokeInvite 在事务中撤销一条未使用邀请。
// 撤销时清空 token 原文；token_hash 保留用于审计和避免结构性删除。
func (s *MySQLStore) RevokeInvite(ctx context.Context, familyID int64, inviteID int64, now time.Time) (FamilyInvite, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return FamilyInvite{}, err
	}
	defer rollbackUnlessCommitted(tx)

	var invite FamilyInvite
	var tokenPlaintext sql.NullString
	var memberDisplayName sql.NullString
	var usedBy sql.NullInt64
	var usedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
SELECT id, family_id, token_hash, token_plaintext, created_by, member_display_name, status, expires_at, used_by, used_at
FROM family_invites
WHERE id = ? AND family_id = ?
FOR UPDATE
`, inviteID, familyID).Scan(&invite.ID, &invite.FamilyID, &invite.TokenHash, &tokenPlaintext, &invite.CreatedBy, &memberDisplayName, &invite.Status, &invite.ExpiresAt, &usedBy, &usedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return FamilyInvite{}, ErrNotFound
	}
	if err != nil {
		return FamilyInvite{}, err
	}
	if tokenPlaintext.Valid {
		invite.TokenPlaintext = tokenPlaintext.String
	}
	if memberDisplayName.Valid {
		invite.MemberDisplayName = memberDisplayName.String
	}
	if usedBy.Valid {
		invite.UsedBy = usedBy.Int64
	}
	if usedAt.Valid {
		invite.UsedAt = usedAt.Time
	}
	if invite.Status != InviteStatusPending || !invite.ExpiresAt.After(now.UTC()) {
		return FamilyInvite{}, ErrInvalidInvite
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE family_invites
SET status = ?, token_plaintext = NULL
WHERE id = ?
`, InviteStatusRevoked, invite.ID); err != nil {
		return FamilyInvite{}, err
	}
	if err := tx.Commit(); err != nil {
		return FamilyInvite{}, err
	}
	invite.Status = InviteStatusRevoked
	invite.TokenPlaintext = ""
	return invite, nil
}

// JoinInvite 在事务内完成“锁定邀请、创建/恢复成员、消耗邀请”。
// 邀请行和成员行都使用 FOR UPDATE，避免并发点击邀请链接造成重复消费或重复成员。
func (s *MySQLStore) JoinInvite(ctx context.Context, tokenHash string, userID int64, fallbackDisplayName string, now time.Time) (FamilyMember, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return FamilyMember{}, err
	}
	defer rollbackUnlessCommitted(tx)

	var invite FamilyInvite
	var memberDisplayName sql.NullString
	err = tx.QueryRowContext(ctx, `
SELECT id, family_id, token_hash, created_by, member_display_name, status, expires_at
FROM family_invites
WHERE token_hash = ?
FOR UPDATE
`, tokenHash).Scan(&invite.ID, &invite.FamilyID, &invite.TokenHash, &invite.CreatedBy, &memberDisplayName, &invite.Status, &invite.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return FamilyMember{}, ErrNotFound
	}
	if err != nil {
		return FamilyMember{}, err
	}
	if memberDisplayName.Valid {
		invite.MemberDisplayName = memberDisplayName.String
	}
	if invite.Status != InviteStatusPending || !invite.ExpiresAt.After(now.UTC()) {
		return FamilyMember{}, ErrInvalidInvite
	}

	member, exists, err := findMemberInTx(ctx, tx, invite.FamilyID, userID)
	if err != nil {
		return FamilyMember{}, err
	}
	if exists && member.Status == MemberStatusActive {
		// 已经是活跃成员时不消耗邀请，方便用户重复打开同一个邀请链接。
		return member, tx.Commit()
	}

	displayName := fallbackDisplayName
	if invite.MemberDisplayName != "" {
		displayName = invite.MemberDisplayName
	}

	if exists {
		_, err = tx.ExecContext(ctx, `
UPDATE family_members
SET display_name = ?, role = ?, status = ?, removed_at = NULL
WHERE id = ?
`, displayName, MemberRoleMember, MemberStatusActive, member.ID)
		if err != nil {
			return FamilyMember{}, err
		}
		member.DisplayName = displayName
		member.Role = MemberRoleMember
		member.Status = MemberStatusActive
	} else {
		result, err := tx.ExecContext(ctx, `
INSERT INTO family_members (family_id, user_id, display_name, role, status)
VALUES (?, ?, ?, ?, ?)
`, invite.FamilyID, userID, displayName, MemberRoleMember, MemberStatusActive)
		if err != nil {
			return FamilyMember{}, err
		}
		memberID, err := result.LastInsertId()
		if err != nil {
			return FamilyMember{}, err
		}
		member = FamilyMember{ID: memberID, FamilyID: invite.FamilyID, UserID: userID, DisplayName: displayName, Role: MemberRoleMember, Status: MemberStatusActive}
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE family_invites
SET status = ?, token_plaintext = NULL, used_by = ?, used_at = ?
WHERE id = ?
`, InviteStatusUsed, userID, now.UTC(), invite.ID); err != nil {
		return FamilyMember{}, err
	}
	if err := tx.Commit(); err != nil {
		return FamilyMember{}, err
	}
	return member, nil
}

// FindActiveUploadBatch 读取同一用户同一家庭中的 active 上传任务。
// API 用它支持“再次上传时回到当前任务”的产品约束。
func (s *MySQLStore) FindActiveUploadBatch(ctx context.Context, familyID int64, userID int64) (UploadBatch, bool, error) {
	batch, err := scanUploadBatch(s.db.QueryRowContext(ctx, `
SELECT id, family_id, created_by, status, active_slot, total_count, ready_count, failed_count, cancelled_count, created_at, completed_at, stopped_at
FROM upload_batches
WHERE family_id = ? AND created_by = ? AND active_slot = 1
`, familyID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return UploadBatch{}, false, nil
	}
	if err != nil {
		return UploadBatch{}, false, err
	}
	return batch, true, nil
}

// ListUploadBatches 按创建时间倒序返回最近上传任务。
// IncludeFamily=false 时只返回当前用户创建的任务，用于普通成员的权限边界。
func (s *MySQLStore) ListUploadBatches(ctx context.Context, input ListUploadBatchesInput) ([]UploadBatch, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, family_id, created_by, status, active_slot, total_count, ready_count, failed_count, cancelled_count, created_at, completed_at, stopped_at
FROM upload_batches
WHERE family_id = ? AND (? OR created_by = ?)
ORDER BY created_at DESC, id DESC
LIMIT ?
`, input.FamilyID, input.IncludeFamily, input.ActorUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []UploadBatch{}
	for rows.Next() {
		batch, err := scanUploadBatch(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, batch)
	}
	return result, rows.Err()
}

// ListTimelineMedia 返回已经处理完成、可以在家庭时间线展示的媒体。
// SQL 层只取 display 派生资源为 ready 的资产，避免主时间线混入仍在 worker 处理中的原文件。
func (s *MySQLStore) ListTimelineMedia(ctx context.Context, input ListTimelineMediaInput) ([]TimelineMedia, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 60
	}
	cursorEnabled := 0
	if !input.AfterTime.IsZero() {
		cursorEnabled = 1
	}
	monthFilterEnabled := 0
	if !input.MonthFrom.IsZero() {
		monthFilterEnabled = 1
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT
  ma.id, ma.family_id, ma.uploaded_by, ma.media_type, ma.status, ma.rendition_status, ma.captured_at, ma.uploaded_at, ma.deleted_at,
  COALESCE(NULLIF(fm.display_name, ''), u.display_name) AS uploaded_by_display_name,
  d.id, d.media_asset_id, d.rendition_type, d.object_key, d.content_type, d.byte_size, d.width, d.height, d.duration_millis, d.status, d.error_message,
  t.id, t.media_asset_id, t.rendition_type, t.object_key, t.content_type, t.byte_size, t.width, t.height, t.duration_millis, t.status, t.error_message
FROM media_assets ma
JOIN users u ON u.id = ma.uploaded_by
LEFT JOIN family_members fm ON fm.family_id = ma.family_id AND fm.user_id = ma.uploaded_by
JOIN media_renditions d ON d.media_asset_id = ma.id
  AND d.status = ?
  AND d.rendition_type = CASE WHEN ma.media_type = ? THEN ? ELSE ? END
LEFT JOIN media_renditions t ON t.media_asset_id = ma.id
  AND t.status = ?
  AND t.rendition_type = ?
WHERE ma.family_id = ?
  AND ma.status = ?
  AND ma.rendition_status = ?
  AND ma.deleted_at IS NULL
  AND (? = '' OR ma.media_type = ?)
  AND (? = 0 OR (COALESCE(ma.captured_at, ma.uploaded_at) >= ? AND COALESCE(ma.captured_at, ma.uploaded_at) < ?))
  AND (? = 0 OR COALESCE(ma.captured_at, ma.uploaded_at) < ? OR (COALESCE(ma.captured_at, ma.uploaded_at) = ? AND ma.id < ?))
ORDER BY COALESCE(ma.captured_at, ma.uploaded_at) DESC, ma.id DESC
LIMIT ?
`,
		RenditionStatusReady,
		MediaTypeVideo,
		RenditionTypeDisplayVideo,
		RenditionTypeDisplayImage,
		RenditionStatusReady,
		RenditionTypeThumbnail,
		input.FamilyID,
		MediaStatusActive,
		RenditionStatusReady,
		input.MediaType,
		input.MediaType,
		monthFilterEnabled,
		input.MonthFrom.UTC(),
		input.MonthTo.UTC(),
		cursorEnabled,
		input.AfterTime.UTC(),
		input.AfterTime.UTC(),
		input.AfterID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []TimelineMedia{}
	for rows.Next() {
		item, err := scanTimelineMedia(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// FindMediaDetail 返回单个已可展示媒体的详情数据。
// SQL 层过滤 deleted、processing 和 failed 媒体，避免 HTTP 层误暴露不可见资源。
func (s *MySQLStore) FindMediaDetail(ctx context.Context, input FindMediaDetailInput) (MediaDetail, bool, error) {
	item, err := scanTimelineMedia(s.db.QueryRowContext(ctx, `
SELECT
  ma.id, ma.family_id, ma.uploaded_by, ma.media_type, ma.status, ma.rendition_status, ma.captured_at, ma.uploaded_at, ma.deleted_at,
  COALESCE(NULLIF(fm.display_name, ''), u.display_name) AS uploaded_by_display_name,
  d.id, d.media_asset_id, d.rendition_type, d.object_key, d.content_type, d.byte_size, d.width, d.height, d.duration_millis, d.status, d.error_message,
  t.id, t.media_asset_id, t.rendition_type, t.object_key, t.content_type, t.byte_size, t.width, t.height, t.duration_millis, t.status, t.error_message
FROM media_assets ma
JOIN users u ON u.id = ma.uploaded_by
LEFT JOIN family_members fm ON fm.family_id = ma.family_id AND fm.user_id = ma.uploaded_by
JOIN media_renditions d ON d.media_asset_id = ma.id
  AND d.status = ?
  AND d.rendition_type = CASE WHEN ma.media_type = ? THEN ? ELSE ? END
LEFT JOIN media_renditions t ON t.media_asset_id = ma.id
  AND t.status = ?
  AND t.rendition_type = ?
WHERE ma.id = ?
  AND ma.family_id = ?
  AND ma.status = ?
  AND ma.rendition_status = ?
  AND ma.deleted_at IS NULL
`,
		RenditionStatusReady,
		MediaTypeVideo,
		RenditionTypeDisplayVideo,
		RenditionTypeDisplayImage,
		RenditionStatusReady,
		RenditionTypeThumbnail,
		input.MediaID,
		input.FamilyID,
		MediaStatusActive,
		RenditionStatusReady,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return MediaDetail{}, false, nil
	}
	if err != nil {
		return MediaDetail{}, false, err
	}
	return MediaDetail{
		Asset:                 item.Asset,
		UploadedByDisplayName: item.UploadedByDisplayName,
		Thumbnail:             item.Thumbnail,
		Display:               item.Display,
	}, true, nil
}

// FindUploadBatch 按家庭和任务 ID 读取上传任务。
// familyID 必须参与查询，避免跨家庭 ID 枚举绕过权限边界。
func (s *MySQLStore) FindUploadBatch(ctx context.Context, familyID int64, batchID int64) (UploadBatch, bool, error) {
	batch, err := scanUploadBatch(s.db.QueryRowContext(ctx, `
SELECT id, family_id, created_by, status, active_slot, total_count, ready_count, failed_count, cancelled_count, created_at, completed_at, stopped_at
FROM upload_batches
WHERE family_id = ? AND id = ?
`, familyID, batchID))
	if errors.Is(err, sql.ErrNoRows) {
		return UploadBatch{}, false, nil
	}
	if err != nil {
		return UploadBatch{}, false, err
	}
	return batch, true, nil
}

// ListUploadItems 返回上传任务中的文件条目。
// object_key 只用于后端存储访问，HTTP 响应层不会把它暴露给前端。
func (s *MySQLStore) ListUploadItems(ctx context.Context, batchID int64) ([]UploadItem, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, upload_batch_id, media_asset_id, original_type, original_filename, content_type, byte_size, object_key, status, error_message, created_at, updated_at, completed_at
FROM upload_items
WHERE upload_batch_id = ?
ORDER BY id ASC
`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []UploadItem{}
	for rows.Next() {
		var item UploadItem
		var mediaAssetID sql.NullInt64
		var errorMessage sql.NullString
		var completedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.UploadBatchID,
			&mediaAssetID,
			&item.OriginalType,
			&item.OriginalFilename,
			&item.ContentType,
			&item.ByteSize,
			&item.ObjectKey,
			&item.Status,
			&errorMessage,
			&item.CreatedAt,
			&item.UpdatedAt,
			&completedAt,
		); err != nil {
			return nil, err
		}
		if mediaAssetID.Valid {
			item.MediaAssetID = mediaAssetID.Int64
		}
		if errorMessage.Valid {
			item.ErrorMessage = errorMessage.String
		}
		if completedAt.Valid {
			item.CompletedAt = completedAt.Time
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// StopUploadBatch 停止上传任务，并取消尚未完成的文件项。
// 事务保证任务状态、active slot 和条目状态同时更新。
func (s *MySQLStore) StopUploadBatch(ctx context.Context, batchID int64, now time.Time) (UploadBatch, []UploadItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return UploadBatch{}, nil, err
	}
	defer rollbackUnlessCommitted(tx)

	if now.IsZero() {
		now = time.Now()
	}
	var existingID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM upload_batches WHERE id = ? FOR UPDATE`, batchID).Scan(&existingID); errors.Is(err, sql.ErrNoRows) {
		return UploadBatch{}, nil, ErrNotFound
	} else if err != nil {
		return UploadBatch{}, nil, err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE upload_items
SET status = ?, updated_at = ?, completed_at = ?
WHERE upload_batch_id = ? AND status IN (?, ?, ?, ?)
`,
		UploadItemStatusCancelled,
		now.UTC(),
		now.UTC(),
		batchID,
		UploadItemStatusWaiting,
		UploadItemStatusUploading,
		UploadItemStatusUploaded,
		UploadItemStatusProcessing,
	); err != nil {
		return UploadBatch{}, nil, err
	}

	var readyCount int
	var failedCount int
	var cancelledCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status = ?`, batchID, UploadItemStatusReady).Scan(&readyCount); err != nil {
		return UploadBatch{}, nil, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status IN (?, ?)`, batchID, UploadItemStatusUploadFailed, UploadItemStatusProcessingFailed).Scan(&failedCount); err != nil {
		return UploadBatch{}, nil, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status = ?`, batchID, UploadItemStatusCancelled).Scan(&cancelledCount); err != nil {
		return UploadBatch{}, nil, err
	}

	if _, err := tx.ExecContext(ctx, `
UPDATE upload_batches
SET status = ?, active_slot = NULL, ready_count = ?, failed_count = ?, cancelled_count = ?, stopped_at = ?
WHERE id = ?
`, UploadBatchStatusStopped, readyCount, failedCount, cancelledCount, now.UTC(), batchID); err != nil {
		return UploadBatch{}, nil, err
	}
	if err := tx.Commit(); err != nil {
		return UploadBatch{}, nil, err
	}

	batch, ok, err := s.findUploadBatchByID(ctx, batchID)
	if err != nil {
		return UploadBatch{}, nil, err
	}
	if !ok {
		return UploadBatch{}, nil, ErrNotFound
	}
	items, err := s.ListUploadItems(ctx, batchID)
	if err != nil {
		return UploadBatch{}, nil, err
	}
	return batch, items, nil
}

// CompleteUploadItem 在确认对象存储已有原文件后创建媒体资产。
// MediaAsset 只由后端在这个阶段创建，避免前端在原文件未上传前拿到时间线对象。
func (s *MySQLStore) CompleteUploadItem(ctx context.Context, input CompleteUploadItemInput) (UploadBatch, UploadItem, MediaAsset, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}
	defer rollbackUnlessCommitted(tx)

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	batch, err := scanUploadBatch(tx.QueryRowContext(ctx, `
SELECT id, family_id, created_by, status, active_slot, total_count, ready_count, failed_count, cancelled_count, created_at, completed_at, stopped_at
FROM upload_batches
WHERE id = ? AND family_id = ? AND created_by = ?
FOR UPDATE
`, input.BatchID, input.FamilyID, input.UploadedBy))
	if errors.Is(err, sql.ErrNoRows) {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, ErrNotFound
	}
	if err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}

	item, err := scanUploadItem(tx.QueryRowContext(ctx, `
SELECT id, upload_batch_id, media_asset_id, original_type, original_filename, content_type, byte_size, object_key, status, error_message, created_at, updated_at, completed_at
FROM upload_items
WHERE id = ? AND upload_batch_id = ?
FOR UPDATE
`, input.ItemID, batch.ID))
	if errors.Is(err, sql.ErrNoRows) {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, ErrNotFound
	}
	if err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}
	if item.MediaAssetID != 0 || item.Status == UploadItemStatusCancelled {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, ErrInvalidUpload
	}

	mediaType, err := mediaTypeForOriginalType(item.OriginalType)
	if err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}
	mediaResult, err := tx.ExecContext(ctx, `
INSERT INTO media_assets (family_id, uploaded_by, media_type, status, rendition_status, uploaded_at)
VALUES (?, ?, ?, ?, ?, ?)
`, batch.FamilyID, batch.CreatedBy, mediaType, MediaStatusActive, RenditionStatusPending, now.UTC())
	if err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}
	mediaID, err := mediaResult.LastInsertId()
	if err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}
	byteSize := input.ObjectSize
	if byteSize <= 0 {
		byteSize = item.ByteSize
	}
	contentType := fallbackString(input.ObjectType, item.ContentType)
	if _, err := tx.ExecContext(ctx, `
INSERT INTO media_originals (media_asset_id, original_type, object_key, original_filename, content_type, byte_size, checksum_sha256, uploaded_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, mediaID, item.OriginalType, item.ObjectKey, item.OriginalFilename, contentType, byteSize, nullableString(input.ChecksumSHA256), now.UTC()); err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE upload_items
SET media_asset_id = ?, status = ?, updated_at = ?
WHERE id = ?
`, mediaID, UploadItemStatusProcessing, now.UTC(), item.ID); err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}
	batch, err = recalculateUploadBatchInTx(ctx, tx, batch)
	if err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}
	if err := tx.Commit(); err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}

	item.MediaAssetID = mediaID
	item.Status = UploadItemStatusProcessing
	item.UpdatedAt = now
	asset := MediaAsset{
		ID:              mediaID,
		FamilyID:        batch.FamilyID,
		UploadedBy:      batch.CreatedBy,
		MediaType:       mediaType,
		Status:          MediaStatusActive,
		RenditionStatus: RenditionStatusPending,
		UploadedAt:      now,
	}
	return batch, item, asset, nil
}

// MarkUploadItemFailed 记录浏览器直传对象存储失败。
// 失败不会创建媒体资产，只让 UploadItem 保持可重试。
func (s *MySQLStore) MarkUploadItemFailed(ctx context.Context, input UpdateUploadItemStatusInput) (UploadBatch, UploadItem, error) {
	return s.updateUploadItemForRetry(ctx, input, func(item UploadItem, now time.Time) (UploadItem, error) {
		if item.MediaAssetID != 0 || item.Status == UploadItemStatusProcessing || item.Status == UploadItemStatusReady || item.Status == UploadItemStatusCancelled {
			return UploadItem{}, ErrInvalidUpload
		}
		item.Status = UploadItemStatusUploadFailed
		item.ErrorMessage = input.ErrorMessage
		item.UpdatedAt = now
		item.CompletedAt = now
		return item, nil
	})
}

// RetryUploadItem 重置可重传条目，HTTP 层会基于返回条目重新签发上传 URL。
func (s *MySQLStore) RetryUploadItem(ctx context.Context, input UpdateUploadItemStatusInput) (UploadBatch, UploadItem, error) {
	return s.updateUploadItemForRetry(ctx, input, func(item UploadItem, now time.Time) (UploadItem, error) {
		if item.MediaAssetID != 0 || (item.Status != UploadItemStatusWaiting && item.Status != UploadItemStatusUploadFailed) {
			return UploadItem{}, ErrInvalidUpload
		}
		item.Status = UploadItemStatusWaiting
		item.ErrorMessage = ""
		item.UpdatedAt = now
		item.CompletedAt = time.Time{}
		return item, nil
	})
}

func (s *MySQLStore) updateUploadItemForRetry(
	ctx context.Context,
	input UpdateUploadItemStatusInput,
	mutate func(UploadItem, time.Time) (UploadItem, error),
) (UploadBatch, UploadItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	defer rollbackUnlessCommitted(tx)

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	batch, err := scanUploadBatch(tx.QueryRowContext(ctx, `
SELECT id, family_id, created_by, status, active_slot, total_count, ready_count, failed_count, cancelled_count, created_at, completed_at, stopped_at
FROM upload_batches
WHERE id = ? AND family_id = ? AND created_by = ?
FOR UPDATE
`, input.BatchID, input.FamilyID, input.ActorUserID))
	if errors.Is(err, sql.ErrNoRows) {
		return UploadBatch{}, UploadItem{}, ErrNotFound
	}
	if err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	item, err := scanUploadItem(tx.QueryRowContext(ctx, `
SELECT id, upload_batch_id, media_asset_id, original_type, original_filename, content_type, byte_size, object_key, status, error_message, created_at, updated_at, completed_at
FROM upload_items
WHERE id = ? AND upload_batch_id = ?
FOR UPDATE
`, input.ItemID, batch.ID))
	if errors.Is(err, sql.ErrNoRows) {
		return UploadBatch{}, UploadItem{}, ErrNotFound
	}
	if err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	item, err = mutate(item, now)
	if err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	var completedAt any
	if item.CompletedAt.IsZero() {
		completedAt = nil
	} else {
		completedAt = item.CompletedAt.UTC()
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE upload_items
SET status = ?, error_message = ?, updated_at = ?, completed_at = ?
WHERE id = ?
`, item.Status, nullableString(item.ErrorMessage), item.UpdatedAt.UTC(), completedAt, item.ID); err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	batch, err = recalculateUploadBatchInTx(ctx, tx, batch)
	if err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	if err := tx.Commit(); err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	return batch, item, nil
}

func (s *MySQLStore) findUploadBatchByID(ctx context.Context, batchID int64) (UploadBatch, bool, error) {
	batch, err := scanUploadBatch(s.db.QueryRowContext(ctx, `
SELECT id, family_id, created_by, status, active_slot, total_count, ready_count, failed_count, cancelled_count, created_at, completed_at, stopped_at
FROM upload_batches
WHERE id = ?
`, batchID))
	if errors.Is(err, sql.ErrNoRows) {
		return UploadBatch{}, false, nil
	}
	if err != nil {
		return UploadBatch{}, false, err
	}
	return batch, true, nil
}

func recalculateUploadBatchInTx(ctx context.Context, tx *sql.Tx, batch UploadBatch) (UploadBatch, error) {
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status = ?`, batch.ID, UploadItemStatusReady).Scan(&batch.ReadyCount); err != nil {
		return UploadBatch{}, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status IN (?, ?)`, batch.ID, UploadItemStatusUploadFailed, UploadItemStatusProcessingFailed).Scan(&batch.FailedCount); err != nil {
		return UploadBatch{}, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status = ?`, batch.ID, UploadItemStatusCancelled).Scan(&batch.CancelledCount); err != nil {
		return UploadBatch{}, err
	}
	var processingCount int
	var pendingCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status = ?`, batch.ID, UploadItemStatusProcessing).Scan(&processingCount); err != nil {
		return UploadBatch{}, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status IN (?, ?, ?)`, batch.ID, UploadItemStatusWaiting, UploadItemStatusUploading, UploadItemStatusUploaded).Scan(&pendingCount); err != nil {
		return UploadBatch{}, err
	}
	switch {
	case batch.CancelledCount == batch.TotalCount:
		batch.Status = UploadBatchStatusStopped
	case batch.FailedCount > 0:
		batch.Status = UploadBatchStatusPartiallyFailed
	case batch.ReadyCount == batch.TotalCount && batch.TotalCount > 0:
		batch.Status = UploadBatchStatusCompleted
	case processingCount > 0:
		batch.Status = UploadBatchStatusProcessing
	case pendingCount > 0:
		batch.Status = UploadBatchStatusCreated
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE upload_batches
SET status = ?, ready_count = ?, failed_count = ?, cancelled_count = ?
WHERE id = ?
`, batch.Status, batch.ReadyCount, batch.FailedCount, batch.CancelledCount, batch.ID); err != nil {
		return UploadBatch{}, err
	}
	return batch, nil
}

type uploadBatchScanner interface {
	Scan(dest ...any) error
}

type uploadItemScanner interface {
	Scan(dest ...any) error
}

func scanUploadBatch(scanner uploadBatchScanner) (UploadBatch, error) {
	var batch UploadBatch
	var activeSlot sql.NullInt64
	var completedAt sql.NullTime
	var stoppedAt sql.NullTime
	err := scanner.Scan(
		&batch.ID,
		&batch.FamilyID,
		&batch.CreatedBy,
		&batch.Status,
		&activeSlot,
		&batch.TotalCount,
		&batch.ReadyCount,
		&batch.FailedCount,
		&batch.CancelledCount,
		&batch.CreatedAt,
		&completedAt,
		&stoppedAt,
	)
	if err != nil {
		return UploadBatch{}, err
	}
	if activeSlot.Valid {
		batch.ActiveSlot = int(activeSlot.Int64)
	}
	if completedAt.Valid {
		batch.CompletedAt = completedAt.Time
	}
	if stoppedAt.Valid {
		batch.StoppedAt = stoppedAt.Time
	}
	return batch, nil
}

func scanUploadItem(scanner uploadItemScanner) (UploadItem, error) {
	var item UploadItem
	var mediaAssetID sql.NullInt64
	var errorMessage sql.NullString
	var completedAt sql.NullTime
	err := scanner.Scan(
		&item.ID,
		&item.UploadBatchID,
		&mediaAssetID,
		&item.OriginalType,
		&item.OriginalFilename,
		&item.ContentType,
		&item.ByteSize,
		&item.ObjectKey,
		&item.Status,
		&errorMessage,
		&item.CreatedAt,
		&item.UpdatedAt,
		&completedAt,
	)
	if err != nil {
		return UploadItem{}, err
	}
	if mediaAssetID.Valid {
		item.MediaAssetID = mediaAssetID.Int64
	}
	if errorMessage.Valid {
		item.ErrorMessage = errorMessage.String
	}
	if completedAt.Valid {
		item.CompletedAt = completedAt.Time
	}
	return item, nil
}

func scanTimelineMedia(scanner uploadItemScanner) (TimelineMedia, error) {
	var item TimelineMedia
	var capturedAt sql.NullTime
	var deletedAt sql.NullTime
	var display MediaRendition
	var displayWidth sql.NullInt64
	var displayHeight sql.NullInt64
	var displayDurationMillis sql.NullInt64
	var displayErrorMessage sql.NullString
	var thumbID sql.NullInt64
	var thumbMediaAssetID sql.NullInt64
	var thumbRenditionType sql.NullString
	var thumbObjectKey sql.NullString
	var thumbContentType sql.NullString
	var thumbByteSize sql.NullInt64
	var thumbWidth sql.NullInt64
	var thumbHeight sql.NullInt64
	var thumbDurationMillis sql.NullInt64
	var thumbStatus sql.NullString
	var thumbErrorMessage sql.NullString

	err := scanner.Scan(
		&item.Asset.ID,
		&item.Asset.FamilyID,
		&item.Asset.UploadedBy,
		&item.Asset.MediaType,
		&item.Asset.Status,
		&item.Asset.RenditionStatus,
		&capturedAt,
		&item.Asset.UploadedAt,
		&deletedAt,
		&item.UploadedByDisplayName,
		&display.ID,
		&display.MediaAssetID,
		&display.RenditionType,
		&display.ObjectKey,
		&display.ContentType,
		&display.ByteSize,
		&displayWidth,
		&displayHeight,
		&displayDurationMillis,
		&display.Status,
		&displayErrorMessage,
		&thumbID,
		&thumbMediaAssetID,
		&thumbRenditionType,
		&thumbObjectKey,
		&thumbContentType,
		&thumbByteSize,
		&thumbWidth,
		&thumbHeight,
		&thumbDurationMillis,
		&thumbStatus,
		&thumbErrorMessage,
	)
	if err != nil {
		return TimelineMedia{}, err
	}
	if capturedAt.Valid {
		item.Asset.CapturedAt = capturedAt.Time
	}
	if deletedAt.Valid {
		item.Asset.DeletedAt = deletedAt.Time
	}
	if displayErrorMessage.Valid {
		display.ErrorMessage = displayErrorMessage.String
	}
	if displayWidth.Valid {
		display.Width = int(displayWidth.Int64)
	}
	if displayHeight.Valid {
		display.Height = int(displayHeight.Int64)
	}
	if displayDurationMillis.Valid {
		display.DurationMillis = displayDurationMillis.Int64
	}
	item.Display = display
	item.Thumbnail = display
	if thumbID.Valid {
		item.Thumbnail = MediaRendition{
			ID:             thumbID.Int64,
			MediaAssetID:   thumbMediaAssetID.Int64,
			RenditionType:  thumbRenditionType.String,
			ObjectKey:      thumbObjectKey.String,
			ContentType:    thumbContentType.String,
			ByteSize:       thumbByteSize.Int64,
			Width:          int(thumbWidth.Int64),
			Height:         int(thumbHeight.Int64),
			DurationMillis: thumbDurationMillis.Int64,
			Status:         thumbStatus.String,
			ErrorMessage:   thumbErrorMessage.String,
		}
	}
	return item, nil
}

// CreateUploadBatch 在事务中创建上传任务和文件条目。
// active_slot=1 配合唯一键保证同一用户同一家庭最多只有一个 active 上传任务。
func (s *MySQLStore) CreateUploadBatch(ctx context.Context, input CreateUploadBatchInput) (UploadBatch, []UploadItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return UploadBatch{}, nil, err
	}
	defer rollbackUnlessCommitted(tx)

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	result, err := tx.ExecContext(ctx, `
INSERT INTO upload_batches (family_id, created_by, status, active_slot, total_count, created_at)
VALUES (?, ?, ?, ?, ?, ?)
`, input.FamilyID, input.CreatedBy, UploadBatchStatusCreated, 1, len(input.Items), now.UTC())
	if err != nil {
		if isDuplicateKey(err) {
			return UploadBatch{}, nil, ErrAlreadyExists
		}
		return UploadBatch{}, nil, err
	}
	batchID, err := result.LastInsertId()
	if err != nil {
		return UploadBatch{}, nil, err
	}

	items := make([]UploadItem, 0, len(input.Items))
	for _, itemInput := range input.Items {
		itemResult, err := tx.ExecContext(ctx, `
INSERT INTO upload_items (upload_batch_id, original_type, original_filename, content_type, byte_size, object_key, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, batchID, itemInput.OriginalType, itemInput.OriginalFilename, itemInput.ContentType, itemInput.ByteSize, itemInput.ObjectKey, UploadItemStatusWaiting, now.UTC(), now.UTC())
		if err != nil {
			if isDuplicateKey(err) {
				return UploadBatch{}, nil, ErrAlreadyExists
			}
			return UploadBatch{}, nil, err
		}
		itemID, err := itemResult.LastInsertId()
		if err != nil {
			return UploadBatch{}, nil, err
		}
		items = append(items, UploadItem{
			ID:               itemID,
			UploadBatchID:    batchID,
			OriginalType:     itemInput.OriginalType,
			OriginalFilename: itemInput.OriginalFilename,
			ContentType:      itemInput.ContentType,
			ByteSize:         itemInput.ByteSize,
			ObjectKey:        itemInput.ObjectKey,
			Status:           UploadItemStatusWaiting,
			CreatedAt:        now,
			UpdatedAt:        now,
		})
	}
	if err := tx.Commit(); err != nil {
		return UploadBatch{}, nil, err
	}

	return UploadBatch{
		ID:         batchID,
		FamilyID:   input.FamilyID,
		CreatedBy:  input.CreatedBy,
		Status:     UploadBatchStatusCreated,
		ActiveSlot: 1,
		TotalCount: len(input.Items),
		CreatedAt:  now,
	}, items, nil
}

// findMemberInTx 在加入邀请事务中读取并锁定成员记录。
func findMemberInTx(ctx context.Context, tx *sql.Tx, familyID int64, userID int64) (FamilyMember, bool, error) {
	var member FamilyMember
	err := tx.QueryRowContext(ctx, `
SELECT id, family_id, user_id, display_name, role, status
FROM family_members
WHERE family_id = ? AND user_id = ?
FOR UPDATE
`, familyID, userID).Scan(&member.ID, &member.FamilyID, &member.UserID, &member.DisplayName, &member.Role, &member.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return FamilyMember{}, false, nil
	}
	if err != nil {
		return FamilyMember{}, false, err
	}
	return member, true, nil
}

// rollbackUnlessCommitted 作为事务 defer 使用；事务已提交时 Rollback 会安全失败。
func rollbackUnlessCommitted(tx *sql.Tx) {
	_ = tx.Rollback()
}

// isDuplicateKey 把 MySQL 唯一键冲突转换成 store 层统一错误。
func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

// nullableString 把空字符串保存为 NULL，区分“未预填显示名”和“预填了空白”。
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
