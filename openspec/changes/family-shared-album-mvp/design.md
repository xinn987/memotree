## Context

MemoTree 的第一阶段面向家庭宝宝照片和视频精选共享场景。当前目标不是构建 AI 相册、网盘、全量备份工具或 NAS 管理器，而是先解决家庭成员跨平台查看精选近况、上传精选照片视频和必要时下载原文件的核心痛点。

关键约束：

- 宝宝照片和视频属于敏感家庭数据，私密访问边界必须从 MVP 开始成立。
- 家庭成员可能使用 iPhone、Android 或电脑，因此第一版应优先考虑跨平台移动端体验。
- 全量照片备份不由 MemoTree 承担，MVP 只承载家庭成员愿意共享给家人的精选内容。
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
- 不做全量照片备份和备份完整性管理。
- 不做宝宝模型、按宝宝筛选或上传时选择具体宝宝。
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

### Decision: Use Invitation-Gated Password Accounts For MVP

MVP 使用持久化用户账号和密码登录。家庭邀请只授予加入某个家庭的资格，不替代用户账号；用户后续通过登录名和密码证明自己是同一个全局账号，浏览器会话只用于免重复登录。

Rationale:

- 魔法链接和长期 cookie 对家庭用户并不一定更直观；一旦浏览器 cookie 丢失，用户和管理员都难以理解恢复路径。
- 账号密码是更传统但更可解释的模型：用户知道自己有一个账号，忘记密码时可以走管理员重置或后续找回流程。
- 邀请负责控制首次加入家庭，账号密码负责后续恢复登录，两者职责清晰。
- 家庭权限来自邀请加入后的 membership，而不是来自“拿到链接即可访问”的匿名状态。

Domain model:

- `User` 表示全局唯一用户身份。
- `UserCredential` 表示用户的登录名和密码哈希，用于证明用户身份。
- `UserSession` 表示某个浏览器或设备当前已经登录的会话，用于后续请求识别。
- `Family` 表示一个家庭空间。
- `FamilyMember` 表示某个用户在某个家庭里的称呼、角色和成员状态。
- `FamilyInvite` 表示加入家庭的一次性授权凭据。

MVP flow:

1. 管理员创建一次性 `FamilyInvite`。
2. 被邀请人打开邀请链接，输入家庭内称呼、登录名和密码。
3. 系统创建 `User`、`UserCredential`、`FamilyMember` 和 `UserSession`，并标记邀请已使用。
4. 后续访问通过 `UserSession` 识别用户；会话丢失后，用户可以通过登录名和密码重新登录。

Notes:

- MVP 的登录名可以优先使用手机号，但在未接入短信验证码前，手机号只作为登录名，不代表已完成手机号所有权验证。
- 忘记密码的 MVP 兜底方式可以是家庭管理员重置密码或重新邀请；短信、邮箱、微信和 Apple 登录等恢复方式后置。

### Decision: Use One-Time Invitations For Family Membership

管理员创建的一条邀请在 MVP 中最多允许一个账号成功加入家庭。

Rationale:

- 一人一次的邀请更容易解释，也更容易撤销和审计。
- 可以避免家庭邀请链接被转发后长期成为公共入口。
- 被移除成员不能依赖旧邀请重新加入，必须由管理员重新邀请。

### Decision: Treat Family As The Primary Sharing Space

MVP 使用 `Family` 表示一个私密家庭精选近况空间，并将其作为媒体内容和权限判断的边界。`Family` 不表示某一个宝宝，MVP 不引入 `Child`、相册或文件夹模型。

Rationale:

- 家庭成员加入的是一个家庭共享空间，而不是分别加入某个宝宝的相册。
- 许多照片和视频并不适合强制归属到某个宝宝，上传时要求选择宝宝会增加负担。
- 当前产品主心智是家庭成员查看精选近况，而不是整理宝宝资料库。
- `User` 和 `Family` 是独立领域对象，通过 `FamilyMember` 形成多对多关系，可以支持一个用户加入多个家庭。

Domain model:

- `Family` 表示家庭私密共享空间。
- `FamilyMember` 表示某个用户在某个家庭里的称呼、角色和成员状态。
- `MediaAsset` 归属于 `Family`，并记录上传者作为贡献者和审计信息。
- MVP UI 默认按单家庭体验设计；如果用户加入多个家庭，可以显示简单家庭选择页。

### Decision: Use Minimal Member Permissions

MVP 只区分 `admin` 和 `member` 两种成员角色，并使用 `active` 和 `removed` 表示成员状态。

Rationale:

- 家庭相册第一版不需要 owner、viewer、uploader 或只读成员等复杂角色。
- 媒体浏览、上传和下载的权限边界可以统一为 active family membership。
- 管理邀请和成员需要更高权限，因此只额外区分 admin。

Rules:

- `active member` 可以查看时间线、查看媒体详情、上传照片/视频和下载原文件。
- `active admin` 可以邀请成员、撤销未使用邀请、移除成员、管理家庭基础信息和成员称呼。
- 移除成员不会删除全局 `User`，也不会删除其历史上传内容。
- 系统应避免家庭空间失去最后一个 active admin。

### Decision: Optimize The First Screen For Elder Browsing

首屏应以家庭精选近况时间线为主，上传作为明显但次要的入口存在。

Rationale:

- 第一批用户包含爸妈和爷爷奶奶，核心体验是打开就看到宝宝和家庭近况。
- 主要上传者可能是爷爷，因此上传入口不能隐藏过深，但不应抢占浏览内容。
- 下载原图原视频是兜底能力，应放在详情或工具入口中，而不是成为首页主动作。

