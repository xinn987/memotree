package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"memotree/server/api/internal/config"
	"memotree/server/api/internal/storage"
	"memotree/server/api/internal/store"
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

type nilFamiliesStore struct {
	*store.MemoryStore
}

func (s nilFamiliesStore) ListFamiliesForUser(_ context.Context, _ int64) ([]store.FamilySummary, error) {
	return nil, nil
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
	return storage.ObjectInfo{Bucket: bucket, ObjectKey: objectKey}, nil
}

func (fakeStorageService) DeleteObject(_ context.Context, _ string, _ string) error {
	return nil
}
