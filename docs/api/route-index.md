---
status: active
doc_type: api
owner: engineering
last_reviewed: 2026-03-22
source_of_truth:
  - server/internal/router/router.go
---

# API 路由索引

## 说明

本文件只记录当前 `server/internal/router/router.go` 已注册的路由分组、路径与主要权限边界。
权限列说明：

- `JWT`：任意持有有效 JWT 的已认证用户可访问，包含 `guest`
- `Login`：任意已认证且非 `guest` 的产品用户可访问

## Public

### EVE SSO

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/sso/eve/login` | 获取 SSO 登录地址 | Public |
| GET | `/sso/eve/callback` | 处理 SSO 回调 | Public |
| GET | `/sso/eve/scopes` | 获取当前注册 scopes | Public |

### SDE

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/sde/version` | 当前 SDE 版本 | Public |
| POST | `/sde/types` | 批量查询 type 信息 | Public |
| POST | `/sde/names` | 批量查询名称映射 | Public |
| POST | `/sde/search` | 模糊搜索物品 / 成员 | Public |

## Authenticated Base

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/sso/eve/characters` | 当前用户绑定角色 | JWT |
| GET | `/sso/eve/bind` | 获取绑定新角色的 SSO 地址 | JWT |
| PUT | `/sso/eve/primary/:character_id` | 设为主角色 | JWT |
| DELETE | `/sso/eve/characters/:character_id` | 解绑角色 | JWT |
| GET | `/me` | 当前用户、角色、权限、绑定角色 | JWT |
| PUT | `/me` | 更新当前用户昵称 / QQ / Discord ID | JWT |
| POST | `/dashboard` | Dashboard 聚合数据 | JWT |
| POST | `/notification/list` | 通知列表 | JWT |
| POST | `/notification/unread-count` | 未读数 | JWT |
| POST | `/notification/read` | 标记已读 | Login |
| POST | `/notification/read-all` | 全部已读 | Login |
| GET | `/menu/list` | 当前用户菜单树 | JWT |

## Operation

### Fleets

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/operation/fleets` | 创建舰队 | `RequireRole(admin, fc)` |
| GET | `/operation/fleets` | 舰队列表 | `RequireRole(admin, fc)` |
| GET | `/operation/fleets/me` | 我的舰队 | Login |
| GET | `/operation/fleets/:id` | 舰队详情 | `RequireRole(admin, fc)` |
| PUT | `/operation/fleets/:id` | 更新舰队 | `RequireRole(admin, fc)` |
| DELETE | `/operation/fleets/:id` | 删除舰队 | `RequireRole(admin)` |
| POST | `/operation/fleets/:id/refresh-esi` | 刷新舰队 ESI 数据 | `RequireRole(admin, fc)` |
| GET | `/operation/fleets/:id/members` | 舰队成员 | `RequireRole(admin, fc)` |
| GET | `/operation/fleets/:id/members-pap` | 舰队成员与 PAP | `RequireRole(admin, fc)` |
| POST | `/operation/fleets/:id/members/manual` | 手动添加成员 | `RequireRole(admin, fc)` |
| POST | `/operation/fleets/:id/members/sync` | 同步 ESI 成员 | `RequireRole(admin, fc)` |
| POST | `/operation/fleets/:id/pap` | 发放 PAP | `RequireRole(admin, fc)` |
| GET | `/operation/fleets/:id/pap` | PAP 日志 | `RequireRole(admin, fc)` |
| GET | `/operation/fleets/pap/me` | 我的 PAP 日志 | Login |
| GET | `/operation/fleets/pap/corporation` | 军团 PAP 汇总 | Login |
| GET | `/operation/fleets/pap/alliance` | 我的联盟 PAP | Login |
| POST | `/operation/fleets/:id/invites` | 创建邀请 | `RequireRole(admin, fc)` |
| GET | `/operation/fleets/:id/invites` | 邀请列表 | `RequireRole(admin, fc)` |
| DELETE | `/operation/fleets/invites/:invite_id` | 停用邀请 | `RequireRole(admin, fc)` |
| POST | `/operation/fleets/join` | 加入舰队 | Login |
| GET | `/operation/fleets/esi/:character_id` | 查询角色当前舰队 | Login |
| POST | `/operation/fleets/:id/ping` | 发送 Webhook Ping | `RequireRole(admin, fc)` |

