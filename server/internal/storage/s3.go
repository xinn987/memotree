package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Service 通过 S3-compatible API 访问私有对象存储。
// 业务层只依赖 Service 或 Worker 自己的窄接口，避免 R2、MinIO 等厂商细节扩散。
type S3Service struct {
	client    *s3.Client
	presigner *s3.PresignClient
}

func NewS3Service(cfg S3Config) (*S3Service, error) {
	if strings.TrimSpace(cfg.AccessKeyID) == "" || strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return nil, fmt.Errorf("storage credentials are required")
	}
	region := strings.TrimSpace(cfg.Region)
	if region == "" {
		region = "auto"
	}

	options := s3.Options{
		Region:       region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		UsePathStyle: cfg.UsePathStyle,
	}
	if strings.TrimSpace(cfg.Endpoint) != "" {
		options.BaseEndpoint = aws.String(cfg.Endpoint)
	}

	client := s3.New(options)
	return &S3Service{
		client:    client,
		presigner: s3.NewPresignClient(client),
	}, nil
}

func (s *S3Service) GetSignedUploadURL(ctx context.Context, request SignedURLRequest) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(request.Bucket),
		Key:         aws.String(request.ObjectKey),
		ContentType: aws.String(request.ContentType),
	}
	signed, err := s.presigner.PresignPutObject(ctx, input, s3.WithPresignExpires(request.ExpiresIn))
	if err != nil {
		return "", err
	}
	return signed.URL, nil
}

func (s *S3Service) GetSignedDownloadURL(ctx context.Context, request SignedURLRequest) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(request.Bucket),
		Key:    aws.String(request.ObjectKey),
	}
	signed, err := s.presigner.PresignGetObject(ctx, input, s3.WithPresignExpires(request.ExpiresIn))
	if err != nil {
		return "", err
	}
	return signed.URL, nil
}

func (s *S3Service) HeadObject(ctx context.Context, bucket string, objectKey string) (ObjectInfo, error) {
	output, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return ObjectInfo{}, err
	}
	contentType := ""
	if output.ContentType != nil {
		contentType = *output.ContentType
	}
	return ObjectInfo{
		Bucket:      bucket,
		ObjectKey:   objectKey,
		ContentType: contentType,
		SizeBytes:   aws.ToInt64(output.ContentLength),
	}, nil
}

func (s *S3Service) GetObject(ctx context.Context, bucket string, objectKey string) (Object, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return Object{}, err
	}
	defer output.Body.Close()
	body, err := io.ReadAll(output.Body)
	if err != nil {
		return Object{}, err
	}
	contentType := ""
	if output.ContentType != nil {
		contentType = *output.ContentType
	}
	return Object{
		Body:        body,
		ContentType: contentType,
		SizeBytes:   int64(len(body)),
	}, nil
}

func (s *S3Service) PutObject(ctx context.Context, bucket string, objectKey string, contentType string, body []byte) (Object, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
		Body:        bytes.NewReader(body),
	})
	if err != nil {
		return Object{}, err
	}
	return Object{
		Body:        body,
		ContentType: contentType,
		SizeBytes:   int64(len(body)),
	}, nil
}

// EnsureBucket 幂等确保 bucket 存在，供本地开发和运维初始化脚本使用。
func (s *S3Service) EnsureBucket(ctx context.Context, bucket string) error {
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		return fmt.Errorf("bucket name is required")
	}
	if _, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err == nil {
		return nil
	}
	_, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return nil
	}
	var ownedByYou *types.BucketAlreadyOwnedByYou
	if errors.As(err, &ownedByYou) {
		return nil
	}
	var alreadyExists *types.BucketAlreadyExists
	if errors.As(err, &alreadyExists) {
		return nil
	}
	return err
}

func (s *S3Service) DeleteObject(ctx context.Context, bucket string, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	return err
}
