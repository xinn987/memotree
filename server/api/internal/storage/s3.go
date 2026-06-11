package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Service 通过 S3-compatible API 访问私有对象存储。
// 业务层只依赖 Service 接口，避免 R2、MinIO 或其他云厂商细节扩散到权限逻辑中。
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

func (s *S3Service) DeleteObject(ctx context.Context, bucket string, objectKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})
	return err
}
