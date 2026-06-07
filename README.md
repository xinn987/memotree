# MemoTree

MemoTree 是一个面向家庭成员的私密宝宝照片和视频共享相册。MVP 的核心闭环是：账号登录、邀请加入家庭空间、上传精选照片/视频、按宝宝成长时间线浏览。

## Project Layout

```text
web/               React + Vite + TypeScript 移动端 PWA
server/            Go module 根目录
server/api/        Go HTTP API，负责认证、家庭权限、媒体元数据和签名授权
server/worker/     Go Worker，负责缩略图、展示图和视频封面等异步媒体处理
server/migrations/ MySQL schema migrations
deploy/            本地开发与部署相关配置
docs/              项目 wiki，沉淀产品设计、协议、模块设计和运维方案
openspec/          OpenSpec change、requirements 和任务拆分
tools/             跨平台本地检查和运行脚本
```

## Local Development

首次拉取项目后，先安装：

- Node.js 22+
- Go 1.24+
- OpenSpec CLI
- Docker Desktop，可选，用于本地 MySQL 和 MinIO

查看可用工具命令：

```bash
node tools/help.mjs
```

一键启动本地开发环境：

```bash
node tools/dev.mjs
```

如果 `8080` 或 `5173` 被旧开发进程占用，脚本会先打印占用端口的 PID 和进程名。确认是旧进程后，可以显式清理端口再启动：

```bash
node tools/dev.mjs --kill-ports
```

统一检查入口：

```bash
node tools/check.mjs
```

API 启动入口：

```bash
node tools/run-api.mjs
node tools/run-api.mjs --mysql
```

脚本会在运行时把 Go/npm 缓存收束到项目内的忽略目录，并设置稳定的 Go module proxy，不需要手动设置系统环境变量。

默认 API 使用内存 store；加 `--mysql` 后会连接本地 Docker MySQL 并自动建表。本地 Docker MySQL 的连接样例见 [.env.example](.env.example)。

更多本地开发说明见 [docs/wiki/local-development.md](docs/wiki/local-development.md)。

## Wiki

项目设计文档从 [docs/wiki/index.md](docs/wiki/index.md) 开始。现阶段优先使用仓库内 Markdown，便于和代码、OpenSpec、CI 一起 review；后续文档规模变大后再评估是否接 Docusaurus 或 VitePress 发布成静态站点。
