---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
  - server/internal/model/menu.go
  - static/src/router
---

# 架构总览

## 项目定位

`AmiyaEden` 是一个面向 EVE Online 联盟 / 军团运营的全栈平台，当前活跃实现覆盖：

- EVE SSO 登录与多角色绑定
- RBAC 角色、菜单、按钮权限
- 动态菜单与动态路由
- 舰队、PAP、舰队配置、自动 SRP 模式
- EVE 角色信息查询与 NPC 刷怪报表
- SRP 价格、申请、审核、发放
- 系统钱包、商店、抽奖
- 联盟 PAP、自动权限映射、Webhook、ESI 刷新队列
- SDE 查询接口

## 技术栈

| 层 | 当前实现 |
| --- | --- |
| Backend | Go, Gin, GORM, PostgreSQL, Redis, cron |
| Frontend | Vue 3, TypeScript, Vite, Pinia, Vue Router, Element Plus |
| Auth | EVE SSO + JWT |
| i18n | `vue-i18n` |

## 分层约束

### Backend

`router -> middleware -> handler -> service -> repository -> model`

### Frontend

`view -> api -> backend`

补充层：

- `hooks` 复用行为
- `store` 跨页面状态
- `router` 路由与守卫
- `types` 合同类型

## 当前模块切分

### 用户入口

- `/auth/login`
- `/auth/callback`
- `/api/v1/sso/eve/*`
- `/api/v1/me`

### 业务模块

- Dashboard
- Operation
- EveInfo
- SRP
- Shop
- System
- Result

这些模块同时存在于：

- 后端路由分组
- 菜单种子
- 前端路由模块或后端菜单模式

## 关键不变量

- 当前产品认证主路径是 EVE SSO，不是用户名密码
- 角色编码以 `server/internal/model/role.go` 为准
- 菜单 / 按钮权限以 `server/internal/model/menu.go` 与 `/api/v1/menu/list` 为准
- 前端支持 `frontend` 与 `backend` 两种菜单模式，修改时不能只改一边
- `AGENTS.md` 与 `docs/` 是唯一维护中的文档源
