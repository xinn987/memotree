## Context

MemoTree 的第一阶段面向家庭宝宝照片和视频共享场景。当前目标不是构建 AI 相册、网盘或 NAS 管理器，而是先解决家庭成员跨平台上传、查看和下载照片视频的核心痛点。

关键约束：

- 宝宝照片和视频属于敏感家庭数据，私密访问边界必须从 MVP 开始成立。
- 家庭成员可能使用 iPhone、Android 或电脑，因此第一版应优先考虑跨平台移动端体验。
- 时间线浏览必须快速，不能直接依赖原图原视频加载。
- 功能范围需要收束，不做收藏、评论、AI、标签、地图、成长报告或复杂权限系统。

## Goals / Non-Goals

**Goals:**

- 提供一个私密家庭空间，让成员通过邀请加入。
- 支持照片、视频和批量上传。
- 将原文件保存在私有对象存储中，并生成轻量预览资源。
- 按时间线快速浏览家庭媒体。
- 支持查看详情并下载原图原视频。
- 以移动端 Web/PWA 体验优先，兼容 iOS、Android 和桌面浏览器。

**Non-Goals:**

- 不做收藏。
- 不做评论、点赞或社交互动。
- 不做 AI 整理、人脸识别、自动成长报告。
- 不做文件夹、标签、复杂搜索或网盘式管理。
- 不做自动同步手机相册。
- 不做复杂角色权限，MVP 仅区分管理员和普通家庭成员。

## Decisions

### Decision: Use React + Vite Mobile-First PWA For MVP

第一版采用 React + Vite + TypeScript + Tailwind 构建移动端优先的 PWA，而不是先做 iOS/Android 原生应用，也不优先使用 Next.js。

Rationale:

- 可以同时覆盖 iPhone、Android 和电脑，直接解决跨平台共享问题。
- 独立开发成本更低，适合先验证家庭真实使用闭环。
- 项目后端由 Go 承担，前端只需要作为 API client，React + Vite 比 Next.js 的服务端渲染和全栈约定更简单。
- React、Vite、Tailwind 和 shadcn/ui 的资料多、模式稳定，更适合前端主要依赖 vibe coding 的开发方式。
- 后续如果上传体验或后台能力成为瓶颈，再基于已验证的产品形态补原生 App。

Alternatives considered:

- 原生 App 优先：体验潜力更高，但开发、审核、双端维护和后台上传复杂度都更高。
- 极简网页优先：实现最快，但如果不按 PWA 和移动端体验设计，可能无法满足老人和家庭日常使用。
- Next.js：适合全栈 React 和服务端渲染场景，但本项目已有 Go 后端，MVP 不需要额外引入 Server Components、API Routes 或 SSR 复杂度。

### Decision: Use Go API With MySQL For Core Backend

后端采用 Go API，数据库优先使用 MySQL。

Rationale:

- 核心后端职责是鉴权、家庭权限、上传签名、下载签名、媒体元数据和时间线查询，Go 很适合实现清晰可控的 API 服务。
- 用户已有 Go 和 MySQL 经验，熟悉度能降低独立开发风险。
- MVP 的关系模型清晰，MySQL 足以支撑用户、家庭、成员、邀请和媒体资产等数据。
- 数据访问优先考虑 sqlc 或明确 SQL；如果开发效率优先，可评估 GORM，但权限关键查询需要保持可读和可测试。

Alternatives considered:

- PostgreSQL：JSONB、复杂索引和 Supabase 生态更强，但不是当前 MVP 的必要条件。
- Supabase 一体化：认证和数据库启动快，但会引入用户不熟悉的后端平台，且与 Go + MySQL 的经验路径不完全一致。

### Decision: Keep Original Files Private And Serve Authorized Temporary Access

原图和原视频保存在 Cloudflare R2 私有对象存储中，前端不持有长期公开 URL。查看预览和下载原文件都需要经过 Go API 权限校验。

Rationale:

- 防止复制出的媒体 URL 长期有效。
- 家庭成员被移除后，可以立即切断后续访问。
- 下载原文件可以通过短期签名 URL 或后端授权响应实现。
- R2 提供 S3 兼容接口和较低的媒体存储成本，适合作为开发和正式早期的对象存储。

Alternatives considered:

- 公开对象 URL：实现简单，但不符合家庭隐私数据的安全要求。
- 完全由应用服务器转发所有文件：权限控制强，但带宽成本和服务压力更高。
- 国内 OSS/COS：如果主要用户在中国大陆且 R2 访问速度不理想，后续可以通过 S3-compatible storage adapter 迁移或增加国内存储实现。

Update:

- MVP 可以先试 Cloudflare R2，但由于主要用户在中国大陆，系统 SHALL NOT 在业务逻辑中写死 R2。
- 后端应通过 S3-compatible storage adapter 访问对象存储，至少保留切换到阿里云 OSS、腾讯云 COS 或 MinIO 的配置边界。
- 本地开发使用 MinIO 模拟私有对象存储。

