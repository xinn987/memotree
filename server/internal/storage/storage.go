package storage

import (
	"context"
	"time"
)

type SignedURLRequest struct {
	Bucket      string
	ObjectKey   string
	ContentType string
	ExpiresIn   time.Duration
}

type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
}

type Service interface {
	// 上传和下载都必须由 API 先完成家庭成员权限校验，再生成短期授权 URL。
	GetSignedUploadURL(ctx context.Context, request SignedURLRequest) (string, error)
	GetSignedDownloadURL(ctx context.Context, request SignedURLRequest) (string, error)
	HeadObject(ctx context.Context, bucket string, objectKey string) (ObjectInfo, error)
	DeleteObject(ctx context.Context, bucket string, objectKey string) error
}

type ObjectInfo struct {
	Bucket      string
	ObjectKey   string
	ContentType string
	SizeBytes   int64
}

type Object struct {
	Body        []byte
	ContentType string
	SizeBytes   int64
}
