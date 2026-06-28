# MemoTree Design Demo Frontend Rebuild Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完整重写 MemoTree React 前端，使七个生产页面严格遵循 `design/demo`，同时保留现有后端 API 契约并诚实占位暂缺能力。

**Architecture:** 使用 React Router 建立正式页面路由；会话与家庭访问由 app provider/guard 管理；API contracts、HTTP client、共享 UI 和六个 feature 模块相互隔离。生产 CSS 从 demo tokens 映射，页面 JSX 以对应 demo 为结构基线，旧 `App.tsx` 和 `styles.css` 不作为兼容边界。

**Tech Stack:** React 19、TypeScript、Vite、React Router、Lucide React、Fraunces/Inter 字体包、Vitest、Testing Library、CSS。

---

### Task 1: 建立测试、依赖和共享 API 基线

**Files:**
- Modify: `web/package.json`
- Modify: `web/vite.config.ts`
- Create: `web/src/test/setup.ts`
- Create: `web/src/api/contracts.ts`
- Create: `web/src/api/client.ts`
- Create: `web/src/api/client.test.ts`
- Create: `web/src/app/capabilities.ts`
- Create: `web/src/app/capabilities.test.ts`
- Create: `web/src/utils/format.ts`
- Create: `web/src/utils/format.test.ts`

- [x] **Step 1: 写失败测试**

覆盖 JSON 请求、204 响应、错误规范化、可选 AbortSignal、占位能力查询、日期/字节/上传状态中文映射：

```ts
expect(await requestJSON("/health")).toEqual({ ok: true });
expect(capabilityMessage("transferAdmin")).toBe("管理员转让暂未开放");
expect(uploadItemStatusText("processing")).toBe("正在整理");
```

- [x] **Step 2: 运行测试并确认因模块缺失失败**

Run: `npm test -- --run src/api/client.test.ts src/app/capabilities.test.ts src/utils/format.test.ts`

Expected: FAIL，原因是待实现模块或导出不存在。

- [x] **Step 3: 安装依赖并实现最小共享模块**

新增 `react-router-dom`、`@fontsource/fraunces`、`@fontsource/inter`、`vitest`、`jsdom`、`@testing-library/react`、`@testing-library/jest-dom`、`@testing-library/user-event`。HTTP client 固定使用 `/api` 默认前缀和 `credentials: "include"`；capability registry 首批包含 `transferAdmin`、`downloadOriginal`、`editCapturedAt`、`adjacentMedia`、`familyNotes`。

- [x] **Step 4: 运行共享模块测试**

Run: `npm test -- --run src/api/client.test.ts src/app/capabilities.test.ts src/utils/format.test.ts`

Expected: PASS。

- [x] **Step 5: 运行类型检查**

Run: `npm run check`

Expected: PASS。

### Task 2: 建立视觉令牌、共享 UI 和应用布局

**Files:**
- Create: `web/src/styles/tokens.css`
- Create: `web/src/styles/global.css`
- Create: `web/src/styles/components.css`
- Create: `web/src/components/ui/Button.tsx`
- Create: `web/src/components/ui/Feedback.tsx`
- Create: `web/src/components/ui/Field.tsx`
- Create: `web/src/components/ui/Chip.tsx`
- Create: `web/src/components/ui/Dialog.tsx`
- Create: `web/src/components/ui/ui.test.tsx`
- Create: `web/src/components/layout/AppBar.tsx`
- Create: `web/src/components/layout/PageShell.tsx`
- Create: `web/src/components/layout/AuthShell.tsx`

- [ ] **Step 1: 写共享组件失败测试**

