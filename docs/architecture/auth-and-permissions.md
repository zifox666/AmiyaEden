---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-04-02
source_of_truth:
  - server/internal/router/router.go
  - server/internal/middleware/auth.go
  - server/internal/model/role.go
  - static/src/store/modules/user.ts
---

# 认证与权限

## 登录架构

当前受支持的产品登录路径是 EVE SSO：

- 公共入口：`GET /api/v1/sso/eve/login`
- 回调入口：`GET /api/v1/sso/eve/callback`
- 前端页面：`/auth/login`、`/auth/callback`

登录成功后，前端通过 `GET /api/v1/me` 获取：

- 当前用户
- 绑定人物
- 职权列表
- 按钮权限列表
- 当前新人资格快照 `is_currently_newbro`
- 当前导师学员资格快照 `is_mentor_mentee_eligible`

`guest` 职权当前仍是已认证用户，但不是 `RequireLoginUser` 意义上的产品用户。
因此需要把 guest onboarding / self-service 能力单独挂在仅需 `JWTAuth()` 的路由上，而不是 `RequireLoginUser()`。
当用户已拥有任一非 `guest` 职权时，不应再同时保留 `guest`；`guest` 只作为无更高产品职权时的 fallback / onboarding 状态。
首次 SSO 登录时，如果主人物所属军团命中 `allow_corporations`，系统应直接落为 `user`；未命中时才落为 `guest`。
自动权限同步时，如果账号当前仍是纯 `guest`（或尚无有效职权）且任一绑定人物命中 `allow_corporations`，也应补为至少 `user`；已拥有 `admin`、`fc` 等非 `guest` 职权的账号不应因这条基线规则被改写。

文档上应把这两类边界区分开：

- `JWT`：任意持有有效 JWT 的已认证用户，包含 `guest`
- `Login`：任意已认证且非 `guest` 的产品用户

## 绑定人物能力

登录后可继续管理绑定人物：

- `GET /api/v1/sso/eve/characters`
- `GET /api/v1/sso/eve/bind`
- `PUT /api/v1/sso/eve/primary/:character_id`
- `DELETE /api/v1/sso/eve/characters/:character_id`

这些接口与 `/api/v1/me` 一样，当前都属于 guest 可用的自助能力，权限边界应记为 `JWT`，不是 `Login`。

## 当前系统职权

`server/internal/model/role.go` 定义的 canonical 职权编码（按优先级降序）：

| 编码 | 名称 | Sort |
|---|---|---|
| `super_admin` | 超级管理员 | 100 |
| `admin` | 管理员 | 90 |
| `senior_fc` | 高级FC | 85 |
| `fc` | FC | 70 |
| `srp` | SRP 官员 | 60 |
| `welfare` | 福利官 | 50 |
| `captain` | 新人队长 | 30 |
| `mentor` | 导师 | 25 |
| `user` | 认证用户 | 10 |
| `guest` | 访客 | 0 |

不要再使用旧文档里的 `Administrator` 之类别名。

## 职权分配权限矩阵

职权分配接口 `PUT /api/v1/system/user/:id/roles` 位于 `admin` 路由组下，仅 `super_admin` 和 `admin` 可访问。

分配规则（`server/internal/service/role.go` → `SetUserRoles` + `validateSetUserRolesPermission`）：

- `super_admin` 可为任何用户（包括自己）分配除 `super_admin` 以外的任意职权；请求中若包含 `super_admin` 会被后端静默剥离，目标用户已有的 `super_admin` 职权自动保留不被覆盖
- `admin` 可管理自己的职权（包括移除自身 admin 职权），可为其他用户分配除 `admin` 以外的任意职权；`admin` 不可为非 admin 用户新增 `admin` 职权，但可为已有 admin 身份的用户保留/调整非 admin 职权
- 非 admin 用户（包括 `senior_fc`、`fc`、`srp`、`welfare`、`captain`、`mentor`、`user`、`guest`）无权分配任何职权
- `super_admin` 职权不可通过 API 分配或撤销，仅通过配置文件管理；`super_admin` 操作者提交的职权列表中的 `super_admin` 会被静默剥离而非报错
- 非 `super_admin` 不可修改已有 `super_admin` 职权用户的任何职权

