// 邀请管理测试：固定 inline 生成结果、邀请状态和作废操作。
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Invite, SessionResponse } from "../../api/contracts";
import { AppRoutes } from "../../app/AppRouter";
import { SessionProvider } from "../../app/SessionProvider";
import { FeedbackProvider } from "../../components/ui/Feedback";
import * as invitesApi from "./invites.api";
import * as membersApi from "../members/members.api";

vi.mock("./invites.api", () => ({
  listInvites: vi.fn(),
  createInvite: vi.fn(),
  revokeInvite: vi.fn(),
}));

vi.mock("../members/members.api", () => ({
  listMembers: vi.fn(),
}));

const session: SessionResponse = {
  authenticated: true,
  user: { id: 1, loginName: "mama", displayName: "妈妈", isSystemAdmin: false },
  families: [
    { id: 7, displayName: "小满的家", timezone: "Asia/Shanghai", role: "admin", memberDisplayName: "妈妈" },
  ],
};

const pendingInvite: Invite = {
  id: 5,
  familyId: 7,
  token: "token-5",
  memberDisplayName: "外婆",
  status: "pending",
  expiresAt: "2026-07-05T00:00:00Z",
};

function renderInvites() {
  return render(
    <FeedbackProvider>
      <SessionProvider initialSession={session}>
        <MemoryRouter initialEntries={["/families/7/invites"]}>
          <AppRoutes />
        </MemoryRouter>
      </SessionProvider>
    </FeedbackProvider>,
  );
}

describe("InvitesPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(invitesApi.listInvites).mockResolvedValue({ invites: [pendingInvite] });
    vi.mocked(invitesApi.createInvite).mockResolvedValue({ ...pendingInvite, id: 6, token: "new-token" });
    vi.mocked(invitesApi.revokeInvite).mockResolvedValue({ ...pendingInvite, status: "revoked" });
    vi.mocked(membersApi.listMembers).mockResolvedValue({ members: [] });
  });

  it("新邀请在生成位置 inline 展开", async () => {
    const user = userEvent.setup();
    renderInvites();

    await user.type(await screen.findByRole("textbox", { name: "家人的称呼" }), "外婆");
    await user.click(screen.getByRole("button", { name: "生成邀请" }));

    expect(await screen.findByText(/链接已生成/)).toBeInTheDocument();
    expect(screen.getByText(/new-token/)).toBeInTheDocument();
  });

  it("待使用邀请可以调用现有作废接口", async () => {
    const user = userEvent.setup();
    renderInvites();

    await user.click(await screen.findByRole("button", { name: "作废" }));

    expect(invitesApi.revokeInvite).toHaveBeenCalledWith(7, 5);
  });
});
