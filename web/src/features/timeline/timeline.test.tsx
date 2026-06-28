// 时间线行为测试：固定月度故事流、真实图片比例、筛选和分页合并。
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { SessionResponse, TimelineResponse } from "../../api/contracts";
import { SessionProvider } from "../../app/SessionProvider";
import { FeedbackProvider } from "../../components/ui/Feedback";
import { AppRoutes } from "../../app/AppRouter";
import * as timelineApi from "./timeline.api";

vi.mock("./timeline.api", () => ({
  listTimeline: vi.fn(),
}));

const session: SessionResponse = {
  authenticated: true,
  user: { id: 1, loginName: "mama", displayName: "妈妈", isSystemAdmin: false },
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

const firstPage: TimelineResponse = {
  groups: [
    {
      month: "2026-06",
      date: "2026-06-15",
      dateLabel: "06月15日",
      items: [
        {
          id: 11,
          mediaType: "photo",
          capturedAt: "2026-06-15T08:20:00Z",
          uploadedAt: "2026-06-15T09:00:00Z",
          uploadedBy: { id: 1, displayName: "妈妈" },
          thumbnail: { url: "/thumb-11.jpg", contentType: "image/jpeg", width: 600, height: 800, durationMillis: 0 },
          display: { url: "/display-11.jpg", contentType: "image/jpeg", width: 1200, height: 1600, durationMillis: 0 },
        },
      ],
    },
  ],
  nextCursor: "next",
};

function renderTimeline() {
  return render(
    <FeedbackProvider>
      <SessionProvider initialSession={session}>
        <MemoryRouter initialEntries={["/families/7/timeline"]}>
          <AppRoutes />
        </MemoryRouter>
      </SessionProvider>
    </FeedbackProvider>,
  );
}

describe("TimelinePage", () => {
  beforeEach(() => {
    vi.mocked(timelineApi.listTimeline).mockResolvedValue(firstPage);
  });

  it("按月份呈现照片并保留 rendition 比例", async () => {
    renderTimeline();

    expect(await screen.findByRole("heading", { name: /六月/ })).toBeInTheDocument();
    expect(screen.getByRole("img", { name: /妈妈上传的照片/ })).toHaveStyle({ aspectRatio: "600 / 800" });
    expect(screen.getByText("这个月的家庭手记稍后补上")).toBeInTheDocument();
  });

  it("媒体类型筛选使用现有 timeline query", async () => {
    const user = userEvent.setup();
    renderTimeline();
    await screen.findByRole("heading", { name: /六月/ });

    await user.click(screen.getAllByRole("tab", { name: "照片" })[0]);

    await waitFor(() =>
      expect(timelineApi.listTimeline).toHaveBeenLastCalledWith(
        7,
        expect.objectContaining({ mediaType: "photo" }),
        expect.any(AbortSignal),
      ),
    );
  });

  it("加载更早内容时合并已有月份", async () => {
    const user = userEvent.setup();
    vi.mocked(timelineApi.listTimeline)
      .mockResolvedValueOnce(firstPage)
      .mockResolvedValueOnce({
        groups: [
          {
            month: "2026-05",
            date: "2026-05-20",
            dateLabel: "05月20日",
            items: [
              {
                ...firstPage.groups[0].items[0],
                id: 10,
                uploadedAt: "2026-05-20T09:00:00Z",
              },
            ],
          },
        ],
      });
    renderTimeline();

    await user.click(await screen.findByRole("button", { name: "再看更早的" }));

    expect(await screen.findByRole("heading", { name: /五月/ })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /六月/ })).toBeInTheDocument();
  });
});
