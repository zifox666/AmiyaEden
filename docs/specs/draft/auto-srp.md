---
status: draft
doc_type: draft
owner: engineering
last_reviewed: 2026-03-20
source_of_truth:
  - server/internal/service/auto_srp.go
  - server/internal/model/fleet.go
---

# 自动 SRP 后续设计

## 当前状态

自动 SRP 已经不是纯想法，当前仓库已存在：

- `fleet.auto_srp_mode`
- 前端舰队页的模式选择
- 后台 `AutoSrpService`
- 舰队 PAP 后触发的自动处理钩子

本文件只记录“当前实现之外仍想增强的部分”。

## 仍在草案中的目标

- 更完整的可替换装备规则
- 更细致的装备槽位归一化与数量判断
- 审核备注中输出更完整的不符原因
- 更明确的舰队配置价格与全局价格回退策略说明

## 约束

- 本文件不能覆盖 `docs/features/current/operation.md`
- 如果未来这些规则真正落地，应把已实现部分迁回 current feature doc，并删除对应草案段落
