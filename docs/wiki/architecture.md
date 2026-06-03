# 系统架构

```text
Mobile PWA
   |
   | HTTPS JSON API
   v
Go API  ---- MySQL
   |
   | signed upload/download URL
   v
Private Object Storage (R2 first, S3-compatible)
   ^
   |
Go Worker ---- FFmpeg / image processing
```

## Runtime Modules

- `web`: 老人友好的移动端 PWA。
- `server`: Go module 根目录。
- `server/api`: 同步 HTTP API 和权限边界。
- `server/worker`: 异步媒体处理。
- `deploy`: 本地和线上部署配置。

## Key Boundaries

- 前端不持有长期公开媒体 URL。
- 原图和原视频只存在私有对象存储。
- API 每次生成上传/下载授权前必须校验家庭成员身份。
- Worker 不决定权限，只处理已入库且需要生成预览的媒体任务。
