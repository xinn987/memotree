# MemoTree

MemoTree 是一个面向家庭成员的私密宝宝相册。当前 MVP 目标是先把账号、家庭邀请、图片上传、异步图片处理、时间线浏览和基础运营能力跑稳。

## Current Status

已完成并归档的 OpenSpec change：

- `family-shared-album-mvp`：账号注册/登录、家庭创建、邀请加入、媒体上传任务、时间线和媒体详情。
- `harden-mvp-family-operations`：管理员成员管理、移除成员保护、媒体软删除、处理失败重试和相关前端入口。

当前进行中的 change：

- `prepare-deployable-runtime`：把项目整理成可部署运行时，包括 API/Worker/Web 镜像、单机 Docker Compose staging 模板、R2/S3 兼容对象存储配置、CI 镜像构建检查和部署 runbook。

MVP 部署路线：

- 云服务器运行 Web、API、Worker、MySQL。
- 对象存储使用 Cloudflare R2 或其他 S3 兼容服务。
- 初版只开放 JPG、PNG、GIF 图片上传；视频处理留到后续扩容阶段。

## Project Layout

```text
web/                 React + Vite + TypeScript PWA
server/api/          Go HTTP API，负责认证、家庭权限、媒体元数据和签名 URL
server/worker/       Go Worker，负责异步生成图片缩略图和展示图
server/migrations/   MySQL schema
deploy/              本地开发和部署相关配置
docs/wiki/           项目 wiki、部署和运维文档
openspec/            OpenSpec changes、requirements 和任务拆分
tools/               跨平台本地检查、运行和镜像构建脚本
```

## Local Development

前置依赖：

- Node.js 22+
- Go 1.24+
- OpenSpec CLI
- Docker Desktop，用于本地 MySQL 和 MinIO

查看可用命令：

```bash
node tools/help.mjs
```

一键启动本地开发环境：

```bash
node tools/dev.mjs
```

完整检查：

```bash
node tools/check.mjs
```

构建本地运行镜像：

```bash
node tools/build-images.mjs
```

更多说明见：

- [本地开发](docs/wiki/local-development.md)
- [发布与持续集成](docs/wiki/release-and-ci.md)
- [部署运行手册](docs/wiki/deployment.md)
