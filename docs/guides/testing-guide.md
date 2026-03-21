---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-21
source_of_truth:
  - AGENTS.md
  - docs/standards/testing-and-verification.md
  - docs/guides/regression-test-plan.md
  - server/go.mod
  - static/package.json
---

# 测试实操指南

## 目的

本文件补充 `docs/standards/testing-and-verification.md` 的“原则”，回答更实际的问题：

- 测试通常写在哪
- 什么时候该写 unit test，什么时候只做构建验证
- 这个仓库目前适合怎样写测试

如果你要规划“接下来该优先补哪些回归测试、分几阶段推进”，看 `docs/guides/regression-test-plan.md`。

## 当前测试落点

### Backend

- 包级测试与被测代码放在同目录，文件名使用 `*_test.go`
- 纯逻辑、权限判断、归一化、映射合并，优先在对应 package 下直接写测试
- 目前已有的重点测试主要在：
  - `server/internal/service/`
  - `server/internal/handler/`
  - `server/internal/repository/`

### Frontend

- 纯 helper / hook 测试优先放在被测文件旁边
- 当前轻量测试入口为 `cd static && pnpm test:unit`
- 当前更适合测试：
  - 纯函数
  - namespace / dedupe / fallback / merge 逻辑
  - 不依赖完整 DOM 环境的状态转换

## 当前仓库更推荐测什么

### Backend 推荐

- service 层权限判断
- 输入归一化
- 时间范围、筛选参数、枚举映射
- repository 中的分支选择、merge、fallback
- handler 中可抽离的纯 helper 或 contract merge 逻辑

### Frontend 推荐

- hook 内的纯 helper
- API 响应 merge 逻辑
- 名称解析、缓存 key、去重逻辑
- 表格列定义或筛选参数转换中的纯逻辑

## 当前仓库暂不优先投入什么

- 为了一个很小的逻辑改动，引入沉重的组件测试基础设施
- 为 repository 临时搭建一套复杂测试数据库，只为覆盖一个低风险分支
- 只验证实现细节、却不验证外部行为的测试

## 测试命名建议

### Go

- 推荐：`TestFunctionNameScenario`
- 例如：
  - `TestParseEFTHeader`
  - `TestNormalizeSkillPlanName`
  - `TestMergeGetNamesNamespacesPreservesNamespacesAndFlatFirstWins`

### Frontend

- 推荐：按行为命名
- 例如：
  - `mergeNamesResponse keeps namespace-specific values`
  - `buildPendingRequest keeps type and solar_system ids separate`

## 测试编写建议

1. 优先测公开行为或稳定 helper 行为，不要复制一份生产实现到测试里。
2. 用最小输入覆盖最关键分支，不追求一次把所有组合打满。
3. 先锁住 bug，再锁住 contract。
4. 如果改动跨后端与前端，至少在一侧补行为测试，另一侧做构建验证。
5. 如果本次改动发现仓库缺失测试入口，优先补一个轻量、可复用的入口，而不是把验证步骤藏在聊天记录里。

## 常用命令

### Backend

```bash
cd server && go test ./...
cd server && go build ./...
```

### Frontend

```bash
cd static && pnpm lint .
cd static && pnpm exec vue-tsc --noEmit
cd static && pnpm test:unit
```

## 什么时候可以只做验证不补测试

以下情况通常可以接受，但应在变更说明里说明原因：

- 文档改动
- 纯格式化改动
- 明显无行为变化的重命名
- 当前仓库没有合理测试基础设施，且临时搭建成本明显高于变更本身

## 评审时可以快速问自己的问题

- 这次是否修了 bug
- 是否改了 fallback、merge、筛选、权限边界
- 是否改了前后端 contract
- 如果答案是“是”，有没有新增或更新回归测试
