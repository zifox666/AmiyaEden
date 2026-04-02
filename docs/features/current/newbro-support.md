---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-04-03
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/newbro_service.go
  - server/internal/service/newbro_report.go
  - server/internal/service/newbro_settings.go
  - server/internal/service/menu.go
  - server/internal/handler/newbro_admin.go
  - static/src/api/newbro.ts
  - static/src/views/newbro
  - static/src/views/system/newbro-settings
---

# 新人帮扶模块

## 当前能力

- 管理员可把已有用户授予真实系统职权 `captain`
- 当前仍符合新人资格的用户可在 `新人选队长` 页面选择或切换队长
- 该页面默认聚焦“当前队长 + 可选队长”，最近关系历史放在独立 tab 中
- 可选队长列表按最近在线时间倒序展示，并同时显示主人物名与用户昵称
- 若当前用户自己也具备 `captain` 职权，他自己不会出现在可选队长列表中
- 用户与队长的关联按历史记录保存，支持高频切换与再次建立关系
- 队长页分为 `新人列表` 与 `收益归因明细` 两个页签
- 队长页新增 `招募新人` 页签，可搜索符合资格的新用户并直接将其挂到当前队长名下
- `新人列表` 默认筛选当前仍关联中的新人，并展示这些用户的当前主人物名与昵称
- 队长侧招募与新人侧自行选队长共享同一套关系切换规则；新关系记录会保存实际发起变更的用户 ID
- 无论是新人侧选队长，还是队长侧招募新人，都不允许用户把自己与自己建立帮扶关系
- 新增按钮让新人可以主动解除当前与队长的关联，队长也可以在 `新人列表` 中结束某个玩家的帮扶关系；相关历史依旧会写入 `newbro_captain_affiliation` 历史记录
- `收益归因明细` 页签展示概览卡片、赏金归因记录与奖励发放历史
- 管理员可查看全量队长绩效，列表与详情都会展示队长主人物名和昵称，并可手动触发赏金归因同步与奖励处理
- `帮扶管理` 页面新增 `奖励发放历史` tab，按“每次处理、每名队长一行”的粒度展示历史结算结果
- `帮扶管理` 页面新增 `关系变更历史` tab，可按变更时间、队长用户 ID、新人人物 ID 查看全量关系记录，并显示实际创建该关系的人物
- 赏金归因结果会持久化到 ledger 表，供后续队长奖励结算直接使用；奖励处理后会回写 `processed_at`

## 前端金额展示

- 新人帮扶页面里的 ISK 金额统一使用 plain ISK value style。
- 队长奖励结算得到的是伏羲币，不受 ISK 格式化标准约束。


## 新人资格判定

当前规则基于用户的全部绑定人物：

- 只要任一人物 `total_sp >= 20,000,000`，用户就不再属于新人
- `total_sp >= 10,000,000` 的绑定人物数达到 `3`，用户也不再属于新人

实现方式：

- 服务层把结果缓存到 `newbro_player_state`
- 这是“当前资格快照”，不是永久毕业标记
- 已判定为非新人的用户，只有在规则版本变化时才会重新计算
- 当前仍是新人的用户，缓存快照超过配置的刷新间隔后才会重新计算
- 如果刷新后发现用户已不再符合新人资格，服务层会在同一次刷新中结束其当前 active 的队长关联
- 管理员可在 `系统管理 -> 帮扶设置` 调整资格阈值、刷新间隔与队长奖励比例

当前配置面向持久化行为的含义：

- `max_character_sp` 控制“任一人物技能点毕业线”
- `multi_character_sp` 与 `multi_character_threshold` 共同控制“多人物技能点毕业线”
- `refresh_interval_days` 控制当前仍为新人的资格快照刷新频率
- `bonus_rate` 控制队长奖励换算比例
- 赏金归因回溯窗口当前固定为最近 `30` 天，尚未做成后台配置项

## 赏金归因规则

归因时：

- 玩家侧会遍历该新人用户下的全部人物钱包流水
- 队长侧只看该队长用户“归因作业执行时”的当前主人物
- 玩家流水只把 `bounty_prizes` 作为归因输入
- 需要满足：
  - 该玩家流水发生时存在 active 的队长关联
  - 玩家与队长流水在同一 `system_id`
  - 双方 `ref_type` 一致
  - 两条流水时间差不超过 `5` 分钟

持久化规则：

- 每条已归因的玩家钱包流水只写入一条 `captain_bounty_attribution`
- `captain_character_id` 在写入时冻结保存
- `processed_at` 为空代表该条归因尚未参与奖励结算
- 每次同步从头扫描最近 `30` 天内尚未归因的流水，已归因记录通过 LEFT JOIN 自动排除
- 当前只为"现在仍属于新人"的用户计算与回填归因

多候选匹配时，服务层使用稳定排序打破并列：

- 优先选择时间差最小的队长钱包流水
- 再按钱包流水时间升序
- 再按钱包流水 ID 升序

## 队长奖励处理规则

