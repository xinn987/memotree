package store

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStoreActiveMembership(t *testing.T) {
	ctx := context.Background()
	memoryStore := NewMemoryStore()

	root, err := memoryStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	grandma, err := memoryStore.CreateUser(ctx, "grandma", "hash", "奶奶账号")
	if err != nil {
		t.Fatalf("create grandma user: %v", err)
	}
	family, err := memoryStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, root.DisplayName)
	if err != nil {
		t.Fatalf("create family: %v", err)
	}

	isMember, err := memoryStore.IsActiveMember(ctx, family.ID, root.ID)
	if err != nil || !isMember {
		t.Fatalf("creator should be active member, isMember=%v err=%v", isMember, err)
	}
	isMember, err = memoryStore.IsActiveMember(ctx, family.ID, grandma.ID)
	if err != nil || isMember {
		t.Fatalf("not-yet-joined user should not be active member, isMember=%v err=%v", isMember, err)
	}

	invite, err := memoryStore.CreateInvite(ctx, family.ID, "token-hash", "token-plain", root.ID, "奶奶", time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	if _, err := memoryStore.JoinInvite(ctx, invite.TokenHash, grandma.ID, grandma.DisplayName, time.Now()); err != nil {
		t.Fatalf("join invite: %v", err)
	}
	isMember, err = memoryStore.IsActiveMember(ctx, family.ID, grandma.ID)
	if err != nil || !isMember {
		t.Fatalf("joined user should be active member, isMember=%v err=%v", isMember, err)
	}

	memoryStore.mu.Lock()
	for id, member := range memoryStore.members {
		if member.FamilyID == family.ID && member.UserID == grandma.ID {
			member.Status = MemberStatusRemoved
			memoryStore.members[id] = member
		}
	}
	memoryStore.mu.Unlock()
	isMember, err = memoryStore.IsActiveMember(ctx, family.ID, grandma.ID)
	if err != nil || isMember {
		t.Fatalf("removed user should not be active member, isMember=%v err=%v", isMember, err)
	}
}

func TestMemoryStoreInviteEdgeCases(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC)
	memoryStore := NewMemoryStore()

	root, err := memoryStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	grandma, err := memoryStore.CreateUser(ctx, "grandma", "hash", "奶奶账号")
	if err != nil {
		t.Fatalf("create grandma user: %v", err)
	}
	aunt, err := memoryStore.CreateUser(ctx, "aunt", "hash", "姨姨账号")
	if err != nil {
		t.Fatalf("create aunt user: %v", err)
	}
	family, err := memoryStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, root.DisplayName)
	if err != nil {
		t.Fatalf("create family: %v", err)
	}

	expiredInvite, err := memoryStore.CreateInvite(ctx, family.ID, "expired-hash", "expired-token", root.ID, "姨姨", now.Add(-time.Minute))
	if err != nil {
		t.Fatalf("create expired invite: %v", err)
	}
	if _, err := memoryStore.JoinInvite(ctx, expiredInvite.TokenHash, aunt.ID, aunt.DisplayName, now); err != ErrInvalidInvite {
		t.Fatalf("expected expired invite to be invalid, got %v", err)
	}
	if _, err := memoryStore.JoinInvite(ctx, "missing-hash", aunt.ID, aunt.DisplayName, now); err != ErrNotFound {
		t.Fatalf("expected unknown invite to be not found, got %v", err)
	}

	revokedInvite, err := memoryStore.CreateInvite(ctx, family.ID, "revoked-hash", "revoked-token", root.ID, "姨姨", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("create revoked invite: %v", err)
	}
	if _, err := memoryStore.RevokeInvite(ctx, family.ID, revokedInvite.ID, now); err != nil {
		t.Fatalf("revoke invite: %v", err)
	}
	if _, err := memoryStore.JoinInvite(ctx, revokedInvite.TokenHash, aunt.ID, aunt.DisplayName, now); err != ErrInvalidInvite {
		t.Fatalf("expected revoked invite to be invalid, got %v", err)
	}

	firstInvite, err := memoryStore.CreateInvite(ctx, family.ID, "first-active-hash", "first-active-token", root.ID, "奶奶", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("create first active invite: %v", err)
	}
	if _, err := memoryStore.JoinInvite(ctx, firstInvite.TokenHash, grandma.ID, grandma.DisplayName, now); err != nil {
		t.Fatalf("join first active invite: %v", err)
	}
	secondInvite, err := memoryStore.CreateInvite(ctx, family.ID, "second-active-hash", "second-active-token", root.ID, "奶奶", now.Add(time.Hour))
	if err != nil {
		t.Fatalf("create second active invite: %v", err)
	}
	if _, err := memoryStore.JoinInvite(ctx, secondInvite.TokenHash, grandma.ID, grandma.DisplayName, now); err != nil {
		t.Fatalf("active member should be able to reopen invite without consuming it: %v", err)
	}
	invites, err := memoryStore.ListInvitesForFamily(ctx, family.ID)
	if err != nil {
		t.Fatalf("list invites: %v", err)
	}
	for _, invite := range invites {
		if invite.ID == secondInvite.ID && invite.Status != InviteStatusPending {
			t.Fatalf("active member repeat join should leave second invite pending, got %#v", invite)
		}
	}
}

