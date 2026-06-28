// 成员管理测试：真实改称呼/移除与暂缺管理员转让必须严格区分。
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { FamilyMember, SessionResponse } from "../../api/contracts";
import { AppRoutes } from "../../app/AppRouter";
import { SessionProvider } from "../../app/SessionProvider";
import { FeedbackProvider } from "../../components/ui/Feedback";
import * as membersApi from "./members.api";

vi.mock("./members.api", () => ({
  listMembers: vi.fn(),
  updateMemberName: vi.fn(),
  removeMember: vi.fn(),
}));

const member: FamilyMember = {
  id: 12,
  familyId: 7,
  userId: 2,
  displayName: "爸爸",
  role: "member",
  status: "active",
};

const session: SessionResponse = {
  authenticated: true,
  user: { id: 1, loginName: "mama", displayName: "妈妈", isSystemAdmin: false },
  families: [
    { id: 7, displayName: "小满的家", timezone: "Asia/Shanghai", role: "admin", memberDisplayName: "妈妈" },
  ],
};

function renderMembers() {
  return render(
    <FeedbackProvider>
      <SessionProvider initialSession={session}>
        <MemoryRouter initialEntries={["/families/7/members"]}>
          <AppRoutes />
        </MemoryRouter>
      </SessionProvider>
    </FeedbackProvider>,
  );
}

describe("MembersPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(membersApi.listMembers).mockResolvedValue({ members: [member] });
    vi.mocked(membersApi.updateMemberName).mockResolvedValue({ member: { ...member, displayName: "孩他爸" } });
    vi.mocked(membersApi.removeMember).mockResolvedValue({ member: { ...member, status: "removed" } });
  });

  it("改称呼调用现有 PATCH 接口", async () => {
    const user = userEvent.setup();
    renderMembers();
    const input = await screen.findByRole("textbox", { name: "爸爸的称呼" });

    await user.clear(input);
    await user.type(input, "孩他爸");
    await user.click(screen.getByRole("button", { name: "保存称呼" }));

    expect(membersApi.updateMemberName).toHaveBeenCalledWith(7, 12, "孩他爸");
  });

  it("管理员转让保留 demo 控件但只显示占位反馈", async () => {
    const user = userEvent.setup();
    renderMembers();

    await user.click(await screen.findByRole("button", { name: "设为管理员" }));

    expect(screen.getByRole("status")).toHaveTextContent("管理员转让暂未开放");
    expect(membersApi.updateMemberName).not.toHaveBeenCalled();
    expect(membersApi.removeMember).not.toHaveBeenCalled();
  });
});
