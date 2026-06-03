# 本地开发

## Prerequisites

- Node.js 22+
- Go 1.22+
- Docker Desktop

## Services

```powershell
Copy-Item .env.example .env
docker compose -f deploy/docker-compose.dev.yml up -d
```

本地使用 MySQL 和 MinIO 模拟线上 MySQL + R2/S3-compatible object storage。

## Commands

前端：

```powershell
Set-Location web
npm install
npm run dev
```

API：

```powershell
Set-Location server
go run ./api/cmd/api
```

Worker：

```powershell
Set-Location server
go run ./worker/cmd/worker
```

## Checks

```powershell
Set-Location web
npm run check
npm run build
Set-Location ..
Set-Location server
go test ./...
Set-Location ..
```
