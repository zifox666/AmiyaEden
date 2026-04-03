---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-04-03
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/welfare.go
  - server/internal/service/sys_wallet.go
  - static/src/api/welfare.ts
  - static/src/views/welfare
---

# 军团福利模块

## 当前能力

- 管理员福利定义 CRUD（创建、编辑、删除、列表）
- 福利定义支持可选整数配置 `pay_by_fuxi_coin`，用于审批发放时按当前配置发放伏羲币
- 两种发放模式：按自然人（per_user）、按人物（per_character）
- 可选技能计划检查：可关联多个军团技能计划，技能合格才允许申请
- 可选人物最大年龄限制（max_char_age_months）：可与 per_user / per_character 一起使用。系统会先按用户检查任一人物年龄，若任一人物超龄则该福利对该用户不可申请；若通过，再继续按发放模式筛选人物
- 可选军团舰队 PAP 门槛（minimum_pap）：若设置为大于 0 的值，申请人需拥有军团舰队 PAP 总数大于该值才可申请
- 可选证明图片要求（require_evidence）：管理员可要求申请人上传图片作为证明，并可上传示例图片供参考
- 用户自助申请福利，系统自动判断资格
- 我的福利页会把当前可申请的福利和“未达到条件”的技能门槛福利一起展示；未达到条件项会灰显，只有当人物年龄符合限制时才会出现
- 若福利要求证明图片，申请时弹窗提示上传；弹窗内同步展示管理员上传的示例图片
- 若福利因技能规划未满足而暂不可申请，前端提示会列出对应的技能规划名称
- 我的福利页面顶部提供前往技能规划完成度检查页的提醒链接
- 申请状态流转：requested → delivered / rejected
- 当福利当前配置 `pay_by_fuxi_coin > 0` 时，审批端执行 delivered 会同步给申请人钱包入账，流水 `ref_type = welfare_payout`
- 审批端执行 delivered 后，系统会以发放福利官的主人物为发件人尽力发送一封双语游戏内邮件；若发件人未绑定可用主人物、未授权 `esi-mail.send_mail.v1` 或 ESI 发送失败，不影响发放结果
- 若发放已成功但邮件发送失败，审批界面会继续显示成功提示，并额外弹出一条包含后端错误内容的警告提示
- 我的福利页面"已领取福利" tab 展示审批福利官昵称
- 福利审批页面：福利官/管理员浏览待发放申请，执行发放或拒绝操作；审批列表展示申请人上传的证明图片缩略图，并支持原页预览，行悬停时展示福利描述
- 福利审批页面的人物列提供共享内联复制按钮，便于复制申请人物名
- 管理员可导入历史已发放记录，按行粘贴人物名和 QQ 号生成 delivered 记录
- 福利存在申请记录时禁止删除

### 申请资格判断

- **按自然人（per_user）**：若已有任何申请记录的 QQ 或 DiscordID 与当前用户匹配，则不可申请
- **按人物（per_character）**：若已有任何申请记录的 character_id 或 character_name 与人物匹配，则该人物不可申请
- **技能计划检查**：
  - per_user + 需要技能计划：至少一个人物满足至少一个关联技能计划即可
  - per_character + 需要技能计划：仅满足技能计划的人物可申请
- **前端展示规则**：
  - 人物年龄不符合限制的福利不显示
  - 仅因技能不足暂不可申请的福利会保留在“申请福利”页，并以灰显状态展示
- **人物年龄限制**（可选，适用于 per_user / per_character）：先按用户检查任一人物年龄（基于 ESI 生日），若任一人物超过限制月数则该福利对该用户不可申请；若通过，再继续按发放模式筛选人物。人物生日在首次检查时从 ESI 获取并持久化到 eve_character 表
- **军团舰队 PAP 门槛**（可选，适用于 per_user / per_character）：若设置为大于 0 的数值，则申请人需拥有军团舰队 PAP 总数大于该值；未达到时该福利仍保留在“申请福利”页，但会灰显并禁用申请按钮

