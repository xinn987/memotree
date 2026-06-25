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
| `memotree-api` | `deploy/api.Dockerfile` | API 服务，并内置 `memotree-init-storage` 初始化工具 |
| `memotree-worker` | `deploy/worker.Dockerfile` | 图片缩略图/展示图处理 Worker |
| `memotree-web` | `deploy/web.Dockerfile` | Nginx 静态站点和 `/api/*` 反向代理 |

大陆云服务器直接从 Docker Hub 拉基础镜像可能不稳定。推荐流程是：

1. 在开发机或 CI 构建最终镜像。
2. 推送到阿里云 ACR。
3. 服务器只从 ACR 拉 `memotree-api`、`memotree-worker`、`memotree-web`。

## 发布镜像到 ACR

先在阿里云 ACR 创建命名空间和三个镜像仓库：

```text
memotree-api
memotree-worker
memotree-web
```

在开发机仓库根目录登录 ACR。为了不污染全局 Docker 配置，可以把登录信息放在项目内 `.docker-config`：

```bash
export DOCKER_CONFIG="$PWD/.docker-config"
docker login crpi-niameksqgqq8nnw0.cn-hangzhou.personal.cr.aliyuncs.com
```

Windows PowerShell：

```powershell
$env:DOCKER_CONFIG="$PWD\.docker-config"
docker login --username=wlgq987 crpi-niameksqgqq8nnw0.cn-hangzhou.personal.cr.aliyuncs.com
```

发布镜像并生成版本单：

```bash
export ACR_REGISTRY=crpi-niameksqgqq8nnw0.cn-hangzhou.personal.cr.aliyuncs.com
export ACR_NAMESPACE=memotree
export IMAGE_TAG=$(git rev-parse --short=12 HEAD)
node tools/publish-acr-images.mjs
```

Windows PowerShell：

```powershell
$env:ACR_REGISTRY="crpi-niameksqgqq8nnw0.cn-hangzhou.personal.cr.aliyuncs.com"
$env:ACR_NAMESPACE="memotree"
$env:IMAGE_TAG=(git rev-parse --short=12 HEAD)
node tools/publish-acr-images.mjs
```

脚本结束后会写入：

```text
deploy/releases/staging-current.env
```

这个文件只包含镜像地址和发布元信息，不包含 MySQL 密码、R2 key、bucket 凭据等长期密钥。

如果不使用 ACR，也可以在服务器本地构建：

```bash
docker build -f deploy/api.Dockerfile -t memotree-api:local .
docker build -f deploy/worker.Dockerfile -t memotree-worker:local .
docker build -f deploy/web.Dockerfile -t memotree-web:local .
```

本地构建路线要求服务器能稳定拉取 `golang`、`node`、`nginx`、`debian` 等基础镜像。

## 环境分层

| 环境 | 用途 | 数据 |
| --- | --- | --- |
| `local` | 开发机本地联调 | Docker MySQL + MinIO，可随时重建 |
| `staging` | 真实服务器上的上线演练 | 本机 Docker MySQL + R2 测试 bucket |
| `production` | 家人/真实用户使用 | 正式 MySQL 数据和正式 R2 bucket，必须备份 |

早期可以用同一台服务器先跑 `staging`。当备份、日志、初始化和 smoke test 都跑通后，再把同一套环境按操作纪律提升为 `production`。

## 首次 Staging 部署

首次部署仍然建议手动做，因为它包含外部控制台动作和长期密钥配置。

1. 在服务器安装 Docker、Docker Compose plugin、Git。

2. 拉取仓库，并切到部署分支：

```bash
git clone git@github.com:xinn987/memotree.git
cd memotree
git switch feat/family_operations
```

3. 复制环境变量模板：

```bash
cp deploy/.env.staging.example deploy/.env.staging
```

4. 编辑 `deploy/.env.staging`：

- `PUBLIC_BASE_URL` 先填服务器公网地址，例如 `http://120.26.28.65`。
- `MYSQL_PASSWORD` 和 `MYSQL_ROOT_PASSWORD` 必须换成强密码。
- `MYSQL_DSN` 里的密码必须和 `MYSQL_PASSWORD` 一致。
- `STORAGE_ENDPOINT`、`STORAGE_ACCESS_KEY_ID`、`STORAGE_SECRET_ACCESS_KEY` 填 R2 或其他 S3 兼容存储。
- `STORAGE_BUCKET_ORIGINALS` 和 `STORAGE_BUCKET_PREVIEWS` 建议 staging/production 分开。
- `MEDIA_WORKER_CONCURRENCY=1`，适合 2 vCPU / 2 GiB 初期服务器。

