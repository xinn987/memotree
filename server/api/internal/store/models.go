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
	MemberStatusRemoved   = "removed"
	InviteStatusPending   = "pending"
	InviteStatusUsed      = "used"
	InviteStatusRevoked   = "revoked"

	MediaTypePhoto     = "photo"
	MediaTypeVideo     = "video"
	MediaTypeLivePhoto = "live_photo"

	MediaStatusActive  = "active"
	MediaStatusDeleted = "deleted"

	RenditionStatusPending    = "pending"
	RenditionStatusProcessing = "processing"
	RenditionStatusReady      = "ready"
	RenditionStatusFailed     = "failed"

	OriginalTypeImage = "image_original"
	OriginalTypeVideo = "video_original"

	RenditionTypeThumbnail    = "thumbnail"
	RenditionTypeDisplayImage = "display_image"
	RenditionTypeDisplayVideo = "display_video"

	UploadBatchStatusCreated         = "created"
	UploadBatchStatusUploading       = "uploading"
	UploadBatchStatusProcessing      = "processing"
	UploadBatchStatusPartiallyFailed = "partially_failed"
	UploadBatchStatusCompleted       = "completed"
	UploadBatchStatusStopped         = "stopped"

	UploadItemStatusWaiting          = "waiting"
	UploadItemStatusUploading        = "uploading"
	UploadItemStatusUploaded         = "uploaded"
	UploadItemStatusProcessing       = "processing"
	UploadItemStatusReady            = "ready"
	UploadItemStatusUploadFailed     = "upload_failed"
	UploadItemStatusProcessingFailed = "processing_failed"
	UploadItemStatusCancelled        = "cancelled"
)

