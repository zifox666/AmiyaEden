# ---- Build Stage ----
FROM node:22-alpine AS builder

WORKDIR /app

# 安装 pnpm
RUN npm install -g pnpm

# 缓存依赖层
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# 复制源码
COPY . .

# Docker 部署：API 走相对路径 /，由 Nginx 代理到后端
# 可通过 build-arg 覆盖：docker build --build-arg VITE_API_URL=https://your-domain.com
ARG VITE_API_URL=/
ARG VITE_BASE_URL=/
ARG VITE_DROP_CONSOLE=true

ENV VITE_API_URL=$VITE_API_URL \
    VITE_BASE_URL=$VITE_BASE_URL \
    VITE_DROP_CONSOLE=$VITE_DROP_CONSOLE

RUN pnpm build

# ---- Runtime Stage ----
FROM nginx:1.27-alpine

# 复制自定义 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
