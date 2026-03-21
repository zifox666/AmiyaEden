---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-22
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/role.go
  - server/internal/service/menu.go
  - server/internal/service/user.go
  - server/internal/service/auto_role.go
  - server/internal/service/alliance_pap.go
  - server/internal/service/sys_webhook.go
  - static/src/api/system-manage.ts
  - static/src/api/webhook.ts
  - static/src/views/system
---

# System 管理模块

## 当前能力

- 基础配置读取与更新
- 菜单管理
- 角色管理与角色菜单分配
- 用户管理、用户角色分配
- 管理员可维护用户昵称、QQ、Discord ID 与状态
- 超级管理员模拟登录
- 自动权限映射（ESI corp roles / corp titles -> system roles）
- 联盟 PAP 列表、抓取、导入、兑换配置、月度结算
- Webhook 配置与测试

## 入口

### 前端页面

- `static/src/views/system/basic-config`
- `static/src/views/system/menu`
- `static/src/views/system/role`
- `static/src/views/system/user`
- `static/src/views/system/auto-role`
- `static/src/views/system/pap`
- `static/src/views/system/webhook`

### 后端路由

- `/api/v1/system/basic-config`
- `/api/v1/system/menu/*`
- `/api/v1/system/role/*`
- `/api/v1/system/user/*`
- `/api/v1/system/auto-role/*`
- `/api/v1/system/pap/*`
- `/api/v1/system/webhook/*`

## 权限边界

- `/system/*` 默认要求 `admin`
- `/system/user/:id/impersonate` 额外要求 `super_admin`

## 关键不变量

- 角色与菜单定义的 canonical 源在代码常量和菜单种子，不在旧文档
- 管理员侧用户资料维护走 `/api/v1/system/user/:id`，当前支持昵称、QQ、Discord ID、状态
- 管理员侧用户列表 `/api/v1/system/user` 的角色列只以有序 `roles[]` 为准，不再暴露历史单值 `role`
- `/api/v1/system/user/:id` 更新与删除都受后端保护：`admin` 不可编辑或删除 `super_admin` 或其他 `admin`
- `/api/v1/system/user/:id/roles` 仅 `super_admin` 可编辑管理员账号或分配 `admin/super_admin`
- 自动权限映射已经是当前功能，不是纯想法
- 主角色在 `allow_corporations` 中时会自动补 `user`
- 任一 `allow_corporations` 角色拥有 ESI corp role `Director` 时会自动补 `admin`
- corp title 仍可通过 title mapping 表显式映射，但不会因为标题名恰好叫 `Director` 而触发内置快捷规则
- 联盟 PAP 的管理接口与用户查看接口分属不同模块
- Webhook 是系统配置能力，不应散落到页面里直接拼接

## 重要 Caveat

### Director 自动提升规则

自动权限映射里的内置 `Director -> admin` 快捷规则，只认 ESI corporation role `Director`。

这意味着：

- 角色必须在允许的 `allow_corporations` 军团内
- 必须命中 `eve_character_corp_role` 快照中的真实 `corp_role = Director`
- corp title 名称即使显示为 `Director`，也不会触发这条内置快捷规则
- 如果确实想让某个 title 触发系统角色，必须通过 `esi_title_mapping` 显式配置

排查自动权限问题时，先看：

1. `allow_corporations` 是否包含该角色军团
2. 该角色的 `eve_character_corp_role` 快照里是否真的存在 `Director`
3. 是否只是 title 名称叫 `Director`，但并没有真实 corp role

## 主要代码文件

- `server/internal/service/role.go`
- `server/internal/service/menu.go`
- `server/internal/service/user.go`
- `server/internal/service/auto_role.go`
- `server/internal/service/alliance_pap.go`
- `server/internal/service/sys_webhook.go`
- `static/src/api/system-manage.ts`
- `static/src/api/webhook.ts`
- `static/src/views/system`
