---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-27
source_of_truth:
  - server/internal/service/npc_kill.go
  - server/internal/repository/npc_kill.go
  - server/internal/handler/npc_kill.go
  - static/src/api/npc-kill.ts
  - static/src/views/info/npc-kills
  - static/src/views/system/npc-kills
---

# NPC 刷怪报表

## 功能概述

NPC 刷怪报表展示人物通过 PvE 活动获得的 NPC 来源收入，涵盖以下来源：

| 收入来源 | wallet journal ref_type | 说明 |
| --- | --- | --- |
| 标准刷怪悬赏 | `bounty_prizes` | 星域/星系 NPC 悬赏奖励，含血族入侵和深渊 Pochven 内容 |
| ESS 分账 | `ess_escrow_transfer` | 零空 ESS（紧急安全容器）到期结算转账 |

Sansha 血族入侵和 Pochven 三角洲空间的收入均以 `bounty_prizes` 形式记录，通过流水 `reason` 字段中的 NPC ID 可区分活动类型，体现在「按 NPC 统计」分组中。

## 当前能力

- 个人单人物 NPC 刷怪报表
- 个人名下所有人物汇总报表
- 公司全员刷怪报表（管理员）
- 总览数据：总悬赏、ESS、税金、实际收入、记录数、估算有效时长
- 按 NPC 分类统计（悬赏流水 reason 字段解析）
- 按星系分类统计
- 按天趋势统计
- 分页流水明细

## 入口

### 前端页面

- `static/src/views/info/npc-kills` — 个人报表（单人物 / 全人物切换）
- `static/src/views/system/npc-kills` — 管理员公司报表

### 后端路由

- `POST /api/v1/info/npc-kills` — 个人单人物报表
- `POST /api/v1/info/npc-kills/all` — 个人全人物汇总
- `POST /api/v1/system/npc-kills` — 公司全员报表（admin）

## 权限边界

- 个人接口要求 `Login`
- 公司接口属于 `/system` 路由，要求 `admin` 或 `super_admin`
- 个人接口在服务层校验 `character_id` 归属，非本人人物返回错误

## 数据来源

所有数据来自本地持久化的 ESI 钱包流水表（`eve_character_wallet_journals`），不实时调用 CCP API。
星系名称来自本地 SDE 数据库（`mapSolarSystems` 表），NPC 名称通过 SDE `GetTypes` 接口查询，支持中英文。

## API 请求结构

### 个人单人物

```json
POST /api/v1/info/npc-kills
{
  "character_id": 12345,      // 必填
  "start_date": "2026-01-01", // 可选，格式 YYYY-MM-DD
  "end_date":   "2026-03-31", // 可选，end 取当天 23:59:59
  "language":   "zh",         // 可选，默认 zh
  "page":       1,            // 可选，0 = 不分页返回全部
  "page_size":  20            // 可选
}
```

### 个人全人物汇总

```json
POST /api/v1/info/npc-kills/all
{
  "start_date": "2026-01-01",
  "end_date":   "2026-03-31",
  "language":   "zh",
  "page":       1,
  "page_size":  20
}
```

### 公司报表（管理员）

```json
POST /api/v1/system/npc-kills
{
  "start_date": "2026-01-01",
  "end_date":   "2026-03-31",
  "language":   "zh"
}
```

## 响应结构

### 个人报表 `NpcKillResponse`

```text
summary    NpcKillSummary          总览统计
by_npc     []NpcKillByNpc          按 NPC 统计（杀怪数降序）
by_system  []NpcKillBySystem       按星系统计（总金额降序）
trend      []NpcKillTrend          按天趋势（日期升序）
journals   []NpcKillJournalItem    分页流水明细
total      int64                   总记录数（分页用）
page       int
page_size  int
```

### 公司报表 `NpcKillCorpResponse`

```text
summary    NpcKillSummary                  全员汇总总览
members    []NpcKillCorpMemberSummary      按成员统计（实际收入降序）
by_system  []NpcKillBySystem               全员按星系统计
trend      []NpcKillTrend                  全员按天趋势
```

## 核心计算逻辑

