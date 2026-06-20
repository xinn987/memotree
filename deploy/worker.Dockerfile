# syntax=docker/dockerfile:1

# Worker 生产镜像：Go 二进制和 FFmpeg 运行时依赖都固化在镜像中。
FROM golang:1.24-bookworm AS build

WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/memotree-worker ./worker/cmd/worker

FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates ffmpeg \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/memotree-worker /app/memotree-worker

ENTRYPOINT ["/app/memotree-worker"]
