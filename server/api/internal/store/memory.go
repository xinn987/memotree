// MemoryStore 是 Store 的内存实现。
//
// 它用于单元测试和无 MYSQL_DSN 的快速本地运行；不承担真实持久化能力，
// API 进程重启后这里的数据会全部丢失。
package store

import (
	"context"
	"strings"
	"sync"
	"time"
)

// MemoryStore 用 map 保存用户、会话、家庭、成员和邀请。
// 所有公开方法都加锁，避免本地并发请求时破坏内存状态。
type MemoryStore struct {
	mu sync.Mutex

	nextUserID    int64
	nextSessionID int64
	nextFamilyID  int64
	nextMemberID  int64
	nextInviteID  int64

	users          map[int64]User
	passwordHashes map[int64]string
	loginToUserID  map[string]int64
	sessions       map[string]Session
	families       map[int64]family
	members        map[int64]FamilyMember
	invites        map[string]FamilyInvite
}

// family 是内存 store 内部结构；HTTP 响应只暴露 FamilySummary。
type family struct {
	ID          int64
	DisplayName string
	Timezone    string
	CreatedBy   int64
}

// NewMemoryStore 创建一份空的内存 store。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:          map[int64]User{},
		passwordHashes: map[int64]string{},
		loginToUserID:  map[string]int64{},
		sessions:       map[string]Session{},
		families:       map[int64]family{},
		members:        map[int64]FamilyMember{},
		invites:        map[string]FamilyInvite{},
	}
}

// CreateUser 创建全局用户和登录凭证。
// 第一位注册用户自动成为系统初始管理员，用于初始化第一个家庭。
func (s *MemoryStore) CreateUser(_ context.Context, loginName string, passwordHash string, displayName string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	normalized := strings.ToLower(loginName)
	if _, exists := s.loginToUserID[normalized]; exists {
		return User{}, ErrAlreadyExists
	}
	s.nextUserID++
	// 第一个注册用户作为系统初始管理员，用于后续引导创建第一个家庭。
	isSystemAdmin := len(s.users) == 0
	created := User{ID: s.nextUserID, LoginName: loginName, DisplayName: displayName, IsSystemAdmin: isSystemAdmin}
	s.users[created.ID] = created
	s.passwordHashes[created.ID] = passwordHash
	s.loginToUserID[normalized] = created.ID
	return created, nil
}

func (s *MemoryStore) FindUserByLoginName(_ context.Context, loginName string) (User, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.loginToUserID[strings.ToLower(loginName)]
	if !ok {
		return User{}, "", ErrNotFound
	}
	return s.users[id], s.passwordHashes[id], nil
}

func (s *MemoryStore) FindUserByID(_ context.Context, id int64) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	found, ok := s.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return found, nil
}

func (s *MemoryStore) CreateSession(_ context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSessionID++
	s.sessions[tokenHash] = Session{UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt}
	return nil
}

func (s *MemoryStore) FindSession(_ context.Context, tokenHash string, now time.Time) (Session, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	found, ok := s.sessions[tokenHash]
	if !ok || !found.ExpiresAt.After(now) {
		return Session{}, false, nil
	}
	return found, true, nil
}

func (s *MemoryStore) DeleteSession(_ context.Context, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, tokenHash)
	return nil
}

// CreateFamily 创建家庭，并同时把创建者写入 active admin 成员。
// 这两个动作在 MySQL 实现里必须放在同一个事务里。
func (s *MemoryStore) CreateFamily(_ context.Context, displayName string, timezone string, creatorID int64, creatorDisplayName string) (FamilySummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextFamilyID++
	created := family{ID: s.nextFamilyID, DisplayName: displayName, Timezone: timezone, CreatedBy: creatorID}
	s.families[created.ID] = created

	s.nextMemberID++
	s.members[s.nextMemberID] = FamilyMember{
		ID:          s.nextMemberID,
		FamilyID:    created.ID,
		UserID:      creatorID,
		DisplayName: creatorDisplayName,
		Role:        MemberRoleAdmin,
		Status:      MemberStatusActive,
	}
	return FamilySummary{ID: created.ID, DisplayName: created.DisplayName, Timezone: created.Timezone, Role: MemberRoleAdmin, MemberDisplayName: creatorDisplayName}, nil
}

func (s *MemoryStore) ListFamiliesForUser(_ context.Context, userID int64) ([]FamilySummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []FamilySummary{}
	for _, member := range s.members {
		if member.UserID != userID || member.Status != MemberStatusActive {
			continue
		}
		family := s.families[member.FamilyID]
		result = append(result, FamilySummary{
			ID:                family.ID,
			DisplayName:       family.DisplayName,
			Timezone:          family.Timezone,
			Role:              member.Role,
			MemberDisplayName: member.DisplayName,
		})
	}
	return result, nil
}

