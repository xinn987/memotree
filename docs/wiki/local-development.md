# 本地开发

## 前置依赖

- Node.js 22+
- Go 1.24+
- OpenSpec CLI
- Docker Desktop，可选，用于本地 MySQL 和 MinIO

## 工具帮助入口

不确定命令时，先运行：

```bash
node tools/help.mjs
```

`help.mjs` 会按“开发启动、本地依赖、检查、单脚本帮助”分组列出当前可用命令。

## 一键启动

启动 Docker 依赖、API、Worker 和前端：

```bash
node tools/dev.mjs
```

如果只想快速试流程，不使用 MySQL 持久化：

```bash
node tools/dev.mjs --memory
```

`dev.mjs` 会在同一个终端里输出 `[api]`、`[worker]` 和 `[web]` 前缀日志。脚本会先启动 API，等 `http://127.0.0.1:8080/healthz` 返回成功后再启动 Worker 和前端，避免 Vite 代理在 API 未就绪时反复报 `ECONNREFUSED`。按 `Ctrl+C` 会同时停止 API、Worker 和前端进程。

`node tools/dev.mjs --memory` 只启动 API 和前端，不启动 Docker 依赖或 Worker；上传、对象存储和媒体处理流程需要使用普通 MySQL/MinIO 开发模式。

API 默认占用 `8080`，前端默认占用 `5173`。如果端口已经被占用，启动脚本会打印占用端口的 PID 和进程名；确认是旧开发进程后，可以显式清理端口再启动：

```bash
node tools/dev.mjs --kill-ports
```

## 统一检查入口

从任意支持 Node 的 Windows、macOS 或 Linux 机器拉下项目后，在项目根目录执行：

```bash
node tools/check.mjs
```

这个脚本会统一完成：

- 检查 `node`、`npm`、`go`、`openspec` 是否可用。
- 如果 `web/node_modules` 不存在，执行 `npm ci` 安装 lockfile 固定的前端依赖。
- 执行 `go test ./...`。
- 执行 `npm run check`。
- 执行 `npm run build`。
- 执行 `openspec validate family-shared-album-mvp --strict`。

脚本运行时会把缓存目录临时注入给子进程：

```text
GOCACHE       -> .gocache/
GOMODCACHE    -> .gomodcache/
npm cache     -> web/.npm-cache/
GOPROXY       -> https://goproxy.cn,direct
```

这些变量不会写入操作系统环境变量，也不会污染个人机器；规则固化在 `tools/*.mjs` 里。

## 分项检查

只检查后端：

```bash
node tools/test-server.mjs
```

只检查前端：

```bash
node tools/check-web.mjs
```

## 本地依赖服务

启动 MySQL 和 MinIO：

```bash
node tools/dev-up.mjs
```

只启动 MySQL：

本地 Docker MySQL 映射到宿主机 `3307`，容器内部仍是 `3306`，用于避免和开发机已有 MySQL 默认端口冲突。

停止本地依赖服务：

```bash
node tools/dev-down.mjs
```

查看容器状态和日志：

```bash
node tools/dev-status.mjs
node tools/dev-logs.mjs
node tools/dev-logs.mjs --follow
```

`dev-logs.mjs` 只查看 Docker 依赖日志，包括 MySQL 和 MinIO。Go 服务自己的日志由服务进程统一落盘，见下方“服务日志”。

## 启动 API

使用内存 store，重启后数据会丢：

```bash
node tools/run-api.mjs
```

使用本地 Docker MySQL：

```bash
node tools/run-api.mjs --mysql
```

`run-api.mjs` 会自动设置项目内 Go cache、`GOMODCACHE` 和 `GOPROXY`。加 `--mysql` 时会注入：

```text
MYSQL_DSN=memotree:memotree@tcp(127.0.0.1:3307)/memotree?parseTime=true
```

API 检测到 `MYSQL_DSN` 后会连接 MySQL，并自动应用 `server/migrations/0001_initial_schema.sql`。

如果 `8080` 被旧 API 进程占用，可以先看脚本打印的 PID；确认可关闭后运行：

```bash
node tools/run-api.mjs --kill-ports
```

## 服务日志

API、Worker 和本地对象存储初始化命令都使用 Go 侧统一日志模块写文件，同时保留终端输出。默认日志目录是：

```text
server/logs/
```

默认文件名：

```text
server/logs/api.log
server/logs/worker.log
server/logs/init-storage.log
```

