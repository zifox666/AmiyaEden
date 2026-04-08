---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-04-09
source_of_truth:
  - server/bootstrap/db.go
  - server/internal/model
  - server/internal/service/role.go
  - server/internal/service/eve_sso.go
---

# 数据库 Schema

## 目的

本文件描述当前应用自己的数据库结构边界与关键关系。

它回答的是：

- 当前应用把哪些业务数据持久化到 PostgreSQL
- 用户、职权关联、自动权限映射等核心表如何关联
- 哪些列是当前权威关系的一部分，哪些只是兼容字段

它不试图替代代码中的完整字段定义。
精确字段类型、索引与 GORM tag 仍以 `server/internal/model/*` 和 `server/bootstrap/db.go` 为准。

## 真实来源

当前应用 schema 的真实来源是：

1. `server/bootstrap/db.go` 中注册到 `AutoMigrate` 的模型
2. `server/internal/model/*` 中的 GORM 模型定义
3. `server/bootstrap/db.go` 中的启动期 schema 规范化逻辑

`docs/reference/sde-schema.sql` 只是历史 SDE 参考资产，不代表本应用当前 live schema。

## Schema 生成方式

应用启动初始化数据库时，当前会执行：

- `AutoMigrate`
- 必要的 schema 规范化与兼容处理

因此，当前 schema 不是通过单独的 SQL migration 目录维护，而是通过 GORM 模型 + 启动时补偿逻辑维护。

## 核心表分组

当前应用表大致分为这些分组：

- 认证与用户：`user`、`eve_character`
- RBAC：`user_role`
- 自动权限映射：`esi_role_mapping`、`esi_title_mapping`、`eve_character_corp_role`
- ESI 快照：资产、通知、技能、合同、装配、结构、钱包等 `eve_*` / `esi_*` 相关表
- 业务模块：`fleet*`、`srp*`、`shop*`、`skill_plan*`、`welfare*`、`newbro_player_state`、`newbro_captain_affiliation`、`captain_bounty_attribution`、`captain_bounty_sync_state`、`alliance_pap*`、`system_wallet*`、`wallet_log`、`wallet_transaction`
- 基础设施：`operation_log`、`sde_versions`、`sys_config`

## 用户与认证

### `user`

当前产品用户表不承载用户名 / 密码登录模型。

关键列包括：

- `id`
- `nickname`
- `qq`
- `discord_id`
- `avatar`
- `status`
- `role`
- `primary_character_id`
- `last_login_at`
- `last_login_ip`
- `created_at` / `updated_at` / `deleted_at`

说明：

- 当前产品认证入口是 EVE SSO，不是账号密码
- `qq` 与 `discord_id` 是当前资料补全与唯一性校验的一部分
- `primary_character_id` 指向用户当前主人物的 EVE `character_id`
- `role` 仍然保留，但它不是当前 RBAC 的权威来源

### `eve_character`

`eve_character` 表示绑定到平台用户的 EVE 人物。

关键列包括：

- `character_id`
- `character_name`
- `portrait_url`
- `user_id`
- `access_token`
- `refresh_token`
- `token_expiry`
- `scopes`
- `token_invalid`
- `corporation_id`
- `alliance_id`
- `faction_id`

关系上：

- 一个 `user` 可以绑定多个 `eve_character`
- `user.primary_character_id` 记录主人物的 EVE `character_id`

## 职权关联与权限数据

### 当前权威 RBAC 持久化表

当前数据库里与 RBAC 直接相关的持久化表基于：

- `user_role`

这是当前用户职权分配的权威 schema。

当前实现中：

- canonical 系统职权编码定义在代码常量与 `SystemRoleDefinitions` 中，不存储在独立 `role` 表
- 路由级权限边界通过后端中间件 `RequireRole` / `RequireLoginUser` 实现，不依赖数据库权限表
- 前端路由 / 菜单元数据属于代码与前端配置，不属于当前数据库 schema

