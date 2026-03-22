---
status: active
doc_type: standard
owner: engineering
last_reviewed: 2026-03-22
source_of_truth:
  - docs/README.md
  - AGENTS.md
---

# 文档治理标准

## 1. 基本原则

- 一份文档只承担一种职责。
- 一类事实只保留一个 canonical source。
- 当前实现、工程规则、未来提案必须分开存放。
- 不要保留第二套并行文档树。
- repo-level canonical 文档只放在根目录 `AGENTS.md` 与 `docs/`。
- 根目录 `README.md` 可以作为 onboarding / product-facing 入口长期维护，但不裁决工程规则；冲突时以 `AGENTS.md` 与 `docs/` 为准。
- 目录内局部 `README.md` 只可作为实现注释，不能重新定义全局规则、路由面或产品行为。

## 2. 文档类型

| 类型 | 目录 | 内容 |
| --- | --- | --- |
| `standard` | `docs/standards/` | 必须 / 不得 / 推荐 |
| `architecture` | `docs/architecture/` | 当前系统如何工作 |
| `api` | `docs/api/` | 路由、认证、响应约定 |
| `feature` | `docs/features/current/` | 当前模块能力、入口、权限、关键不变量 |
| `guide` | `docs/guides/` | 分步骤操作指南 |
| `reference` | `docs/reference/` | 离线参考资产，不作为当前实现裁决依据 |
| `draft` | `docs/specs/draft/` | 提案、增强、未落地设计 |
| `template` | `docs/templates/` | 创建新文档的模板 |

## 3. Front Matter 要求

所有新的 canonical 文档必须包含 YAML front matter，至少包括：

```yaml
---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/router/router.go
---
```

推荐字段：

- `source_of_truth`
- `supersedes`
- `related_docs`

模板约定：

- `docs/templates/*` 使用 `status: template`
- 模板内应显式声明“这是模板，不代表当前实现”

## 4. 文件命名

- 使用 `kebab-case`
- 文件名描述范围，不描述临时结论
- 不使用 `new-`, `final-`, `latest-`, `v2-` 这类会很快失效的名字

推荐示例：

- `auth-and-permissions.md`
- `runtime-and-startup.md`
- `route-index.md`

## 5. 每类文档的最低结构

### standard

- 适用范围
- 强制规则
- 允许例外
- 检查清单

### architecture

- 目标范围
- 当前实现
- 关键入口文件
- 不变量

### api

- base URL / auth / response
- 路由索引或接口清单
- 权限边界应尽量显式、稳定，避免只靠上下文推断
- 变更同步要求

### feature

- 模块目标
- 当前入口
- 权限边界
- 关键不变量
- 主要代码文件

### reference

- 资产用途
- 文件清单
- 非权威性声明
- 使用限制或刷新说明

### draft

- 背景
- 当前状态
- 提案内容
- 未决问题
- 明确声明“不代表已实现”

## 6. 什么时候新建文档

应该新建：

- 一个新的 feature 模块已经足够独立
- 一个新的标准会被多个模块复用
- 一个提案还未落地，但需要持续讨论

不应该新建：

- 只是另一个视角重复已有路由表
- 只是把同一规则改写一遍
- 只是为了记录一次临时对话结论
- 在子目录放一个 README，然后让它和 `docs/` 并行维护同一套规范

## 7. 更新规则

- 行为变化和文档更新应在同一个改动中完成
- 修改 `status` 或范围时，更新 `last_reviewed`
- 从 `draft` 变为 `active` 时，移动到正确目录，而不是只改标题
- 删除或合并文档时，清理旧引用，避免残留影子入口

## 8. 常见反模式

- 在 README、guide、feature doc 里各写一份角色枚举
- 把根目录 `README.md` 写成另一份与 `AGENTS.md` / `docs/` 并列竞争的工程规范
- 在 current-state 文档中混入“以后准备这样做”
- 建立第二套并行文档树，导致 AI 读取到多份互相矛盾的说明
- 把代码引用写得太泛，导致读者找不到真实入口文件
