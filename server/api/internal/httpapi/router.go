// Package httpapi 提供 MemoTree API 的 HTTP 边界。
//
// 这个包只处理请求解析、认证中间件、响应格式和路由注册；
// 业务规则和持久化细节分别放在 auth / store 包中。
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"memotree/server/api/internal/auth"
	"memotree/server/api/internal/config"
	"memotree/server/api/internal/storage"
	"memotree/server/api/internal/store"
)

type app struct {
	cfg     config.Config
	store   store.Store
	storage storage.Service
	now     func() time.Time
}

type requestContextKey string

const userContextKey requestContextKey = "current_user"

// NewRouter 使用内存 store 创建路由，主要用于无数据库本地试跑和单元测试。
func NewRouter(cfg config.Config) http.Handler {
	return NewRouterWithStore(cfg, store.NewMemoryStore())
}

// NewRouterWithStore 注入外部 store 创建路由。
// API 进程使用它接入 MySQL store；测试使用它复用同一个 MemoryStore。
func NewRouterWithStore(cfg config.Config, appStore store.Store) http.Handler {
	return NewRouterWithDependencies(cfg, appStore, nil)
}

func NewRouterWithDependencies(cfg config.Config, appStore store.Store, storageService storage.Service) http.Handler {
	if cfg.SessionCookieName == "" {
		cfg.SessionCookieName = "memotree_session"
	}
	if cfg.SessionTTL <= 0 {
		cfg.SessionTTL = 30 * 24 * time.Hour
	}

	api := &app{
		cfg:     cfg,
		store:   appStore,
		storage: storageService,
		now:     time.Now,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", api.health)
	mux.HandleFunc("POST /auth/register", api.register)
	mux.HandleFunc("POST /auth/login", api.login)
	mux.HandleFunc("POST /auth/logout", api.logout)
	mux.HandleFunc("GET /auth/session", api.currentSession)
	mux.HandleFunc("POST /families", api.requireAuth(api.createFamily))
	mux.HandleFunc("GET /families", api.requireAuth(api.listFamilies))
	mux.HandleFunc("POST /families/{familyId}/invites", api.requireAuth(api.createInvite))
	mux.HandleFunc("GET /families/{familyId}/invites", api.requireAuth(api.listInvites))
	mux.HandleFunc("DELETE /families/{familyId}/invites/{inviteId}", api.requireAuth(api.revokeInvite))
	mux.HandleFunc("POST /invites/{token}/join", api.requireAuth(api.joinInvite))
	mux.HandleFunc("POST /families/{familyId}/media/upload-intents", api.requireAuth(api.createUploadIntents))
	return withRequestLog(mux)
}

func (a *app) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"env":    a.cfg.AppEnv,
	})
}

func (a *app) register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		LoginName   string `json:"loginName"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if !readJSON(w, r, &input) {
		return
	}

	loginName := strings.TrimSpace(input.LoginName)
	displayName := strings.TrimSpace(input.DisplayName)
	if loginName == "" || len(input.Password) < 6 || displayName == "" {
		writeError(w, http.StatusBadRequest, "登录名、至少 6 位密码和显示名不能为空")
		return
	}

	// HTTP 层只处理明文密码入参，落库前必须先变成不可逆哈希。
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "密码处理失败")
		return
	}

	created, err := a.store.CreateUser(r.Context(), loginName, passwordHash, displayName)
	if errors.Is(err, store.ErrAlreadyExists) {
		writeError(w, http.StatusConflict, "登录名已存在")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建用户失败")
		return
	}

	a.issueSession(w, r, created)
	writeJSON(w, http.StatusCreated, userResponse(created))
}

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		LoginName string `json:"loginName"`
		Password  string `json:"password"`
	}
	if !readJSON(w, r, &input) {
		return
	}

	found, passwordHash, err := a.store.FindUserByLoginName(r.Context(), strings.TrimSpace(input.LoginName))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "登录名或密码错误")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "登录失败")
		return
	}
	if !auth.VerifyPassword(input.Password, passwordHash) {
		writeError(w, http.StatusUnauthorized, "登录名或密码错误")
		return
	}

	a.issueSession(w, r, found)
	writeJSON(w, http.StatusOK, userResponse(found))
}

func (a *app) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(a.cfg.SessionCookieName); err == nil {
		// session cookie 中是 token 原文；store 中只保存 token hash。
		_ = a.store.DeleteSession(r.Context(), auth.HashToken(cookie.Value))
	}
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *app) currentSession(w http.ResponseWriter, r *http.Request) {
	current, ok := a.authenticate(r)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}
	families, err := a.store.ListFamiliesForUser(r.Context(), current.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取家庭列表失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"user":          userResponse(current),
		"families":      familyListResponse(families),
	})
}

func (a *app) createFamily(w http.ResponseWriter, r *http.Request) {
	current := currentUser(r)
	var input struct {
		DisplayName string `json:"displayName"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		writeError(w, http.StatusBadRequest, "家庭名称不能为空")
		return
	}

	created, err := a.store.CreateFamily(r.Context(), displayName, store.DefaultFamilyTimezone, current.ID, current.DisplayName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建家庭失败")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *app) listFamilies(w http.ResponseWriter, r *http.Request) {
	current := currentUser(r)
	families, err := a.store.ListFamiliesForUser(r.Context(), current.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取家庭列表失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"families": familyListResponse(families)})
}

