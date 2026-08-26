# ─────────────────────────────────────────────────────────
# Stage 1: 安装 golang-migrate 工具
# ─────────────────────────────────────────────────────────
FROM alpine:3.19 AS migrate-builder

RUN apk add --no-cache curl tar && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate

# ─────────────────────────────────────────────────────────
# Stage 2: 构建 Go 二进制
# ─────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o starbyte-server ./cmd/server

# ─────────────────────────────────────────────────────────
# Stage 3: 生产运行镜像
# ─────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata bash

ENV TZ=Asia/Shanghai

WORKDIR /app

# 复制二进制和迁移文件
COPY --from=builder /app/starbyte-server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/configs ./configs

# 复制 migrate 工具
COPY --from=migrate-builder /usr/local/bin/migrate /usr/local/bin/migrate

# 入口脚本：自动迁移 + 启动服务
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
