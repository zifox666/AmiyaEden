---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/srp.go
  - static/src/api/srp.ts
  - static/src/views/srp
---

# SRP 模块

## 当前能力

- 舰船价格表查询、维护、删除
- 个人补损申请提交
- 我的补损申请列表
- 我的 KM、按舰队筛选 KM、KM 详情
- 审核列表、审核详情、审核通过 / 拒绝
- 发放补损

## 入口

### 前端页面

- `static/src/views/srp/apply`
- `static/src/views/srp/manage`
- `static/src/views/srp/prices`

### 后端路由

- `/api/v1/srp/prices`
- `/api/v1/srp/applications`
- `/api/v1/srp/applications/me`
- `/api/v1/srp/killmails/me`
- `/api/v1/srp/killmails/fleet/:fleet_id`
- `/api/v1/srp/killmails/detail`
- `/api/v1/srp/open-info-window`
- `/api/v1/srp/applications/:id/review`
- `/api/v1/srp/applications/:id/payout`

## 权限边界

- 价格新增 / 更新要求 `srp:price:add`
- 价格删除要求 `srp:price:delete`
- 审核和发放要求 `srp:review`
- 其余个人能力默认要求登录

## 关键不变量

- 审核与发放是分离的接口，不要假设它们是一步完成
- 价格表、舰队配置金额、自动 SRP 逻辑之间存在耦合
- 涉及 killmail、舰队、SDE 名称映射的改动要跨模块检查

## 主要代码文件

- `server/internal/service/srp.go`
- `server/internal/router/router.go`
- `static/src/api/srp.ts`
- `static/src/views/srp`
