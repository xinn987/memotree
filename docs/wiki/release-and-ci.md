# 发布与持续集成

## CI

GitHub Actions 当前分为两条检查：

- Web checks: 安装前端依赖、TypeScript 检查、Vite 构建。
- Go tests: 运行 API 和 Worker 测试。

## Environments

建议环境分层：

| 环境 | 用途 | 数据 |
| --- | --- | --- |
| local | 本地开发 | Docker MySQL + MinIO |
| staging | 发布前验证 | 独立数据库和对象桶 |
| production | 家庭真实使用 | 正式数据库和私有对象桶 |

## Release Notes

每次发布至少记录：

- 数据库 migration。
- API 兼容性变化。
- 对象存储配置变化。
- 媒体处理策略变化。