func TestMemoryStoreCreateUploadBatch(t *testing.T) {
	ctx := context.Background()
	memoryStore := NewMemoryStore()

	root, err := memoryStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	family, err := memoryStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, root.DisplayName)
	if err != nil {
		t.Fatalf("create family: %v", err)
	}

	batch, items, err := memoryStore.CreateUploadBatch(ctx, CreateUploadBatchInput{
		FamilyID:  family.ID,
		CreatedBy: root.ID,
		Items: []CreateUploadItemInput{
			{
				OriginalType:     OriginalTypeImage,
				OriginalFilename: "baby.jpg",
				ContentType:      "image/jpeg",
				ByteSize:         12345,
				ObjectKey:        "originals/families/1/baby.jpg",
			},
		},
		Now: time.Now(),
	})
	if err != nil {
		t.Fatalf("create upload batch: %v", err)
	}
	if batch.Status != UploadBatchStatusCreated || batch.TotalCount != 1 || len(items) != 1 {
		t.Fatalf("unexpected upload batch result: batch=%#v items=%#v", batch, items)
	}
	if items[0].Status != UploadItemStatusWaiting || items[0].ObjectKey == "" {
		t.Fatalf("unexpected upload item: %#v", items[0])
	}

	_, _, err = memoryStore.CreateUploadBatch(ctx, CreateUploadBatchInput{
		FamilyID:  family.ID,
		CreatedBy: root.ID,
		Items:     []CreateUploadItemInput{},
		Now:       time.Now(),
	})
	if err != ErrAlreadyExists {
		t.Fatalf("expected duplicate active upload batch to fail, got %v", err)
	}
}

