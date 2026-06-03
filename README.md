# MemoTree

MemoTree 是一个面向家庭成员的私密宝宝照片和视频共享相册。MVP 的核心闭环是：邀请加入家庭空间、上传照片/视频、快速按时间线浏览、在权限校验后下载原文件。

## Project Layout

```text
web/               React + Vite + TypeScript 移动端 PWA
server/            Go module 根目录
server/api/        Go HTTP API，负责鉴权、家庭权限、媒体元数据和签名授权
server/worker/     Go Worker，负责缩略图和视频封面等异步媒体处理
server/migrations/ MySQL schema migrations
deploy/            本地开发与部署相关配置
docs/              项目 wiki，沉淀产品设计、协议、模块设计和运维方案
openspec/          OpenSpec change、requirements 和任务拆分
```

## Local Development

当前仓库先提交可审查脚手架。本机需要安装 Node.js、Go 和 Docker 后才能运行完整检查。

```powershell
Copy-Item .env.example .env
docker compose -f deploy/docker-compose.dev.yml up -d
```

前端和后端启动方式见 [docs/wiki/local-development.md](D:/CodexProjects/memotree/docs/wiki/local-development.md)。

说明：当前只有一个前端应用，因此 Node.js 依赖边界收在 `web` 内；根目录不放 npm 配置。等未来出现共享 UI 包、API client 或文档站点时，再评估是否引入 npm workspaces。

Go 依赖边界收在 `server` 内；根目录不作为 Go module。API、Worker 以及未来共享 Go 包都从 `server/go.mod` 管理。

## Wiki

项目设计文档从 [docs/wiki/index.md](D:/CodexProjects/memotree/docs/wiki/index.md) 开始。现阶段优先使用仓库内 Markdown，便于和代码、OpenSpec、CI 一起 review；后续文档规模变大后再接 Docusaurus 或 VitePress 发布成静态站点。
