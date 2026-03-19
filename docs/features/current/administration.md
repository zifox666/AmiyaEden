---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/role.go
  - server/internal/service/menu.go
  - server/internal/service/user.go
  - server/internal/service/auto_role.go
  - server/internal/service/alliance_pap.go
  - server/internal/service/sys_webhook.go
  - static/src/api/system-manage.ts
  - static/src/api/webhook.ts
  - static/src/views/system
---

# System 管理模块

## 当前能力

- 基础配置读取与更新
- 菜单管理
- 角色管理与角色菜单分配
- 用户管理、用户角色分配
- 超级管理员模拟登录
- 自动权限映射（ESI corp roles / corp titles -> system roles）
- 联盟 PAP 列表、抓取、导入、兑换配置、月度结算
- Webhook 配置与测试

## 入口

### 前端页面

- `static/src/views/system/basic-config`
- `static/src/views/system/menu`
- `static/src/views/system/role`
- `static/src/views/system/user`
- `static/src/views/system/auto-role`
- `static/src/views/system/pap`
- `static/src/views/system/webhook`

### 后端路由

- `/api/v1/system/basic-config`
- `/api/v1/system/menu/*`
- `/api/v1/system/role/*`
- `/api/v1/system/user/*`
- `/api/v1/system/auto-role/*`
- `/api/v1/system/pap/*`
- `/api/v1/system/webhook/*`

## 权限边界

- `/system/*` 默认要求 `admin`
- `/system/user/:id/impersonate` 额外要求 `super_admin`

## 关键不变量

- 角色与菜单定义的 canonical 源在代码常量和菜单种子，不在旧文档
- 自动权限映射已经是当前功能，不是纯想法
- 联盟 PAP 的管理接口与用户查看接口分属不同模块
- Webhook 是系统配置能力，不应散落到页面里直接拼接

## 主要代码文件

- `server/internal/service/role.go`
- `server/internal/service/menu.go`
- `server/internal/service/user.go`
- `server/internal/service/auto_role.go`
- `server/internal/service/alliance_pap.go`
- `server/internal/service/sys_webhook.go`
- `static/src/api/system-manage.ts`
- `static/src/api/webhook.ts`
- `static/src/views/system`
