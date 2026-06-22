package media

import "testing"

func TestWorkerClaimStateMatchesProcessingRetryState(t *testing.T) {
	// API 的 processing retry 会把媒体重置为 pending、上传条目重置为 processing；
	// Worker 必须继续认这组状态，才能从已存原文件重新生成预览。
	if dbRenditionStatusPending != "pending" {
		t.Fatalf("unexpected pending rendition status: %q", dbRenditionStatusPending)
	}
	if dbUploadItemStatusProcessing != "processing" {
		t.Fatalf("unexpected processing upload item status: %q", dbUploadItemStatusProcessing)
	}
}
