---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/main.go
  - server/bootstrap
  - server/config/config.example.yaml
---

# 运行与启动

## 依赖

- Go `>= 1.24`
- Node.js `>= 20.19.0`
- pnpm `>= 8.8.0`
- PostgreSQL
- Redis

## 配置入口

### 后端

- 模板：`server/config/config.example.yaml`
- 本地文件：`server/config/config.yaml`

通常至少需要确认：

- `server.port`
- `database.*`
- `redis.*`
- `jwt.secret`
- `eve_sso.client_id`
- `eve_sso.client_secret`
- `eve_sso.callback_url`
- `sde.api_key`

### 前端

仓库当前没有提交前端 `.env.example`。本地开发通常需要至少提供：

- `VITE_PORT`
- `VITE_API_URL`
- `VITE_API_PROXY_URL`
- `VITE_ACCESS_MODE`
- `VITE_BASE_URL`

## 本地启动

### 基础设施

```bash
docker compose -f docker-compose.example.yml up -d postgres redis
```

### 后端

```bash
cp server/config/config.example.yaml server/config/config.yaml
cd server
go mod download
go run main.go
```

### 前端

```bash
cd static
pnpm install
pnpm dev
```

## 后端启动顺序

`server/main.go` 当前启动流程：

1. 初始化配置
2. 初始化日志
3. 初始化 JWT
4. 初始化数据库
5. 初始化 Redis
6. 初始化 cron
7. 异步检查 SDE
8. 注册 ESI scopes
9. 初始化 HTTP 路由
10. 启动服务

## 数据库初始化副作用

数据库初始化不仅建立连接，还会执行：

- `AutoMigrate`
- 系统角色种子初始化
- 系统菜单种子初始化
- 历史 `user.role` 到 `user_role` 的迁移

## 运行时提示

- 新角色 SSO 成功后，后台会触发 ESI 全量刷新与自动权限同步
- ESI 刷新队列按 cron 调度，不要求手工逐个任务启动
- `register` / `forget-password` 页面源码仍在仓库中，但不是当前支持的登录架构
