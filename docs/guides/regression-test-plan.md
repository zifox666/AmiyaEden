---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-22
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/standards/testing-and-verification.md
  - docs/guides/testing-guide.md
  - server/
  - static/
---

# 回归测试落地计划

## 目的

本文件不是新的“强制规则”来源，而是把现有测试标准落成一个可逐步实施的项目计划，回答三个实际问题：

- 这个仓库优先补哪些回归测试
- 每一类 bug 最少需要什么测试来挡住它再次发生
- 在不一次性重构测试基础设施的前提下，如何逐步提高回归保护

适用对象：

- 修 bug 的开发者
- 做模块重构的开发者
- 评审 PR 的 reviewer
- 维护文档与工程规范的人员

## 目标

- 让“修过一次又破一次”的问题尽量被测试挡在本地
- 让测试优先覆盖权限、fallback、查询拼接、contract 这些高风险边界
- 让新增测试可以贴近当前代码结构，不要求先引入大规模新基础设施
- 让每个模块都能逐步积累一组稳定的回归样例

## 非目标

- 不要求一次性补齐全仓所有模块测试
- 不要求为单个小 bug 立即引入完整 e2e 框架
- 不把 build / lint / typecheck 误当成行为回归测试
- 不在当前阶段强推高维护成本的 UI 组件快照测试

## 核心策略

默认遵循“离 bug 最近的一层补测试”：

1. 如果 bug 来自纯逻辑、归一化、权限判断，优先补 service / helper 单元测试。
2. 如果 bug 来自 repository 查询、join、fallback、字段映射，优先补 repository 回归测试。
3. 如果 bug 来自 API contract、响应整形、分页 envelope，优先补 handler / API contract 测试。
4. 如果 bug 来自前端纯 helper、筛选参数转换、名称 fallback，优先补 frontend unit test。
5. 如果 bug 只在页面装配层暴露，但根因在后端 contract，至少先在后端补测试，再做 frontend build 验证。

不要默认跳到最重的测试层。先选最小、最稳定、最能锁住真实风险的那一层。

## 风险分层与推荐测试类型

| 风险类型 | 常见例子 | 最小推荐测试 |
| --- | --- | --- |
| 权限边界 | admin 误编辑 admin、guest 误入 Login 页面 | service 单元测试 |
| 输入校验 / 归一化 | 昵称、QQ、Discord、时间范围、枚举修正 | service 或 helper 单元测试 |
| repository 查询拼接 | join 后列名歧义、筛选条件丢失、排序错误 | repository SQL / query-shape 测试 |
| repository fallback / merge | nickname 回退人物名、职权列表回退 guest | repository 行为测试 |
| API contract | 字段名改动、roles[] / role 差异、分页结构 | handler 或 API contract 测试 |
| frontend 纯逻辑 | 筛选参数转换、fallback 文案、表格 helper | `pnpm test:unit` |
| 本地化回归 | key 缺失、页面显示 raw key | 最少做 JSON 校验 + 页面层改动时人工验证 |
| 页面装配错误 | 列映射错误、错误字段绑定、操作按钮条件错误 | 优先补 helper / contract 测试，必要时补轻量 frontend 测试 |

## 当前仓库的优先级

第一优先级模块：

- `operation`: fleets、fleet-detail、pap、fleet-configs
- `system`: user、role、auto-role、pap、webhook
- `auth-and-characters`: `/api/v1/me`、人物绑定、资料补全

原因：

- 这些模块同时涉及权限、查询拼接、前后端 contract、fallback 展示
- 最近已经出现过 join 查询回归和显示字段 fallback 回归
- 这些模块对日常使用影响大，且 bug 通常不是编译期能发现的

第二优先级模块：

- `srp`
- `commerce`
- `info-and-reporting`
- `skill-planning`

第三优先级模块：

- 文档、静态配置、低风险只读页面

## 分阶段落地

### Phase 1: 新 bug 先锁住

目标：从现在开始，所有新修的 bug 都尽量附带最小回归测试。

要求：

- 只要是 bug fix，先问“最接近根因的是哪一层”
- 能合理测试时，必须补一条针对该 bug 的回归测试
- 如果当前缺少基础设施，至少补 query-shape / helper / service 级测试

完成标准：

- 新增 bug fix 不再只有 `go build` 或 `vue-tsc`
- 最近发生过的回归点开始拥有对应测试

建议优先补的样例：

- 舰队列表 FC 昵称 fallback
- join 后的 `deleted_at`、`status`、`id` 等歧义列问题
- 用户列表职权 fallback 与排序
- admin / super_admin 保护逻辑（super_admin 仅通过配置文件管理，API 不可分配 / 修改 / 删除）
- `/api/v1/me` 资料补全与联系方式唯一性

### Phase 2: 补模块级高频回归点

目标：给高频修改模块建立稳定的“保护带”。

每个高优先级模块至少补齐：

- 2 到 5 个 service / helper 回归测试
- 1 到 3 个 repository 回归测试
- 1 个关键 contract 测试

模块建议：

### Operation

- `fleet list` 查询拼接与 FC 展示 fallback
- PAP 日志展示字段回退逻辑
- auto SRP 模式归一化
- fleet 权限判断：`fc` / `admin` / `super_admin`

### Administration

