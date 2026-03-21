---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-22
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/skill_plan.go
  - static/src/api/skill-plan.ts
  - static/src/views/skill-planning
  - server/internal/model/menu.go
---

# Skill Planning 模块

## 当前能力

- 军团技能计划的列表、详情、创建、编辑、删除
- 创建 / 编辑技能计划时可选一个舰船图标，并在列表中展示
- 粘贴技能文本解析为技能计划条目
- 通过独立顶级菜单“技能规划”进入技能计划页面
- 用户可在“检查完成度”页面保存自己的角色选择，并把角色技能与军团规划逐项比对

## 入口

### 前端页面

- 顶级菜单：`技能规划`
- 页面：
  - `static/src/views/skill-planning/skill-plans`
  - `static/src/views/skill-planning/completion-check`

### 后端路由

- `/api/v1/skill-planning/skill-plans/*`

## 权限边界

- `skill-plans` 的列表、详情、创建、修改、删除要求 `super_admin`、`admin` 或 `fc`
- `check/selection` 与 `check/run` 要求 `Login`
- 当前技能规划只承载军团技能计划，不与 EVE 角色技能查询页面混用

## 关键不变量

- 技能规划是独立顶级导航，不再挂在舰队行动下
- 前端菜单、静态路由与后端 API 当前都归属 `SkillPlanning` 模块
- 页面实现位于 `static/src/views/skill-planning` 目录，修改时保持模块边界一致
- 角色选择会按用户持久化保存，用户再次进入“检查完成度”时不需要重新选择
- 完成度检查只允许比较当前用户自己绑定的角色

## 主要代码文件

- `server/internal/service/skill_plan.go`
- `server/internal/router/router.go`
- `server/internal/model/menu.go`
- `static/src/api/skill-plan.ts`
- `static/src/router/modules/skill-planning.ts`
- `static/src/views/skill-planning/skill-plans`
- `static/src/views/skill-planning/completion-check`
