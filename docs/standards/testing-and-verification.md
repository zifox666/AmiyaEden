---
status: active
doc_type: standard
owner: engineering
last_reviewed: 2026-03-21
source_of_truth:
  - AGENTS.md
  - server/go.mod
  - static/package.json
---

# 测试与验证标准

## 适用范围

适用于本仓库内所有 backend、frontend、contract、repository、hook、handler、service 变更。

## 基本原则

- 验证分为两类：构建级验证与行为级验证。
- `build / lint / typecheck` 只能证明“当前代码能通过静态检查或编译”，不能替代回归测试。
- 只要改动修复了 bug、修改了契约、重排了复杂逻辑，默认就应考虑回归测试。
- 测试应尽量贴近被修改的真实逻辑，而不是复制一份近似实现到测试里。

## 当前仓库的默认工具

### Backend

- Go 原生测试：`cd server && go test ./...`
- Go 构建校验：`cd server && go build ./...`

### Frontend

- ESLint：`cd static && pnpm lint .`
- TypeScript / Vue 类型校验：`cd static && pnpm exec vue-tsc --noEmit`
- 纯 helper / hook 单元测试：`cd static && pnpm test:unit`

说明：

- 当前 frontend 单元测试能力仍然是轻量级的，适合测试纯函数、纯 helper、轻状态转换逻辑。
- 如果某个 frontend 行为必须依赖完整组件挂载、浏览器环境或复杂 mock，先评估是否值得引入更重的测试基础设施，而不是临时拼装半套方案。
- `static/src/types/import/auto-imports.d.ts` 与 `static/src/types/import/components.d.ts` 属于 frontend 自动导入的声明文件，当前默认作为仓库工件保留，以保证干净检出也能通过 lint / typecheck。
- `static/.auto-import.json` 仅是本地开发辅助文件，不应作为 CI lint 的前置依赖。
- 具体测试落点、命名与编写建议，见 `docs/guides/testing-guide.md`。

## 必须遵守

- 修复 bug 时，只要能合理测试，必须补充或更新回归测试。
- 修改 backend 纯逻辑时，优先在对应 Go package 下添加 `_test.go`。
- 修改 repository 的查询拼接、映射合并、过滤规则、fallback 选择等逻辑时，应添加 Go 测试覆盖关键分支。
- 修改 frontend 纯 helper 或 hook 的纯逻辑时，应优先添加 `pnpm test:unit` 覆盖。
- 修改 API contract 时，至少应有一侧增加行为覆盖，并同时验证 backend 与 frontend。
- 文档里新增了模块级测试命令时，应确保命令能直接运行，而不是只写说明。

## 推荐做法

### Backend

- 纯函数、权限判断、归一化逻辑：普通 Go 单元测试
- handler 响应整形、contract merge 逻辑：优先测试纯 helper 或 handler 边界
- repository SQL 大片常量 / CASE / fallback 逻辑：至少测试 SQL 片段生成或分支选择 helper
- 如果仓库未来具备稳定测试数据库，再补 repository integration tests

### Frontend

- 将复杂状态转换提炼为纯 helper，再测试 helper
- 对 namespace、dedupe、fallback、merge 这类逻辑，优先写输入 / 输出明确的纯测试
- 不要为了一个很小的逻辑回归，先引入庞大的组件测试框架，除非该模式会被反复复用

## 允许例外

以下情况可以不新增测试，但需要在变更说明中显式说明原因：

- 纯文档改动
- 纯格式化改动
- 明显无行为变化的重命名
- 当前仓库缺少必要基础设施，且临时搭建成本远高于变更本身
- 外部依赖或运行环境决定该行为在当前仓库内无法稳定测试

## 提交前检查

- 这次改动有没有修 bug、改 contract、改 fallback、改复杂条件分支
- 如果有，是否新增或更新了回归测试
- 是否运行了该层最小必需验证命令
- 如果跳过测试，是否在说明里写清楚原因
- 新增的测试命令是否已经在本地实际跑过
