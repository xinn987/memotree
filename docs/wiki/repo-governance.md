# 仓库治理

MemoTree 当前采用单 Git 仓库，不拆多仓库，也不使用 git submodule。MVP 阶段前端、后端、部署和需求会频繁联动，单仓库更容易保持变更一致。

## Directory Ownership

| 路径 | 职责 | 主要检查 |
| --- | --- | --- |
| `web/` | 移动端 PWA | TypeScript check, Vite build |
| `server/` | Go API、Worker、共享后端包、migrations | `go test ./...` |
| `deploy/` | 本地开发和发布配置 | 配置 review |
| `docs/wiki/` | 产品、架构、协议、运维设计 | 文档 review |
| `openspec/` | 需求、验收场景和任务状态 | OpenSpec review |

## PR Boundaries

优先让一个 PR 聚焦一个边界：

- `web: timeline first screen`
- `server: family invitation model`
- `deploy: local object storage`
- `docs: storage strategy`
- `openspec: refine family access requirements`

如果一个变更必须跨边界，PR 描述里需要说明跨边界原因。例如后端 API 字段变化通常会同时修改 `server/`、`web/` 和 `docs/wiki/module-contracts.md`。

## Dependency Rules

- `web/` 只通过 HTTP API contract 依赖后端。
- `server/api` 是权限校验入口。
- `server/worker` 不做用户权限决策，只处理已授权入库后的异步任务。
- 对象存储只能通过 storage adapter 访问，不让 R2/OSS/COS 细节扩散到业务代码。
- `docs/wiki` 记录设计和协议，`openspec` 记录需求与验收。

## When To Split Repositories

暂不拆仓库。只有当出现以下情况时再评估：

- 前端和后端有独立发布节奏且互相很少联动。
- Worker 需要独立基础设施和团队维护。
- CI 时间明显不可接受，且路径级 CI 无法缓解。
- 有明确的权限或开源边界要求。

## Code Comment Rules

- 代码注释优先使用中文，除非第三方 API、协议名称或固定术语更适合英文。
- Go 文件应在 package 或文件入口附近说明该文件的职责边界。
- 导出的类型、函数和方法应使用 Go doc 风格注释，说明用途和关键约束。
- 只在重要业务规则、安全边界、事务边界、跨模块约定和不直观实现处添加注释。
- 避免解释显而易见的赋值、循环或字段含义，防止注释噪音掩盖真正重要的信息。
- 涉及用户身份、家庭权限、邀请 token、session、原文件访问、数据删除等逻辑时，必须注释设计意图。