func (a *app) createInvite(w http.ResponseWriter, r *http.Request) {
	current := currentUser(r)
	familyID, err := parsePathID(r, "familyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "家庭 ID 不合法")
		return
	}
	isAdmin, err := a.store.IsActiveAdmin(r.Context(), familyID, current.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "校验家庭权限失败")
		return
	}
	if !isAdmin {
		writeError(w, http.StatusForbidden, "只有家庭管理员可以创建邀请")
		return
	}

	var input struct {
		MemberDisplayName string `json:"memberDisplayName"`
	}
	if !readJSON(w, r, &input) {
		return
	}

	// MVP 阶段保存邀请 token 原文，便于管理员后续在邀请管理中重新复制链接。
	token, err := auth.NewToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建邀请失败")
		return
	}
	invite, err := a.store.CreateInvite(r.Context(), familyID, auth.HashToken(token), token, current.ID, strings.TrimSpace(input.MemberDisplayName), a.now().Add(7*24*time.Hour))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建邀请失败")
		return
	}
	writeJSON(w, http.StatusCreated, inviteResponse(invite, a.now()))
}

func (a *app) listInvites(w http.ResponseWriter, r *http.Request) {
	current := currentUser(r)
	familyID, err := parsePathID(r, "familyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "家庭 ID 不合法")
		return
	}
	isAdmin, err := a.store.IsActiveAdmin(r.Context(), familyID, current.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "校验家庭权限失败")
		return
	}
	if !isAdmin {
		writeError(w, http.StatusForbidden, "只有家庭管理员可以查看邀请")
		return
	}

	invites, err := a.store.ListInvitesForFamily(r.Context(), familyID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取邀请列表失败")
		return
	}
	items := make([]map[string]any, 0, len(invites))
	now := a.now()
	for _, invite := range invites {
		items = append(items, inviteResponse(invite, now))
	}
	writeJSON(w, http.StatusOK, map[string]any{"invites": items})
}

func (a *app) revokeInvite(w http.ResponseWriter, r *http.Request) {
	current := currentUser(r)
	familyID, err := parsePathID(r, "familyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "家庭 ID 不合法")
		return
	}
	inviteID, err := parsePathID(r, "inviteId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "邀请 ID 不合法")
		return
	}
	isAdmin, err := a.store.IsActiveAdmin(r.Context(), familyID, current.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "校验家庭权限失败")
		return
	}
	if !isAdmin {
		writeError(w, http.StatusForbidden, "只有家庭管理员可以撤销邀请")
		return
	}

	invite, err := a.store.RevokeInvite(r.Context(), familyID, inviteID, a.now())
	switch {
	case errors.Is(err, store.ErrInvalidInvite):
		writeError(w, http.StatusConflict, "只能撤销待使用邀请")
		return
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "邀请不存在")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "撤销邀请失败")
		return
	}
	writeJSON(w, http.StatusOK, inviteResponse(invite, a.now()))
}

