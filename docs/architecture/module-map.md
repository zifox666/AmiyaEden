---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-03-21
source_of_truth:
  - server/internal
  - server/pkg
  - static/src
  - docs/ai/repo-rules.md
---

# 模块地图

## 目的

本文件回答两个问题：

- 这个仓库里某类职责应该去哪个目录找
- 修改某类能力时，通常要同时看哪些文件

它不是新的规范来源，而是 `docs/ai/repo-rules.md` 与现有 architecture / api / feature 文档的导航图。

## Backend 目录职责

| 路径 | 主要职责 | 修改时通常还要看 |
| --- | --- | --- |
| `server/bootstrap/` | 应用启动、配置装配、路由/日志/任务初始化 | `server/main.go`、`docs/architecture/runtime-and-startup.md` |
| `server/config/` | 配置结构与配置文件模板，包括 `EveSSOConfig`（SSO/ESI 端点、ClientID/Secret 等） | `README.md`、`docs/guides/local-development.md`、`docs/features/current/auth-and-characters.md` |
| `server/internal/router/` | 路由注册、分组、中间件挂载 | `server/internal/handler/`、`docs/api/route-index.md` |
| `server/internal/middleware/` | JWT、日志、CORS、统一请求前置逻辑 | `server/internal/router/`、`docs/architecture/auth-and-permissions.md` |
| `server/internal/handler/` | HTTP 请求解析与响应返回 | `server/internal/service/`、`server/pkg/response/` |
| `server/internal/service/` | 业务规则、权限判断、跨仓储编排、ESI/SSO 集成 | `server/internal/repository/`、对应 feature doc |
| `server/internal/repository/` | 数据访问、查询拼接、结果映射 | `server/internal/model/`、`docs/standards/testing-and-verification.md` |
| `server/internal/model/` | GORM 模型、菜单种子、职权常量 | `docs/architecture/routing-and-menus.md` |
| `server/jobs/` | 定时任务 | `server/bootstrap/`、`docs/architecture/runtime-and-startup.md` |
| `server/pkg/eve/` | EVE SSO / ESI 基础能力 | `server/internal/service/`、`docs/guides/adding-esi-feature.md` |
| `server/pkg/response/` | 统一响应封装 | `server/internal/handler/` |

## Frontend 目录职责

| 路径 | 主要职责 | 修改时通常还要看 |
| --- | --- | --- |
| `static/src/views/` | 页面视图与页面级状态 | `static/src/api/`、`static/src/hooks/`、对应 feature doc |
| `static/src/api/` | 前端 API 包装层 | `server/internal/handler/`、`static/src/types/api/api.d.ts` |
| `static/src/components/` | 可复用 UI 组件 | `static/src/hooks/`、`static/src/locales/` |
| `static/src/hooks/` | 共享逻辑、状态转换、纯 helper | `docs/guides/testing-guide.md` |
| `static/src/store/` | Pinia 状态 | `static/src/router/`、`static/src/hooks/core/useAuth.ts` |
| `static/src/router/` | 路由核心、守卫、菜单模式适配 | `server/internal/model/menu.go`、`docs/architecture/routing-and-menus.md` |
| `static/src/types/` | TS 类型、导入声明、契约类型 | `static/src/api/`、`server/internal/model/` |
| `static/src/locales/` | i18n 文案 | `docs/ai/repo-rules.md`「Non-Negotiable Rules」第 3 条 |

## 文档目录职责

| 路径 | 主要职责 |
| --- | --- |
| `docs/standards/` | 工程标准，回答“必须 / 不得 / 推荐” |
| `docs/architecture/` | 当前已存在结构与运行方式 |
| `docs/api/` | 接口约定、路由与边界 |
| `docs/features/current/` | 已落地功能行为 |
| `docs/guides/` | 操作步骤与实践指南 |
| `docs/reference/` | 参考资产，不作为当前实现裁决依据 |

## 常见任务落点

### 新增后端接口

通常需要一起看：

1. `server/internal/router/`
2. `server/internal/handler/`
3. `server/internal/service/`
4. `server/internal/repository/`
5. `static/src/api/`
6. `static/src/types/api/api.d.ts`
7. `docs/api/route-index.md`

### 修改权限、菜单、按钮点位

通常需要一起看：

1. `server/internal/model/menu.go`
2. `server/internal/router/`
3. `static/src/router/`
4. 页面里的 `v-auth`
5. `docs/architecture/routing-and-menus.md`

### 修改 ESI / SSO 行为

通常需要一起看：

1. `server/pkg/eve/`
2. `server/internal/service/`
3. `server/internal/handler/eve_sso.go`
4. `docs/features/current/auth-and-characters.md`
5. `docs/features/current/esi-refresh.md`
6. `docs/guides/adding-esi-feature.md`

### 修改表格页或复杂筛选页

通常需要一起看：

1. `static/src/views/目标页面`
2. `static/src/components/core/tables/`
3. `static/src/hooks/`
4. `docs/standards/frontend-table-pages.md`

## 读代码顺序建议

对不熟悉的模块，优先按下面顺序建立上下文：

1. 相关 feature doc
2. 页面或 handler 入口
3. 对应 service
4. 对应 repository / api wrapper
5. 关联类型与文档边界
