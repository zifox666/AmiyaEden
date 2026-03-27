---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-23
source_of_truth:
  - Makefile
  - .air.toml
  - server/config/config.example.yaml
  - server/go.mod
  - static/package.json
  - static/.env.development
  - static/.env.development.local
---

# 本地开发指南

## 目的

本指南只回答三件事：

- 首次本地开发要准备什么
- 日常开发怎么一条命令启动
- 出问题时先看哪里

## 快速开始

```bash
cp server/config/config.example.yaml server/config/config.yaml
docker compose -f docker-compose.example.yml up -d postgres redis
cd static && pnpm install && cd ..
make dev
```

说明：

- `make dev` 会同时启动后端（Air 热重载）和前端（Vite dev server）
- `Ctrl-C` 会同时停止前后端

## 依赖要求

- Go `>= 1.24.5`
- Node.js `24`（与 CI 保持一致，见根目录 `.nvmrc`）
- pnpm `10.32.1`（与 CI 保持一致）
- Air（Go 热重载工具）
- golangci-lint `v2.11.4`（与 CI 保持一致）
- PostgreSQL
- Redis

## 首次初始化

### 1. 准备后端配置

```bash
cp server/config/config.example.yaml server/config/config.yaml
```

至少检查这些字段：

- `server.port`
- `database.*`
- `redis.*`
- `jwt.secret`
- `eve_sso.client_id`
- `eve_sso.client_secret`
- `eve_sso.callback_url`
- `eve_sso.esi_base_url`
- `eve_sso.esi_api_prefix`
- `eve_sso.sso_authorize_url`
- `eve_sso.sso_token_url`
- `eve_sso.eve_images_base_url`
- `sde.api_key`

### 2. 启动依赖服务

```bash
docker compose -f docker-compose.example.yml up -d postgres redis
```

### 3. 安装 Air 与 golangci-lint（若未安装）

```bash
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4
export PATH="$PATH:$(go env GOPATH)/bin"
air -v
golangci-lint --version
```

### 4. 安装前端依赖（首次或依赖变更后）

```bash
cd static
pnpm install
cd ..
```

## 日常开发启动

```bash
make dev
```

实际行为：

- 后端：使用仓库根目录 `.air.toml`，监听 `server/` 代码变更并自动重启
- 前端：在 `static/` 下执行 `pnpm dev`
- 后端进程会在 `server/` 目录执行，因此 `./config` 这类相对路径可正常解析

## 可选：分开启动前后端

只在你明确需要分开调试时使用。

后端：

```bash
air -c .air.toml
```

前端：

```bash
cd static
pnpm dev
```

## 前端环境变量说明

仓库已提供默认开发环境文件：

- `static/.env.development`
- `static/.env.development.local`
- `static/.env.production`

本地联调常见关注项：

```bash
VITE_VERSION=dev
VITE_PORT=5173
VITE_BASE_URL=/
VITE_API_URL=http://localhost:8080
VITE_API_PROXY_URL=http://localhost:8080
VITE_WITH_CREDENTIALS=false
VITE_LOCK_ENCRYPT_KEY=change_me
VITE_OPEN_ROUTE_INFO=false
```

## 常用校验命令

参见 `docs/standards/testing-and-verification.md`（`Default Commands` 节）。


## 常见问题

`make dev` 提示找不到 `air`：

- 先执行安装命令
- 确认 `$(go env GOPATH)/bin` 在 `PATH` 里

`make dev` 前端启动失败：

- 在 `static/` 执行 `pnpm install`
- 确认 Node.js 与 pnpm 版本满足本指南要求

后端报配置文件读取失败：

- 检查 `server/config/config.yaml` 是否存在
- 检查文件名是否是 `config.yaml`（不是 `config.yml`）

## 配套文档

- 目录职责：`docs/architecture/module-map.md`
- 架构总览：`docs/architecture/overview.md`
- 接口与路由：`docs/api/route-index.md`
- 测试规范：`docs/standards/testing-and-verification.md`
- 测试实践：`docs/guides/testing-guide.md`
