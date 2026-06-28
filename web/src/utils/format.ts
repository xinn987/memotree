// 面向家庭成员的格式化与状态文案。
// 这些函数保持纯净，供页面和测试共同使用。

import type { InviteStatus, TimelineRendition, UploadBatchStatus, UploadItemStatus } from "../api/contracts";

export function formatBytes(value: number): string {
  if (value < 1024) {
    return `${value} B`;
  }
  if (value < 1024 * 1024) {
    return `${roundReadable(value / 1024)} KB`;
  }
  if (value < 1024 * 1024 * 1024) {
    return `${roundReadable(value / (1024 * 1024))} MB`;
  }
  return `${roundReadable(value / (1024 * 1024 * 1024))} GB`;
}

export function uploadItemStatusText(status: UploadItemStatus, progress = 0): string {
  const labels: Record<UploadItemStatus, string> = {
    waiting: "排队中",
    uploading: `传了 ${Math.round(progress)}%`,
    uploaded: "已经传上去",
    processing: "正在整理",
    ready: "已整理好",
    upload_failed: "没传成",
    processing_failed: "整理失败",
    cancelled: "已停下",
  };
  return labels[status];
}

export function uploadBatchStatusText(status: UploadBatchStatus): string {
  const labels: Record<UploadBatchStatus, string> = {
    created: "准备中",
    uploading: "正在传",
    processing: "整理中",
    partially_failed: "有一些没传成",
    completed: "都好了",
    stopped: "已停下",
  };
  return labels[status];
}

export function inviteStatusText(status: InviteStatus): string {
  const labels: Record<InviteStatus, string> = {
    pending: "待使用",
    used: "已加入",
    expired: "已过期",
    revoked: "已作废",
  };
  return labels[status];
}

export function formatDateTime(value: string): string {
  return new Intl.DateTimeFormat("zh-CN", {
    month: "numeric",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

export function renditionAspectRatio(rendition: TimelineRendition): string {
  return rendition.width > 0 && rendition.height > 0 ? `${rendition.width} / ${rendition.height}` : "4 / 3";
}

export function formatRenditionSize(rendition: TimelineRendition): string {
  return rendition.width > 0 && rendition.height > 0 ? `${rendition.width} × ${rendition.height}` : "未记录";
}

function roundReadable(value: number): string {
  return value >= 10 || Number.isInteger(value) ? String(Math.round(value)) : value.toFixed(1);
}
