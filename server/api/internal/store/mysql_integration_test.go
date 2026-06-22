// MySQL store 的可选集成测试。
//
// 默认单元测试不依赖 Docker/MySQL；只有设置 MEMOTREE_TEST_MYSQL_DSN 时才会真实连库。
package store

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"
)

func TestMySQLStoreIntegration(t *testing.T) {
	dsn := os.Getenv("MEMOTREE_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set MEMOTREE_TEST_MYSQL_DSN to run mysql integration test")
	}

	ctx := context.Background()
	db, mysqlStore, err := OpenMySQL(ctx, dsn)
	if err != nil {
		t.Fatalf("open mysql: %v", err)
	}
	defer db.Close()

	schemaSQL, err := os.ReadFile("../../../migrations/0001_initial_schema.sql")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	if err := ApplySchema(ctx, db, string(schemaSQL)); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	cleanMySQLTables(t, ctx, db)

	root, err := mysqlStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	if !root.IsSystemAdmin {
		t.Fatalf("first user should be system admin")
	}

	second, err := mysqlStore.CreateUser(ctx, "grandma", "hash", "奶奶账号")
	if err != nil {
		t.Fatalf("create second user: %v", err)
	}
	if second.IsSystemAdmin {
		t.Fatalf("second user should not be system admin")
	}

	family, err := mysqlStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, root.DisplayName)
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	assertMySQLTableExists(t, ctx, db, "media_assets")
	assertMySQLTableExists(t, ctx, db, "media_originals")
	assertMySQLTableExists(t, ctx, db, "media_renditions")
	assertMySQLTableExists(t, ctx, db, "upload_batches")
	assertMySQLTableExists(t, ctx, db, "upload_items")
	isMember, err := mysqlStore.IsActiveMember(ctx, family.ID, root.ID)
	if err != nil || !isMember {
		t.Fatalf("creator should be active family member, isMember=%v err=%v", isMember, err)
	}
	isAdmin, err := mysqlStore.IsActiveAdmin(ctx, family.ID, root.ID)
	if err != nil || !isAdmin {
		t.Fatalf("creator should be family admin, isAdmin=%v err=%v", isAdmin, err)
	}
	isMember, err = mysqlStore.IsActiveMember(ctx, family.ID, second.ID)
	if err != nil || isMember {
		t.Fatalf("not-yet-joined user should not be active family member, isMember=%v err=%v", isMember, err)
	}
	batch, uploadItems, err := mysqlStore.CreateUploadBatch(ctx, CreateUploadBatchInput{
		FamilyID:  family.ID,
		CreatedBy: root.ID,
		Items: []CreateUploadItemInput{
			{
				OriginalType:     OriginalTypeImage,
				OriginalFilename: "baby.jpg",
				ContentType:      "image/jpeg",
				ByteSize:         12345,
				ObjectKey:        "originals/families/1/users/1/baby.jpg",
			},
		},
		Now: time.Now(),
	})
	if err != nil {
		t.Fatalf("create upload batch: %v", err)
	}
	if batch.Status != UploadBatchStatusCreated || batch.TotalCount != 1 || len(uploadItems) != 1 {
		t.Fatalf("unexpected upload batch result: batch=%#v items=%#v", batch, uploadItems)
	}
	if _, _, err := mysqlStore.CreateUploadBatch(ctx, CreateUploadBatchInput{FamilyID: family.ID, CreatedBy: root.ID, Now: time.Now()}); err != ErrAlreadyExists {
		t.Fatalf("expected duplicate active upload batch to fail, got %v", err)
	}

	revokedInvite, err := mysqlStore.CreateInvite(ctx, family.ID, "token-hash-revoke", "token-plain-revoke", root.ID, "外公", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("create revoked invite candidate: %v", err)
	}
	revokedInvite, err = mysqlStore.RevokeInvite(ctx, family.ID, revokedInvite.ID, time.Now())
	if err != nil {
		t.Fatalf("revoke invite: %v", err)
	}
	if revokedInvite.Status != InviteStatusRevoked || revokedInvite.TokenPlaintext != "" {
		t.Fatalf("expected revoked invite without plaintext token, got %#v", revokedInvite)
	}
	if _, err := mysqlStore.JoinInvite(ctx, "token-hash-revoke", second.ID, second.DisplayName, time.Now()); err != ErrInvalidInvite {
		t.Fatalf("expected revoked invite to be invalid, got %v", err)
	}

	invite, err := mysqlStore.CreateInvite(ctx, family.ID, "token-hash", "token-plain", root.ID, "奶奶", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	invites, err := mysqlStore.ListInvitesForFamily(ctx, family.ID)
	if err != nil {
		t.Fatalf("list invites: %v", err)
	}
	if len(invites) != 2 || invites[0].TokenPlaintext != "token-plain" {
		t.Fatalf("expected invite list with plaintext token, got %#v", invites)
	}
	member, err := mysqlStore.JoinInvite(ctx, invite.TokenHash, second.ID, second.DisplayName, time.Now())
	if err != nil {
		t.Fatalf("join invite: %v", err)
	}
	if member.FamilyID != family.ID || member.Role != MemberRoleMember || member.DisplayName != "奶奶" {
		t.Fatalf("unexpected joined member: %#v", member)
	}
	renamed, err := mysqlStore.UpdateFamilyMemberDisplayName(ctx, family.ID, member.ID, member.DisplayName)
	if err != nil {
		t.Fatalf("repeat update member display name with same value: %v", err)
	}
	if renamed.DisplayName != member.DisplayName {
		t.Fatalf("expected unchanged display name to round-trip, got %#v", renamed)
	}
	isMember, err = mysqlStore.IsActiveMember(ctx, family.ID, second.ID)
	if err != nil || !isMember {
		t.Fatalf("joined user should be active family member, isMember=%v err=%v", isMember, err)
	}
	invites, err = mysqlStore.ListInvitesForFamily(ctx, family.ID)
	if err != nil {
		t.Fatalf("list invites after join: %v", err)
	}
	if invites[0].Status != InviteStatusUsed || invites[0].TokenPlaintext != "" {
		t.Fatalf("expected used invite without plaintext token, got %#v", invites[0])
	}
}

func cleanMySQLTables(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	// 按外键依赖从子表到父表清理，保证集成测试每次从空库状态开始。
	statements := []string{
		"DELETE FROM upload_items",
		"DELETE FROM media_renditions",
		"DELETE FROM media_originals",
		"DELETE FROM media_assets",
		"DELETE FROM upload_batches",
		"DELETE FROM family_invites",
		"DELETE FROM family_members",
		"DELETE FROM families",
		"DELETE FROM user_sessions",
		"DELETE FROM user_credentials",
		"DELETE FROM users",
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("clean table with %q: %v", statement, err)
		}
	}
}

func assertMySQLTableExists(t *testing.T, ctx context.Context, db *sql.DB, tableName string) {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM information_schema.tables
WHERE table_schema = DATABASE() AND table_name = ?
`, tableName).Scan(&count); err != nil {
		t.Fatalf("check table %s: %v", tableName, err)
	}
	if count != 1 {
		t.Fatalf("expected table %s to exist", tableName)
	}
}
