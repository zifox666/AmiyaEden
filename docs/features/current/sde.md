---
status: active
doc_type: feature
owner: backend
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/handler/sde.go
  - server/internal/service/sde.go
  - static/src/api/sde.ts
---

# SDE 模块

## 当前能力

- 查询当前已导入的 SDE 版本
- 批量查询 type 信息
- 批量查询 ID 到名称映射
- 模糊搜索物品 / 成员名称
- 为 EVE 信息、舰队配置、SRP、自动 SRP 等模块提供名称与静态数据支撑

## 入口

- `GET /api/v1/sde/version`
- `POST /api/v1/sde/types`
- `POST /api/v1/sde/names`
- `POST /api/v1/sde/search`

前端封装位于 `static/src/api/sde.ts`。

## 权限边界

- 当前这些路由在 router 中为 Public
- 语言优先级由 body / header / cookie 决定，最终默认 `en`

## 关键不变量

- 旧文档里“网页 API 需要 API Key 鉴权”的说法不再代表当前实现
- SDE 是共享基础能力，修改返回结构时要检查多个业务模块
- 版本检查与导入流程是运行时基础设施问题，不要只从某个页面角度描述
- `docs/reference/sde-schema.sql` 仅是历史参考资产，不代表当前应用的实时 schema

## 主要代码文件

- `server/internal/handler/sde.go`
- `server/internal/service/sde.go`
- `server/internal/router/router.go`
- `static/src/api/sde.ts`
