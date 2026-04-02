---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-04-02
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/mentor_service.go
  - server/internal/service/mentor_reward.go
  - server/internal/service/mentor_eligibility.go
  - server/internal/service/mentor_settings.go
  - server/internal/handler/mentor_mentee.go
  - server/internal/handler/mentor_mentor.go
  - server/internal/handler/mentor_admin.go
  - static/src/api/mentor.ts
  - static/src/views/newbro/select-mentor
  - static/src/views/newbro/mentor
  - static/src/views/newbro/mentor-manage
  - static/src/views/system/mentor-reward-stages
---

# 导师系统

## 当前能力

- 管理员可把已有用户授予真实系统职权 `mentor`
- 已登录且符合学员资格的用户可在 `新人选导师` 页面查看导师列表并提交申请
- 导师候选卡片会展示导师主人物名、昵称以及联系方式（QQ / Discord）
- 当前导师关系卡片会展示导师联系方式（QQ / Discord）
- 学员同一时间最多只能有一条 `pending` 或 `active` 的导师关系
- 导师可在 `我是导师` 页面查看待处理申请中的学员卡片信息（含 QQ / Discord 联系方式），并接受或拒绝申请
- `我是导师` 页面中的待处理申请与学员列表按学员主人物最近登录时间倒序展示；当前实现使用学员主人物所属账号的 `last_login_at` 作为可用在线时间
- 导师可分页查看自己的学员列表、当前关系状态、学员联系方式（QQ / Discord）、学员总技能点、军团 PAP、活跃天数、已发奖励阶段，以及累计已发伏羲币
- 导师可在 `我是导师` 页面查看当前生效的只读 `导师奖励阶段` 配置，便于理解学员奖励进度
- 当导师存在待处理学员申请时，`新人帮扶` 一级菜单与 `我是导师` 菜单会显示相同的待处理数量徽标
- 管理员可在 `导师管理` 页面查看全部导师关系；对 `pending` 状态可取消学员申请，对 `active` 状态可撤销导师关系
- 管理员可在 `导师管理` 页面查看导师奖励发放记录，按 ledger 方式分页展示，并支持按导师人物名或昵称搜索
- 管理员可在 `导师奖励阶段` 页面配置阶段化奖励规则、学员资格阈值，并手动执行一次奖励处理
- 每日定时任务会自动扫描进行中的导师关系，按阶段顺序发放伏羲币奖励；当全部阶段都已发放后，关系会被标记为 `graduated`

## 学员资格判定

当前规则由系统配置驱动，默认值为：

- 任一绑定人物 `total_sp < 4,000,000`
- 账号注册时间 `<= 7` 天

实现细节：

- 资格判定读取用户全部已绑定人物的技能点快照
- 当前数据模型没有持久化“加入允许军团的时间”，因此导师系统当前使用 `user.created_at` 作为账号年龄窗口，而不是 corp join timestamp
- 管理员可在 `导师奖励阶段` 页面修改上述两个阈值；保存后会立即作用于菜单资格快照、候选导师列表与申请校验
- 候选导师列表与申请接口虽然挂在 `Login` 路由组下，但服务层会再次校验学员资格
- 如果用户没有任何可评估人物，资格判定结果会返回 `no_characters`

## 奖励阶段与发放

- 奖励阶段由 `mentor_reward_stage` 表配置，条件类型当前支持：
  - `skill_points`
  - `pap_count`
  - `days_active`
- 奖励处理只扫描 `active` 关系
- 阶段按 `stage_order` 升序检查，只有前序阶段全部完成后才会继续发放后续阶段
- 发放成功后会写入 `mentor_reward_distribution`
- 钱包流水使用伏羲币 `ref_type = mentor_reward`
- 奖励记录同时保存 `stage_id` 和 `stage_order` 快照；进度判断以 `stage_order` 为准，避免管理员替换阶段配置后丢失已发进度或重复发奖

## 入口

### 前端页面

- `static/src/views/newbro/select-mentor` — 新人选导师
- `static/src/views/newbro/mentor` — 我是导师
- `static/src/views/newbro/mentor-manage` — 导师管理
  - 导师关系 tab：查看全部导师关系，支持按导师/学员人物名或昵称筛选
  - 奖励发放记录 tab：按 ledger 方式分页显示导师奖励发放记录，并支持按导师人物名或昵称搜索
- `static/src/views/system/mentor-reward-stages` — 导师奖励阶段

### 后端路由

学员侧：

- `GET /api/v1/mentor/mentors`
- `GET /api/v1/mentor/me`
- `POST /api/v1/mentor/apply`

导师侧：

- `GET /api/v1/mentor/dashboard/applications`
- `GET /api/v1/mentor/dashboard/mentees`
- `GET /api/v1/mentor/dashboard/reward-stages`
- `POST /api/v1/mentor/dashboard/accept`
- `POST /api/v1/mentor/dashboard/reject`

管理侧：

- `GET /api/v1/system/mentor/relationships`
- `GET /api/v1/system/mentor/reward-distributions`
- `POST /api/v1/system/mentor/revoke`
- `GET /api/v1/system/mentor/settings`
- `PUT /api/v1/system/mentor/settings`
- `GET /api/v1/system/mentor/reward-stages`
- `PUT /api/v1/system/mentor/reward-stages`
- `POST /api/v1/system/mentor/reward/process`

## 权限边界

- `新人选导师` 页面不是单纯的 `Login` 页面
- 用户必须：
  - 是已登录且非 `guest`
  - 当前导师学员资格快照为 true
- 页面入口依赖 `/api/v1/me` 返回的 `is_mentor_mentee_eligible` 做菜单与路由过滤
- 页面加载后仍会读取 `/api/v1/mentor/me` 做二次 UX 校验；导师候选与申请动作也会在后端服务层再次校验
- 当当前导师关系状态变为 `active` 后，页面不再展示 `可选导师` 区块
- `我是导师` 页面要求真实系统职权 `mentor`
- `导师管理` 页面要求 `admin`
- `导师奖励阶段` 页面要求 `admin`
- 管理员不是导师的隐式别名；普通 `admin` 只能使用管理页，不能访问导师 dashboard 接口

## 关键不变量

- 同一时间一个学员只能存在一条 `pending` 或 `active` 的导师关系
- 学员不能选择自己作为导师
- 只有关系状态为 `pending` 时，导师才能接受或拒绝该申请
- 只有 `pending` 或 `active` 关系允许被管理员撤销
- 奖励阶段必须以正整数 `stage_order` 排序，且序号不可重复
- 奖励阶段的 `threshold` 与 `reward_amount` 当前仅支持正整数配置
- 所有阶段奖励都完成后，关系才会被标记为 `graduated`

## 主要代码文件

- `server/internal/model/mentor.go`
- `server/internal/repository/mentor.go`
- `server/internal/service/mentor_eligibility.go`
- `server/internal/service/mentor_settings.go`
- `server/internal/service/mentor_service.go`
- `server/internal/service/mentor_reward.go`
- `server/internal/handler/mentor_mentee.go`
- `server/internal/handler/mentor_mentor.go`
- `server/internal/handler/mentor_admin.go`
- `server/jobs/mentor_reward.go`
- `static/src/api/mentor.ts`
- `static/src/router/modules/newbro.ts`
- `static/src/router/modules/system.ts`
- `static/src/views/newbro/select-mentor/`
- `static/src/views/newbro/mentor/`
- `static/src/views/newbro/mentor-manage/`
- `static/src/views/system/mentor-reward-stages/`
- `static/src/locales/langs/zh.json`
- `static/src/locales/langs/en.json`