### 矩阵（操作者 → 目标职权）

| 操作者 \ 可分配目标职权 | super_admin | admin | senior_fc | fc | srp | welfare | captain | mentor | user | guest |
|---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **super_admin**（操作他人） | ✗ 仅配置文件 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **super_admin**（操作自己） | 自动保留 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **admin** | ✗ | ⚠️ 仅已有admin可保留 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **其他所有职权** | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ | ✗ |

### 用户管理权限

`PUT/DELETE /api/v1/system/user/:id`（`server/internal/service/user.go` → `validateManageUserPermission`）：

| 操作者 \ 目标用户 | super_admin | admin | 其他职权 |
|---|:---:|:---:|:---:|
| **super_admin** | ✅ | ✅ | ✅ |
| **admin** | ✗ | ✗ | ✅ |
| **其他所有职权** | ✗ | ✗ | ✗ |

补充规则：

- `PUT /api/v1/system/user/:id` 仅允许维护昵称与状态；`QQ` / `Discord ID` 只能由用户本人通过 `/api/v1/me` 维护
- 删除已登记 `QQ` 或 `Discord ID` 的用户属于更高敏感度操作；`DELETE /api/v1/system/user/:id` 仅 `super_admin` 可执行

## JWT 中间件行为

`JWTAuth()` 当前会：

- 从 `Authorization: Bearer <token>` 或 `?token=` 提取 JWT
- 解析 `userID`、`characterID`、兼容字段 `userRole`
- 从职权服务加载 `roles`
- 从职权服务加载 `permissions`
- 写入 Gin Context

上下文键包括：

- `userID`
- `characterID`
- `userRole`
- `roles`
- `permissions`

## 权限检查

### RequireRole

- 判断用户是否拥有指定职权之一
- `super_admin` 自动通过
- 当前不是数值等级继承模型

### RequireLoginUser

- 判断请求方是否至少拥有一个非 `guest` 职权
- 用于实现 API 文档中的 `Login` 边界
- 适合"任意产品用户可访问"的能力，不再用 `RequireRole(..., user)` 代替
- 当前 `skill-planning` 模块的列表 / 详情与完成度检查接口都属于这条边界
- 不适用于 SSO 首次登录后的 guest onboarding 页面，例如 `/me`、`/sso/eve/characters` 以及 guest 可访问的自助信息页

### JWT-only 自助能力

当前这类路由的共同点是：用户已经完成 SSO 并拿到平台 JWT，但还可能停留在 `guest`。

典型例子：

- `/api/v1/me`
- `/api/v1/sso/eve/characters`
- `/api/v1/sso/eve/bind`
- `/api/v1/sso/eve/primary/:character_id`
- `/api/v1/sso/eve/characters/:character_id`

这类接口主要用于：

- 完成人物绑定与主人物调整
- 让 guest 在准入完成前仍能查看自己的基础信息或自助完成资料

`/api/v1/info/*` 当前不再属于 JWT-only 自助能力，而是 `RequireLoginUser()` 边界。

### RequirePermission

- 判断用户是否拥有指定权限之一
- `super_admin` 自动通过
- 支持父权限前缀命中，例如持有 `srp` 时可满足 `srp:review`

## 前端路由模式

前端使用静态路由 + `meta.login` / `meta.roles` 模式。

静态路由模式下的约定：

