package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"time"
)

const (
	// 缩略图用于时间线网格，尺寸较小以降低首屏流量。
	thumbnailMaxSide = 320
	// 展示图用于详情页和较大预览，避免前端直接加载原图。
	displayMaxSide = 1600
)

var (
	errObjectNotFound             = errors.New("object not found")
	errVideoTranscoderUnavailable = errors.New("video transcoder is not configured")
)

// Object 是 Worker 从对象存储读取或写入后的轻量对象信息。
type Object struct {
	Body        []byte
	ContentType string
	SizeBytes   int64
}

// ObjectStore 隔离具体 S3/R2/MinIO 实现，便于替换对象存储客户端。
type ObjectStore interface {
	GetObject(ctx context.Context, bucket string, objectKey string) (Object, error)
	PutObject(ctx context.Context, bucket string, objectKey string, contentType string, body []byte) (Object, error)
}

// Repository 隔离 Worker 的任务状态写回，当前实现可以是 MySQL 轮询，后续可替换为队列触发。
type Repository interface {
	CompletePhotoJob(ctx context.Context, input CompletePhotoJobInput) error
	FailPhotoJob(ctx context.Context, input FailPhotoJobInput) error
	CompleteVideoJob(ctx context.Context, input CompleteVideoJobInput) error
	FailVideoJob(ctx context.Context, input FailVideoJobInput) error
}

// Processor 负责把已上传原文件转换为前端可展示的派生资源。
type Processor struct {
	Repository      Repository
	ObjectStore     ObjectStore
	OriginalsBucket string
	PreviewsBucket  string
	VideoTranscoder VideoTranscoder
	Now             func() time.Time
}

// PhotoJob 是数据库中已完成原图上传、等待生成预览资源的照片任务。
type PhotoJob struct {
	MediaAssetID      int64
	UploadItemID      int64
	UploadBatchID     int64
	OriginalObjectKey string
	OriginalFilename  string
}

// VideoJob 是已完成原视频上传、等待生成浏览器可播放资源的任务。
type VideoJob struct {
	MediaAssetID      int64
	UploadItemID      int64
	UploadBatchID     int64
	OriginalObjectKey string
	OriginalFilename  string
}

// RenditionResult 记录 Worker 生成的单个派生资源。
type RenditionResult struct {
	RenditionType  string
	ObjectKey      string
	ContentType    string
	ByteSize       int64
	Width          int
	Height         int
	DurationMillis int64
}

// CompletePhotoJobInput 是照片处理成功后的状态写回载荷。
type CompletePhotoJobInput struct {
	MediaAssetID  int64
	UploadItemID  int64
	UploadBatchID int64
	Width         int
	Height        int
	Renditions    []RenditionResult
	Now           time.Time
}

// FailPhotoJobInput 是照片处理失败后的状态写回载荷。
type FailPhotoJobInput struct {
	MediaAssetID  int64
	UploadItemID  int64
	UploadBatchID int64
	ErrorMessage  string
	Now           time.Time
}

// CompleteVideoJobInput 是视频处理成功后的状态写回载荷。
type CompleteVideoJobInput struct {
	MediaAssetID   int64
	UploadItemID   int64
	UploadBatchID  int64
	Width          int
	Height         int
	DurationMillis int64
	Renditions     []RenditionResult
	Now            time.Time
}

// FailVideoJobInput 是视频处理失败后的状态写回载荷。
type FailVideoJobInput struct {
	MediaAssetID  int64
	UploadItemID  int64
	UploadBatchID int64
	ErrorMessage  string
	Now           time.Time
}

// VideoTranscoder 隔离 FFmpeg 命令行，便于 Processor 做单元测试。
type VideoTranscoder interface {
	TranscodeVideo(ctx context.Context, input VideoTranscodeInput) (VideoTranscodeOutput, error)
}

type VideoTranscodeInput struct {
	Original         []byte
	OriginalFilename string
}

type VideoTranscodeOutput struct {
	Thumbnail       []byte
	DisplayVideo    []byte
	Width           int
	Height          int
	DurationMillis  int64
	ThumbnailWidth  int
	ThumbnailHeight int
}

