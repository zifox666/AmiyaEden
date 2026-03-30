---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-31
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/badge.go
  - server/internal/repository/srp.go
  - server/internal/repository/welfare.go
  - server/internal/repository/shop.go
  - static/src/api/badge.ts
  - static/src/store/modules/badge.ts
  - static/src/store/modules/badge.helpers.ts
  - static/src/router/guards/beforeEach.ts
---

# 导航徽章计数

## 当前能力

- 顶部或侧边导航菜单支持基于后端实时计算结果显示数字徽章
- 徽章数据在登录后的动态路由初始化阶段获取一次
- 返回结果只包含当前用户有权查看且值大于 `0` 的字段
- 前端会将子菜单徽章自动汇总到父菜单，无需为父菜单单独配置后端字段

## 后端入口

- `GET /api/v1/badge-counts`
- 权限：`Login`

响应示例：

```json
{
  "welfare_eligible": 2,
  "srp_pending": 5,
  "welfare_pending": 3,
  "order_pending": 1
}
```

## 字段定义

| 字段 | 可见范围 | 含义 |
| --- | --- | --- |
| `welfare_eligible` | 任意已登录产品用户 | 当前用户现在可申请的福利数量；`per_character` 福利按福利项计数，不按人物条目重复计数 |
| `srp_pending` | `super_admin` / `admin` / `srp` / `fc` | `review_status IN ('submitted', 'approved') AND payout_status = 'notpaid'` 的 SRP 申请数 |
| `welfare_pending` | `super_admin` / `admin` / `welfare` | `status = 'requested'` 的福利申请数 |
| `order_pending` | `super_admin` / `admin` / `welfare` | `status = 'requested'` 的商店订单数 |

## 前端映射

| 路由名 | 计数字段 |
| --- | --- |
| `WelfareMy` | `welfare_eligible` |
| `WelfareApproval` | `welfare_pending` |
| `SrpManage` | `srp_pending` |
| `ShopOrderManage` | `order_pending` |

## 徽章规则

- 叶子菜单：若存在映射字段且值大于 `0`，显示该数字
- 叶子菜单：若字段不存在、字段未返回或值为 `0`，清除徽章
- 父菜单：显示所有子菜单数字徽章之和
- 父菜单：若子菜单总和为 `0`，清除徽章
- 数字徽章不影响现有 `showBadge` 点状徽章能力

## 刷新行为

- 徽章只在登录后的动态路由初始化期间请求一次
- 页面运行期间不轮询、不推送、不增量刷新
- 底层业务数据变化后，用户需要重新刷新页面或重新进入初始化流程才会看到新计数

## 扩展约束

- 新增徽章字段时，必须同步更新后端服务字段逻辑、前端 API 类型、前端路由映射与本文件
- 新计数应优先使用仓储层 `COUNT(*)` 查询；仅在必须复用现有资格判定逻辑时才允许走更重的业务路径
- 权限控制必须以后端字段省略为准，前端仅负责展示