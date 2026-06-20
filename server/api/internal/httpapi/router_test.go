package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"memotree/server/api/internal/config"
	"memotree/server/api/internal/store"
	"memotree/server/internal/storage"
)

func TestHealthz(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test"})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
}

func TestAuthFamilyInviteFlow(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test", SessionCookieName: "memotree_test_session"})

	rootCookie, rootUser := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "初始管理员",
	})
	if rootUser["isSystemAdmin"] != true {
		t.Fatalf("first registered user should be system admin, got %#v", rootUser)
	}

	_, session := getJSON(t, router, rootCookie, http.StatusOK, "/auth/session")
	if session["authenticated"] != true {
		t.Fatalf("expected authenticated session, got %#v", session)
	}

	_, family := postJSON(t, router, rootCookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	familyID := int(family["id"].(float64))
	if family["role"] != "admin" {
		t.Fatalf("family creator should be admin, got %#v", family)
	}

	_, invite := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/invites", map[string]string{
		"memberDisplayName": "奶奶",
	})
	inviteID := int(invite["id"].(float64))
	token := invite["token"].(string)
	if token == "" {
		t.Fatalf("expected invite token, got %#v", invite)
	}

	_, inviteList := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/invites")
	inviteItems := inviteList["invites"].([]any)
	if len(inviteItems) != 1 {
		t.Fatalf("expected admin to see one invite, got %#v", inviteList)
	}
	firstInvite := inviteItems[0].(map[string]any)
	if firstInvite["token"] != token || firstInvite["memberDisplayName"] != "奶奶" || firstInvite["status"] != "pending" {
		t.Fatalf("unexpected invite list item: %#v", firstInvite)
	}

	_, revoked := deleteJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/invites/"+itoa(inviteID))
	if revoked["status"] != "revoked" {
		t.Fatalf("expected revoked invite response, got %#v", revoked)
	}

	memberCookie, memberUser := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "grandma",
		"password":    "secret123",
		"displayName": "奶奶账号",
	})
	if memberUser["isSystemAdmin"] != false {
		t.Fatalf("second registered user should not be system admin, got %#v", memberUser)
	}

	postJSON(t, router, memberCookie, http.StatusBadRequest, "/invites/"+token+"/join", map[string]string{})

	_, activeInvite := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/invites", map[string]string{
		"memberDisplayName": "奶奶",
	})
	activeInviteID := int(activeInvite["id"].(float64))
	activeToken := activeInvite["token"].(string)

	_, joined := postJSON(t, router, memberCookie, http.StatusOK, "/invites/"+activeToken+"/join", map[string]string{})
	if joined["familyId"] != float64(familyID) || joined["role"] != "member" || joined["displayName"] != "奶奶" {
		t.Fatalf("expected joined member response, got %#v", joined)
	}

	_, families := getJSON(t, router, memberCookie, http.StatusOK, "/families")
	items := families["families"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected member to see one family, got %#v", families)
	}

	postJSON(t, router, memberCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/invites", map[string]string{
		"memberDisplayName": "外公",
	})
	getJSON(t, router, memberCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/invites")
	deleteJSON(t, router, memberCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/invites/"+itoa(activeInviteID))
	deleteJSON(t, router, rootCookie, http.StatusConflict, "/families/"+itoa(familyID)+"/invites/"+itoa(activeInviteID))
}

func TestRouterCanUseInjectedStore(t *testing.T) {
	sharedStore := store.NewMemoryStore()
	cfg := config.Config{AppEnv: "test", SessionCookieName: "memotree_test_session"}
	firstRouter := NewRouterWithStore(cfg, sharedStore)

	cookie, _ := postJSON(t, firstRouter, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "初始管理员",
	})
	postJSON(t, firstRouter, cookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})

	secondRouter := NewRouterWithStore(cfg, sharedStore)
	_, families := getJSON(t, secondRouter, cookie, http.StatusOK, "/families")
	items := families["families"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected injected store data to be reused across routers, got %#v", families)
	}
}

func TestCurrentSessionReturnsEmptyFamilyArray(t *testing.T) {
	router := NewRouterWithStore(
		config.Config{AppEnv: "test", SessionCookieName: "memotree_test_session"},
		nilFamiliesStore{MemoryStore: store.NewMemoryStore()},
	)

	cookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "new-user",
		"password":    "secret123",
		"displayName": "新用户",
	})

	_, session := getJSON(t, router, cookie, http.StatusOK, "/auth/session")
	families, ok := session["families"].([]any)
	if !ok {
		t.Fatalf("expected families to be a JSON array, got %#v", session["families"])
	}
	if len(families) != 0 {
		t.Fatalf("expected no families for a newly registered user, got %#v", families)
	}
}

