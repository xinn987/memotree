// 应用路由契约测试：固定会话守卫、家庭范围和深链回退行为。
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router-dom";
import { describe, expect, it } from "vitest";
import type { SessionResponse } from "../api/contracts";
import { AppRoutes } from "./AppRouter";
import { SessionProvider } from "./SessionProvider";

const authenticatedSession: SessionResponse = {
  authenticated: true,
  user: {
    id: 1,
    loginName: "mama",
    displayName: "妈妈",
    isSystemAdmin: false,
  },
  families: [
    {
      id: 7,
      displayName: "小满的家",
      timezone: "Asia/Shanghai",
      role: "admin",
      memberDisplayName: "妈妈",
    },
  ],
};

function LocationProbe() {
  const location = useLocation();
  return <output aria-label="当前位置">{location.pathname}</output>;
}

function renderRoute(path: string, session: SessionResponse) {
  return render(
    <SessionProvider initialSession={session}>
      <MemoryRouter initialEntries={[path]}>
        <AppRoutes />
        <LocationProbe />
      </MemoryRouter>
    </SessionProvider>,
  );
}

describe("AppRoutes", () => {
  it("未登录访问家庭深链时回到登录页", async () => {
    renderRoute("/families/7/timeline", { authenticated: false });

    expect(await screen.findByRole("heading", { name: "回到家里的相册" })).toBeInTheDocument();
    expect(screen.getByLabelText("当前位置")).toHaveTextContent("/login");
  });

  it("已登录成员可以打开自己家庭的时间线深链", async () => {
    renderRoute("/families/7/timeline", authenticatedSession);

    expect(await screen.findByRole("heading", { name: "家庭时间线" })).toBeInTheDocument();
    expect(screen.getAllByText("小满的家").length).toBeGreaterThan(0);
    expect(screen.getByLabelText("当前位置")).toHaveTextContent("/families/7/timeline");
  });

  it("不可见家庭会回退到第一个可见家庭", async () => {
    renderRoute("/families/99/upload", authenticatedSession);

    expect(await screen.findByLabelText("当前位置")).toHaveTextContent("/families/7/timeline");
  });

  it("没有家庭时显示创建或加入引导", async () => {
    renderRoute("/", { ...authenticatedSession, families: [] });

    expect(await screen.findByRole("heading", { name: "先建一个家" })).toBeInTheDocument();
  });

  it("多个家庭可以从应用顶栏切换", async () => {
    const user = userEvent.setup();
    renderRoute("/families/7/timeline", {
      ...authenticatedSession,
      families: [
        ...(authenticatedSession.authenticated ? (authenticatedSession.families ?? []) : []),
        {
          id: 8,
          displayName: "外婆的家",
          timezone: "Asia/Shanghai",
          role: "member",
          memberDisplayName: "妈妈",
        },
      ],
    });

    await user.selectOptions(screen.getByRole("combobox", { name: "切换家庭" }), "8");

    expect(screen.getByLabelText("当前位置")).toHaveTextContent("/families/8/timeline");
  });
});