- 当前由管理员手动触发，不会自动跟随归因同步执行
- 每次处理只读取 `captain_bounty_attribution.processed_at IS NULL` 的记录
- 按 `captain_user_id` 分组汇总未处理归因金额
- 奖励换算公式为：`sum(amount) / 1,000,000 * (bonus_rate / 100)`
- `bonus_rate` 以百分比配置，默认值为 `20`
- 奖励金额四舍五入保留 `2` 位小数后，发放到伏羲币积分
- 每次处理会写入一条 `captain_reward_settlement`，并把参与结算的归因记录统一回写同一个 `processed_at`
- 队长奖励流水使用伏羲币 `ref_type = newbro_captain_reward`

## 入口

### 前端页面

- `static/src/views/newbro/select-captain` — 新人选队长
- `static/src/views/newbro/captain` — 我是队长
- `static/src/views/newbro/manage` — 帮扶管理
- `static/src/views/system/newbro-settings` — 帮扶设置

### 后端路由

用户侧：

- `GET /api/v1/newbro/captains`
- `GET /api/v1/newbro/affiliation/me`
- `POST /api/v1/newbro/affiliation/select`
- `POST /api/v1/newbro/affiliation/end`

队长侧：

- `GET /api/v1/newbro/captain/overview`
- `GET /api/v1/newbro/captain/players`
- `GET /api/v1/newbro/captain/eligible-players`
- `POST /api/v1/newbro/captain/enroll`
- `GET /api/v1/newbro/captain/attributions`
- `GET /api/v1/newbro/captain/rewards`
- `POST /api/v1/newbro/captain/affiliation/end`

管理侧：

- `GET /api/v1/system/newbro/settings`
- `PUT /api/v1/system/newbro/settings`
- `GET /api/v1/system/newbro/captains`
- `GET /api/v1/system/newbro/captains/:user_id`
- `GET /api/v1/system/newbro/affiliations/history`
- `GET /api/v1/system/newbro/rewards`
- `POST /api/v1/system/newbro/attribution/sync`
- `POST /api/v1/system/newbro/reward/process`

## 关键页面行为

### 新人选队长

- 页面加载时先读取 `GET /api/v1/newbro/affiliation/me`
- 若当前已不是新人，前端会提示不符合资格的原因与最近评估时间，并清空候选列表后重定向离开该页面
- 选择同一位当前队长时，后端会复用当前 active 关系，不会重复创建历史记录
- 自选自己为队长会被后端拒绝，即使绕过前端候选列表也不允许

### 队长帮扶

- `新人列表` 支持 `all` / `active` / `historical` 三种状态筛选
- `招募新人` 支持按昵称或主人物名搜索
- `收益归因明细` 的 `start_date` / `end_date` 过滤参数必须使用 `YYYY-MM-DD`；非法日期会返回参数错误
- 招募列表会排除当前队长自己
- 招募列表也会排除当前已经 active 挂在该队长名下的玩家，但允许把符合资格、当前挂在其他队长名下的玩家切换过来
- 队长结束关系时，只能结束当前属于自己的 active 关系

### 帮扶管理

- 绩效页展示全部队长的概览排行，并支持查看单个队长详情；使用普通分页，默认每页 `20` 条
- 绩效页搜索支持按队长当前昵称或任一已绑定人物名检索
- 奖励发放历史页展示汇总卡片和按处理批次展开的结算记录，并支持按队长当前昵称或任一已绑定人物名检索
- 关系变更历史页支持按变更日期、队长人物名或昵称、新人人物名或昵称过滤
- 关系变更历史页名称筛选不区分大小写
- 关系变更历史页的 `change_start_date` / `change_end_date` 必须使用 `YYYY-MM-DD`；非法日期会返回参数错误

## 权限边界

- `新人选队长` 页面不是单纯的 `Login` 页面
- 用户必须：
  - 是已登录且非 `guest`
  - 当前 `newbro_player_state.is_currently_newbro = true`
- `我是队长` 页面要求真实职权 `captain`
- `帮扶管理` 页面要求 `admin`
- 后端服务层会再次校验资格或职权，前端菜单隐藏只属于 UX

## 关键不变量

- 同一时间一个新人只能关联一个队长
- 切换队长时，旧关联必须结束，新关联必须新建
- 历史关联记录不可覆盖为单行“最新状态”
- 用户不能把自己选为自己的队长，队长也不能把自己招募为自己名下的新人
- 队长归因台账写入后不应依赖重新扫描历史来解释或结算
- 非新人用户不应继续看到 `新人选队长` 入口；页面直达时会被重定向
- 非 `captain` 用户不应看到 `我是队长` 页面入口

## 主要代码文件

- `server/internal/model/newbro_player_state.go`
- `server/internal/model/newbro_captain_affiliation.go`
- `server/internal/model/captain_bounty_attribution.go`
- `server/internal/model/captain_bounty_sync_state.go`
- `server/internal/model/captain_reward_settlement.go`
- `server/internal/service/newbro_service.go`
- `server/internal/service/captain_reward_processing.go`
- `server/internal/service/newbro_settings.go`
- `server/internal/service/newbro_report.go`
- `server/internal/service/menu.go`
- `server/internal/handler/newbro_user.go`
- `server/internal/handler/newbro_captain.go`
- `server/internal/handler/newbro_admin.go`
- `static/src/api/newbro.ts`
- `static/src/router/modules/newbro.ts`
- `static/src/router/modules/system.ts`
- `static/src/views/newbro/`
- `static/src/views/system/newbro-settings/`
- `static/src/locales/langs/zh.json`
- `static/src/locales/langs/en.json`
