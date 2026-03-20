---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
  - server/internal/middleware/auth.go
  - server/internal/model/role.go
  - server/internal/model/menu.go
  - static/src/store/modules/user.ts
---

# 认证与权限

## 登录架构

当前受支持的产品登录路径是 EVE SSO：

- 公共入口：`GET /api/v1/sso/eve/login`
- 回调入口：`GET /api/v1/sso/eve/callback`
- 前端页面：`/auth/login`、`/auth/callback`

登录成功后，前端通过 `GET /api/v1/me` 获取：

- 当前用户
- 绑定角色
- 角色列表
- 按钮权限列表

## 绑定角色能力

登录后可继续管理绑定角色：

- `GET /api/v1/sso/eve/characters`
- `GET /api/v1/sso/eve/bind`
- `PUT /api/v1/sso/eve/primary/:character_id`
- `DELETE /api/v1/sso/eve/characters/:character_id`

## 当前系统角色

`server/internal/model/role.go` 定义的 canonical 角色编码：

- `super_admin`
- `admin`
- `srp`
- `fc`
- `user`
- `guest`

不要再使用旧文档里的 `Administrator` 之类别名。

## JWT 中间件行为

`JWTAuth()` 当前会：

- 从 `Authorization: Bearer <token>` 或 `?token=` 提取 JWT
- 解析 `userID`、`characterID`、兼容字段 `userRole`
- 从角色服务加载 `roles`
- 从角色服务加载 `permissions`
- 写入 Gin Context

上下文键包括：

- `userID`
- `characterID`
- `userRole`
- `roles`
- `permissions`

## 权限检查

### RequireRole

- 判断用户是否拥有指定角色之一
- `super_admin` 自动通过
- 当前不是数值等级继承模型

### RequirePermission

- 判断用户是否拥有指定权限之一
- `super_admin` 自动通过
- 支持父权限前缀命中，例如持有 `srp` 时可满足 `srp:review`

## 菜单与按钮权限

权限模型基于：

- `role`
- `menu`
- `role_menu`
- `user_role`

`menu.type` 支持：

- `dir`
- `menu`
- `button`

按钮权限通过 `menu.permission` 进入前端 `meta.authList`，供 `v-auth` 与程序化检查使用。

## 前端模式

前端通过 `VITE_ACCESS_MODE` 支持：

- `frontend`: 静态路由 + `meta.roles`
- `backend`: 后端菜单接口 `/api/v1/menu/list`

修改权限或菜单时，必须同时考虑：

- 后端路由保护
- 角色 / 菜单种子
- 前端路由元数据
- 按钮权限使用点

## 当前不变量

- 当前产品不是用户名 / 密码登录系统
- 角色编码以代码常量为准，不以文档中文称呼为准
- 细粒度权限不能只靠前端控制
- 旧兼容文档不能重新定义角色体系
