// 家庭路由守卫：先验证登录态，再验证 URL 中的家庭是否对当前用户可见。
import { createContext, useContext, type ReactNode } from "react";
import { Navigate, Outlet, useLocation, useParams } from "react-router-dom";
import type { Family, User } from "../api/contracts";
import { AppBar } from "../components/layout/AppBar";
import { InlineError } from "../components/ui/Feedback";
import { OnboardingPage } from "../features/auth/OnboardingPage";
import { useSession } from "./SessionProvider";

export type FamilyRouteContextValue = {
  family: Family;
  user: User;
};

const FamilyRouteContext = createContext<FamilyRouteContextValue | null>(null);

export function ProtectedFamilyRoute() {
  const { familyId } = useParams();
  const location = useLocation();
  const { loading, error, user, families, logout } = useSession();

  if (loading) {
    return <AppLoading />;
  }
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname + location.search }} />;
  }
  if (families.length === 0) {
    return <OnboardingPage />;
  }

  const numericFamilyId = Number(familyId);
  const family = families.find((item) => item.id === numericFamilyId);
  if (!family) {
    return <Navigate to={`/families/${families[0].id}/timeline`} replace />;
  }

  return (
    <FamilyRouteContext.Provider value={{ family, user }}>
      <AppBar family={family} families={families} user={user} onLogout={() => void logout()} />
      {error && (
        <div className="app-message">
          <InlineError>{error}</InlineError>
        </div>
      )}
      <Outlet />
    </FamilyRouteContext.Provider>
  );
}

export function useFamilyRoute(): FamilyRouteContextValue {
  const context = useContext(FamilyRouteContext);
  if (!context) {
    throw new Error("useFamilyRoute 必须在 ProtectedFamilyRoute 内使用");
  }
  return context;
}

export function AppLoading({ children = "正在打开家里的相册…" }: { children?: ReactNode }) {
  return (
    <main className="app-loading" aria-live="polite">
      <span className="app-loading__mark">MemoTree</span>
      <p>{children}</p>
    </main>
  );
}
