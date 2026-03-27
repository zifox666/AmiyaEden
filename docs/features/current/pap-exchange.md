---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-26
source_of_truth:
  - server/internal/model/pap_type_rate.go
  - server/internal/repository/pap_type_rate.go
  - server/internal/service/pap_exchange.go
  - server/internal/service/fleet.go
  - server/internal/handler/pap_exchange.go
  - server/internal/router/router.go
  - static/src/api/pap-exchange.ts
  - static/src/views/system/pap-exchange
---

# PAP 兑换汇率

## 概述

PAP 兑换汇率功能允许管理员为每种舰队行动类型（Skirmish / Strategic / CTA）单独配置系统钱包兑换比率。FC 发放 PAP 时，后端根据舰队重要性自动换算为对应钱包金额，取代原来固定 1:1 的比率。

此功能与联盟 PAP（Alliance PAP）系统完全独立。联盟 PAP 月度归档当前为纯归档操作，不涉及钱包兑换。

## 当前能力

- 管理员可在「系统管理 → PAP兑换」页面查看并修改三种 PAP 类型的钱包兑换汇率
- 默认汇率：Skirmish 10、Strategic 30、CTA 50（系统钱包 / 1 PAP）
- FC 对某舰队执行「发放 PAP」时，系统根据该舰队的 `importance` 字段自动选择对应汇率
- 汇率持久化于数据库 `pap_type_rate` 表；缺失行在首次读取时用默认值自动补全

## PAP 类型映射

| 舰队重要性（`fleet.importance`） | PAP 类型 | 默认汇率 |
| --- | --- | --- |
| `cta` | CTA（全面集结） | 50 |
| `strat_op` | Strategic（战略行动） | 30 |
| `other`（及其他未知值） | Skirmish（游击队） | 10 |

## 关键不变量

- 汇率配置存储在 `pap_type_rate` 表，以 `pap_type`（`skirmish` / `strat_op` / `cta`）为主键
- `pap_type_rate` 与 `system_config` 中的 `pap.wallet_per_pap` 互相独立；后者仅用于联盟 PAP 月度结算
- `papImportanceToWalletRate` 是将舰队重要性映射到汇率的纯函数；当 `pap_type` 不在汇率表中时回退到 1
- 三种 PAP 类型固定不可增删；管理页面仅允许修改汇率数值
- 重新发放 PAP（re-issue）时钱包差量按汇率换算，与首次发放一致

## 入口

### 前端页面

- `static/src/views/system/pap-exchange`

### 后端路由

- `GET /api/v1/system/pap-exchange/rates`
- `PUT /api/v1/system/pap-exchange/rates`

## 主要代码文件

- `server/internal/model/pap_type_rate.go` — 模型、类型常量、`NormalizePAPLevel`
- `server/internal/repository/pap_type_rate.go` — 数据访问、默认值补全
- `server/internal/service/pap_exchange.go` — 汇率 CRUD 业务逻辑
- `server/internal/service/fleet.go` — `papImportanceToWalletRate`（发放 PAP 时调用）
- `server/internal/handler/pap_exchange.go` — HTTP 处理器
- `static/src/api/pap-exchange.ts` — 前端 API 包装层
- `static/src/views/system/pap-exchange/index.vue` — 管理页面

## 回归测试

- `server/internal/model/pap_type_rate_test.go` — `NormalizePAPLevel` 的所有输入分支
- `server/internal/service/fleet_test.go` — `papImportanceToWalletRate` 的所有映射分支与缺失键回退
