---
status: active
doc_type: guide
owner: backend
last_reviewed: 2026-03-20
source_of_truth:
  - server/jobs/esi_refresh.go
  - server/internal/router/router.go
  - static/src/api
---

# 新增 ESI 功能指南

## 适用范围

当你要在仓库中新增一个“从 ESI 拉数据并在前后端消费”的模块时，优先按这份流程走。

## 推荐顺序

1. 定义或扩展数据模型
2. 把模型接入数据库迁移
3. 编写或扩展 ESI 刷新任务
4. 编写 repository
5. 编写 service
6. 编写 handler
7. 注册路由
8. 编写 `static/src/api/*`
9. 更新 `static/src/types/api/api.d.ts`
10. 接入页面、路由、i18n
11. 如有需要，更新菜单种子与权限
12. 更新 `docs/api/route-index.md` 与对应 feature doc

## 必查位置

### 后端

- `server/internal/model/`
- `server/bootstrap/db.go`
- `server/pkg/eve/esi/`
- `server/internal/repository/`
- `server/internal/service/`
- `server/internal/handler/`
- `server/internal/router/router.go`

### 前端

- `static/src/api/`
- `static/src/types/api/api.d.ts`
- `static/src/router/modules/`
- `static/src/views/`
- `static/src/locales/langs/zh.json`
- `static/src/locales/langs/en.json`

### 可选离线参考

- `docs/reference/esi-openapi.json`
- `docs/reference/sde-schema.sql`

这些文件只用于离线查阅，不是当前运行时行为的权威来源。

如果你已经确认改动会落在 `server/pkg/eve/esi/` 内部实现细节，再补读该目录下的 `README.md`。但它只用于解释本地机制，不能覆盖 `docs/ai/repo-rules.md` 与 `docs/` 中的 canonical 规则。

## 任务系统要求

如果是可刷新的 ESI 数据：

- 使用现有队列体系
- 明确任务名、scope、刷新频率
- 考虑新角色首次全量刷新与后台周期调度
- 不要在 handler 里直接调用 CCP API
- **所有 ESI API 端点必须使用 `global.Config.EveSSO.ESIBaseURL` 和 `ESIAPIPrefix` 构建，禁止硬编码 EVE 官方 URL**
- 如果需要直接在 service 层调用 ESI，必须使用 `global.Config.EveSSO.ESIBaseURL` 或通过 `pkg/eve/esi/client.go` 的配置化客户端

## 前端要求

- 页面不直接发 HTTP 请求
- 文本全部本地化
- 如果是标准列表页，遵守 `docs/standards/frontend-table-pages.md`

## 文档要求

落地完成后至少更新：

- `docs/api/route-index.md`
- 对应 `docs/features/current/*.md`
- 若引入新的约束，再更新 `docs/ai/repo-rules.md` 或 `docs/standards`
