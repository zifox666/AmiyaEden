# AmiyaEden

面向 EVE Online 联盟/军团的一体化管理平台，提供 SSO 登录、舰队行动、SRP 补损、ESI 数据同步等功能。

---

## 功能特性

- **EVE SSO 登录**：基于 EVE Online OAuth 2.0 身份认证，支持多角色绑定
- **舰队行动管理**：创建舰队、ESI 成员同步、PAP 出勤记录
- **SRP 补损系统**：关联击杀邮件的补损申请、审批与发放流程
- **ESI 数据自动刷新**：定时从 ESI API 拉取角色资产、击杀、合同等数据
- **SDE 静态数据**：自动从 GitHub Release 下载并导入最新 EVE 静态数据
- **角色权限控制**：多级角色体系（超级管理员 / 管理员 / 补损管理 / 舰队指挥 / 普通用户）
- **动态菜单路由**：后端下发菜单与路由配置，前端动态注册

---

## 技术栈

| 层     | 技术                                          |
| ------ | --------------------------------------------- |
| 后端   | Go · Gin · GORM · PostgreSQL · Redis · robfig/cron |
| 前端   | Vue 3 · TypeScript · Vite · Pinia · Vue Router |
| 认证   | EVE SSO OAuth 2.0 · HMAC-SHA256 JWT           |
| 日志   | zap · lumberjack                              |

---

## 环境要求

- Go >= 1.24
- Node.js >= 20.19.0
- pnpm >= 8.8.0
- PostgreSQL
- Redis

---

## 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/zifox666/AmiyaEden.git
cd AmiyaEden
```

### 2. 配置后端

复制配置模板并填写必要参数：

```bash
cd server
cp config/config.example.yaml config/config.yaml
```

需要修改的关键配置项：

| 配置项 | 说明 |
| --- | --- |
| `database.*` | PostgreSQL 连接信息 |
| `redis.*` | Redis 连接信息 |
| `jwt.secret` | JWT 签名密钥，生产环境务必修改 |
| `eve_sso.client_id` | EVE 开发者控制台申请的 Client ID |
| `eve_sso.client_secret` | EVE 开发者控制台申请的 Client Secret |
| `eve_sso.callback_url` | SSO 回调地址 |
| `sde.api_key` | SDE 查询接口的 API Key |

### 3. 启动后端

```bash
cd server

# 安装依赖
go mod download

# 运行（使用指定配置文件）
go run main.go -c ./config/config.yaml

# 或使用 Makefile
make run
```

首次启动会自动执行数据库迁移（AutoMigrate）。

### 4. 启动前端

```bash
cd static

# 安装依赖
pnpm install

# 开发模式
pnpm dev

# 构建生产包
pnpm build
```

---

## 项目结构

```
AmiyaEden/
├── server/                 # Go 后端
│   ├── main.go
│   ├── bootstrap/          # 启动引导（配置/数据库/Redis/路由/定时任务）
│   ├── config/             # 配置定义与配置文件
│   ├── global/             # 全局变量
│   ├── internal/
│   │   ├── handler/        # HTTP 处理器
│   │   ├── service/        # 业务逻辑
│   │   ├── repository/     # 数据访问层
│   │   ├── model/          # GORM 数据模型
│   │   ├── middleware/      # 中间件
│   │   └── router/         # 路由注册
│   ├── jobs/               # 定时任务
│   └── pkg/                # 公共工具包（JWT / ESI / SSO / Cache）
└── static/                 # Vue 3 前端
    └── src/
        ├── api/            # API 调用层
        ├── router/         # 动态路由核心
        ├── store/          # Pinia 状态管理
        └── views/          # 页面视图
```

详细的开发文档请参阅 [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)。

---

## EVE SSO 配置

前往 [EVE 开发者控制台](https://developers.eveonline.com/) 创建应用，回调地址需与配置文件中的 `eve_sso.callback_url` 保持一致。

本地开发默认回调地址：

```
http://localhost:8080/api/v1/sso/eve/callback
```

---

## 许可证

[LICENSE](LICENSE)