如需查看 canonical 职权编码与权限边界，见 `docs/architecture/auth-and-permissions.md`。

### `user_role`

`user_role` 是用户和职权的多对多关联表：

- `user_id`
- `role_code`

它是当前用户职权分配的权威来源。

## 兼容字段

### `user.role` 的当前定位

`user.role` 是当前 schema 中保留的兼容字段。

它仍然存在的原因是：

- 兼容旧 JWT / 旧前端消费者
- 在 `user_role` 为空时提供 fallback
- 为仍依赖单职权字段的返回结构提供兼容值

但当前设计上：

- 职权真实分配以 `user_role` 为准
- `user.role` 只是镜像 / fallback / 兼容字段

### 运行时同步行为

用户职权被重新分配时，服务层会把 `user.role` 同步为最高优先级职权，以降低旧消费者漂移风险。

### 文档约束

因此，今后讨论“当前职权 schema”时：

- 应优先说 `user_role`
- 不应把 `user.role` 描述成权威职权来源
- 需要明确标注它是兼容历史单职权模型的保留列

## 自动权限映射相关表

### `esi_role_mapping`

ESI 军团职权到系统职权的映射表。

关键列：

- `esi_role`
- `role_code`

### `esi_title_mapping`

ESI 头衔到系统职权的映射表。

关键列：

- `corporation_id`
- `title_id`
- `title_name`
- `role_code`

### `eve_character_corp_role`

人物当前 ESI 军团职权快照表。

关键列：

- `character_id`
- `corp_role`

说明：

- 这是自动权限同步的输入快照之一
- 当前 `admin` 的内置快捷规则会读取允许军团中的 corp role `Director`
- `eve_character_title` 只用于显式 title mapping，不负责 `Director` 的内置快捷抬升

## 新人帮扶相关表

### `newbro_player_state`

缓存用户当前新人资格快照。

关键列包括：

- `user_id`
- `is_currently_newbro`
- `evaluated_at`
- `rule_version`
- `disqualified_reason`

说明：

- 这是快照，不是永久毕业标记
- 当人物资料、技能快照或规则版本变化时，服务层会重新计算

### `newbro_captain_affiliation`

记录新人用户与队长用户之间的关联历史。

关键列包括：

- `player_user_id`
- `player_primary_character_id_at_start`
- `captain_user_id`
- `started_at`
- `ended_at`

说明：

- 同一时间一个新人只能有一条 active 关联
- 更换队长时旧记录写入 `ended_at`，新关系另起一行

### `captain_bounty_attribution`

记录已持久化的队长赏金归因台账。

关键列包括：

- `affiliation_id`
- `player_user_id`
- `player_character_id`
- `captain_user_id`
- `captain_character_id`
- `captain_wallet_journal_id`
- `wallet_journal_id`
- `ref_type`
- `system_id`
- `journal_at`
- `amount`
- `processed_at`

说明：

- `wallet_journal_id` 当前唯一，避免同一条新人赏金被重复归因
- `captain_character_id` 在写入时冻结保存，后续队长主人物变化不会回改旧账
- 当前只直接扫描并写入玩家侧 `bounty_prizes`
- `processed_at IS NULL` 表示该条归因尚未参与队长奖励处理

### `captain_reward_settlement`

记录队长奖励处理历史；每次处理、每名队长写入一行。

关键列包括：

- `captain_user_id`
- `attribution_count`
- `attributed_isk_total`
- `bonus_rate`
- `credited_value`
- `processed_at`
- `wallet_ref_id`

说明：

- `bonus_rate` 按百分比保存，便于历史回溯时准确重建当次计算口径
- `credited_value` 是最终发放到伏羲币的积分，保留两位小数
- `wallet_ref_id` 对应伏羲币流水引用，当前使用 `newbro_captain_reward` 类型

### `captain_bounty_sync_state`

记录队长赏金归因增量同步进度。

关键列包括：

- `sync_key`
- `last_wallet_journal_id`
- `last_journal_at`

