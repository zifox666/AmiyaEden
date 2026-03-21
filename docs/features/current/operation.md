---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-22
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
- 查看军团 PAP 汇总
- 查看我的联盟 PAP
- 舰队配置管理、EFT 导出、从角色装配导入、导出到 ESI
- 舰队级自动 SRP 模式：`disabled` / `submit_only` / `auto_approve`

## 入口

### 前端页面

- `static/src/views/operation/fleets`
- `static/src/views/operation/fleet-detail`
- `static/src/views/operation/fleet-configs`
- `static/src/views/operation/corporation-pap`
- `static/src/views/operation/join`
- `static/src/views/operation/pap`

### 后端路由

- `/api/v1/operation/fleets/*`
- `/api/v1/operation/fleet-configs/*`

技能规划已拆分为独立模块，详见 `docs/features/current/skill-planning.md`。
用户侧钱包页面归属 Commerce / Shop 模块，详见 `docs/features/current/commerce.md`。

## 权限边界

- `fleets`、`fleet-detail` 页面访问要求 `super_admin`、`admin` 或 `fc`
- `fleet-configs` 页面访问要求 `super_admin`、`admin`、`fc` 或 `user`
- 舰队管理动作，包括刷新 ESI、成员同步 / 手动维护、PAP 发放、邀请链接与 Ping，要求 `super_admin`、`admin` 或 `fc`
- 舰队删除仍要求 `super_admin` 或 `admin`
- `fleet-configs` 的只读查询要求 `super_admin`、`admin`、`fc` 或 `user`
- `fleet-configs` 的导出到 ESI、创建、修改、删除、导入装配和物品设置要求 `super_admin`、`admin` 或 `fc`
- `corporation-pap` 在前端静态路由模式下允许 `super_admin`、`admin`、`fc`、`srp`、`user`
- 舰队相关角色边界同时由 router、菜单返回与前端路由元数据保持一致

## 关键不变量

- 舰队、PAP、舰队配置共享同一业务切片，修改时要一起考虑
- 自动 SRP 不是纯草案，当前模型、页面和后台处理逻辑都已存在
- 自动 SRP 的触发与舰队成员、KM 刷新、舰队配置装配有关，不能只改 UI 字段
- 联盟 PAP 的用户侧展示在 Operation，管理员配置与导入在 System
- 军团 PAP 页面属于多块统计 + 表格混排的分析页，当前明确允许不走 `useTable` / `ArtTable` 默认模板

## 主要代码文件

- `server/internal/service/fleet.go`
- `server/internal/service/fleet_config.go`
- `server/internal/service/auto_srp.go`
- `server/internal/router/router.go`
- `static/src/api/fleet.ts`
- `static/src/api/fleet-config.ts`
- `static/src/views/operation`
