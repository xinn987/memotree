package media

import (
	"context"
	"database/sql"
	"time"
)

const (
	dbMediaTypePhoto               = "photo"
	dbMediaTypeVideo               = "video"
	dbMediaStatusActive            = "active"
	dbRenditionStatusPending       = "pending"
	dbRenditionStatusProcessing    = "processing"
	dbRenditionStatusReady         = "ready"
	dbRenditionStatusFailed        = "failed"
	dbOriginalTypeImage            = "image_original"
	dbOriginalTypeVideo            = "video_original"
	dbRenditionTypeThumbnail       = "thumbnail"
	dbRenditionTypeDisplayImage    = "display_image"
	dbRenditionTypeDisplayVideo    = "display_video"
	dbUploadBatchStatusCreated     = "created"
	dbUploadBatchStatusProcessing  = "processing"
	dbUploadBatchStatusPartialFail = "partially_failed"
	dbUploadBatchStatusCompleted   = "completed"
	dbUploadItemStatusWaiting      = "waiting"
	dbUploadItemStatusUploading    = "uploading"
	dbUploadItemStatusUploaded     = "uploaded"
	dbUploadItemStatusProcessing   = "processing"
	dbUploadItemStatusReady        = "ready"
	dbUploadItemStatusUploadFail   = "upload_failed"
	dbUploadItemStatusProcessFail  = "processing_failed"
	dbUploadItemStatusCancelled    = "cancelled"
)

// MySQLRepository 使用数据库状态作为 MVP 阶段的媒体任务队列。
type MySQLRepository struct {
	db                *sql.DB
	processingTimeout time.Duration
}

func NewMySQLRepository(db *sql.DB, processingTimeout time.Duration) *MySQLRepository {
	if processingTimeout <= 0 {
		processingTimeout = 30 * time.Minute
	}
	return &MySQLRepository{
		db:                db,
		processingTimeout: processingTimeout,
	}
}

