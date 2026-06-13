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
	memoryStore.mediaRenditions[1] = MediaRendition{ID: 1, MediaAssetID: 1, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/older-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[2] = MediaRendition{ID: 2, MediaAssetID: 1, RenditionType: RenditionTypeThumbnail, ObjectKey: "previews/older-thumb.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[3] = MediaRendition{ID: 3, MediaAssetID: 2, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/newer-display.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[4] = MediaRendition{ID: 4, MediaAssetID: 2, RenditionType: RenditionTypeThumbnail, ObjectKey: "previews/newer-thumb.jpg", Status: RenditionStatusReady}
	memoryStore.mediaRenditions[5] = MediaRendition{ID: 5, MediaAssetID: 3, RenditionType: RenditionTypeDisplayImage, ObjectKey: "previews/processing-display.jpg", Status: RenditionStatusReady}
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
