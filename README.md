# AmiyaEden

面向 EVE Online 联盟 / 军团的一体化运营平台。

本文件是仓库的 onboarding / product-facing 入口。若与工程规则、当前架构边界或接口裁决相关的说明发生冲突，以 [AGENTS.md](AGENTS.md) 与 [docs/](docs/README.md) 为准。

当前仓库的活跃实现包含：

- EVE SSO 登录与多角色绑定
- RBAC 角色 / 菜单 / 按钮权限
- 动态菜单与动态路由
- 舰队行动、PAP、舰队配置
- 技能规划、军团技能计划与完成度检查
- ESI 角色信息查询（钱包、技能、舰船、植入体、资产、合同、装配）
- SRP 补损申请、审核、价格表
- 系统钱包、商店
- 联盟 PAP、自动权限映射、Webhook 配置
- ESI 刷新队列与 SDE 查询接口

## 当前状态

- 后端：Go + Gin + GORM + PostgreSQL + Redis
- 前端：Vue 3 + TypeScript + Vite + Pinia + Vue Router
- 认证：EVE SSO + JWT
- 菜单模式：支持前端静态模式与后端菜单模式，当前仓库完整实现两种路径
- 登录流程：当前产品登录流以 EVE SSO 为主；仓库中仍有模板化的 `register` / `forget-password` 页面源码，但它们不是当前路由中的受支持流程

## 仓库结构

```text
AmiyaEden/
├── server/                 # Go 后端
│   ├── bootstrap/          # 配置 / 日志 / DB / Redis / 路由 / Cron 初始化
│   ├── config/             # config.yaml 与配置结构
│   ├── internal/
│   │   ├── handler/        # HTTP 处理层
│   │   ├── middleware/     # JWT、响应包装、日志、CORS 等
│   │   ├── model/          # GORM 模型与菜单种子
│   │   ├── repository/     # 数据访问层
│   │   ├── router/         # API 路由注册
│   │   └── service/        # 业务逻辑层
│   ├── jobs/               # 定时任务
│   └── pkg/                # JWT、EVE SSO / ESI、响应工具等
├── static/                 # Vue 前端
│   ├── src/api/            # API 调用层
│   ├── src/components/     # 共享组件
│   ├── src/hooks/          # 共享逻辑
│   ├── src/locales/        # 国际化
│   ├── src/router/         # 路由核心、守卫、模块定义
│   ├── src/store/          # Pinia store
│   ├── src/types/          # TS 类型定义
│   └── src/views/          # 页面视图
├── docs/                   # canonical 文档树（standards / architecture / api / features / drafts）
├── AGENTS.md               # 仓库工程约束（最高优先级）
└── docker-compose.example.yml
```

## 运行依赖

- Go `>= 1.24`
- Air（Go 热重载工具）
- Node.js `>= 20.19.0`
- pnpm `>= 8.8.0`
- PostgreSQL
- Redis

如果本机还没有 `air`，可以直接使用 Go 安装：

```bash
go install github.com/air-verse/air@latest
export PATH="$PATH:$(go env GOPATH)/bin"
air -v
```

## 本地启动

### 1. 后端配置

复制后端配置模板：

```bash
cp server/config/config.example.yaml server/config/config.yaml
```

最少需要确认这些配置：

- `server.port`
- `database.*`
- `redis.*`
- `jwt.secret`
- `eve_sso.client_id`
- `eve_sso.client_secret`
- `eve_sso.callback_url`
- `sde.api_key`

### 2. 准备基础设施

仓库提供了容器部署示例：

```bash
docker compose -f docker-compose.example.yml up -d postgres redis
```

如果你使用本机数据库 / Redis，也可以直接修改 `server/config/config.yaml` 指向对应实例。

### 3. 启动后端

```bash
make dev
```

`make dev` 会同时启动前端和后端。

后端通过 Air 热重载工具运行，源码变更后会自动重新编译并启动服务；前端会在 `static/` 下运行 `pnpm dev`。

按 `Ctrl-C` 会同时停止前端和后端。

启动时会执行：

- 配置加载
- 日志初始化
- PostgreSQL 连接
- Redis 连接
- GORM AutoMigrate
- 系统角色 / 菜单种子初始化
- 定时任务注册

### 4. 启动前端

当前仓库没有提交前端 `.env.example`，但已经提交了可直接作为起点的默认环境文件：

- `static/.env.development`
- `static/.env.development.local`
- `static/.env.production`

本地开发通常不需要从空白开始创建全部 Vite 变量；大多数情况下只需确认或覆盖以下常用项：

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

默认本地联调场景下：

- `VITE_API_PROXY_URL` 已指向 `http://localhost:8080`
- `VITE_API_URL` 在开发环境下可保持为 `/`

然后启动前端：

```bash
cd static
pnpm install
pnpm dev
```

## 认证说明

- 后端 SSO 回调默认路径：`/api/v1/sso/eve/callback`
- 前端登录页当前使用 `/auth/login`
- 前端回调页当前使用 `/auth/callback`
- 活跃登录方式是 EVE SSO；不要把未接入的用户名 / 密码模板页当作当前产品能力

## 常用校验命令

```bash
cd server && go test ./...
cd server && go build ./...
cd static && pnpm lint .
cd static && pnpm build
cd static && pnpm exec vue-tsc --noEmit
cd static && pnpm test:unit
```

## 文档入口

- 文档索引与信任顺序：[docs/README.md](docs/README.md)
- 仓库工程规范：[AGENTS.md](AGENTS.md)
- 测试与验证标准：[docs/standards/testing-and-verification.md](docs/standards/testing-and-verification.md)
- 当前架构：[docs/architecture/overview.md](docs/architecture/overview.md)
- API 路由索引：[docs/api/route-index.md](docs/api/route-index.md)
- Feature 状态说明：[docs/features/README.md](docs/features/README.md)

## 许可证

[LICENSE](LICENSE)
