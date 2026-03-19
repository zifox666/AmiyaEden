---
status: active
doc_type: index
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - AGENTS.md
---

# AmiyaEden 文档索引

## 目的

`docs/` 是仓库内除根目录 `AGENTS.md` 之外的规范化文档树，用来把“工程标准”“当前架构”“API 路由面”“当前功能状态”“未来提案”彻底分开，避免文档互相覆盖、过期后继续误导人或 AI。

## 信任顺序

当多个文件描述同一件事时，按以下顺序判断权威性：

1. 根目录 `AGENTS.md`
2. `docs/standards/*.md`
3. `docs/architecture/*.md`
4. `docs/api/*.md`
5. `docs/features/current/*.md`
6. `docs/guides/*.md`
7. `docs/specs/draft/*.md`

说明：

- 第 7 层只表示“计划 / 草案 / 未完成设计”，不能覆盖当前实现。

## 目录结构

| 路径 | 类型 | 作用 |
| --- | --- | --- |
| `docs/ai/` | agent guide | 给 AI / 自动化代理的阅读顺序、冲突处理、更新要求 |
| `docs/standards/` | standard | 约束性标准，描述“必须 / 不得 / 推荐” |
| `docs/architecture/` | architecture | 只描述当前已经存在的系统结构与运行方式 |
| `docs/api/` | api | 接口约定、响应格式、路由索引 |
| `docs/features/current/` | feature | 当前已落地功能的模块级说明 |
| `docs/guides/` | guide | 过程型指南，例如新增一个 ESI 模块 |
| `docs/reference/` | reference | 离线参考资产，不作为当前实现的 source of truth |
| `docs/specs/draft/` | draft | 提案、未来增强、未完成设计 |
| `docs/templates/` | template | 新建文档时复用的模板 |

## 状态字段

所有新的规范性文档都应包含 front matter，并至少声明：

- `status`: `active` / `draft` / `deprecated` / `template`
- `doc_type`: `standard` / `architecture` / `api` / `feature` / `guide` / `draft` / `template` / `index`
- `owner`
- `last_reviewed`

约定：

- `docs/templates/*` 使用 `status: template`，不要写成 `active`
- `docs/specs/draft/*` 使用 `status: draft`

## 文档更新规则

- 当前行为变化时，优先更新对应的 `docs/architecture`、`docs/api`、`docs/features/current`。
- 新增工程约束时，更新 `AGENTS.md` 或 `docs/standards`，不要把规则写进 feature doc。
- 新增尚未落地的设计时，只放进 `docs/specs/draft`。
- 不要在多个文件里重复维护同一份角色定义、路由表、权限矩阵。
- 不要保留并行的“第二套文档入口”。
- 仓库内允许存在少量模块级 `README.md` 作为局部实现说明，但它们不是 repo-level canonical doc，不能覆盖 `AGENTS.md` 与 `docs/`。

## 推荐阅读顺序

### 对人类开发者

1. `README.md`
2. `AGENTS.md`
3. 本文件
4. 相关架构文档
5. 相关 feature doc
6. 相关 API / guide

### 对 AI Agent

1. `AGENTS.md`
2. `docs/ai/agent-onboarding.md`
3. `docs/architecture/overview.md`
4. 任务对应的标准文档
5. 任务对应的 feature / API 文档
6. 只有在明确做规划工作时才读取 `docs/specs/draft/`
7. 如任务已明确落在某个子目录，再补读该目录下的局部 `README.md`，但只把它当作实现注释而不是规范裁决来源

## 维护原则

`AGENTS.md` 与 `docs/` 是唯一维护中的 repo-level canonical Markdown 文档体系。局部 `README.md` 可以存在，但只能补充子目录实现细节，不能重新建立影子规范树。
