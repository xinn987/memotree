// Package store 定义 MemoTree API 的持久化边界。
//
// HTTP 层只依赖 Store 接口，不依赖 MySQL 或内存实现细节。
// 这样测试可以使用 MemoryStore，真实运行可以通过 MYSQL_DSN 切到 MySQLStore。
package store

import (
	"context"
	"errors"
	"time"
)

const (
	// DefaultFamilyTimezone 是 MVP 阶段的固定家庭时区。
	// 所有用户可见的家庭日期分组先按 Asia/Shanghai 处理，后续再开放设置。
	DefaultFamilyTimezone = "Asia/Shanghai"
	MemberRoleAdmin       = "admin"
	MemberRoleMember      = "member"
	MemberStatusActive    = "active"
	InviteStatusPending   = "pending"
	InviteStatusUsed      = "used"
	InviteStatusRevoked   = "revoked"
)

var (
	// ErrNotFound 是 store 层统一的“记录不存在”语义，HTTP 层负责转换为对应响应。
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInvite = errors.New("invalid invite")
)

// User 是全局唯一用户身份。
// 用户可以加入多个家庭；家庭内展示名和权限放在 FamilyMember 中。
type User struct {
	ID            int64
	LoginName     string
	DisplayName   string
	IsSystemAdmin bool
}

// Session 表示后端可识别的浏览器会话。
// TokenHash 持久化 token hash，不保存 cookie 中的 token 原文。
type Session struct {
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
}

// FamilySummary 是前端家庭列表和当前会话返回的轻量视图。
// 它合并了 Family 和当前用户在这个家庭里的 FamilyMember 信息。
type FamilySummary struct {
	ID                int64  `json:"id"`
	DisplayName       string `json:"displayName"`
	Timezone          string `json:"timezone"`
	Role              string `json:"role"`
	MemberDisplayName string `json:"memberDisplayName"`
}

// FamilyInvite 表示一次加入家庭的邀请。
// TokenHash 用于定位邀请，邀请 token 原文只在创建响应中返回一次。
type FamilyInvite struct {
	ID                int64
	FamilyID          int64
	TokenHash         string
	TokenPlaintext    string
	CreatedBy         int64
	MemberDisplayName string
	Status            string
	ExpiresAt         time.Time
	UsedBy            int64
	UsedAt            time.Time
}

// FamilyMember 表示一个用户在某个家庭里的身份、显示名和权限。
type FamilyMember struct {
	ID          int64
	FamilyID    int64
	UserID      int64
	DisplayName string
	Role        string
	Status      string
}

// Store 是 HTTP/领域逻辑访问持久化数据的唯一接口。
// 方法保持偏业务语义，避免 handler 直接拼 SQL 或理解具体表结构。
type Store interface {
	CreateUser(ctx context.Context, loginName string, passwordHash string, displayName string) (User, error)
	FindUserByLoginName(ctx context.Context, loginName string) (User, string, error)
	FindUserByID(ctx context.Context, id int64) (User, error)
	CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	FindSession(ctx context.Context, tokenHash string, now time.Time) (Session, bool, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	CreateFamily(ctx context.Context, displayName string, timezone string, creatorID int64, creatorDisplayName string) (FamilySummary, error)
	ListFamiliesForUser(ctx context.Context, userID int64) ([]FamilySummary, error)
	IsActiveAdmin(ctx context.Context, familyID int64, userID int64) (bool, error)
	CreateInvite(ctx context.Context, familyID int64, tokenHash string, tokenPlaintext string, createdBy int64, memberDisplayName string, expiresAt time.Time) (FamilyInvite, error)
	ListInvitesForFamily(ctx context.Context, familyID int64) ([]FamilyInvite, error)
	RevokeInvite(ctx context.Context, familyID int64, inviteID int64, now time.Time) (FamilyInvite, error)
	JoinInvite(ctx context.Context, tokenHash string, userID int64, fallbackDisplayName string, now time.Time) (FamilyMember, error)
}
