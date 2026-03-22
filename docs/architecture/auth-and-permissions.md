---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-03-22
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

`guest` 角色当前仍是已认证用户，但不是 `RequireLoginUser` 意义上的产品用户。
因此需要把 guest onboarding / self-service 能力单独挂在仅需 `JWTAuth()` 的路由上，而不是 `RequireLoginUser()`。

文档上应把这两类边界区分开：

- `JWT`：任意持有有效 JWT 的已认证用户，包含 `guest`
- `Login`：任意已认证且非 `guest` 的产品用户

## 绑定角色能力

登录后可继续管理绑定角色：

- `GET /api/v1/sso/eve/characters`
- `GET /api/v1/sso/eve/bind`
- `PUT /api/v1/sso/eve/primary/:character_id`
- `DELETE /api/v1/sso/eve/characters/:character_id`

这些接口与 `/api/v1/me` 一样，当前都属于 guest 可用的自助能力，权限边界应记为 `JWT`，不是 `Login`。

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

### RequireLoginUser

- 判断请求方是否至少拥有一个非 `guest` 角色
- 用于实现 API 文档中的 `Login` 边界
- 适合“任意产品用户可访问”的能力，不再用 `RequireRole(..., user)` 代替
- 不适用于 SSO 首次登录后的 guest onboarding 页面，例如 `/me`、`/sso/eve/characters`、`/menu/list` 以及 guest 可访问的自助信息页

### JWT-only 自助能力

当前这类路由的共同点是：用户已经完成 SSO 并拿到平台 JWT，但还可能停留在 `guest`。

典型例子：

- `/api/v1/me`
- `/api/v1/sso/eve/characters`
- `/api/v1/sso/eve/bind`
- `/api/v1/sso/eve/primary/:character_id`
- `/api/v1/sso/eve/characters/:character_id`
- `/api/v1/menu/list`

这类接口主要用于：

- 建立前端权限上下文
- 完成角色绑定与主角色调整
- 让 guest 在准入完成前仍能查看自己的基础信息或自助完成资料

`/api/v1/info/*` 当前不再属于 JWT-only 自助能力，而是 `RequireLoginUser()` 边界。

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

- `frontend`: 静态路由 + `meta.login` / `meta.roles`
- `backend`: 后端菜单接口 `/api/v1/menu/list`

静态路由模式下的约定：

- `meta.login = true` 表示任意非 `guest` 已登录产品用户可访问
- `meta.roles` 只用于真实的显式角色白名单
- 不要用 `meta.roles: ['admin', 'fc', 'user']` 之类写法冒充 `Login`

修改权限或菜单时，必须同时考虑：

- 后端路由保护
- 角色 / 菜单种子
- 前端路由元数据
- 按钮权限使用点

## 当前不变量

- 当前产品不是用户名 / 密码登录系统
- 角色编码以代码常量为准，不以文档中文称呼为准
- `allow_corporations` 的基线准入当前以主角色军团为准，不再按任意绑定角色放行；当列表为空时，当前默认回退到伏羲军团 Fuxi Legion（`98185110`）
- 非 `allow_corporations` 军团角色的 ESI corporation role 信号当前应被整体忽略，不参与权限判断或衍生任务判定
- 自动补 `admin` 的内置快捷规则当前仅接受允许军团中的 ESI corp role `Director`
- corp title 只参与显式 title mapping，不会因为标题名为 `Director` 就自动抬升为 `admin`
- 用户删除当前不是纯路由级能力：即使请求方拥有 `admin`，后端仍会阻止其删除 `super_admin` 或其他 `admin`
- 用户编辑当前也不是纯路由级能力：即使请求方拥有 `admin`，后端仍会阻止其编辑 `super_admin` 或其他 `admin`，且仅 `super_admin` 可分配 `admin/super_admin`
- 管理员用户列表 `/api/v1/system/user` 的角色展示与接口契约当前只认 `roles[]`，不应再依赖历史单值 `role`
- 细粒度权限不能只靠前端控制
- 旧兼容文档不能重新定义角色体系

## 重要 Caveat

### Auto-role Director Signal

自动权限映射里的 `Director -> admin` 内置快捷规则，使用的是 ESI corporation role 信号，不是 title 文本匹配。

因此：

- 真实判断输入来自 `eve_character_corp_role`
- 当 `allow_corporations` 已配置时，非允许军团角色不会保留 `eve_character_corp_role` 快照供后续判断
- `Director` 只是 corp title 名称时，不应被当作管理员快捷信号
- corp title 仍然可以通过 `esi_title_mapping` 参与显式映射，但那是配置行为，不是内置特判

这个区别是权限边界的一部分，文档和实现都必须保持一致。
