// 暂缺后端能力必须集中声明，页面不能自行伪造成功。
import { describe, expect, it } from "vitest";
import { capabilityMessage, isCapabilityAvailable } from "./capabilities";

describe("frontend capabilities", () => {
  it("把暂缺能力标记为不可用", () => {
    expect(isCapabilityAvailable("transferAdmin")).toBe(false);
    expect(isCapabilityAvailable("downloadOriginal")).toBe(false);
    expect(isCapabilityAvailable("editCapturedAt")).toBe(false);
  });

  it("为占位操作提供一致且具体的中文反馈", () => {
    expect(capabilityMessage("transferAdmin")).toBe("管理员转让暂未开放");
    expect(capabilityMessage("downloadOriginal")).toBe("原图下载暂未开放");
    expect(capabilityMessage("adjacentMedia")).toBe("直接打开详情时暂时不能切换前后照片");
  });
});