func TestCreateUploadIntentRequiresActiveMembership(t *testing.T) {
	sharedStore := store.NewMemoryStore()
	router := NewRouterWithDependencies(
		config.Config{
			AppEnv:              "test",
			SessionCookieName:   "memotree_test_session",
			OriginalsBucket:     "memotree-originals",
			PreviewsBucket:      "memotree-previews",
			SignedURLTTL:        15 * time.Minute,
			UploadMaxFileBytes:  50 * 1024 * 1024,
			UploadMaxBatchCount: 10,
		},
		sharedStore,
		fakeStorageService{},
	)

	rootCookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "初始管理员",
	})
	_, family := postJSON(t, router, rootCookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	familyID := int(family["id"].(float64))

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	batch := uploadIntent["batch"].(map[string]any)
	if batch["status"] != store.UploadBatchStatusCreated || batch["totalCount"] != float64(1) {
		t.Fatalf("unexpected upload batch response: %#v", batch)
	}
	items := uploadIntent["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected one upload item, got %#v", uploadIntent)
	}
	item := items[0].(map[string]any)
	if item["uploadUrl"] == "" || item["method"] != http.MethodPut || item["contentType"] != "image/jpeg" {
		t.Fatalf("unexpected upload item response: %#v", item)
	}

	memberCookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "guest",
		"password":    "secret123",
		"displayName": "访客",
	})
	postJSON(t, router, memberCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "guest.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
}

func TestCreateUploadIntentValidatesRequestBoundaries(t *testing.T) {
	t.Run("requires login before membership lookup", func(t *testing.T) {
		router := NewRouterWithDependencies(
			config.Config{AppEnv: "test", SessionCookieName: "memotree_test_session"},
			store.NewMemoryStore(),
			fakeStorageService{},
		)

		postJSON(t, router, nil, http.StatusUnauthorized, "/families/1/media/upload-intents", map[string]any{
			"files": []map[string]any{
				{"filename": "baby.jpg", "contentType": "image/jpeg", "byteSize": 12345},
			},
		})
	})

	t.Run("requires configured object storage for new uploads", func(t *testing.T) {
		router := NewRouterWithStore(config.Config{AppEnv: "test", SessionCookieName: "memotree_test_session"}, store.NewMemoryStore())
		rootCookie, familyID := createSignedInFamily(t, router, "root")

		postJSON(t, router, rootCookie, http.StatusServiceUnavailable, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
			"files": []map[string]any{
				{"filename": "baby.jpg", "contentType": "image/jpeg", "byteSize": 12345},
			},
		})
	})

	t.Run("rejects unsupported media type", func(t *testing.T) {
		router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
		rootCookie, familyID := createSignedInFamily(t, router, "root")

		postJSON(t, router, rootCookie, http.StatusBadRequest, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
			"files": []map[string]any{
				{"filename": "notes.txt", "contentType": "text/plain", "byteSize": 12345},
			},
		})
	})

	t.Run("rejects batch count and file size limits", func(t *testing.T) {
		router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
		rootCookie, familyID := createSignedInFamily(t, router, "root")

		postJSON(t, router, rootCookie, http.StatusBadRequest, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
			"files": []map[string]any{
				{"filename": "1.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "2.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "3.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "4.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "5.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "6.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "7.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "8.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "9.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "10.jpg", "contentType": "image/jpeg", "byteSize": 12345},
				{"filename": "11.jpg", "contentType": "image/jpeg", "byteSize": 12345},
			},
		})

		postJSON(t, router, rootCookie, http.StatusBadRequest, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
			"files": []map[string]any{
				{"filename": "too-large.jpg", "contentType": "image/jpeg", "byteSize": 51 * 1024 * 1024},
			},
		})
	})
}

func TestCreateUploadIntentResponseDoesNotExposeObjectKeyField(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	assertNoObjectKeyField(t, uploadIntent)
}

func TestCreateUploadIntentReturnsActiveBatchWhenOneAlreadyExists(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, firstIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	firstBatch := firstIntent["batch"].(map[string]any)

	_, secondIntent := postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "another.jpg",
				"contentType": "image/jpeg",
				"byteSize":    67890,
			},
		},
	})
	secondBatch := secondIntent["batch"].(map[string]any)
	if secondIntent["activeExisting"] != true {
		t.Fatalf("expected duplicate upload intent to point at existing active batch, got %#v", secondIntent)
	}
	if secondBatch["id"] != firstBatch["id"] || secondBatch["totalCount"] != firstBatch["totalCount"] {
		t.Fatalf("expected existing batch response, first=%#v second=%#v", firstBatch, secondBatch)
	}
	if items := secondIntent["items"].([]any); len(items) != 0 {
		t.Fatalf("expected no new upload URLs when returning existing active batch, got %#v", items)
	}
}

func TestUploadTaskQueriesReturnActiveBatchAndItems(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	createdBatch := uploadIntent["batch"].(map[string]any)
	batchID := int(createdBatch["id"].(float64))

	_, active := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/active")
	activeBatch := active["batch"].(map[string]any)
	if activeBatch["id"] != createdBatch["id"] || activeBatch["status"] != store.UploadBatchStatusCreated {
		t.Fatalf("expected active batch response, got %#v", active)
	}
	activeItems := active["items"].([]any)
	if len(activeItems) != 1 {
		t.Fatalf("expected one active upload item, got %#v", active)
	}
	firstItem := activeItems[0].(map[string]any)
	if firstItem["originalFilename"] != "baby.jpg" || firstItem["status"] != store.UploadItemStatusWaiting {
		t.Fatalf("unexpected active upload item: %#v", firstItem)
	}
	assertNoObjectKeyField(t, active)

	_, detail := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID))
	detailBatch := detail["batch"].(map[string]any)
	if detailBatch["id"] != createdBatch["id"] {
		t.Fatalf("expected upload detail for batch %d, got %#v", batchID, detail)
	}
	assertNoObjectKeyField(t, detail)
}

