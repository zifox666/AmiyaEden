---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-22
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/eve_info.go
  - server/internal/service/fittings.go
  - server/internal/service/npc_kill.go
  - static/src/api/eve-info.ts
  - static/src/api/npc-kill.ts
  - static/src/views/info
---

# EVE 信息与报表

## 当前能力

- 钱包流水
- 技能列表
- 舰船列表
- 植入体
- 资产
- 合同列表与详情
- 装配列表与保存
- 个人 NPC 刷怪报表
- 全量 NPC 刷怪报表

## 入口

### 前端页面

- `static/src/views/info/wallet`
- `static/src/views/info/skill`
- `static/src/views/info/ships`
- `static/src/views/info/implants`
- `static/src/views/info/assets`
- `static/src/views/info/contracts`
- `static/src/views/info/fittings`
- `static/src/views/info/npc-kills`

### 后端路由

- `/api/v1/info/wallet`
- `/api/v1/info/skills`
- `/api/v1/info/ships`
- `/api/v1/info/implants`
- `/api/v1/info/assets`
- `/api/v1/info/contracts`
- `/api/v1/info/contracts/detail`
- `/api/v1/info/fittings`
- `/api/v1/info/fittings/save`
- `/api/v1/info/npc-kills`
- `/api/v1/info/npc-kills/all`
- `/api/v1/system/npc-kills`

## 权限边界

- 用户侧信息查询要求 `Login`，`guest` 不可访问
- 公司级 NPC 刷怪报表属于 `/system` 管理能力，要求 `admin`

## 关键不变量

- 此模块基于本地持久化的 ESI / SDE 数据与查询服务，不是页面直接调 CCP
- NPC 刷怪既有用户视角也有管理员视角，文档和实现都要区分清楚
- 装配功能属于 Info 模块，但也被舰队配置与自动 SRP 复用

## 主要代码文件

- `server/internal/service/eve_info.go`
- `server/internal/service/fittings.go`
- `server/internal/service/npc_kill.go`
- `server/internal/router/router.go`
- `static/src/api/eve-info.ts`
- `static/src/api/npc-kill.ts`
- `static/src/views/info`
