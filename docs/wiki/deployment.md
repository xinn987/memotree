# 部署运行手册

## 目标拓扑

当前部署路线面向一台小规格云服务器：

```text
用户浏览器
  |
  v
Web/Nginx 容器
  |
  | /api/*
  v
API 容器  --->  MySQL 容器
  |
  v
Cloudflare R2 或其他 S3 兼容对象存储

Worker 容器
  |
  +--> MySQL 领取图片处理任务
  +--> R2 读取原图并写入缩略图/展示图
```

MVP 部署先只支持 JPG、PNG、GIF 图片上传。视频处理保留为后续扩容能力，不作为当前 Worker 镜像和服务器规格的上线要求。

## 镜像

| 镜像 | Dockerfile | 作用 |
| --- | --- | --- |
| `memotree-api:local` | `deploy/api.Dockerfile` | API 服务，并内置 `memotree-init-storage` 初始化工具 |
| `memotree-worker:local` | `deploy/worker.Dockerfile` | 图片缩略图/展示图处理 Worker |
| `memotree-web:local` | `deploy/web.Dockerfile` | Nginx 静态站点和 `/api/*` 反向代理 |

本阶段 CI 只验证镜像能构建，不推送到 GHCR、Docker Hub 或云厂商镜像仓库。首次部署可以在服务器本地构建：

```bash
node tools/build-images.mjs
```

## 环境分层

| 环境 | 用途 | 数据 |
| --- | --- | --- |
| `local` | 开发机本地联调 | Docker MySQL + MinIO，可随时重建 |
| `staging` | 真实服务器上的上线演练 | 本机 Docker MySQL + R2 测试 bucket |
| `production` | 家人/真实用户使用 | 正式 MySQL 数据和正式 R2 bucket，必须备份 |

早期可以用同一台服务器先跑 `staging`。当备份、日志、初始化和 smoke test 都跑通后，再把同一套环境按操作纪律提升为 `production`。

## 首次 Staging 部署

1. 在服务器安装 Docker、Docker Compose plugin、Node.js 22+。
2. 拉取仓库，并复制环境变量模板：

```bash
cp deploy/.env.staging.example deploy/.env.staging
```

3. 编辑 `deploy/.env.staging`：

- `MYSQL_PASSWORD` 和 `MYSQL_ROOT_PASSWORD` 必须换成强密码。
- `MYSQL_DSN` 里的密码必须和 `MYSQL_PASSWORD` 一致。
- `STORAGE_ENDPOINT`、`STORAGE_ACCESS_KEY_ID`、`STORAGE_SECRET_ACCESS_KEY` 填 R2 或其他 S3 兼容存储。
- `STORAGE_BUCKET_ORIGINALS` 和 `STORAGE_BUCKET_PREVIEWS` 建议 staging/production 分开。
- `MEDIA_WORKER_CONCURRENCY=1`，适合 2 vCPU / 2 GiB 初期服务器。

4. 构建镜像：

```bash
node tools/build-images.mjs
```

5. 启动 MySQL：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml up -d mysql
```

6. 显式初始化 schema：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml run --rm schema-init
```

7. 初始化对象存储 bucket：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml run --rm init-storage
```

如果对象存储提供商不允许通过 S3 API 创建 bucket，就在控制台手动创建 `STORAGE_BUCKET_ORIGINALS` 和 `STORAGE_BUCKET_PREVIEWS`。

8. 启动服务：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml up -d api worker web
```

9. 检查健康状态：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml ps
curl -fsS http://127.0.0.1/healthz
curl -fsS http://127.0.0.1/api/healthz
```

## 权限边界

- `schema-init` 使用 MySQL root 密码，只用于建表。
- `init-storage` 需要能创建或确认 bucket。
- API/Worker 长期凭据只需要访问已有 bucket、生成签名 URL、读取/写入/删除业务对象。

不要把 MySQL 暴露到公网。API、Worker 和 MySQL 应通过 Docker 网络通信。

## Smoke Test

首次部署后至少跑一遍：

1. 打开站点，注册第一个账号。
2. 创建家庭，确认第一个账号是管理员。
3. 上传一张 JPG 或 PNG，确认前端拿到 upload intent。
4. 确认浏览器 PUT 到 R2 成功。
5. 确认 `complete-upload` 后上传任务进入处理中。
6. 查看 Worker 日志，确认生成 thumbnail 和 display image。
7. 刷新时间线，确认图片可见。
8. 管理员删除一张误传图片，确认时间线不再显示。
9. 临时制造一次处理失败后，确认“重试处理”入口能重新入队。
10. 尝试上传 MP4，确认 API 返回“当前版本暂不支持上传视频”。

## 日志

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml logs -f api
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml logs -f worker
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml logs -f web
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml logs -f mysql
```

优先看 API 和 Worker 日志；如果连接失败，再看 MySQL 或对象存储配置。

## 回滚

本阶段还不推送远程镜像，回滚方式以本机镜像和 Git 版本为准：

1. 记录当前可用 commit。
2. 新版本部署前不要删除旧镜像。
3. 如果新版本异常，切回旧 commit 后重新构建镜像。
4. 执行：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml up -d --build api worker web
```

数据库回滚不能依赖代码回滚。上线前必须先做 `mysqldump --single-transaction`，并把备份上传到 R2 或 NAS。
