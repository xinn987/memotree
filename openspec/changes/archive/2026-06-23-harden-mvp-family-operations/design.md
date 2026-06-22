## Context

当前 MVP 已经有 `admin` / `member`、`active` / `removed` 的领域概念，时间线和上传接口也已经通过 active family membership 做基本保护。缺口主要在“管理员实际能做什么”和“异常状态如何恢复”：成员管理接口尚不完整，媒体删除缺少可操作 API，处理失败条目虽然在规格里要求可重试，但还没有形成稳定的后端入口和前端动作。

这个 change 的目标不是扩展复杂权限体系，而是把 MVP 试用前必须成立的家庭运营边界补齐。

## Goals / Non-Goals

**Goals:**

- 让 active `admin` 能查看家庭成员、更新成员称呼、移除成员。
- 防止家庭失去最后一个 active `admin`。
- 让 removed member 立即失去家庭数据、媒体、上传任务和上传授权访问。
- 让 active `admin` 能软删除媒体，使其从时间线和详情中消失。
- 让上传者或 active `admin` 能重试 `processing_failed` 条目，不要求重新上传原文件。
- 在前端提供最小可用入口，不做大视觉重构。

**Non-Goals:**

- 不引入 owner、viewer、uploader 等复杂角色。
- 不实现成员角色变更 UI，除非为最后管理员保护测试需要补后端能力。
- 不删除对象存储中的原文件或预览资源；软删除只改变业务可见性。
- 不做原文件下载、批量下载、评论点赞收藏或 Warm Editorial UI 迁移。
- 不实现实况照片合并；该能力继续后置。

## Decisions

### Decision: Keep Admin Operations Minimal And Family-Scoped

新增接口放在 `/families/{familyId}/members` 和 `/families/{familyId}/media/{mediaId}` 下，所有操作先校验当前用户是 active member，再按操作要求校验 active admin。

Rationale:

- 当前 API 已经以 family 作为权限边界，继续沿用可以避免引入全局管理语义。
- 成员管理是家庭内运营动作，不应要求 system admin。
- 普通成员仍然可以查看核心相册，但不能管理成员或删除媒体。

### Decision: Model Removal As Membership Status Change

移除成员只把 `FamilyMember.status` 改为 `removed`，不删除 `User`、历史上传媒体或上传任务记录。

Rationale:

- 家庭历史内容归属于 `Family`，不应因为上传者被移除而消失。
- 保留审计信息和上传者显示名有助于后续排查。
- 当前规格已经用 `removed` 表示成员状态，复用即可。

### Decision: Protect The Last Active Admin At Store Layer

最后管理员保护放在 store 操作中实现，API 层只负责把错误翻译成 409。

Rationale:

- MemoryStore 和 MySQLStore 都需要一致行为，放在 store 层更容易被单元测试覆盖。
- 后续如果增加角色变更，也可以复用同一条约束。

### Decision: Soft Delete Media Without Object Cleanup

删除媒体只更新 `MediaAsset.status=deleted` 或等价状态，时间线和详情查询继续只返回 active 且 preview ready 的媒体。

Rationale:

- 对象存储清理涉及后台任务、失败重试和审计窗口，超出本 change。
- 当前查询已经围绕媒体状态过滤，软删除可保持低风险。
- 误删恢复后续可做，MVP 先保证不再展示。

### Decision: Retry Processing By Resetting Existing Upload Item And Media Asset

处理失败重试不生成新的上传项，不要求重新上传原文件；系统把 `UploadItem.status` 重置为 `processing`，并把关联 `MediaAsset.rendition_status` 重置为 pending/processing，供 Worker 再次拾取。

Rationale:

- 原文件已经在私有对象存储中，重新上传会增加用户负担和失败概率。
- 保留原 upload item 可以维持任务历史和前端刷新体验。
- Worker 继续只处理数据库中等待处理的媒体，不决定权限。

## Risks / Trade-offs

- Removed member 已打开的页面仍可能保留旧签名预览 URL → 签名 URL TTL 控制最大残留窗口；后续如需更强隔离可缩短 TTL 或代理预览访问。
- 软删除不清理对象存储会留下存储成本 → 后续补后台清理任务；MVP 先保证业务不可见。
- 前端最小入口可能仍偏工具感 → 本 change 只补可用性，视觉迁移放到单独 change。
- 处理失败重试依赖 Worker 轮询逻辑正确识别重置状态 → 通过 store/API/worker 测试覆盖。

## Migration Plan

1. 如果现有 schema 已有成员状态、媒体状态和上传项状态字段，则只新增查询/更新 SQL。
2. 如果缺少字段或索引，增加向后兼容 migration，默认保持已有成员 active、媒体 active。
3. 发布 API 和 Worker 后，旧上传任务继续可读；只有 `processing_failed` 条目出现新的重试动作。
4. 回滚时软删除状态仍保留在数据库中；旧代码若不认识新接口不会影响已有浏览路径。
