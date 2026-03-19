---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
  - server/internal/handler/eve_sso.go
  - server/internal/service/eve_sso.go
  - static/src/api/auth.ts
  - static/src/views/auth
---

# 认证与角色绑定

## 当前能力

- 通过 EVE SSO 登录
- 首次登录创建 / 关联用户
- 一个用户绑定多个角色
- 设置主角色
- 解绑角色
- 通过 `/api/v1/me` 获取角色、权限、绑定角色信息

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

## 权限边界

- 登录入口与回调是 Public
- 角色管理接口要求 JWT
- 角色和权限的最终决策在后端完成，前端只消费结果

## 关键不变量

- 当前受支持的产品登录方式是 EVE SSO
- `register` / `forget-password` 页面源码存在，但不是当前产品规范
- `/api/v1/me` 是前端启动权限上下文的关键接口
- 角色编码与权限列表必须与后端返回保持一致，不做前端别名映射
