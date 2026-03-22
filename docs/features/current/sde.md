---
status: active
doc_type: feature
owner: backend
last_reviewed: 2026-03-22
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
- 启动时和定时任务中检查最新 SDE 并导入 PostgreSQL
- 为 EVE 信息、舰队配置、SRP、自动 SRP 等模块提供名称与静态数据支撑

## 入口

- `GET /api/v1/sde/version`
- `POST /api/v1/sde/types`
- `POST /api/v1/sde/names`
- `POST /api/v1/sde/search`

前端封装位于 `static/src/api/sde.ts`。

## 运行时行为

- SDE 版本记录保存在 `sde_versions`
- 启动时会异步执行一次检查更新，cron 也会周期性执行
- 下载地址来自 `sde.download_url`
- `sde.proxy` 是可选配置
- 若配置了代理但代理不可达，下载器会自动回退为直连
- 导入目标是当前业务 PostgreSQL，而不是独立的只读 SDE 库

## 权限边界

- 当前这些路由在 router 中为 `Public`
- 语言优先级由 body / header / cookie 决定，最终默认 `en`

## 验证

- 后端基础校验：`cd server && go test ./...`
- 后端构建校验：`cd server && go build ./...`
- 前端类型校验：`cd static && pnpm exec vue-tsc --noEmit`
- 前端 SDE 名称解析回归测试：`cd static && pnpm test:unit`

## 关键不变量

- 旧文档里“网页 API 需要 API Key 鉴权”的说法不再代表当前实现
- SDE 是共享基础能力，修改返回结构时要检查多个业务模块
- 版本检查与导入流程是运行时基础设施问题，不要只从某个页面角度描述
- 英文名称缺失时，type/group/category/market group 查询会回退到 SDE 基础名称列
- `POST /api/v1/sde/names` 当前返回 `flat` 与 `names` 两套映射
- `names` 是按 namespace 分组的权威结果，`flat` 仅用于兼容单 namespace 调用方
- `docs/reference/sde-schema.sql` 仅是历史参考资产，不代表当前应用的实时 schema

## 主要代码文件

- `server/internal/handler/sde.go`
- `server/internal/service/sde.go`
- `server/internal/repository/sde.go`
- `server/internal/repository/sde_version.go`
- `server/internal/repository/sde_types.go`
- `server/internal/repository/sde_search.go`
- `server/internal/repository/sde_ships.go`
- `server/internal/router/router.go`
- `static/src/api/sde.ts`
