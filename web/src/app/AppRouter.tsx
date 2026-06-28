// 正式路由树：当前先提供稳定页面边界，后续 feature 页面逐个替换占位内容。
import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { AuthPage } from "../features/auth/AuthPage";
import { JoinPage } from "../features/auth/JoinPage";
import { OnboardingPage } from "../features/auth/OnboardingPage";
import { TimelinePage } from "../features/timeline/TimelinePage";
import { UploadPage } from "../features/upload/UploadPage";
import { ProtectedFamilyRoute, useFamilyRoute } from "./ProtectedFamilyRoute";
import { useSession } from "./SessionProvider";

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<SessionLanding />} />
      <Route path="/login" element={<AuthRoutePlaceholder />} />
      <Route path="/join" element={<JoinPage />} />
      <Route path="/families/:familyId" element={<ProtectedFamilyRoute />}>
        <Route path="timeline" element={<TimelinePage />} />
        <Route path="upload" element={<UploadPage />} />
        <Route path="invites" element={<FeaturePlaceholder title="邀请家人加入" />} />
        <Route path="members" element={<FeaturePlaceholder title="家里的成员" />} />
        <Route path="media/:mediaId" element={<FeaturePlaceholder title="照片详情" />} />
        <Route index element={<Navigate to="timeline" replace />} />
      </Route>
      <Route path="*" element={<SessionLanding />} />
    </Routes>
  );
}

function SessionLanding() {
  const { loading, user, families } = useSession();
  if (loading) {
    return null;
  }
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  if (families.length === 0) {
    return <OnboardingPage />;
  }
  return <Navigate to={`/families/${families[0].id}/timeline`} replace />;
}

function AuthRoutePlaceholder() {
  const location = useLocation();
  const { loading, user, families } = useSession();
  if (!loading && user) {
    const requestedPath = (location.state as { from?: string } | null)?.from;
    return <Navigate to={requestedPath || (families[0] ? `/families/${families[0].id}/timeline` : "/")} replace />;
  }
  return <AuthPage />;
}

function FeaturePlaceholder({ title }: { title: string }) {
  const { family } = useFamilyRoute();
  return (
    <main className="page-shell page-shell--content">
      <section className="route-placeholder">
        <span className="eyebrow-brand">{family.displayName}</span>
        <h1>{title}</h1>
      </section>
    </main>
  );
}
