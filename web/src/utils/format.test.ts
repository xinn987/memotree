// 纯格式化逻辑测试：固定家庭化文案，避免页面各自翻译后端状态。
import { describe, expect, it } from "vitest";
import { formatBytes, uploadBatchStatusText, uploadItemStatusText } from "./format";

describe("format helpers", () => {
  it("用易读单位显示文件大小", () => {
    expect(formatBytes(0)).toBe("0 B");
    expect(formatBytes(1024)).toBe("1 KB");
    expect(formatBytes(1_572_864)).toBe("1.5 MB");
  });

  it("把上传项目状态翻译成家庭成员能理解的中文", () => {
    expect(uploadItemStatusText("waiting")).toBe("排队中");
    expect(uploadItemStatusText("uploading", 64)).toBe("传了 64%");
    expect(uploadItemStatusText("processing")).toBe("正在整理");
    expect(uploadItemStatusText("ready")).toBe("已整理好");
    expect(uploadItemStatusText("upload_failed")).toBe("没传成");
  });

  it("把整批状态翻译为页面标题语气", () => {
    expect(uploadBatchStatusText("processing")).toBe("整理中");
    expect(uploadBatchStatusText("completed")).toBe("都好了");
    expect(uploadBatchStatusText("partially_failed")).toBe("有一些没传成");
  });
});