- 用户资料更新校验与唯一性
- 保护管理员账号不能被普通 admin 修改 / 删除
- super_admin 职权不可通过 API 授予、修改或删除
- super_admin 用户不可通过 API 删除
- 登录时 super_admin 职权根据配置文件自动同步
- 职权列表 `roles[]` 与 legacy `role` fallback
- `GET /system/basic-config` 只返回固定系统标识，且无对应写接口
- auto-role 的 `Director -> admin` 规则仅接受伏羲军团 Fuxi Legion（`98185110`）的 corp role 信号
- `allow_corporations` 保存与读取时始终保留 `98185110`

### Auth And Characters

- 资料完整度判定
- 人物绑定 / 主人物切换的权限与输入校验
- `guest` 到 `user` 的边界行为

### Phase 3: 建共享测试夹具

目标：减少每次写测试都要重复搭环境的成本。

建议新增但不要求一次做完：

- backend repository dry-run GORM helper
- backend handler test helper
- frontend locale JSON 验证 helper
- frontend API contract mock helper

说明：

- 当前仓库已经适合先做 dry-run SQL / schema mapping 测试
- 如果后续 repository integration test 变多，再考虑统一测试数据库夹具
- 不要为了“未来可能会用”而先搭一整套复杂测试平台

## 具体测试模式

### 1. Repository Query-Shape Test

适用：

- join 变更
- SQL select / where / order / fallback 变更
- 新增计算字段

目的：

- 保证 SQL 关键片段存在
- 保证不会再出现列歧义
- 保证 fallback 表达式仍然保留

示例：

- `fleet.deleted_at IS NULL`
- `LEFT JOIN "user"`
- `COALESCE(NULLIF("user".nickname, ''), fleet.fc_character_name)`

这类测试特别适合当前仓库，因为：

- 运行快
- 不依赖真实数据库
- 能挡住最近这类 join 回归

### 2. DTO / Schema Mapping Test

适用：

- query 里新增 alias 字段
- 特殊 DTO 字段只用于响应，不落库
- 容易因为 GORM tag 写错导致查得到但映射不到

目的：

- 保证 query alias 真能扫进 DTO
- 保证字段名与 JSON / DBName 对齐

### 3. Service Behavior Test

适用：

- 权限判断
- fallback 规则
- 输入归一化
- 唯一性校验

目的：

- 锁定业务规则
- 避免把策略散落在 handler 或页面后没人保护

### 4. Handler / API Contract Test

适用：

- 改动分页结构
- 改动字段名
- 改动响应 envelope
- 改动重要接口的权限边界

目的：

- 防止后端“改了能编译，但前端 contract 已经变了”

### 5. Frontend Unit Test

适用：

- 纯 helper
- hook 内纯计算
- 筛选参数转换
- fallback 文案选择

目的：

- 用最轻的方式保护前端行为

当前不优先要求：

- 为普通列表页引入重型组件挂载测试
- 为简单文案修改引入端到端浏览器测试

## Bug Fix 的最小回归要求

以后修 bug 时，可直接套下面这张表：

| Bug 根因 | 最少要补什么 |
| --- | --- |
| service 规则错 | 一个 service 测试 |
| repository 查询错 | 一个 repository 回归测试 |
| 响应字段错 | 一个 handler / contract 测试 |
| frontend helper 错 | 一个 frontend unit test |
| 多层共同导致 | 根因层测试 + 另一层构建验证 |

如果一时做不到，应在变更说明里写清楚：

- 为什么现在没补
- 缺的是什么基础设施
- 后续应该补在哪

## Review Checklist

评审 bug fix 时至少问：

1. 这次 bug 的根因在 handler、service、repository 还是 frontend helper？
2. 新测试是否真的锁住了那个根因，而不是只覆盖了表面行为？
3. 如果以后再有人改同一块逻辑，这条测试会不会第一时间报错？
4. 除了 build / lint / typecheck，是否有行为级保护？

## 建议的模块回归清单

下面不是一次性任务清单，而是后续逐步补齐时的优先队列。

### auth-and-characters

- `ProfileComplete()` 与前端资料完成判断保持一致
- QQ / Discord 唯一性
- `/api/v1/me` 返回职权与权限上下文

### operation

- fleet 列表查询与显示 fallback
- fleet 管理权限判断
- PAP 发放前置条件
- auto SRP 模式归一化与触发条件

### administration

- 用户列表 DTO 不再泄漏 legacy `role`
- 用户职权排序与 fallback
- admin 无法操作受保护账号
- auto-role 内置快捷规则与 title mapping 区分

### commerce

- 限购规则
- 订单状态流转
- 钱包交易类型与引用类型映射

### srp

- SRP 申请状态流转
- 舰队 / KM 关联 fallback
- 自动审批与手动审批边界

## 命令建议

验证命令见 `docs/standards/testing-and-verification.md`。

## 文档维护规则

当某个模块开始稳定积累回归测试后，建议同步更新对应 feature doc，至少说明：

- 这个模块当前有哪些关键不变量
- 最近新增的高风险保护点是什么
- 哪一层测试在保护这些不变量

但不要把具体测试文件清单复制到很多文档里重复维护。测试策略以：

- `docs/standards/testing-and-verification.md`
- `docs/guides/testing-guide.md`
- 本文件

为准。
