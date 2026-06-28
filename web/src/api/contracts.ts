// MemoTree 现有后端契约的前端类型镜像。
// 本次仅重写表现层；接口字段变化必须先修改服务端规格，再更新这里。

export type User = {
  id: number;
  loginName: string;
  displayName: string;
  isSystemAdmin: boolean;
};

export type Family = {
  id: number;
  displayName: string;
  timezone: string;
  role: "admin" | "member";
  memberDisplayName: string;
};

export type SessionResponse =
  | { authenticated: false }
  | { authenticated: true; user: User; families: Family[] | null };

export type InviteStatus = "pending" | "used" | "expired" | "revoked";

export type Invite = {
  id: number;
  familyId: number;
  token: string;
  memberDisplayName: string;
  status: InviteStatus;
  expiresAt: string;
  usedAt?: string | null;
};

export type FamilyMember = {
  id: number;
  familyId: number;
  userId: number;
  displayName: string;
  role: "admin" | "member";
  status: "active" | "removed";
  joinedAt?: string | null;
  removedAt?: string | null;
};

export type UploadBatchStatus = "created" | "uploading" | "processing" | "partially_failed" | "completed" | "stopped";

export type UploadItemStatus =
  | "waiting"
  | "uploading"
  | "uploaded"
  | "processing"
  | "ready"
  | "upload_failed"
  | "processing_failed"
  | "cancelled";

export type UploadBatch = {
  id: number;
  familyId: number;
  createdBy: number;
  status: UploadBatchStatus;
  totalCount: number;
  readyCount: number;
  failedCount: number;
  cancelledCount: number;
  createdAt: string;
};

export type UploadItem = {
  id: number;
  uploadBatchId?: number;
  mediaAssetId?: number | null;
  originalType?: "image_original" | "video_original";
  originalFilename: string;
  contentType: string;
  byteSize: number;
  status: UploadItemStatus;
  errorMessage?: string;
  uploadUrl?: string;
  method?: "PUT";
  expiresAt?: string;
};

export type UploadTask = {
  batch: UploadBatch | null;
  items: UploadItem[];
};

export type UploadIntentResponse = UploadTask & {
  activeExisting: boolean;
};

export type UploadTaskListResponse = {
  tasks: UploadTask[];
};

export type TimelineRendition = {
  url: string;
  contentType: string;
  width: number;
  height: number;
  durationMillis: number;
};

export type TimelineMedia = {
  id: number;
  mediaType: "photo" | "video" | "live_photo";
  capturedAt?: string | null;
  uploadedAt: string;
  uploadedBy: {
    id: number;
    displayName: string;
  };
  thumbnail: TimelineRendition;
  display: TimelineRendition;
};

export type TimelineMediaTypeFilter = "" | TimelineMedia["mediaType"];

export type TimelineGroup = {
  month: string;
  date: string;
  dateLabel: string;
  items: TimelineMedia[];
};

export type TimelineResponse = {
  groups: TimelineGroup[];
  nextCursor?: string;
};

export type MediaDetailResponse = {
  media: TimelineMedia;
};
