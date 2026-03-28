---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-28
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/srp.go
  - server/internal/service/auto_srp.go
  - server/internal/repository/srp.go
  - server/internal/repository/killmail.go
  - static/src/api/srp.ts
  - static/src/views/srp
---

# SRP 模块

## 当前能力

- 舰船价格表查询、维护、删除
- 个人补损申请提交
- 我的补损申请列表
- 我的 KM、按舰队筛选 KM、KM 详情
- 管理端手动自动审批符合规则的待审批申请，`admin` 也可操作
- 审核列表、审核详情、审核通过 / 拒绝，`admin` 也可操作
- 管理端审核列表按「待处理 / 发放记录」tab 分组，并将 tab 条件传给 `/srp/applications`
- 单条发放补损，`admin` 也可操作
- 管理端批量发放补损汇总、按用户批量发放补损，`admin` 也可操作
- 管理端待处理 tab 支持切换「伏羲币补损 / 手动打钱」两种发放模式，默认使用伏羲币补损

## 入口

### 前端页面

- `static/src/views/srp/apply`
- `static/src/views/srp/manage`
- `static/src/views/srp/prices`

### 后端路由

- `/api/v1/srp/prices`
- `/api/v1/srp/applications`
- `/api/v1/srp/applications/me`
- `/api/v1/srp/killmails/me`
- `/api/v1/srp/killmails/fleet/:fleet_id`
- `/api/v1/srp/killmails/detail`
- `/api/v1/srp/open-info-window`
- `/api/v1/srp/applications/auto-approve`
- `/api/v1/srp/applications/batch-payout-summary`
- `/api/v1/srp/applications/fuxi-payout`
- `/api/v1/srp/applications/:id/review`
- `/api/v1/srp/applications/:id/payout`
- `/api/v1/srp/applications/users/:user_id/payout`

## 权限边界

- 价格新增 / 更新要求 `srp:price:add`
- 价格删除要求 `srp:price:delete`
- 审核列表、详情、审批（approve/reject）要求 `srp` 或 `fc`
- 发放、批量发放、自动审批要求 `srp` 或 `admin`
- 其余个人能力默认要求 `Login`

## 计算 SRP 推荐金额

SRP 推荐金额同时由手动SRP机制和自动SRP机制使用，用于计算SRP申请时的推荐推荐金额和默认最终金额

### 在创建申请前额外执行装配验证

#### 前置条件

- 舰队的 `auto_srp_mode` 不为 `disabled`（当前支持 `submit_only` 和 `auto_approve`）
- 舰队必须关联一个 `FleetConfigID`，该配置下有装配（fittings）定义

- **基础金额确定**：配置装配的 `srp_amount > 0` 则用此金额。若无，则从全局舰船价格表按 `ship_type_id` 查询获得推荐金额，再无则为0
- **按槽位类别验证**：将 KM 物品按 flag 名称归一化到类别（`HiSlot`、`MedSlot`、`LoSlot` 等），与配置物品逐项比对
- **跳过的类别**：`DroneBay`、`FighterBay`、`Cargo` 不参与验证
- **可选物品**（`importance = optional`）：不参与验证
- **可替换物品**（`importance = replaceable`）：原始 type_id 数量不足时，检查配置的替代品
- **惩罚规则**：
  - 数量不足时，按该物品的 `penalty` 字段判定：`none` →  推荐金额为0，`half` → 推荐金额为半价
  - 使用替代品时，按 `replacement_penalty` 字段判定，规则同上
  - 多项不符时，`none` 优先于 `half`（即只要有一项是 `none`，推荐金额为0）

#### 若不符合额外装配验证条件

- 从全局舰船价格表按 `ship_type_id` 查询获得推荐金额，再无则为0

## 手动 SRP

手动 SRP 是用户自行提交补损申请，管理员审核后发放的标准流程。

### 申请提交

由 `SrpService.SubmitApplication` 处理，验证顺序：

1. **角色归属**：`character_id` 必须属于当前登录用户
2. **备注要求**：未关联舰队时，`note` 不能为空
3. **重复检查**：同一 `killmail_id` + `character_id` 不能重复提交
4. **KM 关联**：通过 `EveCharacterKillmail` 表验证角色与 KM 的关联关系，并加载 KM 详情
5. **受害者确认**：KM 的 `character_id` 必须与申请角色一致
6. **舰队验证**（关联舰队时）：
   - 舰队必须存在
   - KM 时间必须在舰队 `start_at` ~ `end_at` 范围内
   - 角色必须是该舰队的成员
7. **金额设定**：在提交申请时，系统自动根据上方推荐金额计算方法计算推荐金额和初始最终金额
8. **创建申请**：初始状态 `review_status = submitted`，`payout_status = notpaid`

### 审批

由 `SrpService.ReviewApplication` 处理：

- 支持 `approve`（批准）和 `reject`（拒绝）两种操作
- 拒绝时必须填写 `review_note`
- 批准时可以修改 `final_amount`
- **已发放的申请不能重新审批**
- 已批准或已拒绝的申请可以重新审批（编辑/重新拒绝），只要未发放

### 发放

