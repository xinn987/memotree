// 会话提供器是前端唯一的登录态与家庭列表来源，feature 不自行缓存 session。
import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import type { Family, SessionResponse, User } from "../api/contracts";
import { getSession, logout as logoutRequest } from "../features/auth/auth.api";

type SessionContextValue = {
  loading: boolean;
  error: string;
  user: User | null;
  families: Family[];
  refresh: () => Promise<void>;
  logout: () => Promise<void>;
  updateFamilyMemberName: (familyId: number, displayName: string) => void;
};

const SessionContext = createContext<SessionContextValue | null>(null);

type SessionProviderProps = {
  children: ReactNode;
  initialSession?: SessionResponse;
};

export function SessionProvider({ children, initialSession }: SessionProviderProps) {
  const [loading, setLoading] = useState(initialSession === undefined);
  const [error, setError] = useState("");
  const [user, setUser] = useState<User | null>(
    initialSession?.authenticated ? initialSession.user : null,
  );
  const [families, setFamilies] = useState<Family[]>(
    initialSession?.authenticated ? (initialSession.families ?? []) : [],
  );

  const applySession = useCallback((session: SessionResponse) => {
    if (session.authenticated) {
      setUser(session.user);
      setFamilies(session.families ?? []);
    } else {
      setUser(null);
      setFamilies([]);
    }
  }, []);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      applySession(await getSession());
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "暂时无法读取登录状态");
    } finally {
      setLoading(false);
    }
  }, [applySession]);

  useEffect(() => {
    if (initialSession === undefined) {
      void refresh();
    }
  }, [initialSession, refresh]);

  const logout = useCallback(async () => {
    await logoutRequest();
    applySession({ authenticated: false });
  }, [applySession]);

  const updateFamilyMemberName = useCallback((familyId: number, displayName: string) => {
    setFamilies((current) =>
      current.map((family) => (family.id === familyId ? { ...family, memberDisplayName: displayName } : family)),
    );
  }, []);

  const value = useMemo<SessionContextValue>(
    () => ({ loading, error, user, families, refresh, logout, updateFamilyMemberName }),
    [error, families, loading, logout, refresh, updateFamilyMemberName, user],
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionContextValue {
  const context = useContext(SessionContext);
  if (!context) {
    throw new Error("useSession 必须在 SessionProvider 内使用");
  }
  return context;
}
