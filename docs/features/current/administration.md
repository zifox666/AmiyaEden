---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-29
source_of_truth:
  - server/internal/model/system_identity.go
  - server/internal/router/router.go
  - server/internal/service/role.go
  - server/internal/service/user.go
  - server/internal/service/auto_role.go
  - server/internal/service/alliance_pap.go
  - server/internal/service/sys_webhook.go
  - server/internal/handler/sys_config.go
  - server/internal/utils/allow_corporations.go
  - static/src/constants/system-identity.ts
  - static/src/api/sys-config.ts
  - static/src/api/system-manage.ts
  - static/src/api/webhook.ts
  - static/src/views/system
---

# System 管理模块

## 当前能力

- 固定系统标识读取
- 系统职权定义只读查询
- 用户管理、用户职权分配
- 管理员可维护用户昵称、QQ、Discord ID 与状态
- 用户管理列表默认按最后登录时间倒序，并支持按昵称、QQ、任意已绑定人物名搜索
- 用户管理列表可展开每个用户行，查看该用户全部已绑定人物的头像、人物 ID、人物名与人物总技能点
- 超级管理员模拟登录
- 自动权限映射（ESI corp roles / corp titles -> system roles）
- 联盟 PAP 列表、抓取、导入、月度归档（钱包兑换暂不启用）
- 军团 PAP 兑换汇率配置（Skirmish / Strategic / CTA 三种类型，外加 FC 工资与每月工资上限，影响舰队 PAP 发放时的钱包换算）
- Webhook 配置与测试

## 入口

### 前端页面

- `static/src/views/system/basic-config`
- `static/src/views/system/user`
- `static/src/views/system/auto-role`
- `static/src/views/system/pap`
- `static/src/views/system/pap-exchange`
- `static/src/views/system/webhook`

### 后端路由

- `/api/v1/system/basic-config`
- `/api/v1/system/basic-config/allow-corporations`
- `/api/v1/system/role/definitions`
- `/api/v1/system/user/*`
- `/api/v1/system/auto-role/*`
- `/api/v1/system/pap/*`
- `/api/v1/system/pap-exchange/*`
- `/api/v1/system/webhook/*`

## 权限边界

- `/system/*` 默认要求 `admin`
- `/system/user/:id/impersonate` 额外要求 `super_admin`
- `GET /system/role/definitions` 仅用于前端加载系统职权定义，属于只读数据源
- `GET /system/basic-config` 仅返回固定系统标识，不提供写接口

## 关键不变量

- 职权定义的 canonical 源在代码常量，不在旧文档
- 系统军团 ID 与网站标题由代码常量提供，当前不通过数据库、API 或 UI 修改
- 管理员侧用户资料维护走 `/api/v1/system/user/:id`，当前支持昵称、QQ、Discord ID、状态
- 管理员侧用户列表 `/api/v1/system/user` 的职权列只以有序 `roles[]` 为准，不再暴露历史单值 `role`
- 管理员侧用户列表 `/api/v1/system/user` 同时返回该用户全部已绑定人物及每个人物的 `total_sp` 快照，供前端展开行展示
- 管理员侧用户列表 `/api/v1/system/user` 支持单职权筛选；职权匹配只以当前 `user_role` 关联为准，不读取历史单值 `role`
- `/api/v1/system/user/:id` 更新与删除都受后端保护：`admin` 不可编辑或删除其他 `admin`
- `/api/v1/system/user/:id/roles` 职权分配规则：`super_admin` 可为任何用户（包括自己）分配除 `super_admin` 以外的任意职权，请求中包含的 `super_admin` 被静默剥离，目标用户已有的 `super_admin` 职权自动保留；`admin` 可管理自己的职权（包括移除自身 admin 职权），可为其他用户分配除 `admin` 以外的任意职权，但不可为非 admin 用户新增 `admin` 职权；非 admin 用户无权分配任何职权
- `super_admin` 职权不可通过 API 分配或撤销，仅通过配置文件管理；`super_admin` 操作者提交的职权列表中的 `super_admin` 会被静默剥离而非报错；非 `super_admin` 不可修改已有 `super_admin` 用户的任何职权
- `super_admin` 职权由 `config.yaml` 的 `app.super_admins` 配置驱动，每次 SSO 登录时自动同步，不可通过任何 API 或 UI 授予、修改或撤销
- `super_admin` 用户不可通过 API 删除
- 自动权限映射已经是当前功能，不是纯想法
- 当账号当前仅为 `guest`（或尚无有效职权）且任一绑定人物在 `allow_corporations` 中时，自动权限同步会补 `user`
- 只有伏羲军团 Fuxi Legion（`98185110`）人物拥有 ESI corp role `Director` 时，才会自动补 `admin`
- 非 `allow_corporations` 军团人物的 ESI corp role 信号不会参与权限判定或相关刷新任务
- `allow_corporations` 配置存储在数据库 `system_config` 表（键名 `app.allow_corporations`），通过基础配置页面管理
- 运行时 `allow_corporations` 总会强制包含伏羲军团 Fuxi Legion（`98185110`），管理员无法通过 API 或 UI 将其移除
- 基础配置页不再允许编辑军团 ID 或网站标题
- corp title 仍可通过 title mapping 表显式映射，但不会因为标题名恰好叫 `Director` 而触发内置快捷规则
- 联盟 PAP 的管理接口与用户查看接口分属不同模块
- Webhook 是系统配置能力，不应散落到页面里直接拼接

## 重要 Caveat

### Director 自动提升规则

自动权限映射里的内置 `Director -> admin` 快捷规则，只认 ESI corporation role `Director`。

这意味着：

- 人物必须属于伏羲军团 Fuxi Legion（`98185110`）
- 必须命中 `eve_character_corp_role` 快照中的真实 `corp_role = Director`
- 当人物不在允许军团时，其 `eve_character_corp_role` 快照会被清空，不再供后续逻辑使用
- corp title 名称即使显示为 `Director`，也不会触发这条内置快捷规则
- 如果确实想让某个 title 触发系统职权，必须通过 `esi_title_mapping` 显式配置

排查自动权限问题时，先看：

1. 该人物是否属于伏羲军团 Fuxi Legion（`98185110`）
2. 该人物的 `eve_character_corp_role` 快照里是否真的存在 `Director`
3. 是否只是 title 名称叫 `Director`，但并没有真实 corp role

## 主要代码文件

- `server/internal/service/role.go`
- `server/internal/service/user.go`
- `server/internal/service/auto_role.go`
- `server/internal/service/alliance_pap.go`
- `server/internal/service/pap_exchange.go`
- `server/internal/service/sys_webhook.go`
- `static/src/api/system-manage.ts`
- `static/src/api/pap-exchange.ts`
- `static/src/api/webhook.ts`
- `static/src/views/system`