var (
	// ErrNotFound 是 store 层统一的“记录不存在”语义，HTTP 层负责转换为对应响应。
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidInvite = errors.New("invalid invite")
	ErrInvalidUpload = errors.New("invalid upload")
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

// MediaAsset 是时间线和详情页展示的核心对象。
// 它归属于家庭，uploaded_by 仅表示贡献者和审计信息，不表示个人所有权。
type MediaAsset struct {
	ID              int64
	FamilyID        int64
	UploadedBy      int64
	MediaType       string
	Status          string
	RenditionStatus string
	CapturedAt      time.Time
	UploadedAt      time.Time
	DeletedAt       time.Time
}

// MediaOriginal 记录私有对象存储中的原文件。
// 原文件 object key 永远不应直接暴露给未经过权限校验的前端响应。
type MediaOriginal struct {
	ID               int64
	MediaAssetID     int64
	OriginalType     string
	ObjectKey        string
	OriginalFilename string
	ContentType      string
	ByteSize         int64
	ChecksumSHA256   string
	Width            int
	Height           int
	DurationMillis   int64
	CapturedAt       time.Time
	UploadedAt       time.Time
}

// MediaRendition 记录浏览器展示用的派生资源，例如缩略图、展示图或展示视频。
type MediaRendition struct {
	ID             int64
	MediaAssetID   int64
	RenditionType  string
	ObjectKey      string
	ContentType    string
	ByteSize       int64
	Width          int
	Height         int
	DurationMillis int64
	Status         string
	ErrorMessage   string
}

// TimelineMedia 是时间线读取的聚合结果。
// 它只包含已经可展示的媒体资产和预览派生资源，HTTP 层会基于 ObjectKey 再签发短期下载 URL。
type TimelineMedia struct {
	Asset                 MediaAsset
	UploadedByDisplayName string
	Thumbnail             MediaRendition
	Display               MediaRendition
}

// MediaDetail 是媒体详情页读取的聚合结果。
// 它和时间线一样只返回展示资源，不携带原文件 object key。
type MediaDetail struct {
	Asset                 MediaAsset
	UploadedByDisplayName string
	Thumbnail             MediaRendition
	Display               MediaRendition
}

// UploadBatch 表示一次可返回查看的上传任务，不作为时间线浏览对象。
// ActiveSlot 用于 MySQL 唯一键约束同一用户同一家庭最多一个 active 任务。
type UploadBatch struct {
	ID             int64
	FamilyID       int64
	CreatedBy      int64
	Status         string
	ActiveSlot     int
	TotalCount     int
	ReadyCount     int
	FailedCount    int
	CancelledCount int
	CreatedAt      time.Time
	CompletedAt    time.Time
	StoppedAt      time.Time
}

// UploadItem 表示上传任务中的单个文件状态。
// 它可以先于 MediaAsset 存在；原文件完成并入库后再关联 media_asset_id。
type UploadItem struct {
	ID               int64
	UploadBatchID    int64
	MediaAssetID     int64
	OriginalType     string
	OriginalFilename string
	ContentType      string
	ByteSize         int64
	ObjectKey        string
	Status           string
	ErrorMessage     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CompletedAt      time.Time
}

type CreateUploadBatchInput struct {
	FamilyID  int64
	CreatedBy int64
	Items     []CreateUploadItemInput
	Now       time.Time
}

type CreateUploadItemInput struct {
	OriginalType     string
	OriginalFilename string
	ContentType      string
	ByteSize         int64
	ObjectKey        string
}

type CompleteUploadItemInput struct {
	FamilyID       int64
	BatchID        int64
	ItemID         int64
	UploadedBy     int64
	ObjectSize     int64
	ObjectType     string
	ChecksumSHA256 string
	Now            time.Time
}

type UpdateUploadItemStatusInput struct {
	FamilyID     int64
	BatchID      int64
	ItemID       int64
	ActorUserID  int64
	ErrorMessage string
	Now          time.Time
}

// ListUploadBatchesInput 表示上传任务列表查询边界。
// IncludeFamily 为 true 时返回整个家庭任务；否则只返回 ActorUserID 创建的任务。
type ListUploadBatchesInput struct {
	FamilyID      int64
	ActorUserID   int64
	IncludeFamily bool
	Limit         int
}

// ListTimelineMediaInput 表示时间线读取边界。
// AfterTime/AfterID 是基于时间线倒序排序的游标，用于稳定读取下一页。
type ListTimelineMediaInput struct {
	FamilyID  int64
	Limit     int
	AfterTime time.Time
	AfterID   int64
	MediaType string
	MonthFrom time.Time
	MonthTo   time.Time
}

type FindMediaDetailInput struct {
	FamilyID int64
	MediaID  int64
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
	IsActiveMember(ctx context.Context, familyID int64, userID int64) (bool, error)
	IsActiveAdmin(ctx context.Context, familyID int64, userID int64) (bool, error)
	ListTimelineMedia(ctx context.Context, input ListTimelineMediaInput) ([]TimelineMedia, error)
	FindMediaDetail(ctx context.Context, input FindMediaDetailInput) (MediaDetail, bool, error)
	CreateInvite(ctx context.Context, familyID int64, tokenHash string, tokenPlaintext string, createdBy int64, memberDisplayName string, expiresAt time.Time) (FamilyInvite, error)
	ListInvitesForFamily(ctx context.Context, familyID int64) ([]FamilyInvite, error)
	RevokeInvite(ctx context.Context, familyID int64, inviteID int64, now time.Time) (FamilyInvite, error)
	JoinInvite(ctx context.Context, tokenHash string, userID int64, fallbackDisplayName string, now time.Time) (FamilyMember, error)
	FindActiveUploadBatch(ctx context.Context, familyID int64, userID int64) (UploadBatch, bool, error)
	ListUploadBatches(ctx context.Context, input ListUploadBatchesInput) ([]UploadBatch, error)
	FindUploadBatch(ctx context.Context, familyID int64, batchID int64) (UploadBatch, bool, error)
	ListUploadItems(ctx context.Context, batchID int64) ([]UploadItem, error)
	StopUploadBatch(ctx context.Context, batchID int64, now time.Time) (UploadBatch, []UploadItem, error)
	CompleteUploadItem(ctx context.Context, input CompleteUploadItemInput) (UploadBatch, UploadItem, MediaAsset, error)
	MarkUploadItemFailed(ctx context.Context, input UpdateUploadItemStatusInput) (UploadBatch, UploadItem, error)
	RetryUploadItem(ctx context.Context, input UpdateUploadItemStatusInput) (UploadBatch, UploadItem, error)
	CreateUploadBatch(ctx context.Context, input CreateUploadBatchInput) (UploadBatch, []UploadItem, error)
}
