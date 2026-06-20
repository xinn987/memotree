package media

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestRunnerRunOnceClaimsAndProcessesPhotoJobs(t *testing.T) {
	ctx := context.Background()
	repository := &fakeJobRepository{
		photoJobs: []PhotoJob{
			{MediaAssetID: 1, UploadItemID: 10, UploadBatchID: 100},
			{MediaAssetID: 2, UploadItemID: 20, UploadBatchID: 200},
		},
		videoJobs: []VideoJob{
			{MediaAssetID: 3, UploadItemID: 30, UploadBatchID: 300},
		},
	}
	processor := &fakeMediaProcessor{}
	runner := Runner{
		Repository: repository,
		Processor:  processor,
		BatchSize:  5,
	}

	processed, err := runner.RunOnce(ctx)

	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if processed != 3 {
		t.Fatalf("expected 3 processed jobs, got %d", processed)
	}
	if repository.claimLimit != 5 {
		t.Fatalf("expected claim limit 5, got %d", repository.claimLimit)
	}
	if len(processor.photoJobs) != 2 || processor.photoJobs[0].MediaAssetID != 1 || processor.photoJobs[1].MediaAssetID != 2 {
		t.Fatalf("expected processor to receive photo jobs, got %#v", processor.photoJobs)
	}
	if len(processor.videoJobs) != 1 || processor.videoJobs[0].MediaAssetID != 3 {
		t.Fatalf("expected processor to receive video jobs, got %#v", processor.videoJobs)
	}
}

func TestRunnerRunOnceLogsClaimedJobsAndProgress(t *testing.T) {
	ctx := context.Background()
	logger := &fakeLogger{}
	runner := Runner{
		Repository: &fakeJobRepository{
			photoJobs: []PhotoJob{
				{MediaAssetID: 1, UploadItemID: 10, UploadBatchID: 100},
			},
		},
		Processor: &fakeMediaProcessor{},
		Logger:    logger,
	}

	processed, err := runner.RunOnce(ctx)

	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if processed != 1 {
		t.Fatalf("expected one processed job, got %d", processed)
	}
	logs := strings.Join(logger.lines, "\n")
	for _, expected := range []string{"tick claimedPhotos=1 claimedVideos=0", "job start mediaType=photo", "mediaAssetID=1", "uploadItemID=10", "job success mediaType=photo"} {
		if !strings.Contains(logs, expected) {
			t.Fatalf("expected logs to contain %q, got:\n%s", expected, logs)
		}
	}
}

type fakeJobRepository struct {
	fakeRepository
	photoJobs  []PhotoJob
	videoJobs  []VideoJob
	claimLimit int
}

func (f *fakeJobRepository) ClaimPhotoJobs(_ context.Context, limit int) ([]PhotoJob, error) {
	f.claimLimit = limit
	return f.photoJobs, nil
}

func (f *fakeJobRepository) ClaimVideoJobs(_ context.Context, limit int) ([]VideoJob, error) {
	f.claimLimit = limit
	return f.videoJobs, nil
}

type fakeMediaProcessor struct {
	photoJobs []PhotoJob
	videoJobs []VideoJob
}

func (f *fakeMediaProcessor) ProcessPhotoJob(_ context.Context, job PhotoJob) error {
	f.photoJobs = append(f.photoJobs, job)
	return nil
}

func (f *fakeMediaProcessor) ProcessVideoJob(_ context.Context, job VideoJob) error {
	f.videoJobs = append(f.videoJobs, job)
	return nil
}

type fakeLogger struct {
	lines []string
}

func (f *fakeLogger) Printf(format string, args ...any) {
	f.lines = append(f.lines, fmt.Sprintf(format, args...))
}