可以通过 `LOG_DIR` 覆盖日志目录。排查接口 500、Worker 处理卡住、MinIO bucket 初始化失败时，优先查看这些服务日志；排查 MySQL/MinIO 容器自身问题时，再用 `node tools/dev-logs.mjs`。

## 启动前端

```bash
node tools/run-web.mjs
```

前端默认地址是 `http://localhost:5173`，Vite 会把 `/auth`、`/families`、`/invites` 代理到本地 API。

如果 `5173` 被旧 Vite 进程占用，可以先看脚本打印的 PID；确认可关闭后运行：

```bash
node tools/run-web.mjs --kill-ports
```

## 当前可测试流程

Family Access 当前已经可以在本地完整测试：

1. 运行 `node tools/dev.mjs --kill-ports` 启动本地依赖、API、Worker 和前端。
2. 打开 `http://localhost:5173`。
3. 注册第一个账号；第一个账号会成为系统初始管理员。
4. 创建一个家庭；创建者会自动成为该家庭的 `admin`。
5. 在家庭页生成邀请，复制邀请链接。
6. 用无痕窗口或另一个浏览器打开邀请链接。
7. 注册或登录另一个账号，并使用 URL 中的邀请加入家庭。
8. 回到管理员页面刷新“最近邀请”，可以看到邀请变成“已使用”。

邀请管理当前支持：

- `admin` 创建邀请、查看最近邀请、复制待使用邀请、撤销待使用邀请。
- 普通 `member` 不能创建、查看或撤销邀请。
- `pending` 邀请可复制和撤销。
- `used`、`revoked` 和 `expired` 邀请不可复制。
- MVP 当前会为待使用邀请保存 `token_plaintext`，用于刷新后重新复制；邀请被使用或撤销后会清空 token 原文。

媒体上传和时间线当前可以在本地测试照片和视频流程：

1. 使用普通 `node tools/dev.mjs --kill-ports` 启动，不使用 `--memory`。
2. 登录并进入 `http://localhost:5173/families/{familyId}/timeline`。
3. 在“上传精选”中选择 JPG、PNG 等照片文件。
4. 前端会先请求 `upload-intents` 创建上传任务和短期 PUT URL，再直接 PUT 到 MinIO。
5. PUT 成功后前端调用 `complete-upload`，后端确认对象存在并创建媒体元数据。
6. Worker 轮询待处理照片，生成缩略图和展示图，成功后写入 `server/logs/worker.log`。
7. 时间线只展示处理完成的媒体；处理中或失败项留在上传任务视图里。
8. 刷新页面后，“最近上传”会从后端读取最近上传任务，不依赖前端内存状态。
9. 可在时间线顶部按媒体类型或月份筛选；前端会把 `mediaType` 和 `month=YYYY-MM` 传给后端重新读取第一页。
10. 点击时间线中的照片或视频，可进入 `/families/{familyId}/media/{mediaId}` 媒体详情页，查看展示资源和基础元数据。

当前限制：

- 照片缩略图/展示图、视频缩略图/展示视频流程已实现；`node tools/run-worker.mjs` 会自动准备项目根目录 npm 托管的 FFmpeg/FFprobe，并注入 `FFMPEG_PATH`、`FFPROBE_PATH`。如需使用自定义二进制，也可以显式设置这两个环境变量。
- 时间线后端/API 已支持 `limit` + `cursor` 稳定分页；前端时间线底部会在存在下一页时显示“加载更多”。
- 移动端上传、时间线浏览和视频详情仍需要执行一次浏览器 smoke test。
- 原文件下载入口和下载 API 明确后置，前端不会拿到原文件 object key 或永久公开 URL。

## Worker

Worker 负责轮询 MySQL 中待处理的照片和视频媒体，照片会生成缩略图/展示图，视频会通过 FFmpeg 生成缩略图/展示视频，并写入预览 bucket。普通 MySQL 开发模式会由 `node tools/dev.mjs` 自动启动；如需单独启动：

```bash
node tools/run-worker.mjs
```

Worker 目前使用进程内 ticker 轮询 MySQL。后续如果切换到 Redis 或队列，API 会在上传完成后投递处理任务，Worker 消费队列并用数据库状态保证幂等；在那之前，MySQL 状态仍是处理进度和重试的事实来源。

生产部署时，Worker 镜像应内置 FFmpeg/FFprobe，参考 `deploy/worker.Dockerfile`。Worker 启动阶段会执行版本检查并写入日志；依赖不可用时直接启动失败，避免领取视频任务后把用户媒体标为处理失败。
