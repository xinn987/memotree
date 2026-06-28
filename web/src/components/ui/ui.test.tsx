// 共享 UI 契约测试：确保视觉组件不会牺牲可访问语义和占位反馈。
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { Button } from "./Button";
import { Chip } from "./Chip";
import { FeedbackProvider, PlaceholderButton } from "./Feedback";
import { Field } from "./Field";
import { EmptyState, ProgressBar, Skeleton } from "./StatusViews";

describe("shared UI", () => {
  it("加载中的按钮保留名称并阻止重复提交", () => {
    render(<Button loading>保存</Button>);

    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "保存" })).toHaveAttribute("aria-busy", "true");
  });

  it("占位按钮不伪造操作并给出统一反馈", async () => {
    const user = userEvent.setup();
    render(
      <FeedbackProvider>
        <PlaceholderButton capability="downloadOriginal">下载原图</PlaceholderButton>
      </FeedbackProvider>,
    );

    await user.click(screen.getByRole("button", { name: "下载原图" }));

    expect(screen.getByRole("status")).toHaveTextContent("原图下载暂未开放");
  });

  it("字段标签与输入框建立原生关联", () => {
    render(
      <Field label="登录名" hint="手机号或用户名">
        <input name="loginName" />
      </Field>,
    );

    expect(screen.getByRole("textbox", { name: "登录名" })).toHaveAccessibleDescription();
  });

  it("状态标签同时保留文字语义", () => {
    render(<Chip tone="ok">已整理好</Chip>);

    expect(screen.getByText("已整理好")).toHaveClass("chip--ok");
  });

  it("进度、加载骨架和空状态具有可读语义", () => {
    render(
      <>
        <ProgressBar value={64} label="birthday.mp4 上传进度" />
        <Skeleton label="正在读取照片" />
        <EmptyState title="还没有照片">上传几张想给家人看的吧。</EmptyState>
      </>,
    );

    expect(screen.getByRole("progressbar", { name: "birthday.mp4 上传进度" })).toHaveAttribute("aria-valuenow", "64");
    expect(screen.getByLabelText("正在读取照片")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "还没有照片" })).toBeInTheDocument();
  });
});