// ProcessPhotoJob 生成 thumbnail 和 display image，并将处理结果写回数据库。
func (p Processor) ProcessPhotoJob(ctx context.Context, job PhotoJob) error {
	now := p.now()
	original, err := p.ObjectStore.GetObject(ctx, p.OriginalsBucket, job.OriginalObjectKey)
	if err != nil {
		return p.failPhoto(ctx, job, err, now)
	}

	src, _, err := image.Decode(bytes.NewReader(original.Body))
	if err != nil {
		return p.failPhoto(ctx, job, err, now)
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	thumbnail, err := encodeJPEG(resizeToFit(src, thumbnailMaxSide))
	if err != nil {
		return p.failPhoto(ctx, job, err, now)
	}
	display, err := encodeJPEG(resizeToFit(src, displayMaxSide))
	if err != nil {
		return p.failPhoto(ctx, job, err, now)
	}

	thumbnailKey := fmt.Sprintf("previews/media/%d/thumbnail.jpg", job.MediaAssetID)
	thumbnailObject, err := p.ObjectStore.PutObject(ctx, p.PreviewsBucket, thumbnailKey, "image/jpeg", thumbnail)
	if err != nil {
		return p.failPhoto(ctx, job, err, now)
	}
	displayKey := fmt.Sprintf("previews/media/%d/display.jpg", job.MediaAssetID)
	displayObject, err := p.ObjectStore.PutObject(ctx, p.PreviewsBucket, displayKey, "image/jpeg", display)
	if err != nil {
		return p.failPhoto(ctx, job, err, now)
	}

	err = p.Repository.CompletePhotoJob(ctx, CompletePhotoJobInput{
		MediaAssetID:  job.MediaAssetID,
		UploadItemID:  job.UploadItemID,
		UploadBatchID: job.UploadBatchID,
		Width:         width,
		Height:        height,
		Renditions: []RenditionResult{
			{
				RenditionType: "thumbnail",
				ObjectKey:     thumbnailKey,
				ContentType:   "image/jpeg",
				ByteSize:      thumbnailObject.SizeBytes,
				Width:         resizeWidth(width, height, thumbnailMaxSide),
				Height:        resizeHeight(width, height, thumbnailMaxSide),
			},
			{
				RenditionType: "display_image",
				ObjectKey:     displayKey,
				ContentType:   "image/jpeg",
				ByteSize:      displayObject.SizeBytes,
				Width:         resizeWidth(width, height, displayMaxSide),
				Height:        resizeHeight(width, height, displayMaxSide),
			},
		},
		Now: now,
	})
	if err != nil {
		return err
	}
	return nil
}

// ProcessVideoJob 生成视频缩略图和 MP4 展示视频，并将处理结果写回数据库。
func (p Processor) ProcessVideoJob(ctx context.Context, job VideoJob) error {
	now := p.now()
	if p.VideoTranscoder == nil {
		return p.failVideo(ctx, job, errVideoTranscoderUnavailable, now)
	}
	original, err := p.ObjectStore.GetObject(ctx, p.OriginalsBucket, job.OriginalObjectKey)
	if err != nil {
		return p.failVideo(ctx, job, err, now)
	}
	output, err := p.VideoTranscoder.TranscodeVideo(ctx, VideoTranscodeInput{
		Original:         original.Body,
		OriginalFilename: job.OriginalFilename,
	})
	if err != nil {
		return p.failVideo(ctx, job, err, now)
	}

	thumbnailKey := fmt.Sprintf("previews/media/%d/thumbnail.jpg", job.MediaAssetID)
	thumbnailObject, err := p.ObjectStore.PutObject(ctx, p.PreviewsBucket, thumbnailKey, "image/jpeg", output.Thumbnail)
	if err != nil {
		return p.failVideo(ctx, job, err, now)
	}
	displayKey := fmt.Sprintf("previews/media/%d/display.mp4", job.MediaAssetID)
	displayObject, err := p.ObjectStore.PutObject(ctx, p.PreviewsBucket, displayKey, "video/mp4", output.DisplayVideo)
	if err != nil {
		return p.failVideo(ctx, job, err, now)
	}

	err = p.Repository.CompleteVideoJob(ctx, CompleteVideoJobInput{
		MediaAssetID:   job.MediaAssetID,
		UploadItemID:   job.UploadItemID,
		UploadBatchID:  job.UploadBatchID,
		Width:          output.Width,
		Height:         output.Height,
		DurationMillis: output.DurationMillis,
		Renditions: []RenditionResult{
			{
				RenditionType:  "thumbnail",
				ObjectKey:      thumbnailKey,
				ContentType:    "image/jpeg",
				ByteSize:       thumbnailObject.SizeBytes,
				Width:          output.ThumbnailWidth,
				Height:         output.ThumbnailHeight,
				DurationMillis: output.DurationMillis,
			},
			{
				RenditionType:  "display_video",
				ObjectKey:      displayKey,
				ContentType:    "video/mp4",
				ByteSize:       displayObject.SizeBytes,
				Width:          output.Width,
				Height:         output.Height,
				DurationMillis: output.DurationMillis,
			},
		},
		Now: now,
	})
	if err != nil {
		return err
	}
	return nil
}

func (p Processor) failPhoto(ctx context.Context, job PhotoJob, cause error, now time.Time) error {
	message := cause.Error()
	if err := p.Repository.FailPhotoJob(ctx, FailPhotoJobInput{
		MediaAssetID:  job.MediaAssetID,
		UploadItemID:  job.UploadItemID,
		UploadBatchID: job.UploadBatchID,
		ErrorMessage:  message,
		Now:           now,
	}); err != nil {
		return err
	}
	return cause
}

func (p Processor) failVideo(ctx context.Context, job VideoJob, cause error, now time.Time) error {
	message := cause.Error()
	if err := p.Repository.FailVideoJob(ctx, FailVideoJobInput{
		MediaAssetID:  job.MediaAssetID,
		UploadItemID:  job.UploadItemID,
		UploadBatchID: job.UploadBatchID,
		ErrorMessage:  message,
		Now:           now,
	}); err != nil {
		return err
	}
	return cause
}

func (p Processor) now() time.Time {
	if p.Now != nil {
		return p.Now()
	}
	return time.Now()
}

func resizeToFit(src image.Image, maxSide int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	targetWidth := resizeWidth(width, height, maxSide)
	targetHeight := resizeHeight(width, height, maxSide)
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	for y := 0; y < targetHeight; y++ {
		sourceY := bounds.Min.Y + y*height/targetHeight
		for x := 0; x < targetWidth; x++ {
			sourceX := bounds.Min.X + x*width/targetWidth
			dst.Set(x, y, src.At(sourceX, sourceY))
		}
	}
	return dst
}

func resizeWidth(width int, height int, maxSide int) int {
	if width <= maxSide && height <= maxSide {
		return width
	}
	if width >= height {
		return maxSide
	}
	return max(1, width*maxSide/height)
}

func resizeHeight(width int, height int, maxSide int) int {
	if width <= maxSide && height <= maxSide {
		return height
	}
	if height >= width {
		return maxSide
	}
	return max(1, height*maxSide/width)
}

func encodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 82}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