```tsx
render(<Button loading>保存</Button>);
expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
render(<PlaceholderButton capability="downloadOriginal">下载原图</PlaceholderButton>);
await user.click(screen.getByRole("button", { name: "下载原图" }));
expect(screen.getByRole("status")).toHaveTextContent("原图下载暂未开放");
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `npm test -- --run src/components/ui/ui.test.tsx`

Expected: FAIL，原因是组件不存在。

- [ ] **Step 3: 映射 demo tokens 并实现共享组件**

完整复制 `design/tokens.css` 的有效令牌值；全局样式定义字体、focus-visible、减少动效和基础排版；组件提供 default/hover/focus/active/disabled/loading/error/placeholder 状态。

- [ ] **Step 4: 运行组件测试与类型检查**

Run: `npm test -- --run src/components/ui/ui.test.tsx`

Expected: PASS。

Run: `npm run check`

Expected: PASS。

### Task 3: 建立会话、路由和无家庭引导

**Files:**
- Replace: `web/src/app/App.tsx`
- Create: `web/src/app/AppRouter.tsx`
- Create: `web/src/app/SessionProvider.tsx`
- Create: `web/src/app/ProtectedFamilyRoute.tsx`
- Create: `web/src/app/router.test.tsx`
- Create: `web/src/features/auth/auth.api.ts`
- Create: `web/src/features/auth/OnboardingPage.tsx`
- Modify: `web/src/main.tsx`
- Modify: `web/index.html`

- [ ] **Step 1: 写路由与守卫失败测试**

覆盖未登录跳转 `/login`、合法家庭深链、非法家庭回退、无家庭显示 onboarding：

```tsx
expect(await screen.findByText("回到家里的相册")).toBeInTheDocument();
expect(router.state.location.pathname).toBe("/families/7/timeline");
```

- [ ] **Step 2: 运行路由测试并确认失败**

Run: `npm test -- --run src/app/router.test.tsx`

Expected: FAIL，原因是 provider/route tree 不存在。

- [ ] **Step 3: 实现会话 provider、守卫和正式路由**

路由固定为 `/login`、`/join`、`/families/:familyId/timeline|upload|invites|members|media/:mediaId`；根路由按 session/family 重定向；不再调用手写 `history.pushState`。

- [ ] **Step 4: 实现无家庭创建/加入引导**

创建家庭和已有登录用户凭 token 加入继续调用现有 API，视觉使用温暖纸面和 demo 表单语言。

- [ ] **Step 5: 运行路由测试与构建**

Run: `npm test -- --run src/app/router.test.tsx`

Expected: PASS。

Run: `npm run build`

Expected: PASS。

### Task 4: 重写登录、注册与邀请加入

**Files:**
- Create: `web/src/features/auth/AuthPage.tsx`
- Create: `web/src/features/auth/JoinPage.tsx`
- Create: `web/src/features/auth/auth.test.tsx`
- Create: `web/src/styles/pages/auth.css`

- [ ] **Step 1: 写认证流程失败测试**

测试登录 payload、注册后刷新 session、邀请 token 注册后调用 join、错误保持页面结构：

```ts
expect(fetchCalls[0].body).toEqual({ loginName: "妈妈", password: "secret1" });
expect(fetchCalls.at(-1)?.path).toBe("/invites/token-1/join");
```

- [ ] **Step 2: 运行认证测试并确认失败**

Run: `npm test -- --run src/features/auth/auth.test.tsx`

Expected: FAIL。

- [ ] **Step 3: 按 auth/join demo 实现页面**

桌面使用图片情绪板 + 380px 表单列；移动端保留 25vh banner；登录、注册、加入使用真实 API，邀请异常显示家庭化错误。

- [ ] **Step 4: 运行认证测试和类型检查**

Run: `npm test -- --run src/features/auth/auth.test.tsx`

Expected: PASS。

Run: `npm run check`

Expected: PASS。

### Task 5: 重写时间线

**Files:**
- Create: `web/src/features/timeline/timeline.api.ts`
- Create: `web/src/features/timeline/useTimeline.ts`
- Create: `web/src/features/timeline/TimelinePage.tsx`
- Create: `web/src/features/timeline/MonthSection.tsx`
- Create: `web/src/features/timeline/PhotoTile.tsx`
- Create: `web/src/features/timeline/timeline.test.tsx`
- Create: `web/src/styles/pages/timeline.css`

- [ ] **Step 1: 写时间线失败测试**

覆盖月份合并、筛选 query、分页、空/错状态、媒体导航和真实 aspect-ratio：

```tsx
expect(screen.getByRole("heading", { name: /六月/ })).toBeInTheDocument();
expect(screen.getByRole("img")).toHaveStyle({ aspectRatio: "600 / 800" });
expect(lastRequest).toContain("mediaType=photo");
```

- [ ] **Step 2: 运行时间线测试并确认失败**

Run: `npm test -- --run src/features/timeline/timeline.test.tsx`

Expected: FAIL。

- [ ] **Step 3: 实现 demo 时间线结构与数据流**

桌面保留 hero、双组 segmented toolbar 和月度相册；移动端隐藏 hero、使用 44px app bar 和 sticky month/filter；照片使用 CSS columns 保留原比例，hover/focus 提供元信息。

- [ ] **Step 4: 接入筛选、分页、空状态和上传 FAB**

月份手记字段缺失时显示中性“这个月的家庭手记稍后补上”占位，不伪造回忆；FAB 导航到 upload route。

- [ ] **Step 5: 运行时间线测试、类型和构建**

Run: `npm test -- --run src/features/timeline/timeline.test.tsx`

Expected: PASS。

Run: `npm run check && npm run build`

Expected: PASS。

### Task 6: 重写上传

**Files:**
- Create: `web/src/features/upload/upload.api.ts`
- Create: `web/src/features/upload/useUploadTasks.ts`
- Create: `web/src/features/upload/UploadPage.tsx`
- Create: `web/src/features/upload/UploadTaskView.tsx`
- Create: `web/src/features/upload/upload.test.tsx`
- Create: `web/src/styles/pages/upload.css`

- [ ] **Step 1: 写上传状态与流程失败测试**

覆盖图片校验、intent、XHR 进度、完成/失败、处理轮询、两种重试、停止和 beforeunload：

```ts
expect(uploadItemStatusText("upload_failed")).toBe("没传成");
expect(screen.getByText("这批正在传")).toBeInTheDocument();
expect(screen.getByRole("button", { name: "重新整理" })).toBeEnabled();
```

- [ ] **Step 2: 运行上传测试并确认失败**

Run: `npm test -- --run src/features/upload/upload.test.tsx`

Expected: FAIL。

- [ ] **Step 3: 迁移现有上传业务到独立 hook/api**

保留当前 API 路径、顺序、2 秒轮询、本地 File 重试映射和只允许 JPG/PNG/GIF 的 MVP 限制。

- [ ] **Step 4: 按 upload demo 实现桌面与移动布局**

桌面 dropzone、任务面板和最近记录；移动端隐藏 dropzone，使用底部文件选择按钮；所有状态使用 demo chip 和细进度条。

- [ ] **Step 5: 运行上传测试与构建**

Run: `npm test -- --run src/features/upload/upload.test.tsx`

Expected: PASS。

Run: `npm run build`

Expected: PASS。

### Task 7: 重写邀请和成员管理

**Files:**
- Create: `web/src/features/invites/invites.api.ts`
- Create: `web/src/features/invites/InvitesPage.tsx`
- Create: `web/src/features/invites/invites.test.tsx`
- Create: `web/src/features/members/members.api.ts`
- Create: `web/src/features/members/MembersPage.tsx`
- Create: `web/src/features/members/members.test.tsx`
- Create: `web/src/styles/pages/family.css`

- [ ] **Step 1: 写邀请/成员失败测试**

测试创建后 inline link、复制、作废、resolved 折叠、改称呼、移除确认和管理员转让占位：

```tsx
expect(await screen.findByText(/链接已生成/)).toBeInTheDocument();
await user.click(screen.getByRole("button", { name: "设为管理员" }));
expect(screen.getByRole("status")).toHaveTextContent("管理员转让暂未开放");
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `npm test -- --run src/features/invites/invites.test.tsx src/features/members/members.test.tsx`

