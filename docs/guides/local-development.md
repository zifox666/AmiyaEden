---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-21
source_of_truth:
  - README.md
  - server/config/config.example.yaml
  - server/go.mod
  - static/package.json
---

# 本地开发指南

## 目的

本文件提供一个比 `README.md` 更面向开发者日常工作的入口，覆盖：

- 本地依赖
- 最小启动步骤
- 常用命令
- 常见前后端联调方式

## 运行依赖

- Go `>= 1.24`
- Node.js `>= 20.19.0`
- pnpm `>= 8.8.0`
- PostgreSQL
- Redis

## 后端启动

### 1. 准备配置

```bash
cp server/config/config.example.yaml server/config/config.yaml
```

至少检查：

- `server.port`
- `database.*`
- `redis.*`
- `jwt.secret`
- `eve_sso.client_id`
- `eve_sso.client_secret`
- `eve_sso.callback_url`
- `sde.api_key`

### 2. 准备依赖服务

```bash
docker compose -f docker-compose.example.yml up -d postgres redis
```

### 3. 启动后端

```bash
cd server
go run main.go
```

## 前端启动

### 1. 准备依赖

```bash
cd static
pnpm install
```

### 2. 准备 Vite 环境变量

当前仓库没有提交前端 `.env.example`。本地开发通常至少需要：

```bash
VITE_VERSION=dev
VITE_PORT=5173
VITE_BASE_URL=/
VITE_API_URL=http://localhost:8080
VITE_API_PROXY_URL=http://localhost:8080
VITE_ACCESS_MODE=backend
VITE_WITH_CREDENTIALS=false
VITE_LOCK_ENCRYPT_KEY=change_me
VITE_OPEN_ROUTE_INFO=false
```

### 3. 启动前端

```bash
cd static
pnpm dev
```

## 常用开发命令

### Backend

```bash
cd server && go test ./...
cd server && go build ./...
```

### Frontend

```bash
cd static && pnpm lint .
cd static && pnpm exec vue-tsc --noEmit
cd static && pnpm test:unit
cd static && pnpm build
```

## 常见开发场景

### 改后端接口

通常至少做这些事：

1. 修改 `handler -> service -> repository`
2. 更新 `static/src/api/`
3. 更新 `static/src/types/api/api.d.ts`
4. 跑 `go test ./...`、`go build ./...`
5. 跑 `pnpm exec vue-tsc --noEmit`

### 改前端页面

通常至少做这些事：

1. 修改 `view`
2. 如有共享逻辑，抽到 `hooks` 或 `components`
3. 补齐 i18n 文案
4. 跑 `pnpm lint .`
5. 跑 `pnpm exec vue-tsc --noEmit`

### 改风险逻辑或修 bug

除构建验证外，优先补回归测试，细则见：

- `AGENTS.md`
- `docs/standards/testing-and-verification.md`
- `docs/guides/testing-guide.md`

## 与文档配套阅读

- 想知道目录职责：看 `docs/architecture/module-map.md`
- 想知道架构规则：看 `docs/architecture/overview.md`
- 想知道接口与路由面：看 `docs/api/route-index.md`
- 想知道当前功能边界：看 `docs/features/README.md`
