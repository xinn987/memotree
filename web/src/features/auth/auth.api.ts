// 认证与家庭入口 API：保持现有服务端路径和请求字段，不在页面中重复拼接。
import type { Family, SessionResponse, User } from "../../api/contracts";
import { requestJSON } from "../../api/client";

export type AuthCredentials = {
  loginName: string;
  password: string;
  displayName?: string;
};

export function getSession(signal?: AbortSignal) {
  return requestJSON<SessionResponse>("/auth/session", { signal });
}

export function login(credentials: AuthCredentials) {
  return requestJSON<User>("/auth/login", {
    method: "POST",
    body: JSON.stringify({
      loginName: credentials.loginName,
      password: credentials.password,
    }),
  });
}

export function register(credentials: Required<AuthCredentials>) {
  return requestJSON<User>("/auth/register", {
    method: "POST",
    body: JSON.stringify(credentials),
  });
}

export function logout() {
  return requestJSON<void>("/auth/logout", { method: "POST" });
}

export function createFamily(displayName: string) {
  return requestJSON<Family>("/families", {
    method: "POST",
    body: JSON.stringify({ displayName }),
  });
}

export function joinFamily(inviteToken: string) {
  return requestJSON<void>(`/invites/${encodeURIComponent(inviteToken)}/join`, {
    method: "POST",
    body: JSON.stringify({}),
  });
}
