FROM golang:1.24-bookworm AS build

WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/memotree-api ./api/cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/memotree-init-storage ./devtools/cmd/init-storage

FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates curl \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /out/memotree-api /app/memotree-api
COPY --from=build /out/memotree-init-storage /app/memotree-init-storage
COPY server/migrations /app/migrations

ENV APP_ENV=production
ENV API_ADDR=:8080

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD curl -fsS http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/memotree-api"]