### Fleet Configs

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/operation/fleet-configs` | 配置列表 | Login |
| GET | `/operation/fleet-configs/:id` | 配置详情 | Login |
| GET | `/operation/fleet-configs/:id/eft` | 获取 EFT 文本 | Login |
| POST | `/operation/fleet-configs` | 创建配置 | `RequireRole(admin, fc)` |
| PUT | `/operation/fleet-configs/:id` | 更新配置 | `RequireRole(admin, fc)` |
| DELETE | `/operation/fleet-configs/:id` | 删除配置 | `RequireRole(admin, fc)` |
| POST | `/operation/fleet-configs/import-fitting` | 从角色装配导入 | `RequireRole(admin, fc)` |
| POST | `/operation/fleet-configs/export-esi` | 导出到 ESI | Login |
| GET | `/operation/fleet-configs/:id/fittings/:fitting_id/items` | 装配物品 | Login |
| PUT | `/operation/fleet-configs/:id/fittings/:fitting_id/items/settings` | 更新物品设置 | `RequireRole(admin, fc)` |

## Skill Planning

### Skill Plans

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/skill-planning/skill-plans/check/selection` | 获取当前用户保存的完成度检查角色选择 | Login |
| PUT | `/skill-planning/skill-plans/check/selection` | 保存当前用户的完成度检查角色选择 | Login |
| POST | `/skill-planning/skill-plans/check/run` | 执行技能规划完成度检查 | Login |
| GET | `/skill-planning/skill-plans/check/plan-selection` | 获取当前用户保存的完成度检查规划选择 | Login |
| PUT | `/skill-planning/skill-plans/check/plan-selection` | 保存当前用户的完成度检查规划选择 | Login |
| GET | `/skill-planning/skill-plans` | 技能计划列表 | `RequireRole(admin, fc)` |
| GET | `/skill-planning/skill-plans/:id` | 技能计划详情 | `RequireRole(admin, fc)` |
| POST | `/skill-planning/skill-plans` | 创建技能计划 | `RequireRole(admin, fc)` |
| PUT | `/skill-planning/skill-plans/:id` | 更新技能计划 | `RequireRole(admin, fc)` |
| DELETE | `/skill-planning/skill-plans/:id` | 删除技能计划 | `RequireRole(admin, fc)` |

## Info

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/info/wallet` | 钱包流水 | JWT |
| POST | `/info/skills` | 技能列表 | JWT |
| POST | `/info/ships` | 舰船列表 | JWT |
| POST | `/info/implants` | 植入体 | JWT |
| POST | `/info/assets` | 资产 | JWT |
| POST | `/info/contracts` | 合同列表 | JWT |
| POST | `/info/contracts/detail` | 合同详情 | JWT |
| POST | `/info/fittings` | 装配列表 | JWT |
| POST | `/info/fittings/save` | 保存装配 | JWT |
| POST | `/info/npc-kills` | 个人 NPC 刷怪报表 | JWT |
| POST | `/info/npc-kills/all` | 全部 NPC 刷怪报表 | JWT |

## Shop

### User Side

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/shop/wallet/my` | 我的钱包 | Login |
| POST | `/shop/wallet/my/transactions` | 我的钱包流水 | Login |
| POST | `/shop/products` | 商品列表 | Login |
| POST | `/shop/product/detail` | 商品详情 | Login |
| POST | `/shop/buy` | 购买商品 | Login |
| POST | `/shop/orders` | 我的订单 | Login |
| POST | `/shop/redeem/list` | 我的兑换码 | Login |

## Welfare

### User Side

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/welfare/eligible` | 可申请福利列表 | Login |
| POST | `/welfare/apply` | 申请福利 | Login |
| POST | `/welfare/my-applications` | 我的福利申请 | Login |

## Upload

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/upload/image` | 上传图片 | Login |

