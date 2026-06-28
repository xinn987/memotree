// 认证页面流程测试：固定登录、注册和邀请加入对现有 API 的编排。
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { FeedbackProvider } from "../../components/ui/Feedback";
import { SessionProvider } from "../../app/SessionProvider";
import { AuthPage } from "./AuthPage";
import { JoinPage } from "./JoinPage";
import * as authApi from "./auth.api";

vi.mock("./auth.api", async () => {
  const actual = await vi.importActual<typeof import("./auth.api")>("./auth.api");
  return {
    ...actual,
    getSession: vi.fn(),
    login: vi.fn(),
    register: vi.fn(),
    joinFamily: vi.fn(),
  };
});

function renderPage(page: React.ReactNode, path = "/login") {
  return render(
    <FeedbackProvider>
      <SessionProvider initialSession={{ authenticated: false }}>
        <MemoryRouter initialEntries={[path]}>{page}</MemoryRouter>
      </SessionProvider>
    </FeedbackProvider>,
  );
}

describe("authentication pages", () => {
  beforeEach(() => {
    vi.mocked(authApi.getSession).mockResolvedValue({
      authenticated: true,
      user: { id: 1, loginName: "mama", displayName: "妈妈", isSystemAdmin: false },
      families: [],
    });
    vi.mocked(authApi.login).mockResolvedValue({
      id: 1,
      loginName: "mama",
      displayName: "妈妈",
      isSystemAdmin: false,
    });
    vi.mocked(authApi.register).mockResolvedValue({
      id: 1,
      loginName: "mama",
      displayName: "妈妈",
      isSystemAdmin: false,
    });
    vi.mocked(authApi.joinFamily).mockResolvedValue(undefined);
  });

  it("登录表单按现有契约提交登录名和密码", async () => {
    const user = userEvent.setup();
    renderPage(<AuthPage />);

    await user.type(screen.getByRole("textbox", { name: "登录名" }), "mama");
    await user.type(screen.getByLabelText("密码"), "secret1");
    await user.click(screen.getByRole("button", { name: "登录" }));

    expect(authApi.login).toHaveBeenCalledWith({
      loginName: "mama",
      password: "secret1",
    });
  });

  it("注册模式收集显示名并调用现有注册接口", async () => {
    const user = userEvent.setup();
    renderPage(<AuthPage />);

    await user.click(screen.getByRole("button", { name: "创建账号" }));
    await user.type(screen.getByRole("textbox", { name: "家人怎么称呼你" }), "妈妈");
    await user.type(screen.getByRole("textbox", { name: "登录名" }), "mama");
    await user.type(screen.getByLabelText("密码"), "secret1");
    await user.click(screen.getByRole("button", { name: "注册并继续" }));

    expect(authApi.register).toHaveBeenCalledWith({
      displayName: "妈妈",
      loginName: "mama",
      password: "secret1",
    });
  });

  it("邀请加入先注册账号再使用 URL 中的邀请 token", async () => {
    const user = userEvent.setup();
    renderPage(<JoinPage />, "/join?invite=token-1");

    await user.type(screen.getByRole("textbox", { name: "家人怎么称呼你" }), "外婆");
    await user.type(screen.getByRole("textbox", { name: "登录名" }), "waipo");
    await user.type(screen.getByLabelText("设一个密码"), "secret1");
    await user.click(screen.getByRole("button", { name: "加入家人的相册" }));

    expect(authApi.register).toHaveBeenCalled();
    expect(authApi.joinFamily).toHaveBeenCalledWith("token-1");
  });
});
