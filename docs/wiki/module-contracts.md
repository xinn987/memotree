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
POST /families/{familyId}/invitations
POST /invitations/{token}/join
```

原则：

- 注册产生持久化用户、登录凭证和会话。
- 登录凭证用于证明用户身份，会话用于浏览器后续请求的重复识别。
- 邀请只赋予家庭成员权限，不替代账号。
- 邀请加入流程可以和注册流程组合：用户通过有效邀请创建账号后，系统同时创建家庭成员关系。
- 管理邀请必须是管理员成员。
- MVP 使用账号密码登录；短信、邮箱、微信等恢复或第三方登录能力后置。

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

## Original Download

```text
POST /families/{familyId}/media/{mediaId}/download
```

原则：

- 每次下载前校验 active membership。
- 返回短期下载 URL 或等效授权响应。
- 不向未授权用户泄露对象 key 或可用 URL。
