/// <reference types="vite/client" />

// 前端构建时可通过该变量覆盖 API 前缀；本地开发默认使用 /api 代理。
interface ImportMetaEnv {
  readonly VITE_API_BASE_PATH?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