func TestUploadTaskQueriesEnforceOwnerAndAdminAccess(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	adminCookie, familyID := createSignedInFamily(t, router, "root")
	memberCookie := joinFamilyWithInvite(t, router, adminCookie, familyID, "grandma")
	guestCookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "guest",
		"password":    "secret123",
		"displayName": "访客",
	})

	_, memberIntent := postJSON(t, router, memberCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "grandma.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	memberBatch := memberIntent["batch"].(map[string]any)
	memberBatchID := int(memberBatch["id"].(float64))

	getJSON(t, router, memberCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(memberBatchID))
	getJSON(t, router, adminCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(memberBatchID))
	getJSON(t, router, guestCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/uploads/active")
	getJSON(t, router, guestCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/uploads/"+itoa(memberBatchID))
}

func TestUploadTaskActiveQueryReturnsEmptyWhenNoActiveBatch(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, active := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/active")
	if active["batch"] != nil {
		t.Fatalf("expected no active batch, got %#v", active)
	}
	items := active["items"].([]any)
	if len(items) != 0 {
		t.Fatalf("expected empty item array, got %#v", active)
	}
}

func TestStopUploadTaskCancelsUnfinishedItemsAndClearsActiveSlot(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	batch := uploadIntent["batch"].(map[string]any)
	batchID := int(batch["id"].(float64))

	_, stopped := postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/stop", map[string]any{})
	stoppedBatch := stopped["batch"].(map[string]any)
	if stoppedBatch["status"] != store.UploadBatchStatusStopped || stoppedBatch["cancelledCount"] != float64(1) {
		t.Fatalf("expected stopped batch with cancelled item, got %#v", stopped)
	}
	items := stopped["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["status"] != store.UploadItemStatusCancelled {
		t.Fatalf("expected cancelled upload item, got %#v", stopped)
	}

	_, active := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/active")
	if active["batch"] != nil {
		t.Fatalf("expected stopped task to leave no active batch, got %#v", active)
	}
}

func TestRecentUploadTasksReturnStoppedTaskAfterActiveSlotCleared(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	batch := uploadIntent["batch"].(map[string]any)
	batchID := int(batch["id"].(float64))

	postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/stop", map[string]any{})

	_, active := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/active")
	if active["batch"] != nil {
		t.Fatalf("expected stopped task to be absent from active query, got %#v", active)
	}

	_, recent := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/recent")
	tasks := recent["tasks"].([]any)
	if len(tasks) != 1 {
		t.Fatalf("expected stopped task in recent uploads, got %#v", recent)
	}
	recentTask := tasks[0].(map[string]any)
	recentBatch := recentTask["batch"].(map[string]any)
	if recentBatch["id"] != float64(batchID) || recentBatch["status"] != store.UploadBatchStatusStopped {
		t.Fatalf("expected stopped batch in recent uploads, got %#v", recentTask)
	}
	items := recentTask["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["status"] != store.UploadItemStatusCancelled {
		t.Fatalf("expected recent task to include cancelled item, got %#v", recentTask)
	}
	assertNoObjectKeyField(t, recent)
}

func TestRecentUploadTasksEnforceMemberAndAdminVisibility(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	adminCookie, familyID := createSignedInFamily(t, router, "root")
	memberCookie := joinFamilyWithInvite(t, router, adminCookie, familyID, "grandma")
	otherMemberCookie := joinFamilyWithInvite(t, router, adminCookie, familyID, "grandpa")
	guestCookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "guest",
		"password":    "secret123",
		"displayName": "guest",
	})

	_, memberIntent := postJSON(t, router, memberCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "grandma.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	memberBatchID := int(memberIntent["batch"].(map[string]any)["id"].(float64))

	_, adminRecent := getJSON(t, router, adminCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/recent")
	adminTasks := adminRecent["tasks"].([]any)
	if len(adminTasks) != 1 || adminTasks[0].(map[string]any)["batch"].(map[string]any)["id"] != float64(memberBatchID) {
		t.Fatalf("expected admin to see member upload task, got %#v", adminRecent)
	}

	_, otherRecent := getJSON(t, router, otherMemberCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/recent")
	if tasks := otherRecent["tasks"].([]any); len(tasks) != 0 {
		t.Fatalf("expected member to see only own upload tasks, got %#v", otherRecent)
	}
	getJSON(t, router, guestCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/uploads/recent")
}

func TestTimelineReturnsReadyMediaGroupsWithSignedPreviewURLs(t *testing.T) {
	appStore := &timelineStore{MemoryStore: store.NewMemoryStore()}
	router := newUploadTestRouter(appStore, fakeStorageService{})

	rootCookie, rootUser := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "妈妈账号",
	})
	_, family := postJSON(t, router, rootCookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	familyID := int64(family["id"].(float64))
	rootUserID := int64(rootUser["id"].(float64))
	appStore.rows = []store.TimelineMedia{
		{
			Asset: store.MediaAsset{
				ID:              7,
				FamilyID:        familyID,
				UploadedBy:      rootUserID,
				MediaType:       store.MediaTypePhoto,
				Status:          store.MediaStatusActive,
				RenditionStatus: store.RenditionStatusReady,
				CapturedAt:      time.Date(2026, 6, 13, 9, 30, 0, 0, time.UTC),
				UploadedAt:      time.Date(2026, 6, 13, 9, 35, 0, 0, time.UTC),
			},
			UploadedByDisplayName: "妈妈",
			Thumbnail: store.MediaRendition{
				RenditionType: store.RenditionTypeThumbnail,
				ObjectKey:     "previews/families/1/thumb.jpg",
				ContentType:   "image/jpeg",
				Width:         360,
				Height:        240,
				Status:        store.RenditionStatusReady,
			},
			Display: store.MediaRendition{
				RenditionType: store.RenditionTypeDisplayImage,
				ObjectKey:     "previews/families/1/display.jpg",
				ContentType:   "image/jpeg",
				Width:         1600,
				Height:        1067,
				Status:        store.RenditionStatusReady,
			},
		},
	}

	_, timeline := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(int(familyID))+"/timeline")
	groups := timeline["groups"].([]any)
	if len(groups) != 1 {
		t.Fatalf("expected one timeline group, got %#v", timeline)
	}
	group := groups[0].(map[string]any)
	if group["date"] != "2026-06-13" || group["month"] != "2026-06" {
		t.Fatalf("unexpected timeline group key: %#v", group)
	}
	items := group["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected one timeline item, got %#v", group)
	}
	item := items[0].(map[string]any)
	display := item["display"].(map[string]any)
	thumbnail := item["thumbnail"].(map[string]any)
	if item["id"] != float64(7) || display["url"] != "https://storage.example/download/previews/families/1/display.jpg" || thumbnail["url"] != "https://storage.example/download/previews/families/1/thumb.jpg" {
		t.Fatalf("unexpected timeline item response: %#v", item)
	}
	assertNoObjectKeyField(t, timeline)

	guestCookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "guest",
		"password":    "secret123",
		"displayName": "访客",
	})
	getJSON(t, router, guestCookie, http.StatusForbidden, "/families/"+itoa(int(familyID))+"/timeline")
}

