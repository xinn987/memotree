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
		jobs: []PhotoJob{
			{MediaAssetID: 1, UploadItemID: 10, UploadBatchID: 100},
			{MediaAssetID: 2, UploadItemID: 20, UploadBatchID: 200},
		},
	}
	processor := &fakePhotoProcessor{}
	runner := Runner{
		Repository: repository,
		Processor:  processor,
		BatchSize:  5,
	}

	processed, err := runner.RunOnce(ctx)

	if err != nil {
		t.Fatalf("run once: %v", err)
	}
	if processed != 2 {
		t.Fatalf("expected 2 processed jobs, got %d", processed)
	}
	if repository.claimLimit != 5 {
		t.Fatalf("expected claim limit 5, got %d", repository.claimLimit)
	}
	if len(processor.jobs) != 2 || processor.jobs[0].MediaAssetID != 1 || processor.jobs[1].MediaAssetID != 2 {
		t.Fatalf("expected processor to receive jobs, got %#v", processor.jobs)
	}
}

func TestRunnerRunOnceLogsClaimedJobsAndProgress(t *testing.T) {
	ctx := context.Background()
	logger := &fakeLogger{}
	runner := Runner{
		Repository: &fakeJobRepository{
			jobs: []PhotoJob{
				{MediaAssetID: 1, UploadItemID: 10, UploadBatchID: 100},
			},
		},
		Processor: &fakePhotoProcessor{},
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
	for _, expected := range []string{"tick claimed=1", "job start", "mediaAssetID=1", "uploadItemID=10", "job success"} {
		if !strings.Contains(logs, expected) {
			t.Fatalf("expected logs to contain %q, got:\n%s", expected, logs)
		}
	}
}

type fakeJobRepository struct {
	fakeRepository
	jobs       []PhotoJob
	claimLimit int
}

func (f *fakeJobRepository) ClaimPhotoJobs(_ context.Context, limit int) ([]PhotoJob, error) {
	f.claimLimit = limit
	return f.jobs, nil
}

type fakePhotoProcessor struct {
	jobs []PhotoJob
}

func (f *fakePhotoProcessor) ProcessPhotoJob(_ context.Context, job PhotoJob) error {
	f.jobs = append(f.jobs, job)
	return nil
}

type fakeLogger struct {
	lines []string
}

func (f *fakeLogger) Printf(format string, args ...any) {
	f.lines = append(f.lines, fmt.Sprintf(format, args...))
}
