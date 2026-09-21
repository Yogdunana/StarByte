# ─────────────────────────────────────────────────────────
# Stage 1: 构建前端静态资源
# node:22-alpine 来自 docker.io。校园网拉失败时请用 daemon.json 的
# registry-mirrors 拉取后 retag，勿在本文件写死未验证的 registry。
# ─────────────────────────────────────────────────────────
FROM node:22-alpine AS builder

WORKDIR /app

# 先复制依赖文件，利用 Docker 缓存
COPY package*.json ./
RUN npm ci

# 复制源码并构建
COPY . .
RUN npm run build

# ─────────────────────────────────────────────────────────
# Stage 2: 开发镜像（带 Vite dev server + 热重载）
# ─────────────────────────────────────────────────────────
FROM node:22-alpine AS development

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

EXPOSE 5173

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

# ─────────────────────────────────────────────────────────
# Stage 3: 生产镜像（Nginx 托管静态文件 + API 代理）
# nginx:1.27-alpine 同样来自 docker.io；拉失败时用 daemon.json 镜像源
# 或 retag，勿写死未验证的 registry。
# ─────────────────────────────────────────────────────────
FROM nginx:1.27-alpine AS production

RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
# 安全头单独成片，供 server 块与静态资源 location 重复 include
# （nginx 的 add_header 不继承，带 add_header 的 location 会吃掉父级全部头）。
COPY security-headers.conf /etc/nginx/snippets/security-headers.conf

# ── 非特权运行（S-15）─────────────────────────────────────
# nginx 官方镜像自带 nginx 用户（uid 101）与组。以非 root 运行需要：
#   1) 预建/授权它必须写的路径（pid 文件与各级 temp 目录）；
#   2) 监听非特权端口 —— 非 root 无法 bind < 1024，故 nginx.conf 用 8080，
#      宿主机再由 compose 把 80 映射到容器 8080。
RUN mkdir -p /var/cache/nginx/client_temp /var/cache/nginx/proxy_temp \
             /var/cache/nginx/fastcgi_temp /var/cache/nginx/uwsgi_temp \
             /var/cache/nginx/scgi_temp /etc/nginx/snippets \
 && touch /var/run/nginx.pid \
 && chown -R nginx:nginx /var/cache/nginx /var/run/nginx.pid \
       /etc/nginx/conf.d /etc/nginx/snippets /usr/share/nginx/html

USER nginx

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
