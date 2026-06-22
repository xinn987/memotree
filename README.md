# MemoTree

MemoTree 是一个面向家庭成员的私密宝宝照片和视频共享相册。MVP 的核心闭环是：账号登录、邀请加入家庭空间、上传精选照片/视频、按宝宝成长时间线浏览。

## Current Status

当前 `family-shared-album-mvp` OpenSpec change 已完成并归档，MVP 闭环已经覆盖：

- 账号注册、账号密码登录、退出登录和会话恢复。
- 创建家庭，创建者自动成为家庭 `admin`。
- `admin` 创建邀请、查看邀请列表、复制待使用邀请、撤销待使用邀请。
- 成员通过有效邀请加入家庭。
- 普通 `member` 不能创建、查看或撤销邀请。
- 家庭成员创建上传任务，并通过短期授权 URL 直接上传精选照片和视频到私有对象存储。
- 上传任务支持进度、部分失败、重试、停止、刷新后恢复和单 active 任务约束。
- Go Worker 生成照片缩略图、展示图，以及视频缩略图、展示视频。
- 首页按月份和日期展示处理完成的家庭媒体时间线，支持分页、媒体类型筛选和月份筛选。
- 媒体详情页支持查看照片和播放视频预览。
- 前端不暴露原文件 object key 或永久公开 URL，原文件下载入口和下载 API 后置。

`harden-mvp-family-operations` change 已补齐 MVP 试用前的关键运营边界：管理员成员管理、移除成员、最后一个 `admin` 保护、媒体软删除、处理失败后的后端重试入口，以及对应的最小前端入口。

已知后续工作包括：实况照片合并；原文文件下载能力；误删恢复和后台对象清理；以及将当前 `web/` 的旧 MVP 工具感界面迁移到 `design/` 中的 Warm Editorial 目标视觉系统。

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