### 总览（calcSummary）

| 字段 | 计算方式 |
| --- | --- |
| `total_bounty` | 所有 `bounty_prizes` 条目 `amount` 之和 |
| `total_ess` | 所有 `ess_escrow_transfer` 条目 `amount` 之和 |
| `total_tax` | 所有条目 `tax` 之和（通常为负数） |
| `actual_income` | `total_bounty + total_ess + total_tax` |
| `total_records` | `bounty_prizes` 条目数 |
| `estimated_hours` | 见下方说明 |

**估算有效时长**：仅统计 `bounty_prizes` 条目中金额 ≥ 平均值 30% 的「有效记录」，每条有效记录视为约 20 分钟，结果取两位小数。低金额记录（如挂机、偶发）不计入时长，避免虚高。

公式：`estimated_hours = round(valid_count × 20 / 60, 2)`

### 按 NPC 统计（calcByNpc）

仅处理 `bounty_prizes` 条目，解析 `reason` 字段格式：

```text
"npc_type_id: kill_count, npc_type_id: kill_count, ..."
```

相同 NPC ID 跨条目累加击杀数，从 SDE 查询本地化 NPC 名称，按击杀数降序排列。
通过此分组，可区分标准刷怪、血族入侵（Sansha NPC ID）和 Pochven（三角洲 NPC ID）等活动类型。

### 按星系统计（calcBySystem）

仅处理 `bounty_prizes` 条目，使用 `context_id` 作为星系 ID，统计每个星系的记录数和总金额，按金额降序排列。`ess_escrow_transfer` 不参与星系统计。

### 趋势（calcTrend）

仅统计 `bounty_prizes` 条目，按 `YYYY-MM-DD` 聚合每天的总金额和记录数，按日期升序排列。

## 流水明细字段

| 字段 | 说明 |
| --- | --- |
| `ref_type` | `bounty_prizes`（标准悬赏）或 `ess_escrow_transfer`（ESS 转账） |
| `amount` | 本次收入金额（正数） |
| `tax` | 扣税金额（通常为负数） |
| `solar_system_name` | 仅 `bounty_prizes` 有值（来自 context_id） |
| `reason` | NPC ID 原始字符串，格式同上 |
| `character_name` | 全人物汇总和公司报表时填充 |

## UI 呈现

### 个人页面

1. 人物选择器（下拉，含头像）+ 「所有人物」选项 + 日期范围选择
2. 5 卡片总览：总悬赏 / 总税金 / 实际收入 / 记录数 / 估算时长
3. 双列布局：按 NPC 统计表 + 按星系统计表
4. 时间趋势表（有数据时显示）
5. 分页流水明细表（ref_type 以 tag 展示，金额带颜色）

### 管理员页面

1. 日期范围选择
2. 5 卡片总览（同上）
3. 成员列表（按实际收入降序，展示每人悬赏 / ESS / 税 / 实际收入 / 记录数）
4. 双列布局：按星系统计 + 时间趋势
5. 无流水明细（管理视角不展示个人流水）

## 关键不变量

- `bounty_prizes` 和 `ess_escrow_transfer` 是唯一参与计算的 ref_type，其余钱包类型不纳入
- 税金（`tax` 字段）为负数，参与实际收入计算时相当于扣减
- 星系统计和趋势仅基于 `bounty_prizes`，ESS 转账不带星系上下文
- 估算时长不是精确值，仅供参考，基于 30% 阈值过滤低效记录
- 个人接口强制校验人物归属，不可跨用户查询
- 公司接口只涵盖当前已绑定有效 token 的人物

## 主要代码文件

- `server/internal/service/npc_kill.go` — 业务逻辑（汇总计算、NPC/星系/趋势解析）
- `server/internal/repository/npc_kill.go` — 数据查询（钱包流水、星系名称）
- `server/internal/handler/npc_kill.go` — HTTP 处理层
- `static/src/api/npc-kill.ts` — 前端 API 封装
- `static/src/views/info/npc-kills` — 个人报表页面
- `static/src/views/system/npc-kills` — 管理员报表页面