## SRP

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/srp/prices` | 价格表 | Login |
| POST | `/srp/prices` | 新增或更新价格 | `RequirePermission(srp:price:add)` |
| DELETE | `/srp/prices/:id` | 删除价格 | `RequirePermission(srp:price:delete)` |
| POST | `/srp/applications` | 提交补损申请 | Login |
| GET | `/srp/applications/me` | 我的补损申请 | Login |
| GET | `/srp/killmails/me` | 我的 KM | Login |
| GET | `/srp/killmails/fleet/:fleet_id` | 指定舰队 KM | Login |
| POST | `/srp/killmails/detail` | KM 详情 | Login |
| POST | `/srp/open-info-window` | 打开游戏内信息窗口 | Login |
| GET | `/srp/applications` | 审核列表 | `RequirePermission(srp:review)` |
| PUT | `/srp/applications/auto-approve` | 对指定 `fleet_id` 自动审批符合规则的待审批申请 | `RequirePermission(srp:review)` |
| GET | `/srp/applications/batch-payout-summary` | 批量发放汇总 | `RequirePermission(srp:review)` |
| GET | `/srp/applications/:id` | 审核详情 | `RequirePermission(srp:review)` |
| PUT | `/srp/applications/:id/review` | 审核申请 | `RequirePermission(srp:review)` |
| PUT | `/srp/applications/:id/payout` | 发放补损 | `RequirePermission(srp:review)` |
| PUT | `/srp/applications/users/:user_id/payout` | 按用户批量发放补损 | `RequirePermission(srp:review)` |

## ESI Refresh

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/esi/refresh/tasks` | 任务列表 | `RequireRole(admin)` |
| GET | `/esi/refresh/statuses` | 状态汇总 | `RequireRole(admin)` |
| POST | `/esi/refresh/run` | 执行队列调度 | `RequireRole(admin)` |
| POST | `/esi/refresh/run-task` | 按名称执行任务 | `RequireRole(admin)` |
| POST | `/esi/refresh/run-all` | 对角色执行全部任务 | `RequireRole(admin)` |

## System

所有 `/system/*` 路由默认要求 `RequireRole(admin)`。

### Basic Config

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/system/basic-config` | 获取基础配置 | `RequireRole(admin)` |
| PUT | `/system/basic-config` | 更新基础配置 | `RequireRole(admin)` |

### SDE Config

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/system/sde-config` | 获取 SDE 配置 | `RequireRole(admin)` |
| PUT | `/system/sde-config` | 更新 SDE 配置 | `RequireRole(admin)` |

### NPC Kills / Alliance PAP

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/system/npc-kills` | 公司级 NPC 刷怪报表 | `RequireRole(admin)` |
| GET | `/system/pap` | 联盟 PAP 列表 | `RequireRole(admin)` |
| POST | `/system/pap/fetch` | 手动抓取联盟 PAP | `RequireRole(admin)` |
| POST | `/system/pap/import` | 导入联盟 PAP | `RequireRole(admin)` |
| GET | `/system/pap/config` | 获取兑换配置 | `RequireRole(admin)` |
| PUT | `/system/pap/config` | 设置兑换配置 | `RequireRole(admin)` |
| POST | `/system/pap/settle` | 月度结算 | `RequireRole(admin)` |

### Menu / Role / User

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/system/menu/tree` | 菜单树 | `RequireRole(admin)` |
| POST | `/system/menu` | 创建菜单 | `RequireRole(admin)` |
| PUT | `/system/menu/:id` | 更新菜单 | `RequireRole(admin)` |
| DELETE | `/system/menu/:id` | 删除菜单 | `RequireRole(admin)` |
| GET | `/system/role` | 角色列表 | `RequireRole(admin)` |
| GET | `/system/role/all` | 全量角色列表 | `RequireRole(admin)` |
| GET | `/system/role/:id` | 角色详情 | `RequireRole(admin)` |
| POST | `/system/role` | 创建角色 | `RequireRole(admin)` |
| PUT | `/system/role/:id` | 更新角色 | `RequireRole(admin)` |
| DELETE | `/system/role/:id` | 删除角色 | `RequireRole(admin)` |
| GET | `/system/role/:id/menus` | 获取角色菜单 | `RequireRole(admin)` |
| PUT | `/system/role/:id/menus` | 设置角色菜单 | `RequireRole(admin)` |
| GET | `/system/user` | 用户列表；角色字段仅返回有序 `roles[]`，不再返回历史单值 `role` | `RequireRole(admin)` |
| GET | `/system/user/:id` | 用户详情 | `RequireRole(admin)` |
| PUT | `/system/user/:id` | 更新用户昵称 / QQ / Discord ID / 状态；`admin` 不可编辑 `super_admin` 或其他 `admin` | `RequireRole(admin)` |
| DELETE | `/system/user/:id` | 删除用户；`admin` 不可删除 `super_admin` 或其他 `admin` | `RequireRole(admin)` |
| GET | `/system/user/:id/roles` | 获取用户角色 | `RequireRole(admin)` |
| PUT | `/system/user/:id/roles` | 设置用户角色；仅 `super_admin` 可编辑管理员账号或分配 `admin/super_admin` | `RequireRole(admin)` |
| POST | `/system/user/:id/impersonate` | 模拟登录，需 `super_admin` | `RequireRole(admin)` + `super_admin` |