- `meta.login = true` 表示任意非 `guest` 已登录产品用户可访问
- `meta.roles` 只用于真实的显式职权白名单
- `meta.requiresNewbro = true` 表示该页面还要求当前用户的新人资格快照为 true
- `meta.requiresMentorMenteeEligibility = true` 表示该页面还要求当前用户的导师学员资格快照为 true
- 不要用 `meta.roles: ['admin', 'fc', 'user']` 之类写法冒充 `Login`
- `skill-planning/skill-plans` 当前使用 `meta.login = true` 提供只读访问，创建 / 编辑 / 删除 / 排序仍依赖页面内 `canManage` 与后端 `RequireRole(admin, senior_fc)` 双层限制

修改权限时，必须同时考虑：

- 后端路由保护
- 前端路由元数据
- 按钮权限使用点

`新人帮扶 / 导师系统` 模块还有几条额外边界：

- `新人选队长` 不是纯职权权限，而是 `Login + 当前新人资格`
- `新人选导师` 不是纯职权权限，而是 `Login + 当前导师学员资格`
- `新人选导师` 页面入口依赖 `/api/v1/me` 返回的 `is_mentor_mentee_eligible` 做路由与菜单过滤；页面加载后仍会读取 `/api/v1/mentor/me` 做二次 UX 校验，后端服务层也会再次校验
- `队长帮扶` 需要真实系统职权 `captain`；普通 `admin` 应使用 `帮扶管理` 页面，而不是把 `admin` 当作 captain 的别名
- `导师帮扶` 需要真实系统职权 `mentor`；普通 `admin` 不能访问导师 dashboard 接口
- `导师管理` 与 `导师奖励阶段` 仍属于 `admin` 管理面，不是 mentor 的延伸权限

## Super Admin 配置驱动机制

`super_admin` 职权完全由配置文件驱动，不通过任何 API 或 UI 管理：

- 配置位置：`config.yaml` 的 `app.super_admins`，值为 EVE character ID 列表
- 授予时机：首次 SSO 登录创建用户时，若主人物 ID 在配置列表中则直接授予
- 同步时机：每次 SSO 登录时，`SyncConfigSuperAdmins` 检查用户所有绑定人物 ID，任一命中配置则授予，全部未命中则移除
- API 拦截：`SetUserRoles` 对 `super_admin` 职权做静默剥离处理；非 `super_admin` 操作者提交包含 `super_admin` 的请求会被拒绝；`super_admin` 操作者的请求中 `super_admin` 被静默剥离，目标用户已有的 `super_admin` 职权自动保留
- 删除保护：`DeleteUser` 拒绝删除拥有 `super_admin` 职权的用户，且已登记联系方式的用户仅 `super_admin` 可删除
- 前端禁用：职权分配对话框中 `super_admin` 复选框始终 disabled
- ESI 自动映射：自动权限映射逻辑已排除 `super_admin`，不会被 ESI corp role / title 触发

## 当前不变量

The following are not restated from the body above — they are additional constraints or easily missed boundaries:

- 管理员用户列表 `/api/v1/system/user` 的职权展示与接口契约当前只认 `roles[]`，不应再依赖历史单值 `role`
- `/system/basic-config` 页面及 `/api/v1/system/basic-config*`、`/api/v1/system/sde-config` 接口是 `system` 模块中的显式例外，只允许 `super_admin`
- 细粒度权限不能只靠前端控制
- 职权编码以代码常量为准，不以文档中文称呼为准

For super_admin rules, see "Super Admin 配置驱动机制" above. For role assignment rules, see "职权分配权限矩阵" above. For Director auto-role rules, see below.

## 重要 Caveat: Auto-role Director Signal

`Director -> admin` 内置快捷规则使用 ESI corporation role 信号（`eve_character_corp_role`），不是 title 文本匹配。

- 当人物不在 `allow_corporations` 中时，`eve_character_corp_role` 快照不保留
- 只有伏羲军团 Fuxi Legion（`98185110`）的 `Director` corp role 触发此规则
- corp title 名称为 `Director` 不触发；title 映射通过 `esi_title_mapping` 显式配置
