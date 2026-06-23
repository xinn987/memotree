# Worker 生产镜像：MVP 只处理图片，不把 FFmpeg 作为启动依赖。
FROM golang:1.24-bookworm AS build

WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/memotree-worker ./worker/cmd/worker

FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/memotree-worker /app/memotree-worker

ENV APP_ENV=production
ENV MEDIA_WORKER_CONCURRENCY=1
ENV MEDIA_WORKER_POLL_INTERVAL_SECONDS=5

ENTRYPOINT ["/app/memotree-worker"]