func (s *MemoryStore) IsActiveAdmin(_ context.Context, familyID int64, userID int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, member := range s.members {
		if member.FamilyID == familyID && member.UserID == userID && member.Status == MemberStatusActive && member.Role == MemberRoleAdmin {
			return true, nil
		}
	}
	return false, nil
}

// CreateInvite 创建待使用邀请。
// MVP 阶段为了便于管理员重新复制邀请链接，store 同时保存 token hash 和 token 原文。
func (s *MemoryStore) CreateInvite(_ context.Context, familyID int64, tokenHash string, tokenPlaintext string, createdBy int64, memberDisplayName string, expiresAt time.Time) (FamilyInvite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextInviteID++
	created := FamilyInvite{
		ID:                s.nextInviteID,
		FamilyID:          familyID,
		TokenHash:         tokenHash,
		TokenPlaintext:    tokenPlaintext,
		CreatedBy:         createdBy,
		MemberDisplayName: memberDisplayName,
		Status:            InviteStatusPending,
		ExpiresAt:         expiresAt,
	}
	s.invites[tokenHash] = created
	return created, nil
}

func (s *MemoryStore) ListInvitesForFamily(_ context.Context, familyID int64) ([]FamilyInvite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []FamilyInvite{}
	for _, invite := range s.invites {
		if invite.FamilyID == familyID {
			result = append(result, invite)
		}
	}
	return result, nil
}

// RevokeInvite 撤销一条仍然待使用且未过期的邀请。
// 撤销后清空 token 原文，避免管理界面继续暴露已失效链接。
func (s *MemoryStore) RevokeInvite(_ context.Context, familyID int64, inviteID int64, now time.Time) (FamilyInvite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for tokenHash, invite := range s.invites {
		if invite.FamilyID != familyID || invite.ID != inviteID {
			continue
		}
		if invite.Status != InviteStatusPending || !invite.ExpiresAt.After(now) {
			return FamilyInvite{}, ErrInvalidInvite
		}
		invite.Status = InviteStatusRevoked
		invite.TokenPlaintext = ""
		s.invites[tokenHash] = invite
		return invite, nil
	}
	return FamilyInvite{}, ErrNotFound
}

// JoinInvite 使用邀请加入家庭。
// 如果用户已经是 active 成员，重复打开邀请链接不消耗邀请，保持体验可恢复。
func (s *MemoryStore) JoinInvite(_ context.Context, tokenHash string, userID int64, fallbackDisplayName string, now time.Time) (FamilyMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	invite, ok := s.invites[tokenHash]
	if !ok {
		return FamilyMember{}, ErrNotFound
	}
	if invite.Status != InviteStatusPending || !invite.ExpiresAt.After(now) {
		return FamilyMember{}, ErrInvalidInvite
	}

	for id, member := range s.members {
		if member.FamilyID != invite.FamilyID || member.UserID != userID {
			continue
		}
		if member.Status == MemberStatusActive {
			// 已经是活跃成员时不消耗邀请，方便重复打开邀请链接。
			return member, nil
		}
		member.Status = MemberStatusActive
		member.Role = MemberRoleMember
		if invite.MemberDisplayName != "" {
			member.DisplayName = invite.MemberDisplayName
		}
		s.members[id] = member
		s.markInviteUsed(tokenHash, invite, userID, now)
		return member, nil
	}

	displayName := fallbackDisplayName
	if invite.MemberDisplayName != "" {
		displayName = invite.MemberDisplayName
	}
	s.nextMemberID++
	member := FamilyMember{
		ID:          s.nextMemberID,
		FamilyID:    invite.FamilyID,
		UserID:      userID,
		DisplayName: displayName,
		Role:        MemberRoleMember,
		Status:      MemberStatusActive,
	}
	s.members[member.ID] = member
	s.markInviteUsed(tokenHash, invite, userID, now)
	return member, nil
}

// markInviteUsed 标记邀请已消耗。调用方必须已经持有 MemoryStore 的锁。
func (s *MemoryStore) markInviteUsed(tokenHash string, invite FamilyInvite, userID int64, now time.Time) {
	invite.Status = InviteStatusUsed
	invite.TokenPlaintext = ""
	invite.UsedBy = userID
	invite.UsedAt = now
	s.invites[tokenHash] = invite
}
