# ─────────────────────────────────────────────────────────
# Stage 1: 安装 golang-migrate 工具
# ─────────────────────────────────────────────────────────
FROM alpine:3.19 AS migrate-builder

# 校园网默认：经 ghfast.top 拉 GitHub Release，避免直连 github.com
RUN apk add --no-cache curl tar && \
    curl -fsSL https://ghfast.top/https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate

# ─────────────────────────────────────────────────────────
# Stage 2: MinIO mc 占位（不从 dl.min.io 下载）
# 生产 compose 请设 SKIP_BUCKET_CREATE=true；桶可在控制台手动创建一次。
# 入口脚本在 SKIP_BUCKET_CREATE=true 时不会调用 mc。
# ─────────────────────────────────────────────────────────
FROM alpine:3.19 AS mc-builder

RUN printf '%s\n' '#!/bin/sh' 'exit 0' > /usr/local/bin/mc && chmod +x /usr/local/bin/mc

# ─────────────────────────────────────────────────────────
# Stage 3: 构建 Go 二进制
# ─────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
# 校园网默认 GOPROXY，避免直连 proxy.golang.org
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o starbyte-server ./cmd/server

# ─────────────────────────────────────────────────────────
# Stage 4: 生产运行镜像
# ─────────────────────────────────────────────────────────
FROM alpine:3.19 AS production

RUN apk add --no-cache ca-certificates tzdata bash postgresql-client

ENV TZ=Asia/Shanghai

WORKDIR /app

# 复制二进制和迁移文件
COPY --from=builder /app/starbyte-server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/configs ./configs
COPY internal/export/service/fonts/OFL.txt /usr/share/licenses/starbyte/NotoSansSC-OFL.txt

# 复制迁移工具和 MinIO 客户端（mc 为占位脚本，见 Stage 2）
COPY --from=migrate-builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=mc-builder /usr/local/bin/mc /usr/local/bin/mc

# 入口脚本：自动迁移 + 启动服务
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
