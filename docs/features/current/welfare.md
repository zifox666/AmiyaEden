---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-23
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/welfare.go
  - static/src/api/welfare.ts
  - static/src/views/welfare
---

# 军团福利模块

## 当前能力

- 管理员福利定义 CRUD（创建、编辑、删除、列表）
- 两种发放模式：按自然人（per_user）、按人物（per_character）
- 可选技能计划检查：可关联多个军团技能计划，技能合格才允许申请
- 可选角色最大年龄限制（max_char_age_months）：设置后发放模式锁定为 per_user，拥有任何超龄角色的用户不可申请
- 用户自助申请福利，系统自动判断资格
- 申请状态流转：requested → delivered / rejected
- 福利审批页面：福利官/管理员浏览待发放申请，执行发放或拒绝操作
- 管理员可导入历史已发放记录，按行粘贴角色名和 QQ 号生成 delivered 记录
- 福利存在申请记录时禁止删除

### 申请资格判断

- **按自然人（per_user）**：若已有任何申请记录的 QQ 或 DiscordID 与当前用户匹配，则不可申请
- **按人物（per_character）**：若已有任何申请记录的 character_id 或 character_name 与角色匹配，则该角色不可申请
- **技能计划检查**：
  - per_user + 需要技能计划：至少一个角色满足至少一个关联技能计划即可
  - per_character + 需要技能计划：仅满足技能计划的角色可申请
- **角色年龄限制**（可选，仅 per_user）：用户的任一角色年龄（基于 ESI 生日）超过限制月数则不可申请。角色生日在首次检查时从 ESI 获取并持久化到 eve_character 表

### 申请记录模型（WelfareApplication）

- `welfare_id`、`user_id`、`character_id`、`character_name`、`qq`、`discord_id`
- `status`：requested / delivered / rejected
- `reviewed_by`、`reviewed_at`：审批人和审批时间
- 历史导入记录允许 `user_id` 为空；当前导入格式保存 `character_name` 与 `qq`

## 入口

### 前端页面

- `static/src/views/welfare/my` — 我的福利（所有已登录用户）
  - 申请福利 tab：显示可申请的福利，per_character 每个角色独立一行
  - 已领取福利 tab：显示所有申请记录及状态
- `static/src/views/welfare/approval` — 福利审批（福利官、管理员）
  - 待发放 tab：显示 requested 申请，支持发放/拒绝操作
  - 历史记录 tab：显示已发放/已拒绝的申请记录
- `static/src/views/welfare/settings` — 福利设置（管理员）

### 后端路由

管理员端：
- `POST /api/v1/system/welfare/list`
- `POST /api/v1/system/welfare/add`
- `POST /api/v1/system/welfare/edit`
- `POST /api/v1/system/welfare/delete`
- `POST /api/v1/system/welfare/import` — 导入历史福利记录
- `POST /api/v1/system/welfare/applications` — 福利申请列表（审批端，支持按状态筛选）
- `POST /api/v1/system/welfare/review` — 审批福利申请（发放/拒绝）

用户端：
- `POST /api/v1/welfare/eligible` — 获取可申请的福利列表
- `POST /api/v1/welfare/apply` — 提交福利申请
- `POST /api/v1/welfare/my-applications` — 查询我的申请记录

## 权限边界

- 军团福利导航栏要求 `Login`（guest 不可见）
- 我的福利页面及用户端 `/welfare/*` 接口要求 `Login`
- 福利审批页面要求 `welfare` 或 `admin` 
- 福利设置页面及后端 `/system/welfare/*` 接口要求 `admin`
- `welfare` 角色（福利官）为系统默认角色，优先级 50

## 关键不变量

- 不论按自然人还是按人物发放，都不允许重复申请
- per_user 去重基于 QQ / DiscordID 匹配（非 user_id），要求用户至少设置一个联系方式
- per_character 去重基于 character_id 和 character_name
- 申请时服务端二次校验资格，防止并发竞态
- 福利系统是纯记录型，实际发放在外部完成（游戏内合同等），系统只追踪申请和审批
- 技能计划检查复用 skill_plan 模块，福利定义通过 welfare_skill_plans 关联表支持多技能计划

## 主要代码文件

- `server/internal/model/welfare.go`
- `server/internal/repository/welfare.go`
- `server/internal/service/welfare.go`
- `server/internal/handler/welfare.go`
- `server/internal/router/router.go`
- `static/src/api/welfare.ts`
- `static/src/types/api/api.d.ts` (Welfare namespace)
- `static/src/views/welfare/`
- `static/src/locales/langs/zh.json` (welfareMy, welfareApproval namespaces)
- `static/src/locales/langs/en.json` (welfareMy, welfareApproval namespaces)