### Decision: Separate Originals From Browsing Previews

上传保存原文件，同时由独立 Go Worker 调用图像处理工具和 FFmpeg 等工具生成用于时间线浏览和详情展示的 Web-compatible 展示资源。

Rationale:

- 时间线加载速度是 MVP 的核心体验指标。
- 移动端直接加载原图原视频会导致首屏慢、流量高、滚动卡顿。
- iPhone HEIC、MOV、实况照片以及不同 Android 设备产生的媒体格式不应直接暴露给前端兼容性处理。
- 照片展示图、视频缩略图和展示视频可以让前端在 iOS、Android 和桌面浏览器中稳定呈现内容。
- 媒体处理不应运行在主 API 请求链路内，独立 Worker 可以避免缩略图、展示图和展示视频生成拖慢上传、浏览和下载接口。

Domain model:

- `MediaAsset` 表示时间线上的一个媒体条目，类型为 `photo`、`video` 或 `live_photo`。
- `MediaOriginal` 表示用户上传并私有保存的原文件，类型为 `image_original` 或 `video_original`。
- `MediaRendition` 表示给前端展示用的派生资源，类型为 `thumbnail`、`display_image` 或 `display_video`。
- 时间线和详情页以 `MediaAsset` 为单位展示，不直接展示 `MediaOriginal`。

Format rules:

- HEIC、JPG、PNG 等照片原样保存为 `image_original`，并生成 `thumbnail` 和 `display_image`。
- MOV、MP4 等视频原样保存为 `video_original`，并生成 `thumbnail` 和 `display_video`。
- iPhone 实况照片应合并为一个 `live_photo` 类型的 `MediaAsset`，其图片和视频分别保存为 `image_original` 和 `video_original`。
- MVP 默认使用实况照片的 `display_image` 展示静态照片，不把实况照片中的视频原文件作为独立时间线条目展示。
- 实况照片的 `display_video` 作为后续增强能力预留，MVP 可以不实现动态播放。

Alternatives considered:

- 只保存原文件：实现更简单，但浏览体验不可控。
- MVP 就做完整多码率视频转码：体验更好，但会显著增加处理复杂度和成本；第一版只要求生成浏览器可播放的单一展示视频，多码率转码可后置。
- 将实况照片拆成照片和短视频两个时间线条目：实现简单，但会让大多数 iPhone 实况照片在时间线中重复出现，破坏浏览体验。

### Decision: Treat Media As Family-Owned Assets

上传到 MemoTree 的媒体归属于 `Family`，上传者只是贡献者和审计信息。MVP 使用单个媒体作为时间线基本单元，不引入帖子、上传批次、动态分组、相册或文件夹。

Rationale:

- MemoTree 是家庭精选近况流，不是个人网盘或个人相册分享。
- 成员离开家庭或账号出现问题时，不应导致家庭历史内容丢失。
- 单个照片或视频天然适配按拍摄时间排序的成长时间线。
- 当前不做评论、点赞和文字动态，因此没有必要引入“一个帖子包含多张媒体”的模型。

Rules:

- `MediaAsset.family_id` 是媒体访问边界。
- `MediaAsset.uploaded_by` 记录上传者，但不表示个人所有权。
- 一次批量上传可以创建多个独立 `MediaAsset`，批量只是上传流程，不是浏览对象。
- 只有 `active admin` 可以删除媒体，普通 `member` 不能删除媒体。
- 删除使用软删除状态；被删除媒体不出现在时间线、详情或下载接口中。
- 对象存储中的原文件和展示资源可以等待后台清理或人工清理。

### Decision: Use Timeline As The Primary Information Architecture

首页以时间线组织内容，按月份和日期分组展示家庭精选照片和视频。主时间线表达宝宝和家庭近况的成长顺序，优先使用拍摄时间排序；缺失拍摄时间时，使用上传时间混入时间线。

Rationale:

- 家庭成员更关心“最近孩子发生了什么”，而不是管理文件夹。
- 时间线符合宝宝成长记录的自然心智模型。
- 上传者是在分享精选内容，不是在备份手机相册；因此首页主轴应服务成长记录，而不是上传流水。
- 可以避免第一版滑向网盘或复杂资料库。

Rules:

- `captured_at` 表示拍摄时间，决定主时间线位置。
- `uploaded_at` 表示上传/分享时间，用于审计、最近添加和缺失拍摄时间时的兜底排序。
- 首页主时间线按 `captured_at ?? uploaded_at` 倒序排列。
- 上传时间排序后续可以作为“最近添加”视图补充，不进入 MVP 主流程。

Alternatives considered:

- 文件夹/相册管理优先：长期整理能力更强，但会增加上传和查看负担。
- 上传时间流优先：更容易看到最新添加内容，但会削弱宝宝成长记录的主题。

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
5. 实现 Go Worker 的缩略图、展示图和展示视频生成流程。
6. 实现时间线浏览和媒体详情。
7. 实现权限校验后的原文件下载。

Rollback strategy:

- MVP 发布前无生产迁移风险。
- 发布后如媒体处理出现问题，应保留原文件上传成功记录，并允许后台重新生成缩略图、展示图或展示视频。

## Open Questions

- MVP 登录名是否固定为手机号，还是允许用户自定义登录名；当前建议优先手机号，但不在 MVP 强依赖短信验证。
- 上传文件大小上限和单次批量数量上限需要根据目标存储成本和移动端稳定性确定。
