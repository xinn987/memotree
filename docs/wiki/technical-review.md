# 技术选型复审

## Frontend

采用 React + Vite + TypeScript + Tailwind 构建移动端优先 PWA。

理由：

- 一套前端覆盖 iPhone、Android 和桌面浏览器。
- MVP 不需要 Next.js 的 SSR、Server Components 或 API Routes。
- PWA 不能完全替代原生后台同步，但当前明确不做自动同步。

## Backend

采用 Go API。后端职责包括登录会话、家庭权限、邀请、媒体元数据、上传授权、下载授权和时间线查询。

## Data Access

优先采用明确 SQL 或 sqlc 风格，不优先 GORM。家庭权限查询属于安全边界，应该保持查询可读、可测、可审计。

## Database

MVP 使用 MySQL。当前关系模型清晰：用户、家庭、成员、邀请、媒体资产、对象引用和处理状态。

## Storage

第一版可以先试 Cloudflare R2，但代码层面使用 S3-compatible storage adapter，不把业务逻辑写死到 R2。

## Media Processing

Worker 只做 MVP 必需处理：

- 图片缩略图。
- 视频封面。
- 基础 metadata 记录。

视频转码、多码率播放和复杂队列后置。