说明：

- 当前默认维护 `default` 同步游标
- 当前归因回填窗口限制在最近 `30` 天

### Director 快捷规则的输入边界

这是一个需要明确保留的 schema 语义：

- `eve_character_corp_role` 承载的是 ESI 返回的真实 corp permission 快照
- `Director -> admin` 的内置快捷规则只读取这里的 `corp_role`
- `esi_title_mapping` 使用的 title 数据是另一条显式配置链路，不能和 corp role 混为一谈

因此，“title 名称叫 Director” 与 “ESI corp role 是 Director” 在当前系统里不是同一件事。

## 福利与伏羲币关联表

### `welfare`

关键列包括：

- `id`
- `dist_mode`
- `pay_by_fuxi_coin`
- `require_skill_plan`
- `max_char_age_months`
- `minimum_pap`
- `require_evidence`
- `example_evidence`
- `status`
- `sort_order`
- `created_by`

说明：

- `pay_by_fuxi_coin` 是福利定义上的可选整数配置，不会在申请时冻结到 `welfare_application`
- 发放时，服务层会重新读取当前 `welfare` 行来决定是否发放伏羲币；若 `0 < pay_by_fuxi_coin < system_config.welfare.auto_approve_fuxi_coin_threshold`，该发放会在申请提交事务内直接发生
- `system_config.welfare.auto_approve_fuxi_coin_threshold` 默认值为 `500`；管理员可在 `/welfare/settings` 调整，设置为 `0` 时关闭自动审批
- 技能计划关联当前通过 `welfare_skill_plans` 关系表维护，不直接内嵌在 `welfare` 表列里

### `welfare_application`

关键列包括：

- `welfare_id`
- `user_id`
- `character_id`
- `character_name`
- `qq`
- `discord_id`
- `evidence_image`
- `status`
- `reviewed_by`
- `reviewed_at`

说明：

- 常规人工审批路径会在事务内重新读取并锁定申请行，再执行 `requested -> delivered / rejected`
- 若当前 `welfare.pay_by_fuxi_coin` 满足 `0 < value < system_config.welfare.auto_approve_fuxi_coin_threshold` 且申请资格校验通过，申请会直接以 `delivered` 状态写入，并同步写入 `reviewed_at`；此时 `reviewed_by = 0`
- 历史导入允许 `user_id` 为空并直接写入 `delivered` 记录

### `wallet_transaction`

与福利模块相关的当前约定：

- 当 `welfare.pay_by_fuxi_coin > 0` 且福利申请被发放为 `delivered` 时，系统会写入一条钱包流水；这既包括人工审批的 `requested -> delivered`，也包括小额伏羲币福利在申请时的自动发放
- 该流水使用 `ref_type = welfare_payout`
- `ref_id` 当前按 `welfare_application:<application_id>` 生成，用于追踪该次福利发放
- 导入历史福利记录不会补写 `welfare_payout` 钱包流水

## 当前未采用的 schema 设计

以下都不是当前产品 schema 的有效方向：

- 用户名 / 密码 / 盐值认证表
- 以单个 `user.role` 作为唯一职权来源
- 把历史 `role` / `menu` / `role_menu` 继续当作当前权威 RBAC schema
- 把历史 `docs/reference/sde-schema.sql` 当作应用业务表定义

仓库里可能仍有旧页面、旧文案或历史兼容逻辑，但它们不应被重新解释成当前数据库设计要求。

## 变更规则

当你修改应用 schema 时，通常需要同步这些层：

1. `server/internal/model/*`
2. `server/bootstrap/db.go`
3. 相关 `repository` / `service`
4. 前端 API 与类型文件
5. 对应 feature / API 文档
6. 本文件

如果变更涉及：

- 职权模型
- 用户资料字段
- 前端路由 / 菜单权限边界
- 自动权限映射表

应优先检查 `docs/architecture/auth-and-permissions.md` 与 `docs/features/current/administration.md` 是否也需要更新。