func TestMemoryStoreCompleteUploadItemPreservesPrivateOriginalMetadata(t *testing.T) {
	ctx := context.Background()
	memoryStore := NewMemoryStore()

	root, err := memoryStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	family, err := memoryStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, root.DisplayName)
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	batch, items, err := memoryStore.CreateUploadBatch(ctx, CreateUploadBatchInput{
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

	_, item, asset, err := memoryStore.CompleteUploadItem(ctx, CompleteUploadItemInput{
		FamilyID:       family.ID,
		BatchID:        batch.ID,
		ItemID:         items[0].ID,
		UploadedBy:     root.ID,
		ObjectSize:     0,
		ObjectType:     "",
		ChecksumSHA256: "abc123",
		Now:            time.Now(),
	})
	if err != nil {
		t.Fatalf("complete upload item: %v", err)
	}
	if item.MediaAssetID != asset.ID {
		t.Fatalf("expected upload item to point at media asset, item=%#v asset=%#v", item, asset)
	}

	memoryStore.mu.Lock()
	defer memoryStore.mu.Unlock()
	if len(memoryStore.mediaOriginals) != 1 {
		t.Fatalf("expected one private original record, got %#v", memoryStore.mediaOriginals)
	}
	for _, original := range memoryStore.mediaOriginals {
		if original.MediaAssetID != asset.ID ||
			original.ObjectKey != "originals/families/1/users/1/baby.jpg" ||
			original.OriginalFilename != "baby.jpg" ||
			original.ContentType != "image/jpeg" ||
			original.ByteSize != 12345 ||
			original.ChecksumSHA256 != "abc123" {
			t.Fatalf("unexpected private original metadata: %#v", original)
		}
	}
}

func TestMemoryStoreListTimelineMediaReturnsOnlyReadyActiveAssets(t *testing.T) {
	ctx := context.Background()
	memoryStore := NewMemoryStore()

	root, err := memoryStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	family, err := memoryStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, "妈妈")
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	olderUploadedAt := time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)
	newerCapturedAt := time.Date(2026, 6, 13, 8, 30, 0, 0, time.UTC)

	memoryStore.mu.Lock()
	memoryStore.mediaAssets[1] = MediaAsset{
		ID:              1,
		FamilyID:        family.ID,
		UploadedBy:      root.ID,
		MediaType:       MediaTypePhoto,
		Status:          MediaStatusActive,
		RenditionStatus: RenditionStatusReady,
		UploadedAt:      olderUploadedAt,
	}
	memoryStore.mediaAssets[2] = MediaAsset{
		ID:              2,
		FamilyID:        family.ID,
		UploadedBy:      root.ID,
		MediaType:       MediaTypePhoto,
		Status:          MediaStatusActive,
		RenditionStatus: RenditionStatusReady,
		CapturedAt:      newerCapturedAt,
		UploadedAt:      olderUploadedAt,
	}
	memoryStore.mediaAssets[3] = MediaAsset{
		ID:              3,
		FamilyID:        family.ID,
		UploadedBy:      root.ID,
		MediaType:       MediaTypePhoto,
		Status:          MediaStatusActive,
		RenditionStatus: RenditionStatusProcessing,
		UploadedAt:      newerCapturedAt,
	}
	memoryStore.mediaAssets[4] = MediaAsset{
		ID:              4,
		FamilyID:        family.ID,
		UploadedBy:      root.ID,
		MediaType:       MediaTypePhoto,
		Status:          MediaStatusActive,
		RenditionStatus: RenditionStatusReady,
		DeletedAt:       newerCapturedAt,
		UploadedAt:      newerCapturedAt,
	}
	memoryStore.mediaAssets[5] = MediaAsset{
		ID:              5,
		FamilyID:        family.ID,
		UploadedBy:      root.ID,
		MediaType:       MediaTypePhoto,
		Status:          MediaStatusActive,
		RenditionStatus: RenditionStatusFailed,
		UploadedAt:      newerCapturedAt,
	}
	memoryStore.mediaRenditions[1] = MediaRendition{ID: 1, MediaAssetID: 1, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/older-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[2] = MediaRendition{ID: 2, MediaAssetID: 1, RenditionType: RenditionTypeThumbnail, ObjectKey: "previews/older-thumb.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[3] = MediaRendition{ID: 3, MediaAssetID: 2, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/newer-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[4] = MediaRendition{ID: 4, MediaAssetID: 2, RenditionType: RenditionTypeThumbnail, ObjectKey: "previews/newer-thumb.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[5] = MediaRendition{ID: 5, MediaAssetID: 3, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/processing-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[6] = MediaRendition{ID: 6, MediaAssetID: 4, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/deleted-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[7] = MediaRendition{ID: 7, MediaAssetID: 5, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/failed-display.jpg", Status: RenditionStatusReady}
	memoryStore.mu.Unlock()

	items, err := memoryStore.ListTimelineMedia(ctx, ListTimelineMediaInput{FamilyID: family.ID, Limit: 20})
	if err != nil {
		t.Fatalf("list timeline media: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected only ready active assets, got %#v", items)
	}
	if items[0].Asset.ID != 2 || items[1].Asset.ID != 1 {
		t.Fatalf("expected captured/uploaded time descending order, got %#v", items)
	}
	if items[0].UploadedByDisplayName != "妈妈" || items[0].Display.ObjectKey != "previews/newer-display.jpg" || items[0].Thumbnail.ObjectKey != "previews/newer-thumb.jpg" {
		t.Fatalf("unexpected timeline row: %#v", items[0])
	}
}

func TestMemoryStoreListTimelineMediaAppliesMediaTypeAndMonthFilters(t *testing.T) {
	ctx := context.Background()
	memoryStore := NewMemoryStore()

	root, err := memoryStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	family, err := memoryStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, "妈妈")
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	june := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	july := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)

	memoryStore.mu.Lock()
	memoryStore.mediaAssets[1] = MediaAsset{ID: 1, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypePhoto, Status: MediaStatusActive, RenditionStatus: RenditionStatusReady, UploadedAt: june}
	memoryStore.mediaAssets[2] = MediaAsset{ID: 2, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypeVideo, Status: MediaStatusActive, RenditionStatus: RenditionStatusReady, UploadedAt: june}
	memoryStore.mediaAssets[3] = MediaAsset{ID: 3, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypeVideo, Status: MediaStatusActive, RenditionStatus: RenditionStatusReady, UploadedAt: july}
	memoryStore.mediaRenditions[1] = MediaRendition{ID: 1, MediaAssetID: 1, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/photo-june.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[2] = MediaRendition{ID: 2, MediaAssetID: 2, RenditionType: RenditionTypeDisplayVideo, ObjectKey: "previews/video-june.mp4", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[3] = MediaRendition{ID: 3, MediaAssetID: 3, RenditionType: RenditionTypeDisplayVideo, ObjectKey: "previews/video-july.mp4", Status: RenditionStatusReady}
	memoryStore.mu.Unlock()

	location, err := time.LoadLocation(DefaultFamilyTimezone)
	if err != nil {
		t.Fatalf("load default timezone: %v", err)
	}
	monthFrom := time.Date(2026, 6, 1, 0, 0, 0, 0, location)
	items, err := memoryStore.ListTimelineMedia(ctx, ListTimelineMediaInput{
		FamilyID:  family.ID,
		Limit:     20,
		MediaType: MediaTypeVideo,
		MonthFrom: monthFrom,
		MonthTo:   monthFrom.AddDate(0, 1, 0),
	})
	if err != nil {
		t.Fatalf("list filtered timeline media: %v", err)
	}
	if len(items) != 1 || items[0].Asset.ID != 2 {
		t.Fatalf("expected only June video media, got %#v", items)
	}
}

func TestMemoryStoreListTimelineMediaAppliesStableCursor(t *testing.T) {
	ctx := context.Background()
	memoryStore := NewMemoryStore()

	root, err := memoryStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	family, err := memoryStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, "妈妈")
	if err != nil {
		t.Fatalf("create family: %v", err)
	}
	firstTime := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)
	secondTime := time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)
	thirdTime := time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC)

	memoryStore.mu.Lock()
	memoryStore.mediaAssets[1] = MediaAsset{ID: 1, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypePhoto, Status: MediaStatusActive, RenditionStatus: RenditionStatusReady, UploadedAt: firstTime}
	memoryStore.mediaAssets[2] = MediaAsset{ID: 2, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypePhoto, Status: MediaStatusActive, RenditionStatus: RenditionStatusReady, UploadedAt: secondTime}
	memoryStore.mediaAssets[3] = MediaAsset{ID: 3, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypePhoto, Status: MediaStatusActive, RenditionStatus: RenditionStatusReady, UploadedAt: thirdTime}
	for id := int64(1); id <= 3; id++ {
		memoryStore.mediaRenditions[id] = MediaRendition{ID: id, MediaAssetID: id, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/display.jpg", Status: RenditionStatusReady}
	}
	memoryStore.mu.Unlock()

	firstPage, err := memoryStore.ListTimelineMedia(ctx, ListTimelineMediaInput{FamilyID: family.ID, Limit: 2})
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(firstPage) != 2 || firstPage[0].Asset.ID != 1 || firstPage[1].Asset.ID != 2 {
		t.Fatalf("unexpected first page: %#v", firstPage)
	}

	secondPage, err := memoryStore.ListTimelineMedia(ctx, ListTimelineMediaInput{
		FamilyID:  family.ID,
		Limit:     2,
		AfterTime: timelineSortTime(firstPage[1].Asset),
		AfterID:   firstPage[1].Asset.ID,
	})
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(secondPage) != 1 || secondPage[0].Asset.ID != 3 {
		t.Fatalf("expected cursor to return only older media, got %#v", secondPage)
	}
}

func TestMemoryStoreFindMediaDetailReturnsOnlyVisibleReadyAsset(t *testing.T) {
	ctx := context.Background()
	memoryStore := NewMemoryStore()

	root, err := memoryStore.CreateUser(ctx, "root", "hash", "初始管理员")
	if err != nil {
		t.Fatalf("create root user: %v", err)
	}
	family, err := memoryStore.CreateFamily(ctx, "小树之家", DefaultFamilyTimezone, root.ID, "妈妈")
	if err != nil {
		t.Fatalf("create family: %v", err)
	}

	memoryStore.mu.Lock()
	memoryStore.mediaAssets[1] = MediaAsset{ID: 1, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypePhoto, Status: MediaStatusActive, RenditionStatus: RenditionStatusReady, UploadedAt: time.Now()}
	memoryStore.mediaAssets[2] = MediaAsset{ID: 2, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypePhoto, Status: MediaStatusActive, RenditionStatus: RenditionStatusPending, UploadedAt: time.Now()}
	memoryStore.mediaAssets[3] = MediaAsset{ID: 3, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypePhoto, Status: MediaStatusDeleted, RenditionStatus: RenditionStatusReady, DeletedAt: time.Now(), UploadedAt: time.Now()}
	memoryStore.mediaAssets[4] = MediaAsset{ID: 4, FamilyID: family.ID, UploadedBy: root.ID, MediaType: MediaTypePhoto, Status: MediaStatusActive, RenditionStatus: RenditionStatusFailed, UploadedAt: time.Now()}
	memoryStore.mediaRenditions[1] = MediaRendition{ID: 1, MediaAssetID: 1, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/ready-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[2] = MediaRendition{ID: 2, MediaAssetID: 1, RenditionType: RenditionTypeThumbnail, ObjectKey: "previews/ready-thumb.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[3] = MediaRendition{ID: 3, MediaAssetID: 2, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/pending-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[4] = MediaRendition{ID: 4, MediaAssetID: 3, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/deleted-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[5] = MediaRendition{ID: 5, MediaAssetID: 4, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/failed-display.jpg", Status: RenditionStatusReady}
	memoryStore.mu.Unlock()

	detail, ok, err := memoryStore.FindMediaDetail(ctx, FindMediaDetailInput{FamilyID: family.ID, MediaID: 1})
	if err != nil {
		t.Fatalf("find visible media detail: %v", err)
	}
	if !ok || detail.Asset.ID != 1 || detail.Display.ObjectKey != "previews/ready-display.jpg" || detail.Thumbnail.ObjectKey != "previews/ready-thumb.jpg" {
		t.Fatalf("unexpected visible detail: ok=%v detail=%#v", ok, detail)
	}

	for _, mediaID := range []int64{2, 3, 4, 404} {
		if detail, ok, err := memoryStore.FindMediaDetail(ctx, FindMediaDetailInput{FamilyID: family.ID, MediaID: mediaID}); err != nil || ok {
			t.Fatalf("expected media %d to be invisible, ok=%v detail=%#v err=%v", mediaID, ok, detail, err)
		}
	}
}
