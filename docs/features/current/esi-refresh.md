---
status: active
doc_type: feature
owner: backend
last_reviewed: 2026-03-20
source_of_truth:
  - server/jobs/esi_refresh.go
  - server/internal/handler/esi_refresh.go
  - static/src/api/esi-refresh.ts
  - static/src/views/system/esi-refresh
---

# ESI 刷新队列

## 当前能力

- 周期性运行 ESI 刷新队列
- 查看任务列表与状态
- 手动执行队列调度
- 按任务名执行
- 对指定人物执行全部任务
- 新人物登录 / 绑定后触发同步钩子
- 舰队 PAP 触发 KM 刷新
- 舰队自动 SRP 触发后台处理

## 入口

- 管理页面：`static/src/views/system/esi-refresh`
- 路由：`/api/v1/esi/refresh/*`
- 运行时调度：`server/jobs/esi_refresh.go`

## 权限边界

- 所有 `/api/v1/esi/refresh/*` 路由要求 `admin`

## 关键不变量

- 新增 ESI 数据模块时，通常不只改一个 handler，还需要任务注册、scope、持久化、前端消费一起落地
- 队列与登录后同步钩子共享同一套任务体系
- 如果要新增模块，请先遵循 `docs/guides/adding-esi-feature.md`
- 所有 ESI API 端点通过 `server/config/config.go` 中的 `EveSSOConfig.ESIBaseURL` 和 `ESIAPIPrefix` 配置管理，禁止在 service 层硬编码 ESI URL
- ESI 刷新队列通过接口注入（`TokenService`、`CharacterRepository`）避免循环依赖，不直接依赖具体 service / repository 实现

## 主要代码文件

- `server/jobs/esi_refresh.go`
- `server/internal/handler/esi_refresh.go`
- `server/pkg/eve/esi`
- `static/src/api/esi-refresh.ts`
- `static/src/views/system/esi-refresh`
