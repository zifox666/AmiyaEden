---
status: active
doc_type: architecture
owner: frontend
last_reviewed: 2026-03-22
source_of_truth:
  - server/internal/model/menu.go
  - static/src/router/core
  - static/src/router/routes
  - static/src/router/modules
---

# 路由与菜单

## 当前模式

仓库同时支持两种路由装载方式：

- `frontend` 模式：前端静态路由模块
- `backend` 模式：后端菜单接口动态生成

这是当前真实能力，不是过渡代码。修改相关逻辑时必须保证两种模式都不被破坏。

## 后端菜单源

系统菜单种子定义在 `server/internal/model/menu.go`，核心职责：

- 定义目录 / 页面 / 按钮节点
- 为后端菜单接口提供树结构
- 为默认角色分配菜单
- 把按钮节点转换成前端 `meta.authList`

## 前端静态路由源

当前静态模块主要位于：

- `static/src/router/modules/dashboard.ts`
- `static/src/router/modules/operation.ts`
- `static/src/router/modules/skill-planning.ts`
- `static/src/router/modules/info.ts`
- `static/src/router/modules/shop.ts`
- `static/src/router/modules/srp.ts`
- `static/src/router/modules/system.ts`

基础静态路由位于：

- `static/src/router/routes/staticRoutes.ts`

静态路由权限约定：

- `meta.login = true` 对应 API / feature 文档中的 `Login`
- `meta.roles` 只表示显式角色白名单
- 同一路由不要再用 `meta.roles` 伪装“任意非 guest 登录用户”
- guest 可访问的 onboarding / self-service 页面不要错误标成 `meta.login = true`，因为这会把它们提升为“非 guest 才可访问”

## 动态路由核心

后端菜单模式的核心文件：

- `MenuProcessor.ts`
- `RouteTransformer.ts`
- `RouteRegistry.ts`
- `RoutePermissionValidator.ts`
- `guards/beforeEach.ts`

职责大致为：

1. 拉取菜单树
2. 转换为前端 route record
3. 校验组件路径和权限元数据
4. 注册到 router
5. 由守卫控制首次加载和访问

## 按钮权限流

1. 后端 `menu.type=button`
2. 转换为 `meta.authList`
3. 前端通过 `v-auth` 或权限 hook 消费

页面不要自己重新发明一套字符串判断规则。

## 当前不变量

- 菜单名称、路径、组件路径应在前后端保持一致
- 菜单可见性不应硬编码在页面内部
- `/api/v1/menu/list` 本身是 `JWT` 边界；guest 允许拿到其可见的有限菜单树
- 路由改动若涉及权限边界，必须同步更新 API / feature 文档
- 路由架构说明只维护在 `docs/` 中