func TestTimelineSupportsStableCursorPagination(t *testing.T) {
	appStore := &timelineStore{MemoryStore: store.NewMemoryStore()}
	router := newUploadTestRouter(appStore, fakeStorageService{})

	rootCookie, rootUser := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "妈妈账号",
	})
	_, family := postJSON(t, router, rootCookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	familyID := int64(family["id"].(float64))
	rootUserID := int64(rootUser["id"].(float64))
	appStore.rows = []store.TimelineMedia{
		timelineRow(1, familyID, rootUserID, time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)),
		timelineRow(2, familyID, rootUserID, time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC)),
		timelineRow(3, familyID, rootUserID, time.Date(2026, 6, 11, 9, 0, 0, 0, time.UTC)),
	}

	_, firstPage := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(int(familyID))+"/timeline?limit=2")
	if ids := collectTimelineItemIDs(firstPage); len(ids) != 2 || ids[0] != 1 || ids[1] != 2 {
		t.Fatalf("unexpected first timeline page: %#v", firstPage)
	}
	nextCursor, ok := firstPage["nextCursor"].(string)
	if !ok || nextCursor == "" {
		t.Fatalf("expected first page to include nextCursor, got %#v", firstPage)
	}

	_, secondPage := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(int(familyID))+"/timeline?limit=2&cursor="+nextCursor)
	if ids := collectTimelineItemIDs(secondPage); len(ids) != 1 || ids[0] != 3 {
		t.Fatalf("expected cursor page to include only older media, got %#v", secondPage)
	}
	if secondPage["nextCursor"] != nil {
		t.Fatalf("expected final page to omit next cursor, got %#v", secondPage)
	}
}

func TestTimelineRejectsInvalidCursor(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	getJSON(t, router, rootCookie, http.StatusBadRequest, "/families/"+itoa(familyID)+"/timeline?cursor=not-a-cursor")
}

func TestTimelineSupportsMediaTypeAndMonthFilters(t *testing.T) {
	appStore := &timelineStore{MemoryStore: store.NewMemoryStore()}
	router := newUploadTestRouter(appStore, fakeStorageService{})

	rootCookie, rootUser := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "妈妈账号",
	})
	_, family := postJSON(t, router, rootCookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	familyID := int64(family["id"].(float64))
	rootUserID := int64(rootUser["id"].(float64))
	junePhoto := timelineRow(1, familyID, rootUserID, time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC))
	juneVideo := timelineRow(2, familyID, rootUserID, time.Date(2026, 6, 12, 9, 0, 0, 0, time.UTC))
	juneVideo.Asset.MediaType = store.MediaTypeVideo
	juneVideo.Display.RenditionType = store.RenditionTypeDisplayVideo
	julyVideo := timelineRow(3, familyID, rootUserID, time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
	julyVideo.Asset.MediaType = store.MediaTypeVideo
	julyVideo.Display.RenditionType = store.RenditionTypeDisplayVideo
	appStore.rows = []store.TimelineMedia{junePhoto, juneVideo, julyVideo}

	_, filtered := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(int(familyID))+"/timeline?mediaType=video&month=2026-06")
	if ids := collectTimelineItemIDs(filtered); len(ids) != 1 || ids[0] != 2 {
		t.Fatalf("expected filters to return only June video media, got %#v", filtered)
	}

	getJSON(t, router, rootCookie, http.StatusBadRequest, "/families/"+itoa(int(familyID))+"/timeline?mediaType=audio")
	getJSON(t, router, rootCookie, http.StatusBadRequest, "/families/"+itoa(int(familyID))+"/timeline?month=2026-13")
}

