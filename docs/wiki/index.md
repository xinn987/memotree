# MemoTree Project Wiki

这个 wiki 用来承接比 OpenSpec 更细的设计信息：产品决策、模块边界、交互协议、本地开发、部署和持续集成。

## 推荐阅读顺序

1. [产品设计](product-design.md)
2. [技术选型复审](technical-review.md)
3. [系统架构](architecture.md)
4. [模块协议](module-contracts.md)
5. [本地开发](local-development.md)
6. [发布与持续集成](release-and-ci.md)
7. [存储方案](storage-strategy.md)
8. [仓库治理](repo-governance.md)

## 文档方案调研结论

| 方案 | 优点 | 风险 | 当前结论 |
| --- | --- | --- | --- |
| 仓库内 Markdown | review 简单，和代码同生命周期，无额外服务 | 导航和搜索弱 | 现在采用 |
| GitHub Wiki | 使用门槛低 | 和代码 review 脱节，不适合协议级文档 | 暂不采用 |
| Docusaurus | 版本化、导航、搜索和发布完整 | 需要额外 Node 文档站维护 | 后续文档增多时可接入 |
| VitePress | 轻量，适合前端项目 | 需要额外配置发布 | 后续可作为更轻方案 |
| MkDocs Material | Markdown 体验好，搜索成熟 | Python 工具链增加维护面 | 暂不优先 |

当前先用 `docs/wiki`，等模块协议和运维文档稳定后，再决定是否生成静态站点。
