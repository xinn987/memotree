# 模块协议

这里先定义模块之间的交互边界，具体字段会随实现逐步收敛。

## Auth / Family

```text
POST /auth/register
POST /auth/login
POST /auth/logout
GET  /auth/session
POST /families
GET  /families
POST /families/{familyId}/invites
GET  /families/{familyId}/invites
DELETE /families/{familyId}/invites/{inviteId}
POST /invites/{token}/join
```

原则：

- 注册产生持久化用户、登录凭证和会话。
- 登录凭证用于证明用户身份，会话用于浏览器后续请求的重复识别。
- 邀请只赋予家庭成员权限，不替代账号。
- 邀请加入流程可以和注册流程组合：用户通过有效邀请创建账号后，系统同时创建家庭成员关系。
- 管理邀请必须是管理员成员。
- 管理员可以查看家庭邀请列表，并撤销仍处于 `pending` 状态的邀请。
- MVP 为了支持刷新后重新复制邀请链接，创建邀请时会临时保存 `token_plaintext`；邀请被使用或撤销后会清空该字段。
- 前端只能复制仍处于 `pending` 且仍有 token 原文的邀请；`used`、`revoked` 和 `expired` 邀请不可复制。
- MVP 使用账号密码登录；短信、邮箱、微信等恢复或第三方登录能力后置。

当前实现状态：

- 已实现账号注册、登录、退出和会话恢复。
- 已实现创建家庭，并把创建者加入为 `admin` 成员。
- 已实现邀请创建、邀请列表、邀请撤销和邀请加入。
- 已实现管理员权限校验：普通成员不能创建、查看或撤销邀请。
- 成员管理、移除成员和最后一个管理员保护仍未实现。

## Media Upload

```text
POST /families/{familyId}/media/upload-intents
POST /families/{familyId}/media/{mediaId}/complete-upload
```

原则：

- 创建 upload intent 前校验 active membership。
- 每个文件独立记录状态，批量上传允许部分失败。
- 原文件对象 key 由后端生成，前端不自行决定存储路径。

## Timeline

```text
GET /families/{familyId}/timeline
GET /families/{familyId}/media/{mediaId}
```

原则：

- 时间线只返回预览资源授权或可访问引用。
- 默认按拍摄时间排序，缺失时回退上传时间。
- 分页必须稳定，不能因新上传导致重复或跳项。

## Deferred: Original Download

```text
POST /families/{familyId}/media/{mediaId}/download
```

原则：

- MVP 不实现下载入口和下载 API。
- 原文件仍保存在私有对象存储中，为后续下载能力预留。
- 每次下载前校验 active membership。
- 返回短期下载 URL 或等效授权响应。
- 不向未授权用户泄露对象 key 或可用 URL。