func TestOriginalDownloadEndpointIsDeferred(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	request := httptest.NewRequest(http.MethodPost, "/families/"+itoa(familyID)+"/media/1/download", nil)
	request.AddCookie(rootCookie)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected original download endpoint to be deferred with 404, got %d body %s", response.Code, response.Body.String())
	}
}

func TestMediaDetailReturnsReadyMediaWithSignedPreviewURLs(t *testing.T) {
	appStore := &mediaDetailStore{MemoryStore: store.NewMemoryStore(), visible: true}
	router := newUploadTestRouter(appStore, fakeStorageService{})

	rootCookie, rootUser := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "妈妈账号",
	})
	_, family := postJSON(t, router, rootCookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	familyID := int64(family["id"].(float64))
	rootUserID := int64(rootUser["id"].(float64))
	appStore.detail = store.MediaDetail(timelineRow(7, familyID, rootUserID, time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)))

	_, detail := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(int(familyID))+"/media/7")
	media := detail["media"].(map[string]any)
	display := media["display"].(map[string]any)
	thumbnail := media["thumbnail"].(map[string]any)
	if media["id"] != float64(7) || display["url"] != "https://storage.example/download/previews/families/1/7-display.jpg" || thumbnail["url"] != "https://storage.example/download/previews/families/1/7-thumb.jpg" {
		t.Fatalf("unexpected media detail response: %#v", detail)
	}
	assertNoObjectKeyField(t, detail)
}

func TestMediaDetailEnforcesVisibility(t *testing.T) {
	appStore := &mediaDetailStore{MemoryStore: store.NewMemoryStore(), visible: true}
	router := newUploadTestRouter(appStore, fakeStorageService{})

	rootCookie, rootUser := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "妈妈账号",
	})
	_, family := postJSON(t, router, rootCookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	familyID := int64(family["id"].(float64))
	rootUserID := int64(rootUser["id"].(float64))
	appStore.detail = store.MediaDetail(timelineRow(7, familyID, rootUserID, time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC)))

	guestCookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "guest",
		"password":    "secret123",
		"displayName": "访客",
	})
	getJSON(t, router, guestCookie, http.StatusForbidden, "/families/"+itoa(int(familyID))+"/media/7")

	appStore.visible = false
	getJSON(t, router, rootCookie, http.StatusNotFound, "/families/"+itoa(int(familyID))+"/media/7")
}

func TestRemovedMemberCannotAccessTimelineOrMediaDetail(t *testing.T) {
	appStore := &removedMemberAccessStore{MemoryStore: store.NewMemoryStore()}
	router := newUploadTestRouter(appStore, fakeStorageService{})
	adminCookie, adminUser := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "root",
		"password":    "secret123",
		"displayName": "妈妈账号",
	})
	_, family := postJSON(t, router, adminCookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	familyID := int(family["id"].(float64))
	adminUserID := int64(adminUser["id"].(float64))
	memberCookie := joinFamilyWithInvite(t, router, adminCookie, familyID, "removed-grandma")
	_, memberSession := getJSON(t, router, memberCookie, http.StatusOK, "/auth/session")
	memberUser := memberSession["user"].(map[string]any)

	appStore.removedUserID = int64(memberUser["id"].(float64))
	appStore.row = timelineRow(7, int64(familyID), adminUserID, time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC))

	getJSON(t, router, memberCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/timeline")
	getJSON(t, router, memberCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/media/7")
}

func TestTimelineLogsInternalStoreError(t *testing.T) {
	var logs bytes.Buffer
	previousOutput := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() {
		log.SetOutput(previousOutput)
	})

	appStore := &timelineErrorStore{MemoryStore: store.NewMemoryStore()}
	router := newUploadTestRouter(appStore, fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	getJSON(t, router, rootCookie, http.StatusInternalServerError, "/families/"+itoa(familyID)+"/timeline")

	logText := logs.String()
	if !strings.Contains(logText, "读取时间线失败") || !strings.Contains(logText, "timeline exploded") || !strings.Contains(logText, "GET /families/1/timeline") {
		t.Fatalf("expected internal timeline error to be logged with path and cause, got %q", logText)
	}
	if !strings.Contains(logText, "goroutine") {
		t.Fatalf("expected internal error log to include stack trace, got %q", logText)
	}
}

func TestStopUploadTaskEnforcesOwnerAndAdminAccess(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	adminCookie, familyID := createSignedInFamily(t, router, "root")
	memberCookie := joinFamilyWithInvite(t, router, adminCookie, familyID, "grandma")
	otherMemberCookie := joinFamilyWithInvite(t, router, adminCookie, familyID, "grandpa")
	guestCookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   "guest",
		"password":    "secret123",
		"displayName": "访客",
	})

	_, memberIntent := postJSON(t, router, memberCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "grandma.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	memberBatch := memberIntent["batch"].(map[string]any)
	memberBatchID := int(memberBatch["id"].(float64))

	postJSON(t, router, otherMemberCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/uploads/"+itoa(memberBatchID)+"/stop", map[string]any{})
	postJSON(t, router, guestCookie, http.StatusForbidden, "/families/"+itoa(familyID)+"/uploads/"+itoa(memberBatchID)+"/stop", map[string]any{})
	postJSON(t, router, adminCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(memberBatchID)+"/stop", map[string]any{})
}

func TestCompleteUploadItemCreatesMediaAndMovesItemToProcessing(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	batch := uploadIntent["batch"].(map[string]any)
	items := uploadIntent["items"].([]any)
	batchID := int(batch["id"].(float64))
	itemID := int(items[0].(map[string]any)["id"].(float64))

	_, completed := postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/items/"+itoa(itemID)+"/complete-upload", map[string]any{})
	completedBatch := completed["batch"].(map[string]any)
	if completedBatch["status"] != store.UploadBatchStatusProcessing {
		t.Fatalf("expected batch to move to processing, got %#v", completed)
	}
	completedItem := completed["item"].(map[string]any)
	if completedItem["status"] != store.UploadItemStatusProcessing || completedItem["mediaAssetId"] == nil {
		t.Fatalf("expected completed item to have processing status and mediaAssetId, got %#v", completed)
	}
	mediaAsset := completed["mediaAsset"].(map[string]any)
	if mediaAsset["mediaType"] != store.MediaTypePhoto || mediaAsset["renditionStatus"] != store.RenditionStatusPending {
		t.Fatalf("expected pending photo media asset, got %#v", completed)
	}
	assertNoObjectKeyField(t, completed)

	_, detail := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID))
	detailItems := detail["items"].([]any)
	if detailItems[0].(map[string]any)["mediaAssetId"] == nil {
		t.Fatalf("expected upload task detail to include mediaAssetId after completion, got %#v", detail)
	}
}

