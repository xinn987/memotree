// 前端测试公共入口：注册可访问性断言，并在每个用例后清理 React DOM。
import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

afterEach(() => {
  cleanup();
});
