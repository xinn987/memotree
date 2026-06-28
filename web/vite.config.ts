import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  // 单元与组件测试统一使用浏览器语义环境，避免每个测试文件重复初始化。
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test/setup.ts"],
    restoreMocks: true,
  },
  server: {
    port: 5173,
    proxy: {
      // 开发期只代理 /api，避免 /families/:id/timeline 这类前端历史路由被后端 API 抢走。
      "/api": {
        target: "http://localhost:8080",
        rewrite: (path) => path.replace(/^\/api/, ""),
      },
    },
  },
});
