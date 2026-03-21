---
status: active
doc_type: architecture
owner: engineering
last_reviewed: 2026-03-22
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
- 用户、角色、菜单、角色绑定等核心表如何关联
- 哪些列是当前设计的一部分，哪些只是兼容历史实现

它不试图替代代码中的完整字段定义。
精确字段类型、索引与 GORM tag 仍以 `server/internal/model/*` 和 `server/bootstrap/db.go` 为准。

## 真实来源

当前应用 schema 的真实来源是：

1. `server/bootstrap/db.go` 中注册到 `AutoMigrate` 的模型
2. `server/internal/model/*` 中的 GORM 模型定义
3. `dropObsoleteSchema()` 中显式清理的历史列 / 历史表

`docs/reference/sde-schema.sql` 只是历史 SDE 参考资产，不代表本应用当前 live schema。

## Schema 生成方式

应用启动初始化数据库时，当前会执行：

- `AutoMigrate`
- 历史遗留列 / 表清理
- 系统角色种子初始化
- 系统菜单种子初始化
- 历史 `user.role` 到 `user_role` 的迁移

因此，当前 schema 不是通过单独的 SQL migration 目录维护，而是通过 GORM 模型 + 启动时补偿逻辑维护。

## 核心表分组

当前应用表大致分为这些分组：

- 认证与用户：`user`、`eve_character`
- RBAC：`role`、`menu`、`role_menu`、`user_role`
- 自动权限映射：`esi_role_mapping`、`esi_title_mapping`、`eve_character_corp_role`
- ESI 快照：资产、通知、技能、合同、装配、结构、钱包等 `eve_*` / `esi_*` 相关表
- 业务模块：`fleet*`、`srp*`、`shop*`、`skill_plan*`、`alliance_pap*`、`system_wallet*`
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
- `primary_character_id` 指向用户当前主角色的 EVE `character_id`
- `role` 仍然保留，但它不是当前 RBAC 的权威来源

### `eve_character`

`eve_character` 表示绑定到平台用户的 EVE 角色。

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
- `user.primary_character_id` 记录主角色的 EVE `character_id`

## 角色、菜单与权限

### 当前权威 RBAC 表

当前角色与菜单权限模型基于：

- `role`
- `menu`
- `role_menu`
- `user_role`

这是当前实现的权威权限模型。

### `role`

角色表承载系统角色和自定义角色。

关键列包括：

- `id`
- `code`
- `name`
- `description`
- `is_system`
- `sort`
- `status`

当前 canonical 角色编码见代码常量：

- `super_admin`
- `admin`
- `srp`
- `fc`
- `user`
- `guest`

### `menu`

菜单表同时承载目录、页面与按钮节点。

关键列包括：

- `parent_id`
- `type`
- `name`
- `path`
- `component`
- `permission`
- `title`
- `icon`
- `sort`
- `is_hide`
- `keep_alive`
- `is_hide_tab`
- `fixed_tab`
- `status`

`type` 当前支持：

- `dir`
- `menu`
- `button`

### `role_menu`

`role_menu` 是角色和菜单的多对多关联表：

- `role_id`
- `menu_id`

它决定某个角色可见哪些菜单、拥有哪些按钮权限。

### `user_role`

`user_role` 是用户和角色的多对多关联表：

- `user_id`
- `role_id`

它是当前用户角色分配的权威来源。

## 兼容历史设计的列与迁移

### `user.role` 的当前定位

`user.role` 是当前 schema 中最重要的历史兼容列。

它仍然存在的原因是：

- 兼容旧 JWT / 旧前端消费者
- 在 `user_role` 为空时提供 fallback
- 为仍依赖单角色字段的返回结构提供兼容值

但当前设计上：

- 角色真实分配以 `user_role` 为准
- `user.role` 只是镜像 / fallback / 兼容字段

### 启动时的兼容行为

当前启动逻辑会把历史 `user.role` 数据迁移到 `user_role`。

同时，在用户角色被重新分配时，服务层会把 `user.role` 同步为最高优先级角色，以降低旧消费者漂移风险。

### 文档约束

因此，今后讨论“当前角色 schema”时：

- 应优先说 `user_role`
- 不应把 `user.role` 描述成权威角色来源
- 需要明确标注它是兼容历史单角色模型的保留列

## 自动权限映射相关表

### `esi_role_mapping`

ESI 军团角色到系统角色的映射表。

关键列：

- `esi_role`
- `role_id`

### `esi_title_mapping`

ESI 头衔到系统角色的映射表。

关键列：

- `corporation_id`
- `title_id`
- `title_name`
- `role_id`

### `eve_character_corp_role`

角色当前 ESI 军团角色快照表。

关键列：

- `character_id`
- `corp_role`

说明：

- 这是自动权限同步的输入快照之一
- 当前 `admin` 的内置快捷规则会读取允许军团中的 corp role `Director`
- `eve_character_title` 只用于显式 title mapping，不负责 `Director` 的内置快捷抬升

### Director 快捷规则的输入边界

这是一个需要明确保留的 schema 语义：

- `eve_character_corp_role` 承载的是 ESI 返回的真实 corp permission 快照
- `Director -> admin` 的内置快捷规则只读取这里的 `corp_role`
- `esi_title_mapping` 使用的 title 数据是另一条显式配置链路，不能和 corp role 混为一谈

因此，“title 名称叫 Director” 与 “ESI corp role 是 Director” 在当前系统里不是同一件事。

## 当前未采用的 schema 设计

以下都不是当前产品 schema 的有效方向：

- 用户名 / 密码 / 盐值认证表
- 以单个 `user.role` 作为唯一角色来源
- 独立维护一套与 `menu` / `role_menu` 无关的前端权限表
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

- 角色模型
- 用户资料字段
- 菜单 / 权限关联
- 自动权限映射表

应优先检查 `docs/architecture/auth-and-permissions.md` 与 `docs/features/current/administration.md` 是否也需要更新。
