// 上传页面测试：固定任务状态、两类失败恢复和不支持文件的前置拦截。
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { SessionResponse, UploadTask } from "../../api/contracts";
import { AppRoutes } from "../../app/AppRouter";
import { SessionProvider } from "../../app/SessionProvider";
import { FeedbackProvider } from "../../components/ui/Feedback";
import * as uploadApi from "./upload.api";

vi.mock("./upload.api", async () => {
  const actual = await vi.importActual<typeof import("./upload.api")>("./upload.api");
  return {
    ...actual,
    getActiveUploadTask: vi.fn(),
    listRecentUploadTasks: vi.fn(),
    createUploadIntent: vi.fn(),
    retryProcessing: vi.fn(),
    retryUpload: vi.fn(),
    stopUploadTask: vi.fn(),
    completeUpload: vi.fn(),
    failUpload: vi.fn(),
    putOriginalFile: vi.fn(),
  };
});

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

const activeTask: UploadTask = {
  batch: {
    id: 31,
    familyId: 7,
    createdBy: 1,
    status: "processing",
    totalCount: 2,
    readyCount: 1,
    failedCount: 1,
    cancelledCount: 0,
    createdAt: "2026-06-28T08:00:00Z",
  },
  items: [
    {
      id: 41,
      uploadBatchId: 31,
      originalFilename: "IMG_4521.JPG",
      contentType: "image/jpeg",
      byteSize: 3_145_728,
      status: "ready",
    },
    {
      id: 42,
      uploadBatchId: 31,
      originalFilename: "IMG_4522.JPG",
      contentType: "image/jpeg",
      byteSize: 4_194_304,
      status: "processing_failed",
      errorMessage: "preview failed",
    },
  ],
};

function renderUpload() {
  return render(
    <FeedbackProvider>
      <SessionProvider initialSession={session}>
        <MemoryRouter initialEntries={["/families/7/upload"]}>
          <AppRoutes />
        </MemoryRouter>
      </SessionProvider>
    </FeedbackProvider>,
  );
}

describe("UploadPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(uploadApi.getActiveUploadTask).mockResolvedValue(activeTask);
    vi.mocked(uploadApi.listRecentUploadTasks).mockResolvedValue({ tasks: [activeTask] });
    vi.mocked(uploadApi.retryProcessing).mockResolvedValue({
      batch: activeTask.batch!,
      item: { ...activeTask.items[1], status: "processing" },
    });
  });

  it("按 demo 呈现当前任务和家庭化状态", async () => {
    renderUpload();

    expect(await screen.findByRole("heading", { name: /这批正在传/ })).toBeInTheDocument();
    expect(screen.getByText("已整理好")).toBeInTheDocument();
    expect(screen.getByText("整理失败")).toBeInTheDocument();
  });

  it("处理失败可以调用现有重新整理接口", async () => {
    const user = userEvent.setup();
    renderUpload();

    await user.click(await screen.findByRole("button", { name: "重新整理" }));

    expect(uploadApi.retryProcessing).toHaveBeenCalledWith(7, 31, 42);
  });

  it("视频文件在创建 intent 前被拦截", async () => {
    const user = userEvent.setup();
    renderUpload();
    const input = await screen.findByLabelText("选择照片或视频");

    await user.upload(input, new File(["video"], "birthday.mp4", { type: "video/mp4" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("当前版本暂不支持上传视频");
    expect(uploadApi.createUploadIntent).not.toHaveBeenCalled();
  });
});
