---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-04-03
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/skill_plan.go
  - static/src/api/skill-plan.ts
  - static/src/views/skill-planning
---

# Skill Planning 模块

## 当前能力

- 军团技能计划的列表、详情、创建、编辑、删除
- 创建 / 编辑技能计划时可选一个舰船图标，并在列表中展示
- 粘贴技能文本解析为技能计划条目
- 通过独立顶级菜单“技能规划”进入技能计划页面
- 任意 `Login` 用户都可查看技能计划列表 / 详情；管理按钮只对可维护角色显示
- 用户可在”检查完成度”页面保存自己的人物选择和规划选择，并把人物技能与选中的军团规划逐项比对
- 完成度检查页的缺失技能列表支持通过共享内联复制按钮逐项复制技能名；表头悬浮预览仍保持只读文本
- 规划选择默认包含全部规划，用户可取消选择不需要检查的规划，选择持久化保存

## 入口

### 前端页面

- 页面：
  - `static/src/views/skill-planning/skill-plans`
  - `static/src/views/skill-planning/completion-check`

### 后端路由

- `/api/v1/skill-planning/skill-plans/*`

## 权限边界

- `skill-plans` 的列表、详情查询要求 `Login`
- `skill-plans` 的创建、修改、删除、排序要求 `admin` 或 `senior_fc`（`super_admin` 仍会自动通过 `RequireRole`）
- `check/selection`、`check/plan-selection` 与 `check/run` 要求 `Login`
- 当前技能规划只承载军团技能计划，不与 EVE 人物技能查询页面混用

## 关键不变量

- 技能规划是独立顶级导航，不再挂在舰队行动下
- 前端静态路由与后端 API 当前都归属 `SkillPlanning` 模块
- 页面实现位于 `static/src/views/skill-planning` 目录，修改时保持模块边界一致
- 完成度检查页面与技能计划列表页共享同一 `Login` 读权限边界，避免普通登录用户在检查页看到误导性的访问拒绝提示
- 人物选择和规划选择都会按用户持久化保存，用户再次进入”检查完成度”时不需要重新选择
- 完成度检查只允许比较当前用户自己绑定的人物
- 完成度检查只会比对用户选中的规划，未选中的规划不参与检查

## 主要代码文件

- `server/internal/service/skill_plan.go`
- `server/internal/router/router.go`
- `static/src/api/skill-plan.ts`
- `static/src/router/modules/skill-planning.ts`
- `static/src/views/skill-planning/skill-plans`
- `static/src/views/skill-planning/completion-check`