func (a *app) joinInvite(w http.ResponseWriter, r *http.Request) {
	current := currentUser(r)
	token := strings.TrimSpace(r.PathValue("token"))
	if token == "" {
		writeError(w, http.StatusBadRequest, "邀请 token 不能为空")
		return
	}

	member, err := a.store.JoinInvite(r.Context(), auth.HashToken(token), current.ID, current.DisplayName, a.now())
	switch {
	case errors.Is(err, store.ErrInvalidInvite):
		writeError(w, http.StatusBadRequest, "邀请无效或已过期")
		return
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "邀请不存在")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "加入家庭失败")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"familyId":    member.FamilyID,
		"role":        member.Role,
		"status":      member.Status,
		"displayName": member.DisplayName,
	})
}

func (a *app) createUploadIntents(w http.ResponseWriter, r *http.Request) {
	current := currentUser(r)
	familyID, err := parsePathID(r, "familyId")
	if err != nil {
		writeError(w, http.StatusBadRequest, "家庭 ID 不合法")
		return
	}
	isMember, err := a.store.IsActiveMember(r.Context(), familyID, current.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "校验家庭权限失败")
		return
	}
	if !isMember {
		writeError(w, http.StatusForbidden, "只有家庭成员可以上传媒体")
		return
	}
	if a.storage == nil {
		writeError(w, http.StatusServiceUnavailable, "对象存储尚未配置")
		return
	}

	var input struct {
		Files []struct {
			Filename         string `json:"filename"`
			OriginalFilename string `json:"originalFilename"`
			ContentType      string `json:"contentType"`
			ByteSize         int64  `json:"byteSize"`
		} `json:"files"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	if len(input.Files) == 0 {
		writeError(w, http.StatusBadRequest, "请选择要上传的文件")
		return
	}
	if a.cfg.UploadMaxBatchCount > 0 && len(input.Files) > a.cfg.UploadMaxBatchCount {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("单次最多上传 %d 个文件", a.cfg.UploadMaxBatchCount))
		return
	}

	now := a.now()
	items := make([]store.CreateUploadItemInput, 0, len(input.Files))
	for _, file := range input.Files {
		filename := strings.TrimSpace(file.OriginalFilename)
		if filename == "" {
			filename = strings.TrimSpace(file.Filename)
		}
		contentType := strings.TrimSpace(file.ContentType)
		originalType, err := originalTypeForContentType(contentType)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if filename == "" {
			writeError(w, http.StatusBadRequest, "文件名不能为空")
			return
		}
		if file.ByteSize <= 0 {
			writeError(w, http.StatusBadRequest, "文件大小不合法")
			return
		}
		if a.cfg.UploadMaxFileBytes > 0 && file.ByteSize > a.cfg.UploadMaxFileBytes {
			writeError(w, http.StatusBadRequest, "文件超过上传大小限制")
			return
		}
		objectKey, err := newOriginalObjectKey(familyID, current.ID, filename, now)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "生成上传路径失败")
			return
		}
		items = append(items, store.CreateUploadItemInput{
			OriginalType:     originalType,
			OriginalFilename: filename,
			ContentType:      contentType,
			ByteSize:         file.ByteSize,
			ObjectKey:        objectKey,
		})
	}

	batch, createdItems, err := a.store.CreateUploadBatch(r.Context(), store.CreateUploadBatchInput{
		FamilyID:  familyID,
		CreatedBy: current.ID,
		Items:     items,
		Now:       now,
	})
	if errors.Is(err, store.ErrAlreadyExists) {
		writeError(w, http.StatusConflict, "当前家庭已有进行中的上传任务")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建上传任务失败")
		return
	}

	responseItems := make([]map[string]any, 0, len(createdItems))
	expiresAt := now.Add(a.cfg.SignedURLTTL)
	for _, item := range createdItems {
		uploadURL, err := a.storage.GetSignedUploadURL(r.Context(), storage.SignedURLRequest{
			Bucket:      a.cfg.OriginalsBucket,
			ObjectKey:   item.ObjectKey,
			ContentType: item.ContentType,
			ExpiresIn:   a.cfg.SignedURLTTL,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "生成上传授权失败")
			return
		}
		responseItems = append(responseItems, map[string]any{
			"id":               item.ID,
			"uploadUrl":        uploadURL,
			"method":           http.MethodPut,
			"contentType":      item.ContentType,
			"originalFilename": item.OriginalFilename,
			"byteSize":         item.ByteSize,
			"status":           item.Status,
			"expiresAt":        expiresAt.Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"batch": map[string]any{
			"id":         batch.ID,
			"familyId":   batch.FamilyID,
			"status":     batch.Status,
			"totalCount": batch.TotalCount,
			"createdAt":  batch.CreatedAt.Format(time.RFC3339),
		},
		"items": responseItems,
	})
}

func (a *app) requireAuth(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		current, ok := a.authenticate(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "请先登录")
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, current)
		next(w, r.WithContext(ctx))
	}
}

func (a *app) authenticate(r *http.Request) (store.User, bool) {
	cookie, err := r.Cookie(a.cfg.SessionCookieName)
	if err != nil || cookie.Value == "" {
		return store.User{}, false
	}
	found, ok, err := a.store.FindSession(r.Context(), auth.HashToken(cookie.Value), a.now())
	if err != nil || !ok {
		return store.User{}, false
	}
	current, err := a.store.FindUserByID(r.Context(), found.UserID)
	return current, err == nil
}

func (a *app) issueSession(w http.ResponseWriter, r *http.Request, created store.User) {
	token, err := auth.NewToken()
	if err != nil {
		// token 生成失败非常少见；这里让客户端后续通过登录重新获取会话。
		return
	}
	expiresAt := a.now().Add(a.cfg.SessionTTL)
	// Cookie 发给浏览器的是 token 原文；服务端只持久化 hash，降低泄露后的直接利用风险。
	if err := a.store.CreateSession(r.Context(), created.ID, auth.HashToken(token), expiresAt); err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(a.cfg.SessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func currentUser(r *http.Request) store.User {
	current, _ := r.Context().Value(userContextKey).(store.User)
	return current
}

func userResponse(current store.User) map[string]any {
	return map[string]any{
		"id":            current.ID,
		"loginName":     current.LoginName,
		"displayName":   current.DisplayName,
		"isSystemAdmin": current.IsSystemAdmin,
	}
}

func inviteResponse(invite store.FamilyInvite, now time.Time) map[string]any {
	status := invite.Status
	if status == store.InviteStatusPending && !invite.ExpiresAt.After(now) {
		status = "expired"
	}
	var usedAt any
	if !invite.UsedAt.IsZero() {
		usedAt = invite.UsedAt.Format(time.RFC3339)
	}
	return map[string]any{
		"id":                invite.ID,
		"familyId":          invite.FamilyID,
		"token":             invite.TokenPlaintext,
		"memberDisplayName": invite.MemberDisplayName,
		"status":            status,
		"expiresAt":         invite.ExpiresAt.Format(time.RFC3339),
		"usedAt":            usedAt,
	}
}

func parsePathID(r *http.Request, name string) (int64, error) {
	value, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid path id %s", name)
	}
	return value, nil
}

func originalTypeForContentType(contentType string) (string, error) {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return store.OriginalTypeImage, nil
	case strings.HasPrefix(contentType, "video/"):
		return store.OriginalTypeVideo, nil
	default:
		return "", fmt.Errorf("暂不支持该文件类型")
	}
}

func newOriginalObjectKey(familyID int64, userID int64, filename string, now time.Time) (string, error) {
	token, err := auth.NewToken()
	if err != nil {
		return "", err
	}
	extension := strings.ToLower(path.Ext(filename))
	return fmt.Sprintf("originals/families/%d/users/%d/%s/%s%s", familyID, userID, now.UTC().Format("20060102"), token, extension), nil
}

func readJSON(w http.ResponseWriter, r *http.Request, output any) bool {
	if err := json.NewDecoder(r.Body).Decode(output); err != nil {
		writeError(w, http.StatusBadRequest, "请求 JSON 不合法")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// 客户端连接中断时编码可能失败，这里不再二次写响应。
		return
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// familyListResponse 固化 HTTP 契约：没有家庭时返回 JSON []，不能返回 null。
// 前端会按数组处理这个字段；这里兜底可以避免不同 store 实现的 nil slice 差异泄漏到客户端。
func familyListResponse(families []store.FamilySummary) []store.FamilySummary {
	if families == nil {
		return []store.FamilySummary{}
	}
	return families
}

// withRequestLog 记录每个 API 请求的基本结果，便于本地开发时从终端判断请求是否真的打到后端。
func withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, recorder.status, time.Since(startedAt).Round(time.Millisecond))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