func TestCompleteUploadItemFailsWhenObjectIsMissing(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), missingObjectStorage{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	batch := uploadIntent["batch"].(map[string]any)
	items := uploadIntent["items"].([]any)
	batchID := int(batch["id"].(float64))
	itemID := int(items[0].(map[string]any)["id"].(float64))

	postJSON(t, router, rootCookie, http.StatusConflict, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/items/"+itoa(itemID)+"/complete-upload", map[string]any{})
}

func TestFailAndRetryUploadItem(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	batch := uploadIntent["batch"].(map[string]any)
	items := uploadIntent["items"].([]any)
	batchID := int(batch["id"].(float64))
	itemID := int(items[0].(map[string]any)["id"].(float64))

	_, failed := postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/items/"+itoa(itemID)+"/fail-upload", map[string]string{
		"errorMessage": "network interrupted",
	})
	failedBatch := failed["batch"].(map[string]any)
	if failedBatch["status"] != store.UploadBatchStatusPartiallyFailed || failedBatch["failedCount"] != float64(1) {
		t.Fatalf("expected partially failed batch, got %#v", failed)
	}
	failedItem := failed["item"].(map[string]any)
	if failedItem["status"] != store.UploadItemStatusUploadFailed || failedItem["errorMessage"] != "network interrupted" {
		t.Fatalf("expected upload_failed item, got %#v", failed)
	}

	_, retry := postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/items/"+itoa(itemID)+"/retry-upload", map[string]any{})
	retryBatch := retry["batch"].(map[string]any)
	if retryBatch["failedCount"] != float64(0) || retryBatch["status"] != store.UploadBatchStatusCreated {
		t.Fatalf("expected retry to clear failure count and reset batch, got %#v", retry)
	}
	retryItem := retry["item"].(map[string]any)
	if retryItem["status"] != store.UploadItemStatusWaiting || retryItem["errorMessage"] != "" || retryItem["uploadUrl"] == "" {
		t.Fatalf("expected retry response with fresh uploadUrl and waiting item, got %#v", retry)
	}
}

func TestBatchKeepsPartialFailureWhenAnotherItemCompletes(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "failed.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
			{
				"filename":    "completed.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	batch := uploadIntent["batch"].(map[string]any)
	items := uploadIntent["items"].([]any)
	batchID := int(batch["id"].(float64))
	failedItemID := int(items[0].(map[string]any)["id"].(float64))
	completedItemID := int(items[1].(map[string]any)["id"].(float64))

	postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/items/"+itoa(failedItemID)+"/fail-upload", map[string]string{
		"errorMessage": "network interrupted",
	})
	_, completed := postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/items/"+itoa(completedItemID)+"/complete-upload", map[string]any{})

	completedBatch := completed["batch"].(map[string]any)
	if completedBatch["status"] != store.UploadBatchStatusPartiallyFailed || completedBatch["failedCount"] != float64(1) {
		t.Fatalf("expected batch to preserve partial failure after another item completes, got %#v", completed)
	}

	_, detail := getJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID))
	detailItems := detail["items"].([]any)
	statuses := map[string]bool{}
	for _, rawItem := range detailItems {
		statuses[rawItem.(map[string]any)["status"].(string)] = true
	}
	if !statuses[store.UploadItemStatusUploadFailed] || !statuses[store.UploadItemStatusProcessing] {
		t.Fatalf("expected failed and processing items to remain visible, got %#v", detail)
	}
}

