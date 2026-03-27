---
status: active
doc_type: architecture
owner: frontend
last_reviewed: 2026-03-27
source_of_truth:
  - static/src/router/core
  - static/src/router/routes
  - static/src/router/modules
---

# 路由

## 前端路由模式

系统使用前端静态路由模式，所有路由定义在前端代码中。

## 前端路由源

当前静态模块主要位于：

- `static/src/router/modules/dashboard.ts`
- `static/src/router/modules/operation.ts`
- `static/src/router/modules/skill-planning.ts`
- `static/src/router/modules/info.ts`
- `static/src/router/modules/shop.ts`
- `static/src/router/modules/welfare.ts`
- `static/src/router/modules/srp.ts`
- `static/src/router/modules/system.ts`

基础静态路由位于：

- `static/src/router/routes/staticRoutes.ts`

静态路由权限约定：

- `meta.login = true` 对应 API / feature 文档中的 `Login`
- `meta.roles` 只表示显式角色白名单
- 同一路由不要再用 `meta.roles` 伪装"任意非 guest 登录用户"
- guest 可访问的 onboarding / self-service 页面不要错误标成 `meta.login = true`，因为这会把它们提升为"非 guest 才可访问"

## 按钮权限

前端通过 `v-auth` 或权限 hook 消费按钮权限，权限定义在路由的 `meta.authList` 中。

## 当前不变量

- 路由改动若涉及权限边界，必须同步更新 API / feature 文档
- 路由架构说明只维护在 `docs/` 中
