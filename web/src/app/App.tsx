import { Component, FormEvent, ReactNode, useEffect, useMemo, useState } from "react";
import { Check, Copy, Home, KeyRound, Link as LinkIcon, LogOut, Plus, RefreshCw, Users, XCircle } from "lucide-react";

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

  useEffect(() => {
    void refreshSession();
  }, []);

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
        setUser(session.user);
        setFamilies(visibleFamilies);
        setSelectedFamilyId((current) => current ?? visibleFamilies[0]?.id ?? null);
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
            onChange={(event) => setSelectedFamilyId(Number(event.target.value))}
          >
            {families.map((family) => (
              <option key={family.id} value={family.id}>
                {family.displayName}
              </option>
            ))}
          </select>
        </label>
      )}

      {selectedFamily ? (
        <FamilyHome
          family={selectedFamily}
          onMessage={setMessage}
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

function FamilyHome({
  family,
  onMessage,
}: {
  family: Family;
  onMessage: (value: string) => void;
}) {
  const [memberDisplayName, setMemberDisplayName] = useState("");
  const [inviteURL, setInviteURL] = useState("");
  const [inviteError, setInviteError] = useState("");
  const [inviteBusy, setInviteBusy] = useState(false);
  const [invites, setInvites] = useState<Invite[]>([]);
  const [invitesLoading, setInvitesLoading] = useState(false);
  const [copiedInviteId, setCopiedInviteId] = useState<number | "latest" | null>(null);
  const [revokeBusyId, setRevokeBusyId] = useState<number | null>(null);
  const canManageInvites = family.role === "admin";

  useEffect(() => {
    if (!canManageInvites) {
      setInviteURL("");
      setInviteError("");
      setInvites([]);
      return;
    }
    void loadInvites();
  }, [family.id, canManageInvites]);

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

  return (
    <section className="home-layout">
      <div className="empty-timeline">
        <div className="empty-icon">
          <Users aria-hidden="true" size={28} />
        </div>
        <h2>还没有照片</h2>
        <p>媒体上传会在下一轮实现。</p>
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

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : "请求失败";
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