func TestRetryUploadItemRejectsAlreadyProcessingItem(t *testing.T) {
	router := newUploadTestRouter(store.NewMemoryStore(), fakeStorageService{})
	rootCookie, familyID := createSignedInFamily(t, router, "root")

	_, uploadIntent := postJSON(t, router, rootCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/media/upload-intents", map[string]any{
		"files": []map[string]any{
			{
				"filename":    "baby.jpg",
				"contentType": "image/jpeg",
				"byteSize":    12345,
			},
		},
	})
	batch := uploadIntent["batch"].(map[string]any)
	items := uploadIntent["items"].([]any)
	batchID := int(batch["id"].(float64))
	itemID := int(items[0].(map[string]any)["id"].(float64))

	postJSON(t, router, rootCookie, http.StatusOK, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/items/"+itoa(itemID)+"/complete-upload", map[string]any{})
	postJSON(t, router, rootCookie, http.StatusConflict, "/families/"+itoa(familyID)+"/uploads/"+itoa(batchID)+"/items/"+itoa(itemID)+"/retry-upload", map[string]any{})
}

type nilFamiliesStore struct {
	*store.MemoryStore
}

func (s nilFamiliesStore) ListFamiliesForUser(_ context.Context, _ int64) ([]store.FamilySummary, error) {
	return nil, nil
}

type timelineStore struct {
	*store.MemoryStore
	rows []store.TimelineMedia
}

func (s *timelineStore) ListTimelineMedia(_ context.Context, input store.ListTimelineMediaInput) ([]store.TimelineMedia, error) {
	result := []store.TimelineMedia{}
	for _, row := range s.rows {
		if row.Asset.FamilyID == input.FamilyID {
			sortTime := timelineMediaTime(row.Asset)
			if input.MediaType != "" && row.Asset.MediaType != input.MediaType {
				continue
			}
			if !input.MonthFrom.IsZero() && (sortTime.Before(input.MonthFrom) || !sortTime.Before(input.MonthTo)) {
				continue
			}
			if !input.AfterTime.IsZero() {
				if sortTime.After(input.AfterTime) || sortTime.Equal(input.AfterTime) && row.Asset.ID >= input.AfterID {
					continue
				}
			}
			result = append(result, row)
		}
	}
	if input.Limit > 0 && len(result) > input.Limit {
		result = result[:input.Limit]
	}
	return result, nil
}

type mediaDetailStore struct {
	*store.MemoryStore
	detail  store.MediaDetail
	visible bool
}

func (s *mediaDetailStore) FindMediaDetail(_ context.Context, input store.FindMediaDetailInput) (store.MediaDetail, bool, error) {
	if !s.visible {
		return store.MediaDetail{}, false, nil
	}
	if s.detail.Asset.FamilyID != input.FamilyID || s.detail.Asset.ID != input.MediaID {
		return store.MediaDetail{}, false, nil
	}
	return s.detail, true, nil
}

type removedMemberAccessStore struct {
	*store.MemoryStore
	removedUserID int64
	row           store.TimelineMedia
}

func (s *removedMemberAccessStore) IsActiveMember(ctx context.Context, familyID int64, userID int64) (bool, error) {
	if userID == s.removedUserID {
		return false, nil
	}
	return s.MemoryStore.IsActiveMember(ctx, familyID, userID)
}

func (s *removedMemberAccessStore) ListTimelineMedia(_ context.Context, input store.ListTimelineMediaInput) ([]store.TimelineMedia, error) {
	if s.row.Asset.FamilyID != input.FamilyID {
		return []store.TimelineMedia{}, nil
	}
	return []store.TimelineMedia{s.row}, nil
}

func (s *removedMemberAccessStore) FindMediaDetail(_ context.Context, input store.FindMediaDetailInput) (store.MediaDetail, bool, error) {
	if s.row.Asset.FamilyID != input.FamilyID || s.row.Asset.ID != input.MediaID {
		return store.MediaDetail{}, false, nil
	}
	return store.MediaDetail(s.row), true, nil
}

func timelineRow(id int64, familyID int64, uploadedBy int64, uploadedAt time.Time) store.TimelineMedia {
	return store.TimelineMedia{
		Asset: store.MediaAsset{
			ID:              id,
			FamilyID:        familyID,
			UploadedBy:      uploadedBy,
			MediaType:       store.MediaTypePhoto,
			Status:          store.MediaStatusActive,
			RenditionStatus: store.RenditionStatusReady,
			UploadedAt:      uploadedAt,
		},
		UploadedByDisplayName: "妈妈",
		Thumbnail: store.MediaRendition{
			RenditionType: store.RenditionTypeThumbnail,
			ObjectKey:     fmt.Sprintf("previews/families/1/%d-thumb.jpg", id),
			ContentType:   "image/jpeg",
			Status:        store.RenditionStatusReady,
		},
		Display: store.MediaRendition{
			RenditionType: store.RenditionTypeDisplayImage,
			ObjectKey:     fmt.Sprintf("previews/families/1/%d-display.jpg", id),
			ContentType:   "image/jpeg",
			Status:        store.RenditionStatusReady,
		},
	}
}

