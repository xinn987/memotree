import { Component, FormEvent, ReactNode, useEffect, useMemo, useState } from "react";
import {
  AlertCircle,
  ArrowLeft,
  Check,
  Copy,
  FileImage,
  FileVideo,
  Home,
  KeyRound,
  Link as LinkIcon,
  LogOut,
  Plus,
  RefreshCw,
  RotateCw,
  Square,
  Upload,
  Users,
  XCircle,
} from "lucide-react";

type User = {
  id: number;
  loginName: string;
  displayName: string;
  isSystemAdmin: boolean;
};

type Family = {
  id: number;
  displayName: string;
  timezone: string;
  role: "admin" | "member";
  memberDisplayName: string;
};

type Invite = {
  id: number;
  familyId: number;
  token: string;
  memberDisplayName: string;
  status: "pending" | "used" | "expired" | "revoked";
  expiresAt: string;
  usedAt?: string | null;
};

type UploadBatch = {
  id: number;
  familyId: number;
  status: UploadBatchStatus;
  totalCount: number;
  readyCount: number;
  failedCount: number;
  cancelledCount: number;
  createdAt: string;
};

type UploadBatchStatus = "created" | "uploading" | "processing" | "partially_failed" | "completed" | "stopped";

type UploadItemStatus =
  | "waiting"
  | "uploading"
  | "uploaded"
  | "processing"
  | "ready"
  | "upload_failed"
  | "processing_failed"
  | "cancelled";

type UploadItem = {
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

type UploadTask = {
  batch: UploadBatch | null;
  items: UploadItem[];
};

type UploadIntentResponse = UploadTask & {
  activeExisting: boolean;
};

type UploadTaskListResponse = {
  tasks: UploadTask[];
};

type TimelineRendition = {
  url: string;
  contentType: string;
  width: number;
  height: number;
  durationMillis: number;
};

type TimelineMedia = {
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

type TimelineMediaTypeFilter = "" | TimelineMedia["mediaType"];

type TimelineGroup = {
  month: string;
  date: string;
  dateLabel: string;
  items: TimelineMedia[];
};

type TimelineResponse = {
  groups: TimelineGroup[];
  nextCursor?: string;
};

type MediaDetailResponse = {
  media: TimelineMedia;
};

type AppRoute = {
  familyId: number | null;
  mediaId: number | null;
};

// 上传处理由后端 Worker 异步完成，前端用短轮询把处理结果回填到任务视图。
const uploadTaskPollIntervalMs = 2000;
const timelinePageSize = 30;

type SessionResponse =
  | { authenticated: false }
  | { authenticated: true; user: User; families: Family[] | null };

type AuthMode = "login" | "register";

type AppErrorBoundaryState = {
  error: Error | null;
};

// AppErrorBoundary 兜住 React 渲染期错误，避免运行时异常直接变成整页白屏。
export class AppErrorBoundary extends Component<{ children: ReactNode }, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error): AppErrorBoundaryState {
    return { error };
  }

  componentDidCatch(error: Error) {
    console.error("MemoTree render error", error);
  }

  render() {
    if (this.state.error) {
      return (
        <main className="error-shell">
          <section className="panel error-panel">
            <p className="eyebrow">MemoTree</p>
            <h1>页面渲染失败</h1>
            <p>请刷新页面重试；如果仍然出现，请查看浏览器 Console 里的错误信息。</p>
          </section>
        </main>
      );
    }
    return this.props.children;
  }
}

