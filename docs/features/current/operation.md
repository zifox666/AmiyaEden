---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/fleet.go
  - server/internal/service/fleet_config.go
  - server/internal/service/auto_srp.go
  - static/src/api/fleet.ts
  - static/src/api/fleet-config.ts
  - static/src/views/operation
---

# Operation 模块

## 当前能力

- 舰队创建、列表、详情、编辑、删除
- 舰队成员同步、成员与 PAP 查询
- 邀请链接创建 / 停用 / 加入舰队
- 发放 PAP、查看 PAP 日志、查看我的 PAP
- 查看我的联盟 PAP
- 舰队配置管理、EFT 导出、从角色装配导入、导出到 ESI
- 用户侧系统钱包与流水
- 舰队级自动 SRP 模式：`disabled` / `submit_only` / `auto_approve`

## 入口

### 前端页面

- `static/src/views/operation/fleets`
- `static/src/views/operation/fleet-detail`
- `static/src/views/operation/fleet-configs`
- `static/src/views/operation/join`
- `static/src/views/operation/pap`

### 后端路由

- `/api/v1/operation/fleets/*`
- `/api/v1/operation/fleet-configs/*`
- `/api/v1/operation/wallet/*`

## 权限边界

- 路由级别默认要求登录
- `fleet-configs` 的创建、修改、删除和物品设置要求 `fc` 或 `srp`
- 舰队相关的细粒度拥有者 / FC / 管理员判断属于 service 层职责

## 关键不变量

- 舰队、PAP、舰队配置共享同一业务切片，修改时要一起考虑
- 自动 SRP 不是纯草案，当前模型、页面和后台处理逻辑都已存在
- 自动 SRP 的触发与舰队成员、KM 刷新、舰队配置装配有关，不能只改 UI 字段
- 联盟 PAP 的用户侧展示在 Operation，管理员配置与导入在 System

## 主要代码文件

- `server/internal/service/fleet.go`
- `server/internal/service/fleet_config.go`
- `server/internal/service/auto_srp.go`
- `server/internal/router/router.go`
- `static/src/api/fleet.ts`
- `static/src/api/fleet-config.ts`
- `static/src/views/operation`
