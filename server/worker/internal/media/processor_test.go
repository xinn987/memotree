package media

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
	"time"
)

func TestProcessorProcessesPhotoJob(t *testing.T) {
	ctx := context.Background()
	original := jpegBytes(t, 640, 480)
	objectStore := &fakeObjectStore{
		objects: map[string]storedObject{
			"originals:originals/families/1/users/1/baby.jpg": {
				body:        original,
				contentType: "image/jpeg",
			},
		},
	}
	repository := &fakeRepository{}
	processor := Processor{
		Repository:      repository,
		ObjectStore:     objectStore,
		OriginalsBucket: "originals",
		PreviewsBucket:  "previews",
		Now: func() time.Time {
			return time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
		},
	}

	err := processor.ProcessPhotoJob(ctx, PhotoJob{
		MediaAssetID:      42,
		UploadItemID:      7,
		UploadBatchID:     3,
		OriginalObjectKey: "originals/families/1/users/1/baby.jpg",
		OriginalFilename:  "baby.jpg",
	})

	if err != nil {
		t.Fatalf("process photo job: %v", err)
	}
	if len(objectStore.objects) != 3 {
		t.Fatalf("expected original plus two renditions, got %#v", objectStore.objects)
	}
	thumbnail := objectStore.objects["previews:previews/media/42/thumbnail.jpg"]
	if thumbnail.contentType != "image/jpeg" || len(thumbnail.body) == 0 {
		t.Fatalf("expected uploaded thumbnail jpeg, got %#v", thumbnail)
	}
	display := objectStore.objects["previews:previews/media/42/display.jpg"]
	if display.contentType != "image/jpeg" || len(display.body) == 0 {
		t.Fatalf("expected uploaded display jpeg, got %#v", display)
	}
	if repository.completed.MediaAssetID != 42 || repository.completed.UploadItemID != 7 || len(repository.completed.Renditions) != 2 {
		t.Fatalf("expected completed job with two renditions, got %#v", repository.completed)
	}
	if repository.completed.Width != 640 || repository.completed.Height != 480 {
		t.Fatalf("expected original dimensions to be recorded, got %#v", repository.completed)
	}
}

func TestProcessorMarksPhotoJobFailedWhenOriginalCannotDecode(t *testing.T) {
	ctx := context.Background()
	objectStore := &fakeObjectStore{
		objects: map[string]storedObject{
			"originals:bad.jpg": {
				body:        []byte("not an image"),
				contentType: "image/jpeg",
			},
		},
	}
	repository := &fakeRepository{}
	processor := Processor{
		Repository:      repository,
		ObjectStore:     objectStore,
		OriginalsBucket: "originals",
		PreviewsBucket:  "previews",
	}

	err := processor.ProcessPhotoJob(ctx, PhotoJob{
		MediaAssetID:      5,
		UploadItemID:      9,
		UploadBatchID:     2,
		OriginalObjectKey: "bad.jpg",
	})

	if err == nil {
		t.Fatal("expected decode error")
	}
	if repository.failed.MediaAssetID != 5 || repository.failed.UploadItemID != 9 || repository.failed.ErrorMessage == "" {
		t.Fatalf("expected failed job to be recorded, got %#v", repository.failed)
	}
}

func TestProcessorProcessesVideoJob(t *testing.T) {
	ctx := context.Background()
	original := []byte("video-original")
	objectStore := &fakeObjectStore{
		objects: map[string]storedObject{
			"originals:originals/families/1/users/1/baby.mp4": {
				body:        original,
				contentType: "video/mp4",
			},
		},
	}
	repository := &fakeRepository{}
	transcoder := &fakeVideoTranscoder{
		output: VideoTranscodeOutput{
			Thumbnail:       jpegBytes(t, 320, 180),
			DisplayVideo:    []byte("display-video"),
			Width:           1280,
			Height:          720,
			DurationMillis:  4200,
			ThumbnailWidth:  320,
			ThumbnailHeight: 180,
		},
	}
	processor := Processor{
		Repository:      repository,
		ObjectStore:     objectStore,
		OriginalsBucket: "originals",
		PreviewsBucket:  "previews",
		VideoTranscoder: transcoder,
		Now: func() time.Time {
			return time.Date(2026, 6, 13, 12, 0, 0, 0, time.UTC)
		},
	}

	err := processor.ProcessVideoJob(ctx, VideoJob{
		MediaAssetID:      43,
		UploadItemID:      8,
		UploadBatchID:     4,
		OriginalObjectKey: "originals/families/1/users/1/baby.mp4",
		OriginalFilename:  "baby.mp4",
	})

	if err != nil {
		t.Fatalf("process video job: %v", err)
	}
	if string(transcoder.input.Original) != string(original) || transcoder.input.OriginalFilename != "baby.mp4" {
		t.Fatalf("expected transcoder to receive original video, got %#v", transcoder.input)
	}
	thumbnail := objectStore.objects["previews:previews/media/43/thumbnail.jpg"]
	if thumbnail.contentType != "image/jpeg" || len(thumbnail.body) == 0 {
		t.Fatalf("expected uploaded thumbnail jpeg, got %#v", thumbnail)
	}
	display := objectStore.objects["previews:previews/media/43/display.mp4"]
	if display.contentType != "video/mp4" || string(display.body) != "display-video" {
		t.Fatalf("expected uploaded display video, got %#v", display)
	}
	if repository.completedVideo.MediaAssetID != 43 || repository.completedVideo.UploadItemID != 8 || len(repository.completedVideo.Renditions) != 2 {
		t.Fatalf("expected completed video with two renditions, got %#v", repository.completedVideo)
	}
	if repository.completedVideo.Width != 1280 || repository.completedVideo.Height != 720 || repository.completedVideo.DurationMillis != 4200 {
		t.Fatalf("expected video metadata to be recorded, got %#v", repository.completedVideo)
	}
	if repository.completedVideo.Renditions[1].RenditionType != "display_video" || repository.completedVideo.Renditions[1].DurationMillis != 4200 {
		t.Fatalf("expected display video rendition metadata, got %#v", repository.completedVideo.Renditions)
	}
}

