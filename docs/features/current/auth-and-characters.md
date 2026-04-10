---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-04-10
source_of_truth:
  - server/internal/router/router.go
  - server/internal/handler/eve_sso.go
  - server/internal/handler/me.go
  - server/internal/service/eve_sso.go
  - server/internal/service/user.go
  - static/src/api/auth.ts
  - static/src/router/guards/beforeEach.ts
  - static/src/views/auth
  - static/src/views/dashboard/characters
---

# 认证与人物绑定

## 当前能力

- 通过 EVE SSO 登录
- 首次登录创建 / 关联用户
- 一个用户绑定多个人物
- 设置主人物
- 解绑人物
- 通过 `/api/v1/me` 获取当前用户、职权、权限与绑定人物信息
- 通过 `/api/v1/me` 维护昵称、QQ、Discord ID 资料
- 未填写昵称或未提供 QQ / Discord 任一联系方式时，前端强制停留在 `/dashboard/characters`，且尝试访问其他页面时弹出原因提示
- 可选启用：任一已绑定人物 ESI 失效时，前端强制停留在 `/dashboard/characters`，且尝试访问其他页面时弹出原因提示
- 主人物 ESI 已失效时，`/api/v1/me` 仍返回启动上下文，前端强制停留在 `/dashboard/characters` 直到主人物重新授权，且尝试访问其他页面时弹出原因提示

## 入口

### 前端

- `/auth/login`
- `/auth/callback`

### 后端

- `GET /api/v1/sso/eve/login`
- `GET /api/v1/sso/eve/callback`
- `GET /api/v1/sso/eve/scopes`
- `GET /api/v1/sso/eve/characters`
- `GET /api/v1/sso/eve/bind`
- `PUT /api/v1/sso/eve/primary/:character_id`
- `DELETE /api/v1/sso/eve/characters/:character_id`
- `GET /api/v1/me`
- `PUT /api/v1/me`

## 权限边界

- 登录入口与回调是 `Public`
- `/api/v1/me` 与人物绑定相关接口要求有效 `JWT`，允许 `guest` 使用
- `guest` 通过这些接口完成权限上下文建立、人物绑定与资料补全，再决定是否能进入 `Login` 边界的业务页面
- `/api/v1/me` 会返回主人物 `token_invalid` 状态，供前端将用户锁定在 `/dashboard/characters` 直到主人物重新授权
- `/api/v1/me` 同时返回 `enforce_character_esi_restriction`，供前端路由守卫决定是否对非主人物失效 ESI 启用页面停留限制
- 首次 SSO 登录时，若主人物所属军团在 `allow_corporations` 内，后端会直接赋予 `user`；该列表运行时始终包含代码常量中的伏羲军团 Fuxi Legion（`98185110`）
- 首次 SSO 登录时，若主人物 ID 在 `config.yaml` 的 `app.super_admins` 列表中，后端会直接赋予 `super_admin`
- 每次 SSO 登录时，`SyncConfigSuperAdmins` 会根据配置文件自动同步 `super_admin` 职权
- 职权和权限的最终决策在后端完成，前端只消费结果

## 关键不变量

- 当前受支持的产品登录方式是 EVE SSO
- `register` 页面源码仍存在，但不是当前产品规范；`forget-password` 页面已移除
- `/api/v1/me` 是前端启动权限上下文的关键接口
- `/api/v1/me` 不是"非 guest 才可访问"的业务接口，而是登录后立即可用的自助上下文接口
- 当前登录后必须完成昵称与联系方式资料，才允许继续访问其他业务页面
- QQ / Discord ID 的默认管理入口是 `/api/v1/me`；管理员侧 `/api/v1/system/user/:id` 仅允许 `super_admin` 维护非 `super_admin` 用户的联系方式
- 当前登录后若系统配置 `auth.enforce_character_esi_restriction = true`，则还必须保证所有已绑定人物的 ESI 有效，才允许离开 `/dashboard/characters`
- 无论系统配置是否开启，主人物 ESI 已失效都会强制前端停留在 `/dashboard/characters`，直到主人物重新授权；不会自动退出登录
- 用户仍被锁定在 `/dashboard/characters` 时，尝试导航到其他页面会弹出警告消息框，说明资料未完成、主人物 ESI 失效或其他绑定人物 ESI 失效等原因
- 重新绑定已存在人物会沿用当前 SSO 回调流程刷新该人物 token 并清除 `token_invalid`
- QQ / Discord ID 的唯一性由后端校验
- 职权编码与权限列表必须与后端返回保持一致，不做前端别名映射
- 所有 EVE SSO 端点（授权、令牌、图片服务等）通过 `server/config/config.go` 中的 `EveSSOConfig` 配置管理，禁止硬编码 URL

## 主要代码文件

- `server/internal/handler/eve_sso.go`
- `server/internal/handler/me.go`
- `server/internal/service/eve_sso.go`
- `server/internal/service/user.go`
- `server/internal/router/router.go`
- `static/src/api/auth.ts`
- `static/src/views/auth`
- `static/src/views/dashboard/characters`
