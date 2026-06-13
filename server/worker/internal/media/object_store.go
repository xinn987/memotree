package media

import (
	"context"

	sharedstorage "memotree/server/internal/storage"
)

// S3ObjectStore 把共享 S3 客户端适配到媒体处理器需要的窄接口。
type S3ObjectStore struct {
	Service interface {
		GetObject(ctx context.Context, bucket string, objectKey string) (sharedstorage.Object, error)
		PutObject(ctx context.Context, bucket string, objectKey string, contentType string, body []byte) (sharedstorage.Object, error)
	}
}

func (s S3ObjectStore) GetObject(ctx context.Context, bucket string, objectKey string) (Object, error) {
	object, err := s.Service.GetObject(ctx, bucket, objectKey)
	if err != nil {
		return Object{}, err
	}
	return Object{
		Body:        object.Body,
		ContentType: object.ContentType,
		SizeBytes:   object.SizeBytes,
	}, nil
}

func (s S3ObjectStore) PutObject(ctx context.Context, bucket string, objectKey string, contentType string, body []byte) (Object, error) {
	object, err := s.Service.PutObject(ctx, bucket, objectKey, contentType, body)
	if err != nil {
		return Object{}, err
	}
	return Object{
		Body:        object.Body,
		ContentType: object.ContentType,
		SizeBytes:   object.SizeBytes,
	}, nil
}
