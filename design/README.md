# MemoTree 前端重构设计稿

这是 MemoTree 前端从零重做的**设计探索稿**，用单 HTML 演示 demo 的形式探索页面与交互，**完全不动现有 `web/` 项目代码**。设计方法论遵循 [impeccable](https://impeccable.style/) 的 craft 流程，产品语境来自 [docs/wiki/product-design.md](../docs/wiki/product-design.md)。

## 怎么看

本地起一个静态服务器，根目录指向 `design/`（因为 `demo.css` 通过相对路径 `../../tokens.css` 引用设计系统正本）：

```bash
cd design
npx http-server . -p 8401 -c-1
```

然后打开：

- http://localhost:8401/demo/timeline.html — 家庭时间线（主心智页，定调）
- http://localhost:8401/demo/media-detail.html — 媒体详情 + 上/下一张
- http://localhost:8401/demo/upload.html — 上传任务页
- http://localhost:8401/demo/auth.html — 登录/注册页

页面底部有一个固定的 demo 导航条，可在四个页面间快速切换。占位图来自 picsum.photos，首次加载需联网。

## 设计方向

经过对现有 MVP 前端的诊断和项目文档的研读，确定方向为 **温暖纪实 / Warm Editorial**：

> 这是记录宝宝成长、给爸妈爷爷奶奶看的私密家庭相册。核心情绪是温暖、私密、时间的流逝感。视觉取自胶片相册、纸质手账、午后客厅的暖光，刻意避开原项目冷灰 SaaS 蓝的工具感。

核心信息架构遵循产品文档：首屏直接进时间线看图（不是后台工作台），照片为王，上传可见但次要。时间线按 `captured_at ?? uploaded_at` 倒序、按月份分组——这与"月度故事流"的形态天然契合。

## 设计语言

### 色彩（OKLCH，全部定义在 `tokens.css`）

| 令牌 | 值 | 用途 |
|------|-----|------|
| `--paper` | `oklch(0.965 0.008 60)` | 燕麦暖白主背景 |
| `--ink` | `oklch(0.270 0.020 45)` | 暖深棕主文字（≥4.5:1） |
| `--accent` | `oklch(0.580 0.165 38)` | 赭石砖红，主行动/选中 |
| `--caramel` | `oklch(0.680 0.090 55)` | 焦糖，月份装饰 |
| `--state-ok/warn/err` | 暖域绿/琥珀/砖红 | 语义状态，不出现刺眼标准蓝/红 |

色相卡在暖橙区间（h 38-60），不用冷灰、不用 SaaS 蓝。状态色也收敛到暖色系，避免工具感。背景刻意带极低彩度的焦糖色相，而不是"中性偏暖"的偷懒默认。

### 字体

- **Fraunces**（衬线，光学尺寸）— 月份大字、页面标题、媒体详情的拍摄信息。承载"相册/手账"气质。
- **Inter**（无衬线）— 所有 UI 正文、按钮、数据、元信息。

字号用固定 px 而非 fluid clamp——这是 product UI，用户在固定 DPI 浏览，fluid 标题在侧栏里缩小反而更难看（impeccable product register 建议）。字阶 1.125-1.2 紧凑比例。

### 关键决策（对抗原 MVP 的问题）

1. **移除 ghost-card**：原项目几乎所有面板都是 `border: 1px solid` + `box-shadow: 0 12px 36px`，这是 impeccable 明确禁止的 AI/工程 UI tell。新设计面板默认用纯边框，照片用深沉聚焦阴影，二选一不叠加。
2. **移除 eyebrow tell**：原项目每个 section 顶上都顶一行灰色小字。新设计用月份大衬线标题做天然分隔，不重复"AI eyebrow 语法"。
3. **照片为王**：照片直角、白纸边距，元信息（上传人/时间）只在 hover 浮出，不再每张下面预置碎字。网格用大小混排的"特写 + 小图"编辑式布局，不是统一方格。
4. **文案软化**：原项目 `任务状态/失败/已停止` 这种后台黑话，全部改成家庭化语气 —— `这批正在传 / 没传成 / 再试一次 / 先不传了`。老人是核心用户，工具感要尽量弱。
5. **上传不抢占浏览**：按文档要求，上传入口做成时间线右下角的浮起 FAB，而不是和照片网格并列的大面板。

## 文件结构

```
design/
├── README.md                  ← 本文件
├── tokens.css                 ← 设计令牌正本（色板/字体/间距/圆角/阴影/动效）
└── demo/
    ├── timeline.html          ← 月度故事流主页（核心定调页）
    ├── media-detail.html      ← 媒体详情 + 上/下一张切换
    ├── upload.html            ← 上传任务页（状态机：等待/上传中/整理中/完成/失败）
    ├── auth.html              ← 登录/注册（双栏：左侧情绪板 + 右侧表单）
    └── assets/
        ├── demo.css           ← 共享组件样式（@import tokens.css）
        ├── shoot.mjs          ← playwright 截图自检脚本（可选）
        └── shots/             ← 已截的桌面/移动参考图
```

## 与现有 web/ 项目的对应关系

demo 里的每个页面对应 `web/src/app/App.tsx` 里的一个组件，数据形态严格对齐真实 API 契约（见 [docs/wiki/module-contracts.md](../docs/wiki/module-contracts.md)）：

| demo 页 | 现有组件 | 对应 API |
|---------|---------|---------|
| timeline.html | `FamilyHome` / `TimelinePanel` | `GET /families/{id}/timeline` |
| media-detail.html | `MediaDetailPage` | `GET /families/{id}/media/{id}` |
| upload.html | `UploadPanel` | `POST .../upload-intents` 等 |
| auth.html | `AuthPanel` | `POST /auth/login` / `/auth/register` |

确认方向后，迁移时把 `tokens.css` 的令牌接入 `web/src/styles.css`，把 demo 的组件样式逐步替换 `App.tsx` 里的内联 className 逻辑即可，React 结构和数据流可基本不动。

## 新增的设计要点（相对原 MVP）

- **详情页支持上一张/下一张**：左右箭头按钮 + 键盘方向键位预留，产品文档明确要求，原 MVP 未实现。
- **上传状态语义化**：用 chip + 进度条颜色区分 `排队中 / 传了 X% / 正在整理 / 已整理好 / 没传成`，对应 `UploadItemStatus` 的 8 个状态。
- **月份手记**：时间线每个月份区可有一句斜体衬线手记（"小满过生日那天，外婆特意从老家赶来"），让相册有"编辑过的回忆册"感而非纯流水。这是精选内容的天然 affordance——既然不是全量备份，就有编辑的余地。

## 已知限制 / 后续

- 占位图用 picsum.photos，正式迁移时换成真实 `MediaRendition` 资源 URL。
- demo 用 Google Fonts CDN 引入 Fraunces/Inter，正式项目应 self-host + subset + preload。
- 详情页的"改拍摄时间""下载原图"等 admin/预留功能只做了按钮占位，未做交互流。
- 未做实况照片（live_photo）的专门视觉标记，迁移时按文档"默认展示静态图"处理即可。
- 邀请管理（admin 创建/撤销邀请）未单独出页，可在确认方向后补一个 settings/invite 页。
