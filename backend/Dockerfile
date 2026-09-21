# ─────────────────────────────────────────────────────────
# Stage 1: MinIO mc 占位（不从 dl.min.io 下载）
# 生产 compose 请设 SKIP_BUCKET_CREATE=true；桶可在控制台手动创建一次。
# 入口脚本在 SKIP_BUCKET_CREATE=true 时不会调用 mc。
# ─────────────────────────────────────────────────────────
FROM alpine:3.21 AS mc-builder

RUN printf '%s\n' '#!/bin/sh' 'exit 0' > /usr/local/bin/mc && chmod +x /usr/local/bin/mc

# ─────────────────────────────────────────────────────────
# Stage 2: 构建 Go 二进制与 golang-migrate
#
# 工具链下限说明：go.mod 的 go 指令已提升到 1.25 —— 这是
# github.com/xuri/excelize/v2 v2.11.0（CVE-2026-54063 / CVE-2026-59161）
# 与 golang.org/x/net v0.55+（GO-2026-5030）的最低要求。
# 本镜像必须 >= 1.25，否则 go build 会直接拒绝。
# ─────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY go.mod go.sum ./
# 校园网默认 GOPROXY，避免直连 proxy.golang.org。migrate 也走模块代理，
# 不依赖 ghfast.top 拉 GitHub Release（该镜像在 CI 上会 500）。
RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

# 复制源码并构建
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o starbyte-server ./cmd/server

# ─────────────────────────────────────────────────────────
# Stage 3: 生产运行镜像
# ─────────────────────────────────────────────────────────
FROM alpine:3.21 AS production

RUN apk add --no-cache ca-certificates tzdata bash postgresql-client

ENV TZ=Asia/Shanghai

WORKDIR /app

# 复制二进制和迁移文件
COPY --from=builder /app/starbyte-server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/configs ./configs
COPY internal/export/service/fonts/OFL.txt /usr/share/licenses/starbyte/NotoSansSC-OFL.txt

# 复制迁移工具和 MinIO 客户端（mc 为占位脚本，见 Stage 1）
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=mc-builder /usr/local/bin/mc /usr/local/bin/mc

# 入口脚本：自动迁移 + 启动服务
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# ── 非特权运行（S-15）─────────────────────────────────────
# 进程不再以 root 启动：一旦应用被 RCE，攻击者拿到的是 uid 10001 而非容器 root，
# 无法直接改系统文件 / 装包 / 读取其它进程信息，横向移动成本显著上升。
# 需要预建并授权的可写路径：
#   /app/logs      logger 默认 logs/starbyte.log（相对 WORKDIR /app）
#   $HOME/.mc      entrypoint 的 `mc alias set` 会写配置文件
ENV HOME=/home/starbyte
RUN addgroup -S -g 10001 starbyte \
 && adduser -S -u 10001 -G starbyte -h /home/starbyte starbyte \
 && mkdir -p /app/logs /home/starbyte/.mc \
 && chown -R starbyte:starbyte /app /home/starbyte

USER starbyte

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
