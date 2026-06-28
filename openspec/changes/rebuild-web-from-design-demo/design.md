## Context

MemoTree 当前生产前端集中在 `web/src/app/App.tsx` 和 `web/src/styles.css`：路由、会话、API 类型、上传编排、时间线、邀请、成员和媒体详情相互耦合，页面结构仍然是冷灰后台面板。该实现可以调用现有 API，但不是本次重构需要兼容的前端基线。

目标设计已经完整存在于 `design/`：`tokens.css` 定义视觉令牌，`demo/assets/demo.css` 定义共享组件，七个 HTML 页面及桌面/移动截图定义最终布局和交互。生产实现必须把这些静态设计映射到真实 React 状态，而不是在旧页面上继续打补丁。

约束如下：

- 只重构前端模块，不轻易修改 Go API、Worker、数据库或部署行为。
- `design/` 是视觉真源，生产代码不得自行发明另一套视觉语言。
- demo 中存在但 API 暂不支持的能力需要保留版位，但不得伪造数据写入或成功反馈。
- 核心用户包含老人，中文可读性、触控尺寸、键盘焦点和移动端稳定性是产品要求。

## Goals / Non-Goals

**Goals:**

- 完整重写 `web/src/`，建立正式、可维护、按产品功能垂直拆分的 React 工程。
- 让登录/注册、邀请加入、时间线、上传、邀请、成员和媒体详情在桌面与移动端严格遵循对应 demo。
- 保留当前已实现 API 的请求和业务语义，并通过清晰的 API 边界隔离服务端契约。
- 为后端暂缺能力提供视觉完整、行为诚实、未来容易接线的占位实现。
- 建立可验证的加载、空、错误、禁用、焦点、减少动效和响应式状态。

**Non-Goals:**

- 不兼容旧 React 组件、旧 className、旧 CSS、旧页面组合或旧内部状态结构。
- 不在本 change 中新增或修改后端端点、数据模型、权限规则、媒体处理和存储逻辑。
- 不实现公开分享、评论、点赞、AI 整理、全量备份或文件管理器能力。
- 不为了抽象而建立独立设计系统包，也不引入与本次页面重写无关的重型状态框架。

## Decisions

### Decision: Clean Rewrite Of The Frontend Presentation Layer

删除或整体替换旧页面实现，以真实 API 契约和 `design/demo` 为输入重新建立生产前端。旧 `FamilyHome`、`TimelinePanel`、`UploadPanel` 等组件不作为兼容边界。

原因：旧实现的问题同时存在于信息架构、组件边界、文案和视觉层。保留旧结构只换 CSS 会继续制造耦合，并阻碍 demo 的页面级布局。

备选方案：

- 在旧组件上渐进换肤：改动较小，但无法消除首页堆叠和单文件耦合。
- 重写后同时改造 API：可以一次完成更多能力，但把视觉升级和服务端风险混在一起，不符合本次边界。

### Decision: Feature-Oriented Production Structure

生产结构按职责组织：

```text
web/src/
  app/                 应用入口、providers、route tree、session guard
  api/                 HTTP client、error model、共享 API contracts
  components/
    layout/            app bar、page shell、auth split shell
    ui/                button、field、chip、empty、error、skeleton、dialog
  features/
    auth/
    timeline/
    upload/
    invites/
    members/
    media/
  styles/
    tokens.css
    global.css
    components.css
    pages/
  utils/               日期、字节、状态文案等纯函数
```

每个 feature 自己拥有 page、局部 components、hooks 和 API adapter。只有被多个 feature 使用的元素才进入共享层。

原因：产品页面之间有明确业务边界；垂直模块可以同时保持局部内聚和共享 UI 一致性。

备选方案：

- 按 `components/hooks/services` 水平分层：文件看似整齐，但一个功能会散落在全项目。
- 一开始发布独立组件库：对 MVP 过度设计，拖慢页面保真。

### Decision: Use React Router For Canonical Product Routes

新增 `react-router-dom`，建立以下正式路由：

```text
/login
/join?invite=<token>
/families/:familyId/timeline
/families/:familyId/upload
/families/:familyId/invites
/families/:familyId/members
/families/:familyId/media/:mediaId
```

根路径根据会话和家庭状态重定向。会话守卫负责未登录、无家庭、家庭不可见和深链接恢复；页面组件不再手写 `pushState` 或解析 pathname。

原因：本次已经从单页组合升级为七个独立产品表面，正式路由可以提供可测试的导航、返回和深链接语义。

备选方案：

- 延续自定义 History 解析：依赖少，但随着页面增加会重复实现路由库能力。

### Decision: Preserve API Contracts Behind Typed Feature Adapters

统一 HTTP client 负责 base path、credentials、JSON、错误规范化和中止信号。各 feature 通过自己的 adapter 调用现有端点，页面不直接拼接请求。

现有能力继续真实接入：

- 注册、登录、退出和会话；
- 创建家庭、通过邀请加入家庭；
- 时间线筛选、分页、详情和管理员删除；
- 上传 intent、直传、完成/失败回报、重试、停止和轮询；
- 创建、查询、复制和作废邀请；
- 查询、改称呼和移除成员。

原因：重写表现层时保持服务端稳定，同时让未来 API 变化局限在 adapter。

备选方案：

- 引入 TanStack Query 等数据框架：有缓存优势，但当前流程包含自定义上传直传和轮询，先用明确 hooks 更容易控制迁移风险。

### Decision: Treat Demo Files As A Verifiable Visual Contract

生产样式从 `design/tokens.css` 映射到 `web/src/styles/tokens.css`，页面结构逐一对齐 demo：