func collectTimelineItemIDs(response map[string]any) []int {
	ids := []int{}
	for _, rawGroup := range response["groups"].([]any) {
		group := rawGroup.(map[string]any)
		for _, rawItem := range group["items"].([]any) {
			item := rawItem.(map[string]any)
			ids = append(ids, int(item["id"].(float64)))
		}
	}
	return ids
}

type timelineErrorStore struct {
	*store.MemoryStore
}

func (s *timelineErrorStore) ListTimelineMedia(_ context.Context, _ store.ListTimelineMediaInput) ([]store.TimelineMedia, error) {
	return nil, errors.New("timeline exploded")
}

func postJSON(t *testing.T, router http.Handler, cookie *http.Cookie, expectedStatus int, path string, payload any) (*http.Cookie, map[string]any) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != expectedStatus {
		t.Fatalf("POST %s expected status %d, got %d with body %s", path, expectedStatus, response.Code, response.Body.String())
	}

	var decoded map[string]any
	if response.Body.Len() > 0 {
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("decode response body: %v", err)
		}
	}

	for _, responseCookie := range response.Result().Cookies() {
		if responseCookie.Name == "memotree_test_session" {
			return responseCookie, decoded
		}
	}
	return cookie, decoded
}

func getJSON(t *testing.T, router http.Handler, cookie *http.Cookie, expectedStatus int, path string) (*http.Cookie, map[string]any) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != expectedStatus {
		t.Fatalf("GET %s expected status %d, got %d with body %s", path, expectedStatus, response.Code, response.Body.String())
	}

	var decoded map[string]any
	if response.Body.Len() > 0 {
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("decode response body: %v", err)
		}
	}
	return cookie, decoded
}

func deleteJSON(t *testing.T, router http.Handler, cookie *http.Cookie, expectedStatus int, path string) (*http.Cookie, map[string]any) {
	t.Helper()
	request := httptest.NewRequest(http.MethodDelete, path, nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != expectedStatus {
		t.Fatalf("DELETE %s expected status %d, got %d with body %s", path, expectedStatus, response.Code, response.Body.String())
	}

	var decoded map[string]any
	if response.Body.Len() > 0 {
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("decode response body: %v", err)
		}
	}
	return cookie, decoded
}

func newUploadTestRouter(appStore store.Store, storageService storage.Service) http.Handler {
	return NewRouterWithDependencies(
		config.Config{
			AppEnv:              "test",
			SessionCookieName:   "memotree_test_session",
			OriginalsBucket:     "memotree-originals",
			PreviewsBucket:      "memotree-previews",
			SignedURLTTL:        15 * time.Minute,
			UploadMaxFileBytes:  50 * 1024 * 1024,
			UploadMaxBatchCount: 10,
		},
		appStore,
		storageService,
	)
}

func createSignedInFamily(t *testing.T, router http.Handler, loginName string) (*http.Cookie, int) {
	t.Helper()
	cookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   loginName,
		"password":    "secret123",
		"displayName": "初始管理员",
	})
	_, family := postJSON(t, router, cookie, http.StatusCreated, "/families", map[string]string{
		"displayName": "小树之家",
	})
	return cookie, int(family["id"].(float64))
}

func joinFamilyWithInvite(t *testing.T, router http.Handler, adminCookie *http.Cookie, familyID int, loginName string) *http.Cookie {
	t.Helper()
	_, invite := postJSON(t, router, adminCookie, http.StatusCreated, "/families/"+itoa(familyID)+"/invites", map[string]string{
		"memberDisplayName": loginName,
	})
	token := invite["token"].(string)
	memberCookie, _ := postJSON(t, router, nil, http.StatusCreated, "/auth/register", map[string]string{
		"loginName":   loginName,
		"password":    "secret123",
		"displayName": loginName,
	})
	postJSON(t, router, memberCookie, http.StatusOK, "/invites/"+token+"/join", map[string]string{})
	return memberCookie
}

func assertNoObjectKeyField(t *testing.T, value any) {
	t.Helper()
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if key == "objectKey" {
				t.Fatalf("response should not expose objectKey field: %#v", value)
			}
			assertNoObjectKeyField(t, child)
		}
	case []any:
		for _, child := range typed {
			assertNoObjectKeyField(t, child)
		}
	}
}

func itoa(value int) string {
	return strconv.Itoa(value)
}

type fakeStorageService struct{}

func (fakeStorageService) GetSignedUploadURL(_ context.Context, request storage.SignedURLRequest) (string, error) {
	return "https://storage.example/upload/" + request.ObjectKey, nil
}

func (fakeStorageService) GetSignedDownloadURL(_ context.Context, request storage.SignedURLRequest) (string, error) {
	return "https://storage.example/download/" + request.ObjectKey, nil
}

func (fakeStorageService) HeadObject(_ context.Context, bucket string, objectKey string) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{Bucket: bucket, ObjectKey: objectKey, ContentType: "image/jpeg", SizeBytes: 12345}, nil
}

func (fakeStorageService) DeleteObject(_ context.Context, _ string, _ string) error {
	return nil
}

type missingObjectStorage struct {
	fakeStorageService
}

func (missingObjectStorage) HeadObject(_ context.Context, _ string, objectKey string) (storage.ObjectInfo, error) {
	return storage.ObjectInfo{}, fmt.Errorf("object %s not found", objectKey)
}