Expected: FAIL。

- [ ] **Step 3: 接入已有邀请与成员 API**

只接 create/list/revoke invite、list/patch/delete member；普通成员访问管理路由显示权限状态。

- [ ] **Step 4: 按 invite/members demo 实现页面**

邀请页使用纵向 action blocks；成员页使用桌面行操作、移动 overflow 菜单和原生 dialog；转让管理员不发请求，只显示统一占位反馈。

- [ ] **Step 5: 运行测试与构建**

Run: `npm test -- --run src/features/invites/invites.test.tsx src/features/members/members.test.tsx`

Expected: PASS。

Run: `npm run build`

Expected: PASS。

### Task 8: 重写媒体详情

**Files:**
- Create: `web/src/features/media/media.api.ts`
- Create: `web/src/features/media/MediaDetailPage.tsx`
- Create: `web/src/features/media/media.test.tsx`
- Create: `web/src/features/media/navigationContext.ts`
- Create: `web/src/styles/pages/media-detail.css`

- [ ] **Step 1: 写详情失败测试**

覆盖照片/视频、管理员删除、timeline state 上下张、深链占位、下载/改时间占位：

```tsx
expect(screen.getByRole("img", { name: /家人上传/ })).toBeInTheDocument();
await user.click(screen.getByRole("button", { name: "下载原图" }));
expect(screen.getByRole("status")).toHaveTextContent("原图下载暂未开放");
```