func TestProcessorMarksVideoJobFailedWhenTranscoderFails(t *testing.T) {
	ctx := context.Background()
	objectStore := &fakeObjectStore{
		objects: map[string]storedObject{
			"originals:bad.mp4": {
				body:        []byte("bad video"),
				contentType: "video/mp4",
			},
		},
	}
	repository := &fakeRepository{}
	processor := Processor{
		Repository:      repository,
		ObjectStore:     objectStore,
		OriginalsBucket: "originals",
		PreviewsBucket:  "previews",
		VideoTranscoder: &fakeVideoTranscoder{err: errObjectNotFound},
	}

	err := processor.ProcessVideoJob(ctx, VideoJob{
		MediaAssetID:      6,
		UploadItemID:      10,
		UploadBatchID:     3,
		OriginalObjectKey: "bad.mp4",
		OriginalFilename:  "bad.mp4",
	})

	if err == nil {
		t.Fatal("expected transcoder error")
	}
	if repository.failedVideo.MediaAssetID != 6 || repository.failedVideo.UploadItemID != 10 || repository.failedVideo.ErrorMessage == "" {
		t.Fatalf("expected failed video job to be recorded, got %#v", repository.failedVideo)
	}
}

type fakeRepository struct {
	completed      CompletePhotoJobInput
	failed         FailPhotoJobInput
	completedVideo CompleteVideoJobInput
	failedVideo    FailVideoJobInput
}

func (f *fakeRepository) CompletePhotoJob(_ context.Context, input CompletePhotoJobInput) error {
	f.completed = input
	return nil
}

func (f *fakeRepository) FailPhotoJob(_ context.Context, input FailPhotoJobInput) error {
	f.failed = input
	return nil
}

func (f *fakeRepository) CompleteVideoJob(_ context.Context, input CompleteVideoJobInput) error {
	f.completedVideo = input
	return nil
}

func (f *fakeRepository) FailVideoJob(_ context.Context, input FailVideoJobInput) error {
	f.failedVideo = input
	return nil
}

type fakeVideoTranscoder struct {
	input  VideoTranscodeInput
	output VideoTranscodeOutput
	err    error
}

func (f *fakeVideoTranscoder) TranscodeVideo(_ context.Context, input VideoTranscodeInput) (VideoTranscodeOutput, error) {
	f.input = input
	if f.err != nil {
		return VideoTranscodeOutput{}, f.err
	}
	return f.output, nil
}

type storedObject struct {
	body        []byte
	contentType string
}

type fakeObjectStore struct {
	objects map[string]storedObject
}

func (f *fakeObjectStore) GetObject(_ context.Context, bucket string, objectKey string) (Object, error) {
	object, ok := f.objects[bucket+":"+objectKey]
	if !ok {
		return Object{}, errObjectNotFound
	}
	return Object{Body: object.body, ContentType: object.contentType, SizeBytes: int64(len(object.body))}, nil
}

func (f *fakeObjectStore) PutObject(_ context.Context, bucket string, objectKey string, contentType string, body []byte) (Object, error) {
	if f.objects == nil {
		f.objects = map[string]storedObject{}
	}
	copied := append([]byte(nil), body...)
	f.objects[bucket+":"+objectKey] = storedObject{body: copied, contentType: contentType}
	return Object{Body: copied, ContentType: contentType, SizeBytes: int64(len(copied))}, nil
}

func jpegBytes(t *testing.T, width int, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 120, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}
