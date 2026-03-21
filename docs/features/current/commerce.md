---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-03-22
source_of_truth:
  - server/internal/router/router.go
  - server/internal/service/shop.go
  - server/internal/service/sys_wallet.go
  - static/src/api/shop.ts
  - static/src/api/sys-wallet.ts
  - static/src/views/shop
  - static/src/views/system/wallet
---

# 商店、钱包

## 当前能力

- 用户侧系统钱包与流水
- 商品浏览、购买、订单、兑换码
- 管理员商品管理、订单审批、兑换码列表
- 管理员系统钱包调整、流水、日志

## 入口

### 前端页面

- `static/src/views/shop/browse`
- `static/src/views/shop/manage`
- `static/src/views/shop/wallet`
- `static/src/views/system/wallet`

### 后端路由

- `/api/v1/shop/*`
- `/api/v1/system/wallet/*`
- `/api/v1/system/shop/*`

## 权限边界

- 用户侧能力要求 `Login`
- `/system/wallet/*`、`/system/shop/*` 默认要求 `admin`

## 关键不变量

- 系统钱包是多个模块共享的资金载体，不能按单一页面理解
- 钱包流水与调整日志属于不同查询面
- 用户侧钱包页面与其后端接口当前都归属 `Shop` 模块
- 商店、兑换码虽然都在 `Shop` 目录下，但用户态与管理态接口是分开的

## 主要代码文件

- `server/internal/service/shop.go`
- `server/internal/service/sys_wallet.go`
- `server/internal/router/router.go`
- `static/src/api/shop.ts`
- `static/src/api/sys-wallet.ts`
- `static/src/views/shop`
- `static/src/views/system/wallet`
