# ============================================================
# Stage 1: 构建前端
# ============================================================
FROM node:22-alpine AS frontend-builder

WORKDIR /app

RUN npm install -g pnpm

# 先复制依赖描述文件，利用 Docker 层缓存
COPY static/package.json static/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# 复制前端源码
COPY static/ .

# Docker 部署：API 走相对路径 /，由 Nginx 代理到后端
ARG VITE_API_URL=/
ARG VITE_BASE_URL=/
ARG VITE_ACCESS_MODE=backend
ARG VITE_DROP_CONSOLE=true

ENV VITE_API_URL=$VITE_API_URL \
    VITE_BASE_URL=$VITE_BASE_URL \
    VITE_ACCESS_MODE=$VITE_ACCESS_MODE \
    VITE_DROP_CONSOLE=$VITE_DROP_CONSOLE

RUN pnpm build

# ============================================================
# Stage 2: 构建后端
# ============================================================
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

RUN apk add --no-cache git

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X amiya-eden/internal/service.currentVersion=${VERSION}" \
    -o server ./main.go

# ============================================================
# Stage 3: 运行时（nginx + supervisor + Go server）
# ============================================================
FROM alpine:3.20

# 安装运行时依赖
RUN apk add --no-cache tzdata ca-certificates nginx supervisor \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && mkdir -p /run/nginx /run/supervisord

WORKDIR /app

# 复制后端二进制
COPY --from=backend-builder /app/server ./server

# 将前端构建产物暂存在 /app/html-init
# entrypoint.sh 在首次运行时（volume 为空）将其复制到 /usr/share/nginx/html
COPY --from=frontend-builder /app/dist ./html-init

# 复制 nginx 和 supervisor 配置
COPY docker/nginx.conf /etc/nginx/http.d/default.conf
COPY docker/supervisord.conf /etc/supervisord.conf
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# 创建持久化目录
RUN mkdir -p logs uploads /usr/share/nginx/html

# 80: nginx（前端 + API 代理）
# 8080: Go server（内部，不建议直接暴露）
EXPOSE 80

ENTRYPOINT ["/entrypoint.sh"]
