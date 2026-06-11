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