### 申请记录模型（WelfareApplication）

- `welfare_id`、`user_id`、`character_id`、`character_name`、`qq`、`discord_id`
- `evidence_image`：申请人上传的证明图片 URL（可选，当福利 require_evidence=true 时必填）
- `status`：requested / delivered / rejected
- `reviewed_by`、`reviewed_at`：审批人和审批时间
- 历史导入记录允许 `user_id` 为空；当前导入格式保存 `character_name` 与 `qq`

## 入口

### 前端页面

- `static/src/views/welfare/my` — 我的福利（所有已登录用户）
  - 申请福利 tab：显示可申请的福利，per_character 每个人物独立一行
  - 已领取福利 tab：分页显示申请记录及状态
- `static/src/views/welfare/approval` — 福利审批（福利官、管理员）
  - 待发放 tab：显示 requested 申请，支持发放/拒绝操作
  - 历史记录 tab：显示已发放/已拒绝的申请记录，并支持按人物名、昵称或 QQ 搜索
- `static/src/views/welfare/settings` — 福利设置（管理员）

### 后端路由

管理员端：

- `POST /api/v1/system/welfare/list`
- `POST /api/v1/system/welfare/add`
- `POST /api/v1/system/welfare/edit`
- `POST /api/v1/system/welfare/delete`
- `POST /api/v1/system/welfare/reorder`
- `POST /api/v1/system/welfare/import` — 导入历史福利记录
- `POST /api/v1/system/welfare/applications` — 福利申请列表（审批端，支持按状态与人物名/昵称/QQ 关键词筛选）
- `POST /api/v1/system/welfare/applications/delete` — 删除单条申请记录（仅 admin）
- `POST /api/v1/system/welfare/review` — 审批福利申请（发放/拒绝）

用户端：

- `POST /api/v1/welfare/eligible` — 获取可申请的福利列表
- `POST /api/v1/welfare/apply` — 提交福利申请
- `POST /api/v1/welfare/my-applications` — 查询我的申请记录
- `POST /api/v1/welfare/upload-evidence` — 上传证明图片（multipart），返回 base64 data URL；最大 2MB，仅支持 jpeg/png/webp；不写入文件系统，直接存库

## 权限边界

- 军团福利导航栏要求 `Login`（guest 不可见）
- 我的福利页面及用户端 `/welfare/*` 接口要求 `Login`
- 福利审批页面要求 `welfare` 或 `admin`；历史记录删除按钮及对应接口仅 `admin`
- 福利设置页面及后端 `/system/welfare/list` 接口要求 `welfare` 或 `admin`
- 福利设置页面写操作（创建、编辑、删除、导入、排序）及对应后端接口要求 `admin`
- `welfare` 职权（福利官）为系统默认职权，优先级 50

## 关键不变量

- 不论按自然人还是按人物发放，都不允许重复申请
- per_user 去重基于 QQ / DiscordID 匹配（非 user_id），要求用户至少设置一个联系方式
- per_character 去重基于 character_id 和 character_name
- 申请时服务端二次校验资格，防止并发竞态
- `pay_by_fuxi_coin` 使用审批当下的福利配置，不在申请记录里冻结快照
- 当 `pay_by_fuxi_coin > 0` 且申请记录包含 `user_id` 时，`requested -> delivered` 会在同一事务内写入一条 `wallet_transaction`，`ref_type = welfare_payout`
- `requested -> delivered` 提交成功后，服务会尽力向申请人主人物发送一封双语发放通知邮件，发件人为执行发放的福利官主人物；邮件失败只记录告警、不回滚发放，并在成功响应里附带 `mail_error` 供前端提示
- 若 ESI 接受了发信请求，成功响应还可能附带邮件调试信息；具体字段以代码契约为准
- 导入历史福利记录只写福利申请历史，不补写 `welfare_payout` 钱包流水
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
