---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-04-03
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
- 管理员可维护用户昵称与状态
- 用户 QQ / Discord ID 在管理端只读展示，仍由用户本人通过 `/api/v1/me` 维护
- 用户管理列表默认按最后登录时间倒序，并支持按昵称、QQ、任意已绑定人物名搜索
- 用户管理列表可展开每个用户行，查看该用户全部已绑定人物的头像、人物 ID、人物名、ESI 状态与人物总技能点
- 用户管理展开人物列表在人物名右侧提供共享内联复制按钮，便于复制已绑定人物名
- `super_admin` 可在 `/system/user` 切换“已失效人物 ESI 限制”，决定是否强制用户在任一已绑定人物 ESI 失效时停留在人物管理页
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
- `/api/v1/system/sde-config`
- `/api/v1/system/basic-config/allow-corporations`
- `/api/v1/system/basic-config/character-esi-restriction`
- `/api/v1/system/role/definitions`
- `/api/v1/system/user/*`
- `/api/v1/system/auto-role/*`
- `/api/v1/system/pap/*`
- `/api/v1/system/pap-exchange/*`
- `/api/v1/system/webhook/*`

## 权限边界

- `/system/*` 默认要求 `admin`
- `/system/basic-config` 页面及 `/api/v1/system/basic-config*`、`/api/v1/system/sde-config` 接口仅 `super_admin` 可见且可用
- `/system/user/:id/impersonate` 额外要求 `super_admin`
- `/system/user/:id/impersonate` 在目标用户主人物 ESI 已失效时拒绝签发模拟登录 token
- `/system/user/:id` 删除用户时，若目标用户已登记 QQ 或 Discord ID，则仅 `super_admin` 可执行
- `GET /system/role/definitions` 仅用于前端加载系统职权定义，属于只读数据源
- `GET /system/basic-config` 仅返回固定系统标识，不提供写接口

## 关键不变量

For role assignment rules, super_admin protection, and Director auto-role rules, see `docs/architecture/auth-and-permissions.md`. Below are invariants specific to the administration module:

- 系统军团 ID 与网站标题由代码常量提供，当前不通过数据库、API 或 UI 修改
- 基础配置页不再允许编辑军团 ID 或网站标题
- 管理员侧用户列表 `/api/v1/system/user` 的职权列只以有序 `roles[]` 为准，不再暴露历史单值 `role`
- 管理员侧用户列表同时返回该用户全部已绑定人物及每个人物的 `total_sp` 快照与 `token_invalid` 状态
- 管理员侧用户列表支持单职权筛选；职权匹配只以当前 `user_role` 关联为准
- `allow_corporations` 配置存储在数据库 `system_config` 表（键名 `app.allow_corporations`），运行时总会强制包含伏羲军团 Fuxi Legion（`98185110`）
- 失效人物 ESI 页面停留限制存储在数据库 `system_config` 表（键名 `auth.enforce_character_esi_restriction`），默认开启
- 联盟 PAP 的管理接口与用户查看接口分属不同模块
- Webhook 是系统配置能力，不应散落到页面里直接拼接

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