5. 在开发机发布镜像并把版本单复制到服务器：

```bash
scp deploy/releases/staging-current.env root@120.26.28.65:/root/repos/memotree/deploy/releases/staging-current.env
```

6. 服务器登录 ACR：

```bash
docker login crpi-niameksqgqq8nnw0.cn-hangzhou.personal.cr.aliyuncs.com
```

7. 检查 compose 配置：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml config
```

8. 启动 MySQL：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml up -d mysql
```

9. 显式初始化 schema：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml run --rm schema-init
```

10. 初始化对象存储 bucket：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml run --rm init-storage
```

如果对象存储提供商不允许通过 S3 API 创建 bucket，就在控制台手动创建 `STORAGE_BUCKET_ORIGINALS` 和 `STORAGE_BUCKET_PREVIEWS`。

11. 启动业务服务：

```bash
sh deploy/staging-deploy.sh deploy/releases/staging-current.env
```

## 日常更新发布

日常代码更新不再手动编辑 `deploy/.env.staging` 的镜像变量。流程是：

1. 在开发机发布镜像并生成版本单：

```powershell
$env:DOCKER_CONFIG="$PWD\.docker-config"
docker login --username=wlgq987 crpi-niameksqgqq8nnw0.cn-hangzhou.personal.cr.aliyuncs.com

$env:ACR_REGISTRY="crpi-niameksqgqq8nnw0.cn-hangzhou.personal.cr.aliyuncs.com"
$env:ACR_NAMESPACE="memotree"
$env:IMAGE_TAG=(git rev-parse --short=12 HEAD)
node tools/publish-acr-images.mjs
```

2. 把版本单复制到服务器：

```powershell
scp deploy/releases/staging-current.env root@120.26.28.65:/root/repos/memotree/deploy/releases/staging-current.env
```

3. 在服务器执行日常部署：

```bash
cd /root/repos/memotree
git pull
sh deploy/staging-deploy.sh deploy/releases/staging-current.env
```

`staging-deploy.sh` 会：

- 检查 `deploy/.env.staging` 和 release env 文件是否存在。
- 读取 `API_IMAGE`、`WORKER_IMAGE`、`WEB_IMAGE`。
- 校验 Docker Compose 配置。
- 拉取 API、Worker、Web 镜像。
- 使用 `--no-build` 重启 API、Worker、Web。
- 检查 `http://127.0.0.1/healthz` 和 `http://127.0.0.1/api/healthz`。

它不会自动运行 `schema-init` 或 `init-storage`。这两个命令只在首次部署或明确需要迁移/初始化时手动执行。

## 权限边界

- `schema-init` 使用 MySQL root 密码，只用于建表。
- `init-storage` 需要能创建或确认 bucket。
- API/Worker 长期凭据只需要访问已有 bucket、生成签名 URL、读取/写入/删除业务对象。
- release env 只保存镜像版本，不保存长期密钥。

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

查看最近日志：

```bash
sh deploy/staging-logs.sh --tail=120
```

持续跟随日志：

```bash
sh deploy/staging-logs.sh --follow
```

只看部分服务：

```bash
sh deploy/staging-logs.sh --tail=200 api worker
```

也可以直接使用 Docker Compose：

```bash
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml logs -f api
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml logs -f worker
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml logs -f web
docker compose --env-file deploy/.env.staging -f deploy/docker-compose.staging.yml logs -f mysql
```

优先看 API 和 Worker 日志；如果连接失败，再看 MySQL 或对象存储配置。

## 回滚

回滚以旧 release env 文件和 ACR 镜像 tag 为准：

1. 发布新版本前保留旧的 release env 文件。
2. 新版本部署前不要删除旧 ACR 镜像 tag。
3. 如果新版本异常，把旧 release env 文件复制回服务器。
4. 执行：

```bash
sh deploy/staging-deploy.sh deploy/releases/previous-release.env
```

数据库回滚不能依赖代码回滚。上线前必须先做 `mysqldump --single-transaction`，并把备份上传到 R2 或 NAS。