分为单条发放和批量发放两种方式：

**单条发放**（`SrpService.Payout`）：

- 申请必须已批准（`review_status = approved`）
- 不能重复发放（`payout_status` 已为 `paid` 则拒绝）
- 发放时可以最终覆盖 `final_amount`
- 记录 `paid_by`、`paid_at`
- 当模式为 `manual_transfer`（手动打钱）时，仅将申请标记为已发放，保留当前人工线下打款流程
- 当模式为 `fuxi_coin`（伏羲币补损）时，将 `final_amount` 按 `1,000,000 ISK : 1 伏羲币` 换算，四舍五入保留 `2` 位小数，写入系统钱包 `ref_type = srp_payout` 流水，并将申请结案
- 伏羲币流水 `reason` 包含 SRP 申请 ID、舰船名，以及存在时的舰队标题

**按用户批量发放**（`SrpService.BatchPayoutByUser`）：

- 将某用户所有已批准且未发放的申请一次性标记为已发放
- 使用数据库事务 + `SELECT FOR UPDATE` 防止并发发放
- 若发放过程中待发放集合发生变化，事务回滚并要求刷新重试

**伏羲币批量发放**（`SrpService.BatchPayoutAsFuxiCoin`）：

- 处理全部“已批准且未发放”的申请，不再按用户逐个确认
- 对每条申请分别换算伏羲币金额、写系统钱包流水、再标记 SRP 为已发放
- 整体使用数据库事务，任一申请发放失败则整批回滚

### 管理端列表

- 申请列表支持按 tab 分组：`pending`（待处理：submitted/approved + notpaid）和 `history`（发放记录：paid 或 rejected）
- 列表结果附带舰队标题、FC 名称、用户昵称等关联信息
- 批量发放汇总按用户聚合，展示主角色名、昵称、总金额、申请数量
- 待处理 tab 的发放方式单选默认选中「伏羲币补损」；切到「手动打钱」后，顶部“批量发放”和行内“发放”恢复旧的人工打款面板流程

### KM 查询

- **我的 KM**（`GetMyKillmails`）：返回当前用户所有角色作为受害者的最近 30 天 KM，限 200 条
- **按舰队筛选 KM**（`GetFleetKillmails`）：返回当前用户在指定舰队时间范围内、且为舰队成员的角色的受害 KM
- **KM 详情**（`GetKillmailDetail`）：返回 KM 装配详情，按槽位类别分组并合并同类物品，支持中英文名称

## 自动 SRP（Auto-SRP）

自动 SRP 在舰队 PAP 结算后触发，由 `AutoSrpService.ProcessAutoSRP(fleetID)` 入口驱动。

- **手动触发**： 允许SRP管理员针对选定的特定舰队ID为目标重新跑自动审批逻辑。

### 自动 SRP 前置条件

- 舰队的 `auto_srp_mode` 不为 `disabled`（当前支持 `submit_only` 和 `auto_approve`）
- 舰队必须关联一个 `FleetConfigID`，该配置下有装配（fittings）定义

### 处理流程

1. **构建舰队上下文**：加载舰队配置 → 装配列表 → 按 `ship_type_id` 索引 → 预加载配置物品和替代品
2. **遍历舰队成员**：对每个成员查询其作为受害者的 KM（`EveCharacterKillmail.victim = true`）
3. **时间范围过滤**：只处理 `killmail_time` 在舰队 `start_at` ~ `end_at` 之间的 KM
4. **舰船匹配**：KM 的 `ship_type_id` 必须在配置装配中存在，否则跳过
5. **金额确定**：计算 SRP 推荐金额，若推荐金额为0，则跳过。若推荐金额 > 0，则将此金额设定为该SRP申请的推荐金额和最终金额。
6. **创建 SRP 申请**：自动提交，重复申请（唯一约束冲突）静默跳过

- `auto_approve` 模式：验证通过的申请自动标记为 `approved`，备注”补损根据舰队的自动补损设置，已由系统自动批准。”
- `submit_only` 模式：申请保持 `submitted` 状态，等待管理员手动审批
- 不符合规则的配置和KM，依然允许成员手动自行提交补损申请，自动SRP机制会跳过这些KM

## 关键不变量

- 审核与发放是分离的接口，不要假设它们是一步完成
- 批量发放按用户聚合，只处理”已批准且未发放”的申请
- 价格表、舰队配置金额、自动 SRP 逻辑之间存在耦合
- 涉及 killmail、舰队、SDE 名称映射的改动要跨模块检查
- 所有 killmail 数据库访问通过 `KillmailRepository` 进行，不在 service 层直接使用 `global.DB`

## 主要代码文件

- `server/internal/service/srp.go` — 手动 SRP 业务逻辑（申请、审批、发放、KM 查询）
- `server/internal/service/auto_srp.go` — 自动 SRP 处理与装配验证
- `server/internal/repository/srp.go` — SRP 申请数据访问
- `server/internal/repository/killmail.go` — 击杀邮件数据访问（KM 主记录、物品、角色关联）
- `server/internal/router/router.go`
- `static/src/api/srp.ts`
- `static/src/views/srp`