export function App() {
  const [loading, setLoading] = useState(true);
  const [user, setUser] = useState<User | null>(null);
  const [families, setFamilies] = useState<Family[]>([]);
  const [selectedFamilyId, setSelectedFamilyId] = useState<number | null>(null);
  const [message, setMessage] = useState("");
  const [route, setRoute] = useState<AppRoute>(() => parseAppRoute());

  useEffect(() => {
    void refreshSession();
  }, []);

  useEffect(() => {
    const handlePopState = () => setRoute(parseAppRoute());
    window.addEventListener("popstate", handlePopState);
    return () => window.removeEventListener("popstate", handlePopState);
  }, []);

  useEffect(() => {
    if (route.familyId && families.some((family) => family.id === route.familyId)) {
      setSelectedFamilyId(route.familyId);
    }
  }, [families, route.familyId]);

  const selectedFamily = useMemo(
    () => families.find((family) => family.id === selectedFamilyId) ?? families[0] ?? null,
    [families, selectedFamilyId],
  );

  async function refreshSession() {
    setLoading(true);
    try {
      const session = await request<SessionResponse>("/auth/session");
      if (session.authenticated) {
        // 后端契约要求 families 始终是数组；这里仍做防御，避免异常响应导致白屏。
        const visibleFamilies = session.families ?? [];
        const routedFamilyId =
          route.familyId && visibleFamilies.some((family) => family.id === route.familyId) ? route.familyId : null;
        setUser(session.user);
        setFamilies(visibleFamilies);
        setSelectedFamilyId((current) => routedFamilyId ?? current ?? visibleFamilies[0]?.id ?? null);
      } else {
        setUser(null);
        setFamilies([]);
        setSelectedFamilyId(null);
      }
    } catch (error) {
      setMessage(getErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  async function logout() {
    await request("/auth/logout", { method: "POST" });
    setUser(null);
    setFamilies([]);
    setSelectedFamilyId(null);
  }

  function navigateTo(path: string) {
    window.history.pushState({}, "", path);
    setRoute(parseAppRoute());
  }

  if (loading) {
    return <main className="shell">加载中...</main>;
  }

  if (!user) {
    return (
      <main className="auth-shell">
        <AuthPanel
          onDone={async () => {
            setMessage("");
            await refreshSession();
          }}
        />
        {message && <p className="form-message">{message}</p>}
      </main>
    );
  }

  return (
    <main className="shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">MemoTree</p>
          <h1>{selectedFamily?.displayName ?? "家庭空间"}</h1>
        </div>
        <button className="icon-button" type="button" onClick={logout} aria-label="退出登录">
          <LogOut aria-hidden="true" size={21} />
        </button>
      </header>

      <section className="account-bar">
        <span>{user.displayName}</span>
        {user.isSystemAdmin && <strong>初始管理员</strong>}
      </section>

      {families.length > 1 && (
        <label className="field">
          <span>家庭</span>
          <select
            value={selectedFamily?.id ?? ""}
            onChange={(event) => {
              const familyID = Number(event.target.value);
              setSelectedFamilyId(familyID);
              navigateTo(`/families/${familyID}/timeline`);
            }}
          >
            {families.map((family) => (
              <option key={family.id} value={family.id}>
                {family.displayName}
              </option>
            ))}
          </select>
        </label>
      )}

      {selectedFamily && route.mediaId && route.familyId === selectedFamily.id ? (
        <MediaDetailPage
          family={selectedFamily}
          mediaId={route.mediaId}
          onBack={() => navigateTo(`/families/${selectedFamily.id}/timeline`)}
          onMessage={setMessage}
        />
      ) : selectedFamily ? (
        <FamilyHome
          family={selectedFamily}
          onMessage={setMessage}
          onOpenMedia={(mediaId) => navigateTo(`/families/${selectedFamily.id}/media/${mediaId}`)}
        />
      ) : (
        <Onboarding onChanged={refreshSession} onMessage={setMessage} />
      )}

      {message && <p className="form-message">{message}</p>}
    </main>
  );
}

function AuthPanel({ onDone }: { onDone: () => Promise<void> }) {
  const [mode, setMode] = useState<AuthMode>("login");
  const [loginName, setLoginName] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function submit(event: FormEvent) {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      await request<User>(mode === "login" ? "/auth/login" : "/auth/register", {
        method: "POST",
        body: JSON.stringify({ loginName, password, displayName }),
      });
      await onDone();
    } catch (error) {
      setError(getErrorMessage(error));
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="panel auth-panel">
      <div className="brand-mark">
        <Home aria-hidden="true" size={28} />
      </div>
      <div>
        <p className="eyebrow">MemoTree</p>
        <h1>家庭相册</h1>
      </div>

      <div className="segmented" role="tablist" aria-label="登录方式">
        <button className={mode === "login" ? "active" : ""} type="button" onClick={() => setMode("login")}>
          登录
        </button>
        <button className={mode === "register" ? "active" : ""} type="button" onClick={() => setMode("register")}>
          注册
        </button>
      </div>

      <form className="form" onSubmit={submit}>
        <label className="field">
          <span>登录名</span>
          <input value={loginName} onChange={(event) => setLoginName(event.target.value)} autoComplete="username" />
        </label>
        <label className="field">
          <span>密码</span>
          <input
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            type="password"
            autoComplete={mode === "login" ? "current-password" : "new-password"}
          />
        </label>
        {mode === "register" && (
          <label className="field">
            <span>显示名</span>
            <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} autoComplete="name" />
          </label>
        )}
        {error && <p className="form-message">{error}</p>}
        <button className="primary-button" type="submit" disabled={busy}>
          <KeyRound aria-hidden="true" size={18} />
          {mode === "login" ? "登录" : "创建账号"}
        </button>
      </form>
    </section>
  );
}

function Onboarding({ onChanged, onMessage }: { onChanged: () => Promise<void>; onMessage: (value: string) => void }) {
  const [familyName, setFamilyName] = useState("");
  const [inviteToken, setInviteToken] = useState(readInviteTokenFromURL());

  async function createFamily(event: FormEvent) {
    event.preventDefault();
    await request<Family>("/families", {
      method: "POST",
      body: JSON.stringify({ displayName: familyName }),
    });
    onMessage("");
    await onChanged();
  }

  async function joinInvite(event: FormEvent) {
    event.preventDefault();
    await request(`/invites/${encodeURIComponent(inviteToken)}/join`, {
      method: "POST",
      body: JSON.stringify({}),
    });
    onMessage("");
    await onChanged();
  }

  return (
    <section className="split-layout">
      <form className="panel form" onSubmit={createFamily}>
        <div className="panel-title">
          <Plus aria-hidden="true" size={20} />
          <h2>创建家庭</h2>
        </div>
        <label className="field">
          <span>家庭名称</span>
          <input value={familyName} onChange={(event) => setFamilyName(event.target.value)} />
        </label>
        <button className="primary-button" type="submit">
          创建
        </button>
      </form>

      <form className="panel form" onSubmit={joinInvite}>
        <div className="panel-title">
          <LinkIcon aria-hidden="true" size={20} />
          <h2>加入家庭</h2>
        </div>
        <label className="field">
          <span>邀请 token</span>
          <input value={inviteToken} onChange={(event) => setInviteToken(event.target.value)} />
        </label>
        <button className="secondary-button" type="submit">
          加入
        </button>
      </form>
    </section>
  );
}

function MediaDetailPage({
  family,
  mediaId,
  onBack,
  onMessage,
}: {
  family: Family;
  mediaId: number;
  onBack: () => void;
  onMessage: (value: string) => void;
}) {
  const [media, setMedia] = useState<TimelineMedia | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");
    setMedia(null);

    async function loadMediaDetail() {
      try {
        const detail = await request<MediaDetailResponse>(`/families/${family.id}/media/${mediaId}`);
        if (!active) {
          return;
        }
        setMedia(detail.media);
        onMessage("");
      } catch (error) {
        if (active) {
          setError(getErrorMessage(error));
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    }

    void loadMediaDetail();
    return () => {
      active = false;
    };
  }, [family.id, mediaId, onMessage]);

  return (
    <section className="media-detail-shell">
      <button className="text-button detail-back" type="button" onClick={onBack}>
        <ArrowLeft aria-hidden="true" size={17} />
        返回时间线
      </button>

      {loading && (
        <div className="panel media-detail-empty">
          <p>正在读取媒体详情...</p>
        </div>
      )}

      {error && (
        <div className="panel media-detail-empty">
          <p className="upload-error">
            <AlertCircle aria-hidden="true" size={16} />
            {error}
          </p>
        </div>
      )}

      {media && (
        <>
          <div className="media-detail-viewer">
            {isVideoMedia(media) ? (
              <video controls src={media.display.url} poster={media.thumbnail.url} preload="metadata" />
            ) : (
              <img
                src={media.display.url}
                alt={`${media.uploadedBy.displayName || "家人"} 上传的照片`}
                style={{ aspectRatio: renditionAspectRatio(media.display) }}
              />
            )}
          </div>

          <section className="panel media-detail-meta">
            <div className="panel-title">
              {isVideoMedia(media) ? <FileVideo aria-hidden="true" size={19} /> : <FileImage aria-hidden="true" size={19} />}
              <h2>{isVideoMedia(media) ? "视频详情" : "照片详情"}</h2>
            </div>
            <dl>
              <div>
                <dt>上传成员</dt>
                <dd>{media.uploadedBy.displayName || "家人"}</dd>
              </div>
              <div>
                <dt>拍摄时间</dt>
                <dd>{media.capturedAt ? formatDateTime(media.capturedAt) : "未记录"}</dd>
              </div>
              <div>
                <dt>上传时间</dt>
                <dd>{formatDateTime(media.uploadedAt)}</dd>
              </div>
              <div>
                <dt>展示规格</dt>
                <dd>{formatRenditionSize(media.display)}</dd>
              </div>
            </dl>
          </section>
        </>
      )}
    </section>
  );
}

function FamilyHome({
  family,
  onMessage,
  onOpenMedia,
}: {
  family: Family;
  onMessage: (value: string) => void;
  onOpenMedia: (mediaId: number) => void;
}) {
  const [memberDisplayName, setMemberDisplayName] = useState("");
  const [inviteURL, setInviteURL] = useState("");
  const [inviteError, setInviteError] = useState("");
  const [inviteBusy, setInviteBusy] = useState(false);
  const [invites, setInvites] = useState<Invite[]>([]);
  const [invitesLoading, setInvitesLoading] = useState(false);
  const [copiedInviteId, setCopiedInviteId] = useState<number | "latest" | null>(null);
  const [revokeBusyId, setRevokeBusyId] = useState<number | null>(null);
  const [uploadTask, setUploadTask] = useState<UploadTask>({ batch: null, items: [] });
  const [recentUploadTasks, setRecentUploadTasks] = useState<UploadTask[]>([]);
  const [uploadBusy, setUploadBusy] = useState(false);
  const [uploadError, setUploadError] = useState("");
  const [uploadProgress, setUploadProgress] = useState<Record<number, number>>({});
  const [uploadingItemIds, setUploadingItemIds] = useState<Set<number>>(() => new Set());
  const [localFilesByItemId, setLocalFilesByItemId] = useState<Record<number, File>>({});
  const [timelineGroups, setTimelineGroups] = useState<TimelineGroup[]>([]);
  const [timelineCursor, setTimelineCursor] = useState<string | null>(null);
  const [timelineMediaTypeFilter, setTimelineMediaTypeFilter] = useState<TimelineMediaTypeFilter>("");
  const [timelineMonthFilter, setTimelineMonthFilter] = useState("");
  const [timelineLoading, setTimelineLoading] = useState(false);
  const [timelineLoadingMore, setTimelineLoadingMore] = useState(false);
  const [timelineError, setTimelineError] = useState("");
  const canManageInvites = family.role === "admin";

  useEffect(() => {
    setUploadTask({ batch: null, items: [] });
    setRecentUploadTasks([]);
    setUploadError("");
    setUploadProgress({});
    setLocalFilesByItemId({});
    setTimelineMediaTypeFilter("");
    setTimelineMonthFilter("");
    void loadUploadTasks();
  }, [family.id]);

  useEffect(() => {
    setTimelineGroups([]);
    setTimelineCursor(null);
    setTimelineError("");
    void loadTimeline();
  }, [family.id, timelineMediaTypeFilter, timelineMonthFilter]);

  useEffect(() => {
    if (!canManageInvites) {
      setInviteURL("");
      setInviteError("");
      setInvites([]);
      return;
    }
    void loadInvites();
  }, [family.id, canManageInvites]);

  useEffect(() => {
    if (!shouldPollUploadTask(uploadTask)) {
      return;
    }
    const timer = window.setInterval(() => {
      void loadUploadTasks();
    }, uploadTaskPollIntervalMs);
    return () => window.clearInterval(timer);
  }, [family.id, uploadTask.batch?.id, uploadTask.batch?.status]);

  async function loadUploadTasks() {
    try {
      const [activeTask, recentTasks] = await Promise.all([
        request<UploadTask>(`/families/${family.id}/uploads/active`),
        request<UploadTaskListResponse>(`/families/${family.id}/uploads/recent`),
      ]);
      const nextActiveTask = normalizeUploadTask(activeTask);
      const nextRecentTasks = (recentTasks.tasks ?? []).map(normalizeUploadTask);
      setUploadTask(nextActiveTask);
      setRecentUploadTasks(nextRecentTasks);
      // Worker 完成后 active 查询会清空；这里顺手刷新时间线，让新预览不必等用户手动刷新页面。
      if (!shouldPollUploadTask(nextActiveTask) && nextRecentTasks.some((task) => task.batch?.status === "completed")) {
        void loadTimeline();
      }
    } catch (error) {
      setUploadError(getErrorMessage(error));
    }
  }

  async function loadTimeline(cursor?: string) {
    const append = Boolean(cursor);
    if (append) {
      setTimelineLoadingMore(true);
    } else {
      setTimelineLoading(true);
    }
    setTimelineError("");
    try {
      const query = new URLSearchParams({ limit: String(timelinePageSize) });
      if (cursor) {
        query.set("cursor", cursor);
      }
      if (timelineMediaTypeFilter) {
        query.set("mediaType", timelineMediaTypeFilter);
      }
      if (timelineMonthFilter) {
        query.set("month", timelineMonthFilter);
      }
      const timeline = await request<TimelineResponse>(`/families/${family.id}/timeline?${query.toString()}`);
      setTimelineGroups((current) => (append ? mergeTimelineGroups(current, timeline.groups ?? []) : timeline.groups ?? []));
      setTimelineCursor(timeline.nextCursor ?? null);
    } catch (error) {
      setTimelineError(getErrorMessage(error));
    } finally {
      if (append) {
        setTimelineLoadingMore(false);
      } else {
        setTimelineLoading(false);
      }
    }
  }

  async function loadInvites() {
    if (!canManageInvites) {
      return;
    }
    setInvitesLoading(true);
    setInviteError("");
    try {
      const data = await request<{ invites: Invite[] }>(`/families/${family.id}/invites`);
      setInvites(data.invites);
    } catch (error) {
      setInviteError(getErrorMessage(error));
    } finally {
      setInvitesLoading(false);
    }
  }

  async function createInvite(event: FormEvent) {
    event.preventDefault();
    setInviteBusy(true);
    setInviteError("");
    setCopiedInviteId(null);
    try {
      const invite = await request<Invite>(`/families/${family.id}/invites`, {
        method: "POST",
        body: JSON.stringify({ memberDisplayName }),
      });
      const nextURL = inviteLink(invite.token);
      setInviteURL(nextURL);
      setInvites((current) => [invite, ...current.filter((item) => item.id !== invite.id)]);
      setMemberDisplayName("");
      onMessage("");
    } catch (error) {
      setInviteError(getErrorMessage(error));
    } finally {
      setInviteBusy(false);
    }
  }

  async function copyInviteURL(value: string, id: number | "latest") {
    if (!value) {
      return;
    }
    await navigator.clipboard.writeText(value);
    setCopiedInviteId(id);
  }

  async function revokeInvite(inviteId: number) {
    setRevokeBusyId(inviteId);
    setInviteError("");
    try {
      const revoked = await request<Invite>(`/families/${family.id}/invites/${inviteId}`, {
        method: "DELETE",
      });
      setInvites((current) => current.map((item) => (item.id === inviteId ? revoked : item)));
      if (copiedInviteId === inviteId) {
        setCopiedInviteId(null);
      }
    } catch (error) {
      setInviteError(getErrorMessage(error));
    } finally {
      setRevokeBusyId(null);
    }
  }

  async function selectFiles(files: File[]) {
    if (files.length === 0) {
      return;
    }
    setUploadBusy(true);
    setUploadError("");
    setUploadProgress({});
    try {
      const intent = await request<UploadIntentResponse>(`/families/${family.id}/media/upload-intents`, {
        method: "POST",
        body: JSON.stringify({
          files: files.map((file) => ({
            filename: file.name,
            contentType: file.type || fallbackContentType(file.name),
            byteSize: file.size,
          })),
        }),
      });
      const nextTask = normalizeUploadTask(intent);
      setUploadTask(nextTask);
      setRecentUploadTasks((current) => upsertUploadTask(current, nextTask));
      if (intent.activeExisting) {
        setUploadError("已有进行中的上传任务，请先处理当前任务。");
        return;
      }

      const fileMap: Record<number, File> = {};
      nextTask.items.forEach((item, index) => {
        if (files[index]) {
          fileMap[item.id] = files[index];
        }
      });
      setLocalFilesByItemId((current) => ({ ...current, ...fileMap }));
      const batchID = nextTask.batch?.id;
      for (const item of nextTask.items) {
        const file = fileMap[item.id];
        if (file && item.uploadUrl && batchID) {
          await uploadOriginalFile(item, file, batchID);
        }
      }
    } catch (error) {
      setUploadError(getErrorMessage(error));
    } finally {
      setUploadBusy(false);
    }
  }

  async function uploadOriginalFile(item: UploadItem, file: File, batchID: number) {
    markUploading(item.id, true);
    setUploadProgress((current) => ({ ...current, [item.id]: 0 }));
    try {
      await putFile(item, file, (progress) => {
        setUploadProgress((current) => ({ ...current, [item.id]: progress }));
      });
      const completed = await request<{ batch: UploadBatch; item: UploadItem }>(
        `/families/${family.id}/uploads/${batchID}/items/${item.id}/complete-upload`,
        { method: "POST", body: JSON.stringify({}) },
      );
      setUploadTask((current) => upsertUploadItem(current, completed.batch, completed.item));
      setRecentUploadTasks((current) => upsertUploadTaskItem(current, completed.batch, completed.item));
      setUploadProgress((current) => ({ ...current, [item.id]: 100 }));
    } catch (error) {
      await markUploadFailed(item, getErrorMessage(error), batchID);
    } finally {
      markUploading(item.id, false);
    }
  }

  async function markUploadFailed(item: UploadItem, message: string, fallbackBatchID?: number) {
    const batchID = fallbackBatchID ?? uploadTask.batch?.id ?? item.uploadBatchId;
    if (!batchID) {
      setUploadError(message);
      return;
    }
    try {
      const failed = await request<{ batch: UploadBatch; item: UploadItem }>(
        `/families/${family.id}/uploads/${batchID}/items/${item.id}/fail-upload`,
        { method: "POST", body: JSON.stringify({ errorMessage: message }) },
      );
      setUploadTask((current) => upsertUploadItem(current, failed.batch, failed.item));
      setRecentUploadTasks((current) => upsertUploadTaskItem(current, failed.batch, failed.item));
    } catch {
      setUploadError(message);
    }
  }

  async function retryUpload(item: UploadItem) {
    const batchID = uploadTask.batch?.id ?? item.uploadBatchId;
    const file = localFilesByItemId[item.id];
    if (!batchID || !file) {
      setUploadError("当前浏览器没有这个文件，请重新选择后上传。");
      return;
    }
    setUploadBusy(true);
    setUploadError("");
    try {
      const retry = await request<{ batch: UploadBatch; item: UploadItem }>(
        `/families/${family.id}/uploads/${batchID}/items/${item.id}/retry-upload`,
        { method: "POST", body: JSON.stringify({}) },
      );
      setUploadTask((current) => upsertUploadItem(current, retry.batch, retry.item));
      setRecentUploadTasks((current) => upsertUploadTaskItem(current, retry.batch, retry.item));
      await uploadOriginalFile(retry.item, file, retry.batch.id);
    } catch (error) {
      setUploadError(getErrorMessage(error));
    } finally {
      setUploadBusy(false);
    }
  }

  async function stopUploadTask() {
    if (!uploadTask.batch) {
      return;
    }
    setUploadBusy(true);
    setUploadError("");
    try {
      const stopped = await request<UploadTask>(`/families/${family.id}/uploads/${uploadTask.batch.id}/stop`, {
        method: "POST",
        body: JSON.stringify({}),
      });
      const nextTask = normalizeUploadTask(stopped);
      setUploadTask(nextTask);
      setRecentUploadTasks((current) => upsertUploadTask(current, nextTask));
    } catch (error) {
      setUploadError(getErrorMessage(error));
    } finally {
      setUploadBusy(false);
    }
  }

  function markUploading(itemID: number, uploading: boolean) {
    setUploadingItemIds((current) => {
      const next = new Set(current);
      if (uploading) {
        next.add(itemID);
      } else {
        next.delete(itemID);
      }
      return next;
    });
  }

  return (
    <section className="home-layout">
      <div className="main-column">
        <UploadPanel
          task={uploadTask}
          busy={uploadBusy}
          error={uploadError}
          progress={uploadProgress}
          uploadingItemIds={uploadingItemIds}
          localFilesByItemId={localFilesByItemId}
          recentTasks={recentUploadTasks}
          onFilesSelected={selectFiles}
          onRetry={retryUpload}
          onStop={stopUploadTask}
          onRefresh={loadUploadTasks}
        />

        <TimelinePanel
          groups={timelineGroups}
          loading={timelineLoading}
          loadingMore={timelineLoadingMore}
          error={timelineError}
          mediaTypeFilter={timelineMediaTypeFilter}
          monthFilter={timelineMonthFilter}
          onMediaTypeFilterChange={setTimelineMediaTypeFilter}
          onMonthFilterChange={setTimelineMonthFilter}
          onRefresh={loadTimeline}
          onLoadMore={() => (timelineCursor ? loadTimeline(timelineCursor) : Promise.resolve())}
          hasNextPage={Boolean(timelineCursor)}
          onOpenMedia={onOpenMedia}
        />
      </div>

      {canManageInvites && (
        <form className="panel form" onSubmit={createInvite}>
          <div className="panel-title">
            <LinkIcon aria-hidden="true" size={20} />
            <h2>邀请家人</h2>
          </div>
          <label className="field">
            <span>成员显示名</span>
            <input value={memberDisplayName} onChange={(event) => setMemberDisplayName(event.target.value)} />
          </label>
          <button className="secondary-button" type="submit" disabled={inviteBusy}>
            {inviteBusy ? "生成中..." : "生成邀请"}
          </button>
          {inviteError && <p className="form-message">{inviteError}</p>}
          {inviteURL && (
            <div className="invite-box">
              <span>邀请链接</span>
              <code>{inviteURL}</code>
              <button className="copy-button" type="button" onClick={() => copyInviteURL(inviteURL, "latest")}>
                {copiedInviteId === "latest" ? <Check aria-hidden="true" size={17} /> : <Copy aria-hidden="true" size={17} />}
                {copiedInviteId === "latest" ? "已复制" : "复制链接"}
              </button>
            </div>
          )}
          <div className="invite-list-header">
            <h3>最近邀请</h3>
            <button className="text-button" type="button" onClick={loadInvites} disabled={invitesLoading}>
              <RefreshCw aria-hidden="true" size={16} />
              {invitesLoading ? "刷新中" : "刷新"}
            </button>
          </div>
          <div className="invite-list">
            {invites.length === 0 ? (
              <p className="muted-text">还没有邀请记录</p>
            ) : (
              invites.map((invite) => {
                const link = invite.token ? inviteLink(invite.token) : "";
                const copyable = invite.status === "pending" && link !== "";
                return (
                  <div className="invite-row" key={invite.id}>
                    <div>
                      <strong>{invite.memberDisplayName || "未填写"}</strong>
                      <span>{inviteStatusText(invite.status)} · {formatDateTime(invite.expiresAt)}</span>
                    </div>
                    {invite.status === "pending" ? (
                      <div className="invite-actions">
                        {copyable && (
                          <button className="copy-button" type="button" onClick={() => copyInviteURL(link, invite.id)}>
                            {copiedInviteId === invite.id ? <Check aria-hidden="true" size={17} /> : <Copy aria-hidden="true" size={17} />}
                            {copiedInviteId === invite.id ? "已复制" : "复制"}
                          </button>
                        )}
                        <button className="danger-button" type="button" onClick={() => revokeInvite(invite.id)} disabled={revokeBusyId === invite.id}>
                          <XCircle aria-hidden="true" size={17} />
                          {revokeBusyId === invite.id ? "撤销中" : "撤销"}
                        </button>
                      </div>
                    ) : (
                      <span className="invite-unavailable">不可复制</span>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </form>
      )}
    </section>
  );
}

function TimelinePanel({
  groups,
  loading,
  loadingMore,
  error,
  mediaTypeFilter,
  monthFilter,
  onMediaTypeFilterChange,
  onMonthFilterChange,
  onRefresh,
  onLoadMore,
  hasNextPage,
  onOpenMedia,
}: {
  groups: TimelineGroup[];
  loading: boolean;
  loadingMore: boolean;
  error: string;
  mediaTypeFilter: TimelineMediaTypeFilter;
  monthFilter: string;
  onMediaTypeFilterChange: (value: TimelineMediaTypeFilter) => void;
  onMonthFilterChange: (value: string) => void;
  onRefresh: () => Promise<void>;
  onLoadMore: () => Promise<void>;
  hasNextPage: boolean;
  onOpenMedia: (mediaId: number) => void;
}) {
  const hasItems = groups.some((group) => group.items.length > 0);

  return (
    <section className="timeline-section">
      <div className="timeline-head">
        <div>
          <p className="eyebrow">家庭时间线</p>
          <h2>照片</h2>
        </div>
        <button className="text-button" type="button" onClick={() => void onRefresh()} disabled={loading}>
          <RefreshCw aria-hidden="true" size={16} />
          {loading ? "刷新中" : "刷新"}
        </button>
      </div>

      <div className="timeline-filters">
        <label className="field">
          <span>媒体类型</span>
          <select value={mediaTypeFilter} onChange={(event) => onMediaTypeFilterChange(event.target.value as TimelineMediaTypeFilter)}>
            <option value="">全部</option>
            <option value="photo">照片</option>
            <option value="video">视频</option>
            <option value="live_photo">实况照片</option>
          </select>
        </label>
        <label className="field">
          <span>月份</span>
          <input type="month" value={monthFilter} onChange={(event) => onMonthFilterChange(event.target.value)} />
        </label>
      </div>

      {error && (
        <p className="upload-error">
          <AlertCircle aria-hidden="true" size={16} />
          {error}
        </p>
      )}

      {!hasItems ? (
        <div className="empty-timeline">
          <div className="empty-icon">
            <Users aria-hidden="true" size={28} />
          </div>
          <h2>{loading ? "正在读取照片" : "还没有照片"}</h2>
          <p>{loading ? "上传处理完成后会出现在这里。" : "上传完成并生成预览后，家庭时间线会显示在这里。"}</p>
        </div>
      ) : (
        <div className="timeline-groups">
          {groups.map((group) => (
            <section className="timeline-group" key={group.date}>
              <div className="timeline-date">
                <strong>{group.dateLabel || group.date}</strong>
                <span>{group.items.length} 张</span>
              </div>
              <div className="timeline-grid">
                {group.items.map((item) => (
                  <button className="timeline-card" key={item.id} type="button" onClick={() => onOpenMedia(item.id)}>
                    <img
                      src={item.display.url}
                      alt={`${item.uploadedBy.displayName || "家人"} 上传的照片`}
                      loading="lazy"
                      style={{ aspectRatio: renditionAspectRatio(item.display) }}
                    />
                    <div className="timeline-card-meta">
                      <span>{item.uploadedBy.displayName || "家人"}</span>
                      <time dateTime={item.capturedAt ?? item.uploadedAt}>{formatDateTime(item.capturedAt ?? item.uploadedAt)}</time>
                    </div>
                  </button>
                ))}
              </div>
            </section>
          ))}
          {hasNextPage && (
            <button className="secondary-button timeline-more-button" type="button" onClick={() => void onLoadMore()} disabled={loadingMore}>
              {loadingMore ? "加载中..." : "加载更多"}
            </button>
          )}
        </div>
      )}
    </section>
  );
}

function UploadPanel({
  task,
  busy,
  error,
  progress,
  uploadingItemIds,
  localFilesByItemId,
  recentTasks,
  onFilesSelected,
  onRetry,
  onStop,
  onRefresh,
}: {
  task: UploadTask;
  busy: boolean;
  error: string;
  progress: Record<number, number>;
  uploadingItemIds: Set<number>;
  localFilesByItemId: Record<number, File>;
  recentTasks: UploadTask[];
  onFilesSelected: (files: File[]) => Promise<void>;
  onRetry: (item: UploadItem) => Promise<void>;
  onStop: () => Promise<void>;
  onRefresh: () => Promise<void>;
}) {
  const active = task.batch && task.batch.status !== "stopped" && task.batch.status !== "completed";
  const visibleRecentTasks = recentTasks.filter((recentTask) => recentTask.batch && recentTask.batch.id !== task.batch?.id);

  return (
    <section className="panel upload-panel">
      <div className="upload-head">
        <div className="panel-title">
          <Upload aria-hidden="true" size={20} />
          <h2>上传精选</h2>
        </div>
        <button className="text-button" type="button" onClick={onRefresh} disabled={busy}>
          <RefreshCw aria-hidden="true" size={16} />
          刷新
        </button>
      </div>

      <label className="upload-drop">
        <input
          type="file"
          multiple
          accept="image/*,video/*"
          disabled={busy || Boolean(active)}
          onChange={(event) => {
            const files = Array.from(event.target.files ?? []);
            event.target.value = "";
            void onFilesSelected(files);
          }}
        />
        <span className="upload-drop-icon">
          <Upload aria-hidden="true" size={22} />
        </span>
        <strong>{active ? "当前有进行中的上传任务" : "选择照片或视频"}</strong>
        <small>{active ? "完成、停止或刷新当前任务后继续" : "支持多选，文件会直接上传到私有对象存储"}</small>
      </label>

      {task.batch && (
        <div className="upload-summary">
          <div>
            <span>任务状态</span>
            <strong>{uploadBatchStatusText(task.batch.status)}</strong>
          </div>
          <div>
            <span>文件</span>
            <strong>{task.items.length}/{task.batch.totalCount}</strong>
          </div>
          <div>
            <span>失败</span>
            <strong>{task.batch.failedCount}</strong>
          </div>
        </div>
      )}

      {task.items.length > 0 && (
        <div className="upload-list">
          {task.items.map((item) => {
            const isUploading = uploadingItemIds.has(item.id);
            const itemProgress = progress[item.id] ?? statusProgress(item.status);
            const canRetry = item.status === "upload_failed" && Boolean(localFilesByItemId[item.id]);
            return (
              <div className="upload-row" key={item.id}>
                <div className="upload-file-icon">
                  {item.contentType.startsWith("video/") ? <FileVideo aria-hidden="true" size={19} /> : <FileImage aria-hidden="true" size={19} />}
                </div>
                <div className="upload-file-main">
                  <div className="upload-file-title">
                    <strong>{item.originalFilename}</strong>
                    <span>{formatBytes(item.byteSize)}</span>
                  </div>
                  <div className="progress-track" aria-label="上传进度">
                    <span style={{ width: `${itemProgress}%` }} />
                  </div>
                  <div className="upload-file-meta">
                    <span>{isUploading ? "上传中" : uploadItemStatusText(item.status)}</span>
                    {item.errorMessage && <span>{item.errorMessage}</span>}
                  </div>
                </div>
                {item.status === "upload_failed" && (
                  <button className="copy-button compact-button" type="button" onClick={() => void onRetry(item)} disabled={!canRetry || busy}>
                    <RotateCw aria-hidden="true" size={16} />
                    重试
                  </button>
                )}
              </div>
            );
          })}
        </div>
      )}

      {error && (
        <p className="upload-error">
          <AlertCircle aria-hidden="true" size={16} />
          {error}
        </p>
      )}

      {active && (
        <button className="danger-button" type="button" onClick={() => void onStop()} disabled={busy}>
          <Square aria-hidden="true" size={16} />
          停止任务
        </button>
      )}

      {visibleRecentTasks.length > 0 && (
        <div className="recent-upload-section">
          <div className="recent-upload-head">
            <strong>最近上传</strong>
            <span>{visibleRecentTasks.length} 个任务</span>
          </div>
          <div className="recent-upload-list">
            {visibleRecentTasks.map((recentTask) => {
              if (!recentTask.batch) {
                return null;
              }
              return (
                <div className="recent-upload-row" key={recentTask.batch.id}>
                  <div>
                    <strong>{uploadBatchStatusText(recentTask.batch.status)}</strong>
                    <span>{formatDateTime(recentTask.batch.createdAt)}</span>
                  </div>
                  <div>
                    <span>文件 {recentTask.items.length}/{recentTask.batch.totalCount}</span>
                    <span>失败 {recentTask.batch.failedCount}</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </section>
  );
}

async function request<T = unknown>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
      ...init.headers,
    },
  });
  const text = await response.text();
  const data = text ? JSON.parse(text) : {};
  if (!response.ok) {
    throw new Error(data.error ?? "请求失败");
  }
  return data;
}

function putFile(item: UploadItem, file: File, onProgress: (progress: number) => void): Promise<void> {
  return new Promise((resolve, reject) => {
    if (!item.uploadUrl) {
      reject(new Error("缺少上传授权"));
      return;
    }
    const xhr = new XMLHttpRequest();
    xhr.open(item.method ?? "PUT", item.uploadUrl);
    xhr.setRequestHeader("Content-Type", item.contentType || file.type || "application/octet-stream");
    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable && event.total > 0) {
        onProgress(Math.round((event.loaded / event.total) * 100));
      }
    };
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        onProgress(100);
        resolve();
      } else {
        reject(new Error(`对象存储上传失败：${xhr.status}`));
      }
    };
    xhr.onerror = () => reject(new Error("网络中断，上传失败"));
    xhr.onabort = () => reject(new Error("上传已取消"));
    xhr.send(file);
  });
}

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : "请求失败";
}

function normalizeUploadTask(task: UploadTask): UploadTask {
  return {
    batch: task.batch,
    items: task.items ?? [],
  };
}

function upsertUploadItem(current: UploadTask, batch: UploadBatch, item: UploadItem): UploadTask {
  const exists = current.items.some((candidate) => candidate.id === item.id);
  return {
    batch,
    items: exists ? current.items.map((candidate) => (candidate.id === item.id ? item : candidate)) : [...current.items, item],
  };
}

function upsertUploadTask(current: UploadTask[], task: UploadTask): UploadTask[] {
  if (!task.batch) {
    return current;
  }
  const exists = current.some((candidate) => candidate.batch?.id === task.batch?.id);
  const next = exists ? current.map((candidate) => (candidate.batch?.id === task.batch?.id ? task : candidate)) : [task, ...current];
  return next.slice(0, 20);
}

function upsertUploadTaskItem(current: UploadTask[], batch: UploadBatch, item: UploadItem): UploadTask[] {
  const exists = current.some((candidate) => candidate.batch?.id === batch.id);
  if (!exists) {
    return [{ batch, items: [item] }, ...current].slice(0, 20);
  }
  return current.map((candidate) => (candidate.batch?.id === batch.id ? upsertUploadItem(candidate, batch, item) : candidate));
}

function mergeTimelineGroups(current: TimelineGroup[], incoming: TimelineGroup[]) {
  const next = current.map((group) => ({ ...group, items: [...group.items] }));
  const groupIndexByDate = new Map(next.map((group, index) => [group.date, index]));

  incoming.forEach((group) => {
    const existingIndex = groupIndexByDate.get(group.date);
    if (existingIndex === undefined) {
      groupIndexByDate.set(group.date, next.length);
      next.push({ ...group, items: [...group.items] });
      return;
    }

    const existing = next[existingIndex];
    const existingItemIds = new Set(existing.items.map((item) => item.id));
    existing.items.push(...group.items.filter((item) => !existingItemIds.has(item.id)));
  });

  return next;
}

function fallbackContentType(filename: string) {
  const lower = filename.toLowerCase();
  if (lower.endsWith(".mp4")) {
    return "video/mp4";
  }
  if (lower.endsWith(".mov")) {
    return "video/quicktime";
  }
  if (lower.endsWith(".png")) {
    return "image/png";
  }
  if (lower.endsWith(".heic") || lower.endsWith(".heif")) {
    return "image/heic";
  }
  return "image/jpeg";
}

function uploadBatchStatusText(status: UploadBatchStatus) {
  const labels: Record<UploadBatchStatus, string> = {
    created: "等待上传",
    uploading: "上传中",
    processing: "处理中",
    partially_failed: "部分失败",
    completed: "已完成",
    stopped: "已停止",
  };
  return labels[status];
}

function shouldPollUploadTask(task: UploadTask) {
  if (!task.batch) {
    return false;
  }
  return !["completed", "stopped", "partially_failed"].includes(task.batch.status);
}

function uploadItemStatusText(status: UploadItemStatus) {
  const labels: Record<UploadItemStatus, string> = {
    waiting: "等待上传",
    uploading: "上传中",
    uploaded: "已上传",
    processing: "处理中",
    ready: "已完成",
    upload_failed: "上传失败",
    processing_failed: "处理失败",
    cancelled: "已取消",
  };
  return labels[status];
}

function statusProgress(status: UploadItemStatus) {
  if (status === "processing" || status === "uploaded") {
    return 100;
  }
  if (status === "ready") {
    return 100;
  }
  if (status === "upload_failed" || status === "processing_failed" || status === "cancelled") {
    return 100;
  }
  return 0;
}

function formatBytes(value: number) {
  if (value >= 1024 * 1024 * 1024) {
    return `${(value / 1024 / 1024 / 1024).toFixed(1)} GB`;
  }
  if (value >= 1024 * 1024) {
    return `${(value / 1024 / 1024).toFixed(1)} MB`;
  }
  if (value >= 1024) {
    return `${(value / 1024).toFixed(1)} KB`;
  }
  return `${value} B`;
}

function renditionAspectRatio(rendition: TimelineRendition) {
  if (rendition.width > 0 && rendition.height > 0) {
    return `${rendition.width} / ${rendition.height}`;
  }
  return "4 / 3";
}

function isVideoMedia(media: TimelineMedia) {
  return media.mediaType === "video" || media.display.contentType.startsWith("video/");
}

function formatRenditionSize(rendition: TimelineRendition) {
  if (rendition.width > 0 && rendition.height > 0) {
    return `${rendition.width} × ${rendition.height}`;
  }
  return "未记录";
}

function parseAppRoute(): AppRoute {
  const match = window.location.pathname.match(/^\/families\/(\d+)(?:\/(?:timeline|media\/(\d+)))?\/?$/);
  if (!match) {
    return { familyId: null, mediaId: null };
  }
  return {
    familyId: Number(match[1]),
    mediaId: match[2] ? Number(match[2]) : null,
  };
}

function readInviteTokenFromURL() {
  return new URLSearchParams(window.location.search).get("invite") ?? "";
}

function inviteLink(token: string) {
  return `${window.location.origin}${window.location.pathname}?invite=${encodeURIComponent(token)}`;
}

function inviteStatusText(status: Invite["status"]) {
  if (status === "pending") {
    return "待使用";
  }
  if (status === "used") {
    return "已使用";
  }
  if (status === "revoked") {
    return "已撤销";
  }
  return "已过期";
}

function formatDateTime(value: string) {
  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}
