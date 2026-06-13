package media

import (
	"context"
	"time"
)

const (
	defaultBatchSize    = 4
	defaultPollInterval = 5 * time.Second
)

// JobRepository 提供 Worker 领取媒体处理任务的入口。
type JobRepository interface {
	Repository
	ClaimPhotoJobs(ctx context.Context, limit int) ([]PhotoJob, error)
}

// PhotoProcessor 是照片任务处理器的最小接口，便于后续替换成队列驱动。
type PhotoProcessor interface {
	ProcessPhotoJob(ctx context.Context, job PhotoJob) error
}

// Logger 是 Worker 的最小日志接口，兼容标准库 log.Logger。
type Logger interface {
	Printf(format string, args ...any)
}

// Runner 负责调度媒体任务；当前由 ticker 驱动，后续可以替换 JobRepository 的来源。
type Runner struct {
	Repository   JobRepository
	Processor    PhotoProcessor
	BatchSize    int
	PollInterval time.Duration
	Logger       Logger
}

// RunOnce 执行一次数据库轮询和处理，便于测试和手动触发。
func (r Runner) RunOnce(ctx context.Context) (int, error) {
	batchSize := r.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	jobs, err := r.Repository.ClaimPhotoJobs(ctx, batchSize)
	if err != nil {
		return 0, err
	}
	r.logf("media worker tick claimed=%d", len(jobs))
	for _, job := range jobs {
		startedAt := time.Now()
		r.logf("media worker job start mediaAssetID=%d uploadItemID=%d uploadBatchID=%d", job.MediaAssetID, job.UploadItemID, job.UploadBatchID)
		if err := r.Processor.ProcessPhotoJob(ctx, job); err != nil {
			r.logf("media worker job failed mediaAssetID=%d uploadItemID=%d uploadBatchID=%d duration=%s error=%q", job.MediaAssetID, job.UploadItemID, job.UploadBatchID, time.Since(startedAt).Round(time.Millisecond), err.Error())
			return len(jobs), err
		}
		r.logf("media worker job success mediaAssetID=%d uploadItemID=%d uploadBatchID=%d duration=%s", job.MediaAssetID, job.UploadItemID, job.UploadBatchID, time.Since(startedAt).Round(time.Millisecond))
	}
	return len(jobs), nil
}

// Run 常驻运行 Worker：启动时立即扫一次，之后按固定间隔轮询。
func (r Runner) Run(ctx context.Context, onError func(error)) error {
	if _, err := r.RunOnce(ctx); err != nil {
		if onError != nil {
			onError(err)
		}
	}
	interval := r.PollInterval
	if interval <= 0 {
		interval = defaultPollInterval
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := r.RunOnce(ctx); err != nil && onError != nil {
				onError(err)
			}
		}
	}
}

func (r Runner) logf(format string, args ...any) {
	if r.Logger != nil {
		r.Logger.Printf(format, args...)
	}
}