### Decision: Use Persistent Accounts With Passwordless Login

MVP 使用持久化用户账号，并采用魔法链接或验证码作为无密码登录方式。邀请只授予家庭成员权限，不替代用户账号。

Rationale:

- 爸妈和爷爷奶奶不适合从第一版开始承担密码管理、找回密码和复杂登录流程。
- 魔法链接或验证码可以降低登录门槛，但用户、登录身份和家庭成员关系仍然需要持久化落库。
- 家庭权限来自邀请加入后的 membership，而不是来自“拿到链接即可访问”的匿名状态。

### Decision: Use One-Time Invitations For Family Membership

管理员创建的一条邀请在 MVP 中最多允许一个账号成功加入家庭。

Rationale:

- 一人一次的邀请更容易解释，也更容易撤销和审计。
- 可以避免家庭邀请链接被转发后长期成为公共入口。
- 被移除成员不能依赖旧邀请重新加入，必须由管理员重新邀请。

### Decision: Optimize The First Screen For Elder Browsing

首屏应以最近照片和视频时间线为主，上传作为明显但次要的入口存在。

Rationale:

- 第一批用户包含爸妈和爷爷奶奶，核心体验是打开就看到宝宝近况。
- 主要上传者可能是爷爷，因此上传入口不能隐藏过深，但不应抢占浏览内容。
- 下载原图原视频是兜底能力，应放在详情或工具入口中，而不是成为首页主动作。

### Decision: Separate Originals From Browsing Previews

上传保存原文件，同时由独立 Go Worker 调用 FFmpeg 等工具生成用于时间线浏览的照片缩略图和视频封面图。

Rationale:

- 时间线加载速度是 MVP 的核心体验指标。
- 移动端直接加载原图原视频会导致首屏慢、流量高、滚动卡顿。
- 视频封面可以让视频在时间线里稳定呈现，不依赖浏览器预加载视频。
- 媒体处理不应运行在主 API 请求链路内，独立 Worker 可以避免缩略图和视频封面生成拖慢上传、浏览和下载接口。

Alternatives considered:

- 只保存原文件：实现更简单，但浏览体验不可控。
- MVP 就做完整视频转码：体验更好，但会显著增加处理复杂度和成本；第一版只要求视频封面，转码可后置。

### Decision: Use Timeline As The Primary Information Architecture

首页以时间线组织内容，按月份和日期分组展示最近上传的照片和视频。

Rationale:

- 家庭成员更关心“最近孩子发生了什么”，而不是管理文件夹。
- 时间线符合宝宝成长记录的自然心智模型。
- 可以避免第一版滑向网盘或复杂资料库。

Alternatives considered:

- 文件夹/相册管理优先：长期整理能力更强，但会增加上传和查看负担。
- 上传收件箱优先：有利于聚合散落素材，但日常查看吸引力较弱。

## Risks / Trade-offs

- 大视频上传在移动端浏览器中不稳定 -> 使用分文件进度、失败重试和明确错误状态；必要时后续补分片上传。
- 媒体处理会引入异步状态 -> 时间线允许显示处理中占位，不阻塞其他媒体展示。
- PWA 无法完全替代原生后台自动同步 -> MVP 明确不做自动同步，先验证手动上传和家庭浏览闭环。
- 对象存储和带宽成本可能随家庭视频增长而上升 -> 第一版记录文件大小和媒体类型，后续基于真实使用评估配额、压缩或转码策略。
- 轻量筛选可能不足以应对大量历史内容 -> MVP 只保留月份和媒体类型筛选，待真实内容规模增长后再判断是否引入标签或搜索。

## Migration Plan

这是新项目的首个产品 change，无既有数据迁移要求。

建议交付顺序：

1. 建立 React + Vite PWA 和 Go API 的项目边界。
2. 建立 MySQL 数据模型、家庭、成员和邀请的访问控制基础。
3. 建立 R2 私有对象存储和媒体元数据模型。
4. 实现照片/视频上传和处理状态。
5. 实现 Go Worker 的缩略图和视频封面生成流程。
6. 实现时间线浏览和媒体详情。
7. 实现权限校验后的原文件下载。

Rollback strategy:

- MVP 发布前无生产迁移风险。
- 发布后如媒体处理出现问题，应保留原文件上传成功记录，并允许后台重新生成缩略图或视频封面。

## Open Questions

- MVP 的登录身份优先使用手机号还是邮箱；当前仅确定采用持久化账号 + 无密码登录。
- 上传文件大小上限和单次批量数量上限需要根据目标存储成本和移动端稳定性确定。
- 时间线日期优先使用 EXIF/媒体拍摄时间还是上传时间；建议优先拍摄时间，缺失时回退上传时间。