// ClaimPhotoJobs 原子领取一批 pending 照片任务，避免多个 Worker 重复处理同一媒体。
func (r *MySQLRepository) ClaimPhotoJobs(ctx context.Context, limit int) ([]PhotoJob, error) {
	if limit <= 0 {
		limit = defaultBatchSize
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	now := time.Now().UTC()
	staleBefore := now.Add(-r.processingTimeout)
	if _, err := tx.ExecContext(ctx, `
UPDATE media_assets ma
JOIN upload_items ui ON ui.media_asset_id = ma.id
SET ma.rendition_status = ?, ma.updated_at = ?
WHERE ma.media_type = ?
  AND ma.rendition_status = ?
  AND ma.updated_at < ?
  AND ui.status = ?
`, dbRenditionStatusPending, now, dbMediaTypePhoto, dbRenditionStatusProcessing, staleBefore, dbUploadItemStatusProcessing); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT ma.id, ui.id, ui.upload_batch_id, mo.object_key, mo.original_filename
FROM media_assets ma
JOIN media_originals mo ON mo.media_asset_id = ma.id
JOIN upload_items ui ON ui.media_asset_id = ma.id
WHERE ma.status = ?
  AND ma.media_type = ?
  AND ma.rendition_status = ?
  AND mo.original_type = ?
  AND ui.status = ?
ORDER BY ma.id
LIMIT ?
FOR UPDATE SKIP LOCKED
`, dbMediaStatusActive, dbMediaTypePhoto, dbRenditionStatusPending, dbOriginalTypeImage, dbUploadItemStatusProcessing, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []PhotoJob
	var ids []int64
	for rows.Next() {
		var job PhotoJob
		if err := rows.Scan(&job.MediaAssetID, &job.UploadItemID, &job.UploadBatchID, &job.OriginalObjectKey, &job.OriginalFilename); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
		ids = append(ids, job.MediaAssetID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, `
UPDATE media_assets
SET rendition_status = ?, updated_at = ?
WHERE id = ? AND rendition_status = ?
`, dbRenditionStatusProcessing, now, id, dbRenditionStatusPending); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

// CompletePhotoJob 写入照片派生资源，并把媒体和上传任务推进到 ready。
// ClaimVideoJobs 原子领取一批 pending 视频任务，避免多个 Worker 重复处理同一媒体。
func (r *MySQLRepository) ClaimVideoJobs(ctx context.Context, limit int) ([]VideoJob, error) {
	if limit <= 0 {
		limit = defaultBatchSize
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	now := time.Now().UTC()
	staleBefore := now.Add(-r.processingTimeout)
	if _, err := tx.ExecContext(ctx, `
UPDATE media_assets ma
JOIN upload_items ui ON ui.media_asset_id = ma.id
SET ma.rendition_status = ?, ma.updated_at = ?
WHERE ma.media_type = ?
  AND ma.rendition_status = ?
  AND ma.updated_at < ?
  AND ui.status = ?
`, dbRenditionStatusPending, now, dbMediaTypeVideo, dbRenditionStatusProcessing, staleBefore, dbUploadItemStatusProcessing); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT ma.id, ui.id, ui.upload_batch_id, mo.object_key, mo.original_filename
FROM media_assets ma
JOIN media_originals mo ON mo.media_asset_id = ma.id
JOIN upload_items ui ON ui.media_asset_id = ma.id
WHERE ma.status = ?
  AND ma.media_type = ?
  AND ma.rendition_status = ?
  AND mo.original_type = ?
  AND ui.status = ?
ORDER BY ma.id
LIMIT ?
FOR UPDATE SKIP LOCKED
`, dbMediaStatusActive, dbMediaTypeVideo, dbRenditionStatusPending, dbOriginalTypeVideo, dbUploadItemStatusProcessing, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []VideoJob
	var ids []int64
	for rows.Next() {
		var job VideoJob
		if err := rows.Scan(&job.MediaAssetID, &job.UploadItemID, &job.UploadBatchID, &job.OriginalObjectKey, &job.OriginalFilename); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
		ids = append(ids, job.MediaAssetID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, `
UPDATE media_assets
SET rendition_status = ?, updated_at = ?
WHERE id = ? AND rendition_status = ?
`, dbRenditionStatusProcessing, now, id, dbRenditionStatusPending); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *MySQLRepository) CompletePhotoJob(ctx context.Context, input CompletePhotoJobInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	if _, err := tx.ExecContext(ctx, `
UPDATE media_originals
SET width = ?, height = ?
WHERE media_asset_id = ? AND original_type = ?
`, input.Width, input.Height, input.MediaAssetID, dbOriginalTypeImage); err != nil {
		return err
	}
	for _, rendition := range input.Renditions {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO media_renditions (media_asset_id, rendition_type, object_key, content_type, byte_size, width, height, duration_millis, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  object_key = VALUES(object_key),
  content_type = VALUES(content_type),
  byte_size = VALUES(byte_size),
  width = VALUES(width),
  height = VALUES(height),
  duration_millis = VALUES(duration_millis),
  status = VALUES(status),
  error_message = NULL,
  updated_at = VALUES(updated_at)
`, input.MediaAssetID, rendition.RenditionType, rendition.ObjectKey, rendition.ContentType, rendition.ByteSize, rendition.Width, rendition.Height, rendition.DurationMillis, dbRenditionStatusReady, now, now); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE media_assets
SET rendition_status = ?, updated_at = ?
WHERE id = ?
`, dbRenditionStatusReady, now, input.MediaAssetID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE upload_items
SET status = ?, error_message = NULL, updated_at = ?, completed_at = ?
WHERE id = ? AND media_asset_id = ?
`, dbUploadItemStatusReady, now, now, input.UploadItemID, input.MediaAssetID); err != nil {
		return err
	}
	if err := recalculateUploadBatch(ctx, tx, input.UploadBatchID); err != nil {
		return err
	}
	return tx.Commit()
}

// FailPhotoJob 标记派生资源生成失败；原文件保留，后续可以重新进入处理流程。
// CompleteVideoJob 写入视频派生资源，并把媒体和上传任务推进到 ready。
func (r *MySQLRepository) CompleteVideoJob(ctx context.Context, input CompleteVideoJobInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	if _, err := tx.ExecContext(ctx, `
UPDATE media_originals
SET width = ?, height = ?, duration_millis = ?
WHERE media_asset_id = ? AND original_type = ?
`, input.Width, input.Height, input.DurationMillis, input.MediaAssetID, dbOriginalTypeVideo); err != nil {
		return err
	}
	for _, rendition := range input.Renditions {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO media_renditions (media_asset_id, rendition_type, object_key, content_type, byte_size, width, height, duration_millis, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  object_key = VALUES(object_key),
  content_type = VALUES(content_type),
  byte_size = VALUES(byte_size),
  width = VALUES(width),
  height = VALUES(height),
  duration_millis = VALUES(duration_millis),
  status = VALUES(status),
  error_message = NULL,
  updated_at = VALUES(updated_at)
`, input.MediaAssetID, rendition.RenditionType, rendition.ObjectKey, rendition.ContentType, rendition.ByteSize, rendition.Width, rendition.Height, rendition.DurationMillis, dbRenditionStatusReady, now, now); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE media_assets
SET rendition_status = ?, updated_at = ?
WHERE id = ?
`, dbRenditionStatusReady, now, input.MediaAssetID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE upload_items
SET status = ?, error_message = NULL, updated_at = ?, completed_at = ?
WHERE id = ? AND media_asset_id = ?
`, dbUploadItemStatusReady, now, now, input.UploadItemID, input.MediaAssetID); err != nil {
		return err
	}
	if err := recalculateUploadBatch(ctx, tx, input.UploadBatchID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *MySQLRepository) FailPhotoJob(ctx context.Context, input FailPhotoJobInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	if _, err := tx.ExecContext(ctx, `
UPDATE media_assets
SET rendition_status = ?, updated_at = ?
WHERE id = ?
`, dbRenditionStatusFailed, now, input.MediaAssetID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE upload_items
SET status = ?, error_message = ?, updated_at = ?, completed_at = ?
WHERE id = ? AND media_asset_id = ?
`, dbUploadItemStatusProcessFail, input.ErrorMessage, now, now, input.UploadItemID, input.MediaAssetID); err != nil {
		return err
	}
	if err := recalculateUploadBatch(ctx, tx, input.UploadBatchID); err != nil {
		return err
	}
	return tx.Commit()
}

// FailVideoJob 标记视频派生资源生成失败；原文件保留，后续可以重新进入处理流程。
func (r *MySQLRepository) FailVideoJob(ctx context.Context, input FailVideoJobInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	if _, err := tx.ExecContext(ctx, `
UPDATE media_assets
SET rendition_status = ?, updated_at = ?
WHERE id = ?
`, dbRenditionStatusFailed, now, input.MediaAssetID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE upload_items
SET status = ?, error_message = ?, updated_at = ?, completed_at = ?
WHERE id = ? AND media_asset_id = ?
`, dbUploadItemStatusProcessFail, input.ErrorMessage, now, now, input.UploadItemID, input.MediaAssetID); err != nil {
		return err
	}
	if err := recalculateUploadBatch(ctx, tx, input.UploadBatchID); err != nil {
		return err
	}
	return tx.Commit()
}

func recalculateUploadBatch(ctx context.Context, tx *sql.Tx, batchID int64) error {
	var totalCount int
	var readyCount int
	var failedCount int
	var cancelledCount int
	var processingCount int
	var pendingCount int
	if err := tx.QueryRowContext(ctx, `SELECT total_count FROM upload_batches WHERE id = ? FOR UPDATE`, batchID).Scan(&totalCount); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status = ?`, batchID, dbUploadItemStatusReady).Scan(&readyCount); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status IN (?, ?)`, batchID, dbUploadItemStatusUploadFail, dbUploadItemStatusProcessFail).Scan(&failedCount); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status = ?`, batchID, dbUploadItemStatusCancelled).Scan(&cancelledCount); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status = ?`, batchID, dbUploadItemStatusProcessing).Scan(&processingCount); err != nil {
		return err
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM upload_items WHERE upload_batch_id = ? AND status IN (?, ?, ?)`, batchID, dbUploadItemStatusWaiting, dbUploadItemStatusUploading, dbUploadItemStatusUploaded).Scan(&pendingCount); err != nil {
		return err
	}

	status := dbUploadBatchStatusCreated
	completedAt := any(nil)
	activeSlot := any(1)
	now := time.Now().UTC()
	switch {
	case failedCount > 0:
		status = dbUploadBatchStatusPartialFail
	case readyCount == totalCount && totalCount > 0:
		status = dbUploadBatchStatusCompleted
		completedAt = now
		activeSlot = nil
	case processingCount > 0:
		status = dbUploadBatchStatusProcessing
	case pendingCount > 0:
		status = dbUploadBatchStatusCreated
	}
	_, err := tx.ExecContext(ctx, `
UPDATE upload_batches
SET status = ?, active_slot = ?, ready_count = ?, failed_count = ?, cancelled_count = ?, completed_at = COALESCE(?, completed_at)
WHERE id = ?
`, status, activeSlot, readyCount, failedCount, cancelledCount, completedAt, batchID)
	return err
}

func rollbackUnlessCommitted(tx *sql.Tx) {
	_ = tx.Rollback()
}