- `auth.html` → 登录/注册；
- `join.html` → 邀请加入；
- `timeline.html` → 家庭时间线；
- `upload.html` → 上传任务；
- `invite.html` → 邀请管理；
- `members.html` → 成员管理；
- `media-detail.html` → 媒体详情。

共享组件和页面样式保留 demo 的 BEM 语义或建立一一对应的生产命名。不会直接导入 demo CSS，因为其中包含静态导航、内联样式和演示资源。Fraunces 与 Inter 使用可控的前端字体包或本地静态资源提供，避免生产布局依赖运行时 Google Fonts。

视觉验收同时比较桌面和移动截图，检查布局、字体、颜色、间距、圆角、阴影、图像比例、固定控件和交互状态。真实数据为空或长度变化时仍需保持同一视觉语法。

原因：用户已经确认 demo 是目标，不需要第二轮视觉再设计。

备选方案：

- 直接导入 demo CSS：初期接近，但会把 demo 导航和静态页面假设带入生产。
- 只抽取颜色变量：不能保证布局和组件保真。

### Decision: Preserve Unsupported Demo Capabilities As Honest Placeholders

后端暂缺但 demo 已展示的操作可以保留相同位置和视觉层级。占位控件必须遵守以下规则：

- 不发送不存在的请求；
- 不修改本地状态来伪造持久成功；
- 触发时显示统一的“暂未开放”轻提示，或在明确不应触发时使用可访问的禁用状态；
- 代码中通过 capability registry 集中标记，不把判断散落在 JSX；
- 后续接入真实 API 时只替换 feature adapter 和 capability 状态，不重排页面。

首批可能的占位包括：管理员转让、原图下载、修改拍摄时间、没有相邻媒体上下文时的上一张/下一张，以及后端没有提供的家庭手记/照片标题字段。静态占位不得伪造具体家庭回忆或用户数据。

原因：既保持 demo 完整布局，也避免让用户误以为操作已保存。

备选方案：

- 完全隐藏缺失能力：会破坏 demo 的结构和未来接线位置。
- 纯前端模拟成功：会产生数据一致性和信任问题。

### Decision: State, Error And Motion Behavior Are Part Of The Design

- 会话状态由 app provider 管理，feature 数据由各自 hooks 管理。
- 上传生命周期保留轮询、XHR 进度和本地文件重试约束；路由离开时明确提示仍在浏览器上传的项目。
- 每个页面实现 loading skeleton、empty、inline error 和 retry。
- 主操作使用明确 disabled/loading 状态；占位提示、API 错误和成功反馈使用统一 toast/inline feedback。
- CSS 动效使用 demo 的 150–250ms 时长和 ease-out 曲线，只表达 hover、进度、展开、切换和反馈。
- `prefers-reduced-motion` 下关闭位移、缩放和非必要过渡，内容始终默认可见。

原因：仅复刻静态截图不足以成为正式产品实现。

### Decision: Verification Covers Structure, Behavior And Visual Fidelity

- TypeScript 严格检查和 Vite production build 必须通过。
- 为 API client、route guard、状态文案、格式化和 capability placeholder 编写单元测试。
- 为认证、时间线、上传、邀请/成员和详情的关键状态编写组件或集成测试；网络使用可控 mock，不访问真实后端。
- 在桌面和移动 viewport 逐页进行浏览器检查，并与 `design/demo/assets/shots` 对照。
- 检查键盘导航、焦点、触控尺寸、文本溢出、图片加载失败、空数据、错误和减少动效。

原因：本次目标包含精确视觉和正式工程质量，两者都需要可重复验证。

## Risks / Trade-offs

- [一次性重写可能造成功能回归] → 先冻结 API contracts，按 feature 逐页接线，每完成一个页面就运行类型、测试和浏览器验证。
- [严格 demo 布局与真实数据长度冲突] → 使用真实数据驱动的弹性容器、截断和换行规则，并验证极短、极长和空值。
- [瀑布流在动态图片尺寸下发生布局跳动] → 使用 rendition 宽高预留 aspect ratio，图片加载失败提供稳定占位。
- [占位操作被误认为已生效] → 统一 capability registry 和“暂未开放”反馈，禁止模拟持久成功。
- [新增路由和字体依赖扩大包体] → 只引入必要依赖，使用字体子集/指定字重，并在 production build 中检查输出。
- [demo 中某些视觉文案依赖后端不存在的内容字段] → 不伪造家庭数据；保留版位并显示中性的占位状态。
- [旧活动 change 与新 change 范围重叠] → 本 change 作为后续实施依据，旧 `refactor-web-warm-editorial` 不参与实现，完成后单独决定是否废弃或归档。

## Migration Plan

1. 建立新路由、应用壳、API client、contracts、共享 UI 和样式令牌。
2. 先完成未登录链路：登录/注册、邀请加入、会话恢复和无家庭引导。
3. 完成时间线与应用顶栏，使登录后的默认页面首先达到 demo 基线。
4. 完成上传页面并恢复现有直传、轮询、失败和重试行为。
5. 完成邀请与成员页面，真实接入已有操作，标记缺失的管理员转让占位。
6. 完成媒体详情、客户端可用的相邻导航上下文以及缺失操作占位。
7. 删除不再引用的旧前端模块和样式，完成测试、构建和逐页视觉验收。

每一步保持前端可构建。回滚方式是恢复本 change 前的 `web/` 代码；没有数据库或服务端迁移需要回滚。

## Open Questions

当前没有阻塞实施的问题。实现过程中若确认某个 demo 能力必须依赖新增后端端点或字段，暂停该能力的真实接入并向用户确认；在确认前维持占位。