### System Wallet

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/system/wallet/list` | 钱包列表 | `RequireRole(admin)` |
| POST | `/system/wallet/detail` | 钱包详情 | `RequireRole(admin)` |
| POST | `/system/wallet/adjust` | 调整余额 | `RequireRole(admin)` |
| POST | `/system/wallet/transactions` | 钱包流水 | `RequireRole(admin)` |
| POST | `/system/wallet/logs` | 调整日志 | `RequireRole(admin)` |

### Welfare Admin

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/system/welfare/list` | 福利列表 | `RequireRole(admin)` |
| POST | `/system/welfare/add` | 创建福利 | `RequireRole(admin)` |
| POST | `/system/welfare/edit` | 编辑福利 | `RequireRole(admin)` |
| POST | `/system/welfare/delete` | 删除福利 | `RequireRole(admin)` |
| POST | `/system/welfare/import` | 导入历史福利记录 | `RequireRole(admin)` |
| POST | `/system/welfare/applications` | 福利申请列表（审批端） | `RequireRole(admin)` |
| POST | `/system/welfare/review` | 审批福利申请（发放/拒绝） | `RequireRole(admin)` |

### Shop Admin

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| POST | `/system/shop/product/list` | 商品列表 | `RequireRole(admin)` |
| POST | `/system/shop/product/add` | 新增商品 | `RequireRole(admin)` |
| POST | `/system/shop/product/edit` | 编辑商品 | `RequireRole(admin)` |
| POST | `/system/shop/product/delete` | 删除商品 | `RequireRole(admin)` |
| POST | `/system/shop/order/list` | 订单列表 | `RequireRole(admin)` |
| POST | `/system/shop/order/approve` | 审批订单 | `RequireRole(admin)` |
| POST | `/system/shop/order/reject` | 驳回订单 | `RequireRole(admin)` |
| POST | `/system/shop/redeem/list` | 兑换码列表 | `RequireRole(admin)` |

### Auto Role / Webhook

| Method | Path | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | `/system/auto-role/esi-roles` | ESI corp roles 列表 | `RequireRole(admin)` |
| GET | `/system/auto-role/esi-role-mappings` | ESI role 映射列表 | `RequireRole(admin)` |
| POST | `/system/auto-role/esi-role-mappings` | 新增 ESI role 映射 | `RequireRole(admin)` |
| DELETE | `/system/auto-role/esi-role-mappings/:id` | 删除 ESI role 映射 | `RequireRole(admin)` |
| GET | `/system/auto-role/corp-titles` | Corp titles 列表（含军团名称） | `RequireRole(admin)` |
| GET | `/system/auto-role/esi-title-mappings` | Title 映射列表 | `RequireRole(admin)` |
| POST | `/system/auto-role/esi-title-mappings` | 新增 title 映射 | `RequireRole(admin)` |
| DELETE | `/system/auto-role/esi-title-mappings/:id` | 删除 title 映射 | `RequireRole(admin)` |
| POST | `/system/auto-role/sync` | 手动触发同步 | `RequireRole(admin)` |
| GET | `/system/webhook/config` | 获取 Webhook 配置 | `RequireRole(admin)` |
| PUT | `/system/webhook/config` | 保存 Webhook 配置 | `RequireRole(admin)` |
| POST | `/system/webhook/test` | 测试 Webhook | `RequireRole(admin)` |
