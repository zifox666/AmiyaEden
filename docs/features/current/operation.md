---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-28
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/fleet.go
  - server/internal/service/fleet_config.go
  - server/internal/service/auto_srp.go
  - server/internal/service/pap_exchange.go
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
- 舰队配置管理、EFT 导出、从人物装配导入、导出到 ESI
- 舰队配置装配物品的独立装备设置：重要性、惩罚、替代品
- 舰队级自动 SRP 模式：`disabled` / `submit_only` / `auto_approve`
- 创建舰队时，选择舰队配置会默认把自动 SRP 模式设为 `auto_approve`

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

- `fleets`、`fleet-detail` 页面访问要求 `super_admin`、`admin`、`fc` 或 `senior_fc`
- `fleet-configs` 页面访问要求 `Login`
- 舰队管理动作，包括刷新 ESI、成员同步 / 手动维护、PAP 发放、邀请链接与 Ping，要求 `super_admin`、`admin`、`fc` 或 `senior_fc`；其中 `fc` 职权仅限操作自己创建的舰队（`fc_user_id == current_user_id`），`senior_fc`、`admin`、`super_admin` 不受此限制
- 舰队删除仍要求 `super_admin` 或 `admin`
- `fleet-configs` 的只读查询要求 `Login`
- `fleet-configs` 的导出到 ESI（保存到自己的游戏装配）要求 `Login`
- `fleet-configs` 的创建、修改、删除、导入装配和物品设置要求 `super_admin`、`admin` 或 `senior_fc`
- `corporation-pap`、`pap`、`join` 按 `Login` 处理
- 舰队相关职权边界由前端路由元数据决定

## 关键不变量

- 舰队、PAP、舰队配置共享同一业务切片，修改时要一起考虑
- 自动 SRP 不是纯草案，当前模型、页面和后台处理逻辑都已存在
- 自动 SRP 的触发与舰队成员、KM 刷新、舰队配置装配有关，不能只改 UI 字段
- 保存到游戏是把现有舰队配置装配导出到当前用户自己的 ESI 人物，不是系统配置写操作
- 装配物品的装备设置通过独立「装备设置」对话框单独保存，不随主配置表单同一次接口提交
- 舰队配置编辑会在同一装配条目内按 `flag + type_id + quantity` 保留未变化物品的装备设置（重要性、惩罚、替代品）；删除、替换、改槽位或改数量后，该物品按新条目处理，并重置为默认设置
- 联盟 PAP 的用户侧展示在 Operation，管理员配置与导入在 System
- 军团 PAP 页面属于多块统计 + 表格混排的分析页，当前明确允许不走 `useTable` / `ArtTable` 默认模板
- 发放 PAP 时的伏羲币换算不再是固定 1:1，而是按舰队 `importance`（`cta` / `strat_op` / `other`）查询 `pap_type_rate` 表中对应汇率；若成员是舰队 FC，则优先发放固定 `FC工资`，并受 `FC工资上限次数` 约束，汇率配置入口在「系统管理 → PAP兑换」，详见 `docs/features/current/pap-exchange.md`
- 联盟 PAP 月度归档为纯归档操作，当前不进行钱包兑换（该能力预留为未来特性）

## 主要代码文件

- `server/internal/service/fleet.go`
- `server/internal/service/fleet_config.go`
- `server/internal/service/auto_srp.go`
- `server/internal/router/router.go`
- `static/src/api/fleet.ts`
- `static/src/api/fleet-config.ts`
- `static/src/views/operation`
