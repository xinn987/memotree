// 应用组合根：错误边界、会话、反馈和浏览器路由只在这里装配一次。
import { Component, type ReactNode } from "react";
import { BrowserRouter } from "react-router-dom";
import { FeedbackProvider } from "../components/ui/Feedback";
import { AppRoutes } from "./AppRouter";
import { SessionProvider } from "./SessionProvider";

type ErrorBoundaryState = {
  error: Error | null;
};

export class AppErrorBoundary extends Component<{ children: ReactNode }, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { error };
  }

  componentDidCatch(error: Error) {
    console.error("MemoTree render error", error);
  }

  render() {
    if (this.state.error) {
      return (
        <main className="fatal-error">
          <span className="eyebrow-brand">MemoTree</span>
          <h1>这一页没有打开</h1>
          <p>刷新后再试一次。如果还是不行，请把页面截图发给家里管事的人。</p>
          <button className="btn btn--primary" type="button" onClick={() => window.location.reload()}>
            刷新页面
          </button>
        </main>
      );
    }
    return this.props.children;
  }
}

export function App() {
  return (
    <FeedbackProvider>
      <SessionProvider>
        <BrowserRouter>
          <AppRoutes />
        </BrowserRouter>
      </SessionProvider>
    </FeedbackProvider>
  );
}