- [ ] **Step 2: 运行详情测试并确认失败**

Run: `npm test -- --run src/features/media/media.test.tsx`

Expected: FAIL。

- [ ] **Step 3: 按 media-detail demo 实现沉浸查看**

使用暖深背景、居中媒体、桌面 metadata rail、移动 disclosure 和顶栏；详情从时间线打开时由 location state 携带已知 media IDs。

- [ ] **Step 4: 接入详情与删除 API，完成占位能力**

深链没有上下文时保留前后箭头但标记不可用；原图下载和改时间统一显示暂未开放。

- [ ] **Step 5: 运行详情测试与构建**

Run: `npm test -- --run src/features/media/media.test.tsx`

Expected: PASS。

Run: `npm run build`

Expected: PASS。

### Task 9: 清理、全量验证与视觉验收

**Files:**
- Delete: `web/src/styles.css`
- Modify: `web/src/main.tsx`
- Modify: `web/public/manifest.webmanifest`
- Modify: `openspec/changes/rebuild-web-from-design-demo/tasks.md`
- Create: `docs/wiki/frontend-demo-rebuild.md`

- [ ] **Step 1: 删除未引用旧实现和样式**

确认 `rg "FamilyHome|TimelinePanel|UploadPanel|parseAppRoute|styles.css" web/src` 无旧生产引用。

- [ ] **Step 2: 运行全量自动检查**

Run: `npm test -- --run`

Expected: 全部 PASS、无未处理 warning。

Run: `node tools/check-web.mjs`

Expected: TypeScript 与 production build PASS。

Run: `openspec validate rebuild-web-from-design-demo --type change --strict --no-interactive`

Expected: change valid。

- [ ] **Step 3: 运行本地应用并逐页浏览器验证**

桌面使用约 1440×1000，移动使用 390×844；逐页对照 `design/demo/assets/shots`，覆盖 auth、join、timeline、upload、invite、members、detail。

- [ ] **Step 4: 校验交互与无障碍**

验证键盘 focus、hover、disabled/loading/error/placeholder、40px 触控目标、长文案、空数据、图片错误和 `prefers-reduced-motion`。

- [ ] **Step 5: 更新 OpenSpec 任务与前端说明**

逐项将真正完成的任务改为 `- [x]`；文档列出仍缺后端的管理员转让、原图下载、拍摄时间修改、相邻媒体深链和家庭手记字段。

- [ ] **Step 6: 最终提交**

```bash
git add web docs openspec/changes/rebuild-web-from-design-demo/tasks.md
git commit -m "feat: 按设计 demo 重写前端"
```
