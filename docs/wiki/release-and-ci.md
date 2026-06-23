# 发布与持续集成

## CI

GitHub Actions 当前分为四类检查：

- Web checks：安装前端依赖、运行 TypeScript 检查、构建 Vite 产物。
- Go tests：使用 Go 1.24.x 运行 API、Worker 和共享包测试。
- OpenSpec validation：校验主 specs，并动态校验当前 active changes。
- Docker image builds：构建 API、Worker、Web 镜像，但不推送、不部署。

CI 的目标是证明代码和运行镜像都能被重复构建。自动发布、镜像仓库推送和远程部署暂不在当前阶段实现。

## 本地完整检查

在仓库根目录运行：

```bash
node tools/check.mjs
```

该脚本会执行：

- 工具可用性检查：`node`、`npm`、`go`、`openspec`
- `node --test tools/shared.test.mjs`
- `go test ./...`
- `npm run check`
- `npm run build`
- `openspec validate --specs --strict`
- 对当前 active OpenSpec changes 逐个执行 strict validate

构建本地运行镜像：

```bash
node tools/build-images.mjs
```

## 环境

| 环境 | 用途 | 数据 |
| --- | --- | --- |
| `local` | 本地开发 | Docker MySQL + MinIO |
| `staging` | 真实服务器上的上线演练 | Docker MySQL + R2 测试 bucket |
| `production` | 家人真实使用 | 正式 MySQL 数据 + 正式 R2 bucket |

## Release Notes

每次发布至少记录：

- 数据库 schema 或 migration 变化
- API 兼容性变化
- 对象存储配置变化
- 媒体处理策略变化
- 部署和回滚注意事项

首次 staging 部署和 smoke test 见 [部署运行手册](deployment.md)。
