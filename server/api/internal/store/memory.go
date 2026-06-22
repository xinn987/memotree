// MemoryStore 是 Store 的内存实现。
//
// 它用于单元测试和无 MYSQL_DSN 的快速本地运行；不承担真实持久化能力，
// API 进程重启后这里的数据会全部丢失。
package store

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStore 用 map 保存用户、会话、家庭、成员和邀请。
// 所有公开方法都加锁，避免本地并发请求时破坏内存状态。
type MemoryStore struct {
	mu sync.Mutex

	nextUserID      int64
	nextSessionID   int64
	nextFamilyID    int64
	nextMemberID    int64
	nextInviteID    int64
	nextBatchID     int64
	nextItemID      int64
	nextMediaID     int64
	nextOriginalID  int64
	nextRenditionID int64

	users           map[int64]User
	passwordHashes  map[int64]string
	loginToUserID   map[string]int64
	sessions        map[string]Session
	families        map[int64]family
	members         map[int64]FamilyMember
	invites         map[string]FamilyInvite
	uploadBatches   map[int64]UploadBatch
	uploadItems     map[int64]UploadItem
	mediaAssets     map[int64]MediaAsset
	mediaOriginals  map[int64]MediaOriginal
	mediaRenditions map[int64]MediaRendition
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
		users:           map[int64]User{},
		passwordHashes:  map[int64]string{},
		loginToUserID:   map[string]int64{},
		sessions:        map[string]Session{},
		families:        map[int64]family{},
		members:         map[int64]FamilyMember{},
		invites:         map[string]FamilyInvite{},
		uploadBatches:   map[int64]UploadBatch{},
		uploadItems:     map[int64]UploadItem{},
		mediaAssets:     map[int64]MediaAsset{},
		mediaOriginals:  map[int64]MediaOriginal{},
		mediaRenditions: map[int64]MediaRendition{},
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
	now := time.Now()
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
		JoinedAt:    now,
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

func (s *MemoryStore) IsActiveMember(_ context.Context, familyID int64, userID int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, member := range s.members {
		if member.FamilyID == familyID && member.UserID == userID && member.Status == MemberStatusActive {
			return true, nil
		}
	}
	return false, nil
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

func (s *MemoryStore) ListFamilyMembers(_ context.Context, familyID int64) ([]FamilyMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []FamilyMember{}
	for _, member := range s.members {
		if member.FamilyID == familyID {
			result = append(result, member)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Status != result[j].Status {
			return result[i].Status == MemberStatusActive
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (s *MemoryStore) UpdateFamilyMemberDisplayName(_ context.Context, familyID int64, memberID int64, displayName string) (FamilyMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	member, ok := s.members[memberID]
	if !ok || member.FamilyID != familyID {
		return FamilyMember{}, ErrNotFound
	}
	member.DisplayName = displayName
	s.members[member.ID] = member
	return member, nil
}

func (s *MemoryStore) RemoveFamilyMember(_ context.Context, familyID int64, memberID int64, actorUserID int64, now time.Time) (FamilyMember, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	member, ok := s.members[memberID]
	if !ok || member.FamilyID != familyID {
		return FamilyMember{}, ErrNotFound
	}
	if member.UserID == actorUserID {
		// 成员管理里的“移除”只用于移除他人；自己离开家庭需要单独的产品动作。
		return FamilyMember{}, ErrSelfRemoval
	}
	if member.Status == MemberStatusRemoved {
		return member, nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	if member.Role == MemberRoleAdmin && s.activeAdminCount(familyID) <= 1 {
		// 最后一个 active admin 不能被移除，否则家庭空间会失去可管理者。
		return FamilyMember{}, ErrLastAdmin
	}
	member.Status = MemberStatusRemoved
	member.RemovedAt = now
	s.members[member.ID] = member
	return member, nil
}

// ListTimelineMedia 返回家庭时间线中已经可展示的媒体。
// 这里故意过滤掉仍在处理中的资产，保证上传任务状态和主时间线是两个清晰入口。
func (s *MemoryStore) ListTimelineMedia(_ context.Context, input ListTimelineMediaInput) ([]TimelineMedia, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	limit := input.Limit
	if limit <= 0 {
		limit = 60
	}
	result := []TimelineMedia{}
	for _, asset := range s.mediaAssets {
		if asset.FamilyID != input.FamilyID || asset.Status != MediaStatusActive || asset.RenditionStatus != RenditionStatusReady || !asset.DeletedAt.IsZero() {
			continue
		}
		if input.MediaType != "" && asset.MediaType != input.MediaType {
			continue
		}
		sortTime := timelineSortTime(asset)
		if !input.MonthFrom.IsZero() && (sortTime.Before(input.MonthFrom) || !sortTime.Before(input.MonthTo)) {
			continue
		}
		if !input.AfterTime.IsZero() && (sortTime.After(input.AfterTime) || sortTime.Equal(input.AfterTime) && asset.ID >= input.AfterID) {
			continue
		}
		display, thumbnail, ok := s.readyPreviewRenditions(asset)
		if !ok {
			continue
		}
		result = append(result, TimelineMedia{
			Asset:                 asset,
			UploadedByDisplayName: s.memberDisplayName(asset.FamilyID, asset.UploadedBy),
			Thumbnail:             thumbnail,
			Display:               display,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		left := timelineSortTime(result[i].Asset)
		right := timelineSortTime(result[j].Asset)
		if left.Equal(right) {
			return result[i].Asset.ID > result[j].Asset.ID
		}
		return left.After(right)
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// FindMediaDetail 返回单个已可展示媒体的详情数据。
// 不可见、删除或仍在处理中的媒体统一返回 ok=false。
func (s *MemoryStore) FindMediaDetail(_ context.Context, input FindMediaDetailInput) (MediaDetail, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	asset, ok := s.mediaAssets[input.MediaID]
	if !ok ||
		asset.FamilyID != input.FamilyID ||
		asset.Status != MediaStatusActive ||
		asset.RenditionStatus != RenditionStatusReady ||
		!asset.DeletedAt.IsZero() {
		return MediaDetail{}, false, nil
	}
	display, thumbnail, ok := s.readyPreviewRenditions(asset)
	if !ok {
		return MediaDetail{}, false, nil
	}
	return MediaDetail{
		Asset:                 asset,
		UploadedByDisplayName: s.memberDisplayName(asset.FamilyID, asset.UploadedBy),
		Thumbnail:             thumbnail,
		Display:               display,
	}, true, nil
}

func (s *MemoryStore) SoftDeleteMedia(_ context.Context, familyID int64, mediaID int64, now time.Time) (MediaAsset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	asset, ok := s.mediaAssets[mediaID]
	if !ok || asset.FamilyID != familyID || asset.Status != MediaStatusActive {
		return MediaAsset{}, ErrNotFound
	}
	if now.IsZero() {
		now = time.Now()
	}
	asset.Status = MediaStatusDeleted
	asset.DeletedAt = now
	s.mediaAssets[asset.ID] = asset
	return asset, nil
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
		member.RemovedAt = time.Time{}
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
		JoinedAt:    now,
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

// FindActiveUploadBatch 查找同一用户在同一家庭中的 active 上传任务。
// API 会用它把“再次上传”导向当前任务页，避免创建第二个 active batch。
func (s *MemoryStore) FindActiveUploadBatch(_ context.Context, familyID int64, userID int64) (UploadBatch, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, batch := range s.uploadBatches {
		if batch.FamilyID == familyID && batch.CreatedBy == userID && batch.ActiveSlot == 1 {
			return batch, true, nil
		}
	}
	return UploadBatch{}, false, nil
}

// ListUploadBatches 按创建时间倒序返回最近上传任务。
// 管理员调用时 IncludeFamily=true，可查看整个家庭；普通成员只读取自己创建的任务。
func (s *MemoryStore) ListUploadBatches(_ context.Context, input ListUploadBatchesInput) ([]UploadBatch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	result := []UploadBatch{}
	for _, batch := range s.uploadBatches {
		if batch.FamilyID != input.FamilyID {
			continue
		}
		if !input.IncludeFamily && batch.CreatedBy != input.ActorUserID {
			continue
		}
		result = append(result, batch)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].ID > result[j].ID
		}
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// FindUploadBatch 按家庭和任务 ID 读取上传任务。
// familyID 是权限边界的一部分，避免跨家庭枚举任务 ID。
func (s *MemoryStore) FindUploadBatch(_ context.Context, familyID int64, batchID int64) (UploadBatch, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.uploadBatches[batchID]
	if !ok || batch.FamilyID != familyID {
		return UploadBatch{}, false, nil
	}
	return batch, true, nil
}

// ListUploadItems 返回上传任务下的文件条目，按 ID 稳定排序，方便前端和测试使用。
func (s *MemoryStore) ListUploadItems(_ context.Context, batchID int64) ([]UploadItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := []UploadItem{}
	for _, item := range s.uploadItems {
		if item.UploadBatchID == batchID {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items, nil
}

// StopUploadBatch 停止上传任务，并取消尚未完成的文件项。
// 已经 ready 或失败的条目保留原状态，便于后续页面展示真实结果。
func (s *MemoryStore) StopUploadBatch(_ context.Context, batchID int64, now time.Time) (UploadBatch, []UploadItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.uploadBatches[batchID]
	if !ok {
		return UploadBatch{}, nil, ErrNotFound
	}
	if now.IsZero() {
		now = time.Now()
	}

	batch.Status = UploadBatchStatusStopped
	batch.ActiveSlot = 0
	batch.StoppedAt = now
	batch.CancelledCount = 0
	batch.ReadyCount = 0
	batch.FailedCount = 0

	items := []UploadItem{}
	for id, item := range s.uploadItems {
		if item.UploadBatchID != batchID {
			continue
		}
		if isCancellableUploadItemStatus(item.Status) {
			item.Status = UploadItemStatusCancelled
			item.UpdatedAt = now
			item.CompletedAt = now
			s.uploadItems[id] = item
		}
		switch item.Status {
		case UploadItemStatusReady:
			batch.ReadyCount++
		case UploadItemStatusUploadFailed, UploadItemStatusProcessingFailed:
			batch.FailedCount++
		case UploadItemStatusCancelled:
			batch.CancelledCount++
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	s.uploadBatches[batch.ID] = batch
	return batch, items, nil
}

// CompleteUploadItem 在原文件 PUT 成功后，把上传条目转成媒体资产。
// 这里是上传状态机的关键边界：前端只能完成 UploadItem，MediaAsset 由后端创建。
func (s *MemoryStore) CompleteUploadItem(_ context.Context, input CompleteUploadItemInput) (UploadBatch, UploadItem, MediaAsset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	batch, ok := s.uploadBatches[input.BatchID]
	if !ok || batch.FamilyID != input.FamilyID || batch.CreatedBy != input.UploadedBy {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, ErrNotFound
	}
	item, ok := s.uploadItems[input.ItemID]
	if !ok || item.UploadBatchID != batch.ID {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, ErrNotFound
	}
	if item.MediaAssetID != 0 || item.Status == UploadItemStatusCancelled {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, ErrInvalidUpload
	}

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	mediaType, err := mediaTypeForOriginalType(item.OriginalType)
	if err != nil {
		return UploadBatch{}, UploadItem{}, MediaAsset{}, err
	}

	s.nextMediaID++
	asset := MediaAsset{
		ID:              s.nextMediaID,
		FamilyID:        batch.FamilyID,
		UploadedBy:      batch.CreatedBy,
		MediaType:       mediaType,
		Status:          MediaStatusActive,
		RenditionStatus: RenditionStatusPending,
		UploadedAt:      now,
	}
	s.mediaAssets[asset.ID] = asset

	s.nextOriginalID++
	byteSize := input.ObjectSize
	if byteSize <= 0 {
		byteSize = item.ByteSize
	}
	original := MediaOriginal{
		ID:               s.nextOriginalID,
		MediaAssetID:     asset.ID,
		OriginalType:     item.OriginalType,
		ObjectKey:        item.ObjectKey,
		OriginalFilename: item.OriginalFilename,
		ContentType:      fallbackString(input.ObjectType, item.ContentType),
		ByteSize:         byteSize,
		ChecksumSHA256:   input.ChecksumSHA256,
		UploadedAt:       now,
	}
	s.mediaOriginals[original.ID] = original

	item.MediaAssetID = asset.ID
	item.Status = UploadItemStatusProcessing
	item.UpdatedAt = now
	s.uploadItems[item.ID] = item

	batch = s.recalculateUploadBatch(batch)
	s.uploadBatches[batch.ID] = batch
	return batch, item, asset, nil
}

// MarkUploadItemFailed 记录浏览器直传对象存储失败。
// 失败只绑定 UploadItem，不创建 MediaAsset，后续可重新生成上传 URL 重试。
func (s *MemoryStore) MarkUploadItemFailed(_ context.Context, input UpdateUploadItemStatusInput) (UploadBatch, UploadItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, item, err := s.findMutableUploadItem(input.BatchID, input.ItemID, input.FamilyID, input.ActorUserID)
	if err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	if item.MediaAssetID != 0 || item.Status == UploadItemStatusProcessing || item.Status == UploadItemStatusReady || item.Status == UploadItemStatusCancelled {
		return UploadBatch{}, UploadItem{}, ErrInvalidUpload
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	item.Status = UploadItemStatusUploadFailed
	item.ErrorMessage = input.ErrorMessage
	item.UpdatedAt = now
	item.CompletedAt = now
	s.uploadItems[item.ID] = item
	batch = s.recalculateUploadBatch(batch)
	s.uploadBatches[batch.ID] = batch
	return batch, item, nil
}

// RetryUploadItem 把可重传的条目重置为 waiting，让 HTTP 层重新签发短期上传 URL。
func (s *MemoryStore) RetryUploadItem(_ context.Context, input UpdateUploadItemStatusInput) (UploadBatch, UploadItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, item, err := s.findMutableUploadItem(input.BatchID, input.ItemID, input.FamilyID, input.ActorUserID)
	if err != nil {
		return UploadBatch{}, UploadItem{}, err
	}
	if item.MediaAssetID != 0 || (item.Status != UploadItemStatusWaiting && item.Status != UploadItemStatusUploadFailed) {
		return UploadBatch{}, UploadItem{}, ErrInvalidUpload
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	item.Status = UploadItemStatusWaiting
	item.ErrorMessage = ""
	item.UpdatedAt = now
	item.CompletedAt = time.Time{}
	s.uploadItems[item.ID] = item
	batch = s.recalculateUploadBatch(batch)
	s.uploadBatches[batch.ID] = batch
	return batch, item, nil
}

func (s *MemoryStore) RetryProcessingItem(_ context.Context, input UpdateUploadItemStatusInput) (UploadBatch, UploadItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch, ok := s.uploadBatches[input.BatchID]
	if !ok || batch.FamilyID != input.FamilyID {
		return UploadBatch{}, UploadItem{}, ErrNotFound
	}
	item, ok := s.uploadItems[input.ItemID]
	if !ok || item.UploadBatchID != batch.ID || item.MediaAssetID == 0 || item.Status != UploadItemStatusProcessingFailed {
		return UploadBatch{}, UploadItem{}, ErrInvalidUpload
	}
	asset, ok := s.mediaAssets[item.MediaAssetID]
	if !ok || asset.FamilyID != input.FamilyID || asset.Status != MediaStatusActive {
		return UploadBatch{}, UploadItem{}, ErrNotFound
	}
	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	// 原文件已经入库，重试只重新排队处理任务，不生成新的 upload item。
	item.Status = UploadItemStatusProcessing
	item.ErrorMessage = ""
	item.UpdatedAt = now
	item.CompletedAt = time.Time{}
	s.uploadItems[item.ID] = item

	asset.RenditionStatus = RenditionStatusPending
	s.mediaAssets[asset.ID] = asset

	batch = s.recalculateUploadBatch(batch)
	s.uploadBatches[batch.ID] = batch
	return batch, item, nil
}

// CreateUploadBatch 创建上传任务和待上传文件条目。
// 同一用户同一家庭只允许一个 active batch，和 MySQL 的唯一键语义保持一致。
func (s *MemoryStore) CreateUploadBatch(_ context.Context, input CreateUploadBatchInput) (UploadBatch, []UploadItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, batch := range s.uploadBatches {
		if batch.FamilyID == input.FamilyID && batch.CreatedBy == input.CreatedBy && batch.ActiveSlot == 1 {
			return UploadBatch{}, nil, ErrAlreadyExists
		}
	}

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	s.nextBatchID++
	batch := UploadBatch{
		ID:         s.nextBatchID,
		FamilyID:   input.FamilyID,
		CreatedBy:  input.CreatedBy,
		Status:     UploadBatchStatusCreated,
		ActiveSlot: 1,
		TotalCount: len(input.Items),
		CreatedAt:  now,
	}
	s.uploadBatches[batch.ID] = batch

	items := make([]UploadItem, 0, len(input.Items))
	for _, itemInput := range input.Items {
		s.nextItemID++
		item := UploadItem{
			ID:               s.nextItemID,
			UploadBatchID:    batch.ID,
			OriginalType:     itemInput.OriginalType,
			OriginalFilename: itemInput.OriginalFilename,
			ContentType:      itemInput.ContentType,
			ByteSize:         itemInput.ByteSize,
			ObjectKey:        itemInput.ObjectKey,
			Status:           UploadItemStatusWaiting,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		s.uploadItems[item.ID] = item
		items = append(items, item)
	}
	return batch, items, nil
}

func (s *MemoryStore) readyPreviewRenditions(asset MediaAsset) (MediaRendition, MediaRendition, bool) {
	var display MediaRendition
	var thumbnail MediaRendition
	for _, rendition := range s.mediaRenditions {
		if rendition.MediaAssetID != asset.ID || rendition.Status != RenditionStatusReady {
			continue
		}
		switch rendition.RenditionType {
		case displayRenditionTypeForMedia(asset.MediaType):
			display = rendition
		case RenditionTypeThumbnail:
			thumbnail = rendition
		}
	}
	if display.ObjectKey == "" {
		return MediaRendition{}, MediaRendition{}, false
	}
	if thumbnail.ObjectKey == "" {
		thumbnail = display
	}
	return display, thumbnail, true
}

func (s *MemoryStore) memberDisplayName(familyID int64, userID int64) string {
	for _, member := range s.members {
		if member.FamilyID == familyID && member.UserID == userID && member.DisplayName != "" {
			return member.DisplayName
		}
	}
	if user, ok := s.users[userID]; ok {
		return user.DisplayName
	}
	return ""
}

func (s *MemoryStore) activeAdminCount(familyID int64) int {
	count := 0
	for _, member := range s.members {
		if member.FamilyID == familyID && member.Status == MemberStatusActive && member.Role == MemberRoleAdmin {
			count++
		}
	}
	return count
}

func timelineSortTime(asset MediaAsset) time.Time {
	if !asset.CapturedAt.IsZero() {
		return asset.CapturedAt
	}
	return asset.UploadedAt
}

func displayRenditionTypeForMedia(mediaType string) string {
	if mediaType == MediaTypeVideo {
		return RenditionTypeDisplayVideo
	}
	return RenditionTypeDisplayImage
}

func isCancellableUploadItemStatus(status string) bool {
	return status == UploadItemStatusWaiting ||
		status == UploadItemStatusUploading ||
		status == UploadItemStatusUploaded ||
		status == UploadItemStatusProcessing
}

func mediaTypeForOriginalType(originalType string) (string, error) {
	switch originalType {
	case OriginalTypeImage:
		return MediaTypePhoto, nil
	case OriginalTypeVideo:
		return MediaTypeVideo, nil
	default:
		return "", ErrInvalidUpload
	}
}

func fallbackString(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func (s *MemoryStore) findMutableUploadItem(batchID int64, itemID int64, familyID int64, actorUserID int64) (UploadBatch, UploadItem, error) {
	batch, ok := s.uploadBatches[batchID]
	if !ok || batch.FamilyID != familyID || batch.CreatedBy != actorUserID {
		return UploadBatch{}, UploadItem{}, ErrNotFound
	}
	item, ok := s.uploadItems[itemID]
	if !ok || item.UploadBatchID != batch.ID {
		return UploadBatch{}, UploadItem{}, ErrNotFound
	}
	return batch, item, nil
}

func (s *MemoryStore) recalculateUploadBatch(batch UploadBatch) UploadBatch {
	batch.ReadyCount = 0
	batch.FailedCount = 0
	batch.CancelledCount = 0
	hasProcessing := false
	hasPending := false
	for _, item := range s.uploadItems {
		if item.UploadBatchID != batch.ID {
			continue
		}
		switch item.Status {
		case UploadItemStatusReady:
			batch.ReadyCount++
		case UploadItemStatusUploadFailed, UploadItemStatusProcessingFailed:
			batch.FailedCount++
		case UploadItemStatusCancelled:
			batch.CancelledCount++
		case UploadItemStatusProcessing:
			hasProcessing = true
		case UploadItemStatusWaiting, UploadItemStatusUploading, UploadItemStatusUploaded:
			hasPending = true
		}
	}
	switch {
	case batch.CancelledCount == batch.TotalCount:
		batch.Status = UploadBatchStatusStopped
	case batch.FailedCount > 0:
		batch.Status = UploadBatchStatusPartiallyFailed
	case batch.ReadyCount == batch.TotalCount && batch.TotalCount > 0:
		batch.Status = UploadBatchStatusCompleted
	case hasProcessing:
		batch.Status = UploadBatchStatusProcessing
	case hasPending:
		batch.Status = UploadBatchStatusCreated
	}
	return batch
}
