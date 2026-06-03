# 存储方案

## 当前结论

第一版可以先试 Cloudflare R2，原因是成本友好且 S3 兼容。但考虑主要用户在中国大陆，R2 访问稳定性不能假设可靠。

## 设计原则

- 业务层依赖 `StorageService`，不直接依赖 R2 SDK。
- 对象 key 由后端生成。
- 原文件和预览资源分桶或至少分前缀保存。
- 所有原文件访问都通过短期授权。

## Provider Candidates

| Provider | 优点 | 风险 |
| --- | --- | --- |
| Cloudflare R2 | 出口流量成本低，S3 兼容 | 大陆访问稳定性不确定 |
| 阿里云 OSS | 大陆访问较稳，生态成熟 | 出口流量成本和备案 |
| 腾讯云 COS | 大陆访问较稳，微信生态附近 | 出口流量成本和备案 |
| MinIO | 本地开发友好 | 不作为生产默认方案 |

## Required Adapter Methods

```text
PutObject
GetSignedUploadURL
GetSignedDownloadURL
HeadObject
DeleteObject
```
