# AmiyaEden API 接口文档

> **Base URL**：`/api/v1`
>
> **认证方式**：需要登录的接口在 `Authorization` 请求头中携带 JWT Token：`Bearer <token>`
>
> **统一响应格式**：
>
> ```json
> { "code": 0, "msg": "ok", "data": {} }
> ```
>
> 分页响应：
>
> ```json
> {
>   "code": 0,
>   "msg": "ok",
>   "data": { "list": [], "total": 0, "page": 1, "pageSize": 20 }
> }
> ```

---

## 目录

- [1. EVE SSO 认证](#1-eve-sso-认证)
- [2. SDE 数据查询](#2-sde-数据查询)
- [3. 个人信息](#3-个人信息)
- [4. 通知](#4-通知)
- [5. 菜单](#5-菜单)
- [6. 舰队](#6-舰队)
- [7. 角色信息 & NPC 刷怪](#7-角色信息--npc-刷怪)
- [8. 系统钱包（用户端）](#8-系统钱包用户端)
- [9. 商店（用户端）](#9-商店用户端)
- [10. SRP 补损](#10-srp-补损)
- [11. ESI 刷新队列](#11-esi-刷新队列)
- [12. 系统管理（Admin）](#12-系统管理admin)

---

## 1. EVE SSO 认证

### 1.1 发起 EVE SSO 登录

```
GET /sso/eve/login
```

| 参数       | 类型         | 必填 | 说明                                 |
| ---------- | ------------ | ---- | ------------------------------------ |
| `redirect` | query string | 否   | 登录成功后前端回调地址               |
| `scopes`   | query string | 否   | 额外申请的 ESI scope，多个以逗号分隔 |

**响应**：`{ "url": "https://login.eveonline.com/..." }`

---

### 1.2 EVE SSO OAuth 回调

```
GET /sso/eve/callback
```

| 参数    | 类型         | 必填 | 说明             |
| ------- | ------------ | ---- | ---------------- |
| `code`  | query string | 是   | EVE 返回的授权码 |
| `state` | query string | 是   | CSRF state 值    |

**响应（首次登录/无 redirect）**：

```json
{ "token": "jwt_token", "user": {}, "character": {} }
```

若有 `redirect` 则 302 重定向到 `redirect?token=<jwt>`

---

### 1.3 获取 ESI Scope 列表

```
GET /sso/eve/scopes
```

> 需要 JWT

**响应**：已注册的 ESI Scope 字符串数组

---

### 1.4 获取当前用户绑定的角色列表

```
GET /sso/eve/characters
```

> 需要 JWT

**响应**：`EveCharacter[]` 数组

---

### 1.5 绑定新角色

```
GET /sso/eve/bind
```

> 需要 JWT

| 参数       | 类型         | 必填 | 说明                 |
| ---------- | ------------ | ---- | -------------------- |
| `redirect` | query string | 否   | 绑定成功后的回调地址 |
| `scopes`   | query string | 否   | 额外申请的 ESI scope |

**响应**：`{ "url": "..." }`

---

### 1.6 设置主角色

```
PUT /sso/eve/primary/:character_id
```

> 需要 JWT

| 参数           | 类型 | 必填 | 说明    |
| -------------- | ---- | ---- | ------- |
| `character_id` | path | 是   | 角色 ID |

---

### 1.7 解绑角色

```
DELETE /sso/eve/characters/:character_id
```

> 需要 JWT

| 参数           | 类型 | 必填 | 说明            |
| -------------- | ---- | ---- | --------------- |
| `character_id` | path | 是   | 要解绑的角色 ID |

---

## 2. SDE 数据查询

### 2.1 获取 SDE 版本

```
GET /sde/version
```

**响应**：当前已导入的 SDE 版本信息

---

### 2.2 批量查询物品信息

```
POST /sde/types
```

**请求体**：

```json
{
  "type_ids": [34, 35, 36],
  "published": true,
  "language_id": "zh"
}
```

| 字段          | 类型     | 必填 | 说明                                                                |
| ------------- | -------- | ---- | ------------------------------------------------------------------- |
| `type_ids`    | `int[]`  | 是   | 物品 TypeID 数组                                                    |
| `published`   | `bool`   | 否   | 过滤是否上架                                                        |
| `language_id` | `string` | 否   | 语言代码（默认 `en`，支持 `zh`/`en`/`de`/`ja`/`ko`/`ru`/`fr`/`es`） |

---

### 2.3 批量查询 ID→Name 映射

```
POST /sde/names
```

**请求体**：

```json
{
  "language": "zh",
  "ids": {
    "type": [34, 35],
    "region": [10000002]
  },
  "esi": [123456789]
}
```

| 字段       | 类型               | 必填 | 说明                                                                                                            |
| ---------- | ------------------ | ---- | --------------------------------------------------------------------------------------------------------------- |
| `language` | `string`           | 否   | 语言代码，也可通过 `Accept-Language` header 或 `language` cookie 指定                                           |
| `ids`      | `map[string][]int` | 否   | key 取值：`type`/`group`/`category`/`region`/`constellation`/`solar_system`/`market_group`/`tech`/`description` |
| `esi`      | `int64[]`          | 否   | character/corporation/alliance ID，调用 ESI universe/names 查询                                                 |

**响应**：`{ "34": "Tritanium", "10000002": "The Forge" }`

---

### 2.4 模糊搜索物品/成员

```
POST /sde/search
```

**请求体**：

```json
{
  "keyword": "Raven",
  "language": "zh",
  "category_ids": [6],
  "exclude_category_ids": [],
  "limit": 20,
  "search_member": false
}
```

| 字段                   | 类型     | 必填 | 说明                    |
| ---------------------- | -------- | ---- | ----------------------- |
| `keyword`              | `string` | 是   | 搜索关键词              |
| `language`             | `string` | 否   | 语言代码（默认 `en`）   |
| `category_ids`         | `int[]`  | 否   | 限定分类 ID             |
| `exclude_category_ids` | `int[]`  | 否   | 排除分类 ID             |
| `limit`                | `int`    | 否   | 最大返回数量（默认 20） |
| `search_member`        | `bool`   | 否   | 是否同时搜索成员名称    |

---

## 3. 个人信息

### 3.1 获取当前用户信息

```
GET /me
```

> 需要 JWT

**响应**：

```json
{
  "user": {},
  "characters": [],
  "roles": ["admin"],
  "permissions": ["srp:review"]
}
```

---

### 3.2 获取 Dashboard 数据

```
POST /dashboard
```

> 需要 JWT

---

## 4. 通知

### 4.1 通知列表

```
POST /notification/list
```

> 需要 JWT

**请求体（分页）**：

```json
{ "current": 1, "size": 20 }
```

---

### 4.2 获取未读通知数

```
POST /notification/unread-count
```

> 需要 JWT

---

### 4.3 标记通知已读

```
POST /notification/read
```

> 需要 JWT

**请求体**：

```json
{ "ids": [1, 2, 3] }
```

---

### 4.4 全部标记已读

```
POST /notification/read-all
```

> 需要 JWT

---

## 5. 菜单

### 5.1 获取当前用户可用菜单

```
GET /menu/list
```

> 需要 JWT

**响应**：菜单列表（按当前用户权限过滤）

---

## 6. 舰队

> 基础路径：`/operation/fleets`，所有接口需要 JWT

### 6.1 创建舰队

```
POST /operation/fleets
```

**请求体**：`CreateFleetRequest`

---

### 6.2 舰队列表

```
GET /operation/fleets
```

| 参数         | 类型  | 必填 | 说明                |
| ------------ | ----- | ---- | ------------------- |
| `current`    | query | 否   | 页码（默认 1）      |
| `size`       | query | 否   | 每页条数（默认 20） |
| `importance` | query | 否   | 按重要性过滤        |
| `fc_user_id` | query | 否   | 按 FC 用户 ID 过滤  |

---

### 6.3 舰队详情

```
GET /operation/fleets/:id
```

---

### 6.4 更新舰队

```
PUT /operation/fleets/:id
```

**请求体**：`UpdateFleetRequest`

---

### 6.5 删除舰队

```
DELETE /operation/fleets/:id
```

---

### 6.6 刷新舰队 ESI Fleet ID

```
POST /operation/fleets/:id/refresh-esi
```

---

### 6.7 获取舰队成员列表

```
GET /operation/fleets/:id/members
```

---

### 6.8 同步 ESI 成员

```
POST /operation/fleets/:id/members/sync
```

从 ESI 拉取当前舰队成员并同步到数据库

---

### 6.9 发放 PAP

```
POST /operation/fleets/:id/pap
```

---

### 6.10 获取舰队 PAP 记录

```
GET /operation/fleets/:id/pap
```

---

### 6.11 获取我的 PAP 记录

```
GET /operation/fleets/pap/me
```

---

### 6.12 查询我的联盟 PAP 数据

```
GET /operation/fleets/pap/alliance
```

| 参数    | 类型  | 必填 | 说明             |
| ------- | ----- | ---- | ---------------- |
| `year`  | query | 否   | 年份（默认当年） |
| `month` | query | 否   | 月份（默认当月） |

---

### 6.13 创建邀请链接

```
POST /operation/fleets/:id/invites
```

---

### 6.14 获取邀请链接列表

```
GET /operation/fleets/:id/invites
```

---

### 6.15 禁用邀请链接

```
DELETE /operation/fleets/invites/:invite_id
```

---

### 6.16 通过邀请码加入舰队

```
POST /operation/fleets/join
```

**请求体**：

```json
{
  "code": "invite_code",
  "character_id": 123456789
}
```

---

### 6.17 获取角色所在的 ESI 舰队

```
GET /operation/fleets/esi/:character_id
```

---

## 7. 角色信息 & NPC 刷怪

> 基础路径：`/info`，所有接口需要 JWT

### 7.1 获取角色钱包流水

```
POST /info/wallet
```

---

### 7.2 获取角色技能

```
POST /info/skills
```

---

### 7.3 获取角色舰船

```
POST /info/ships
```

---

### 7.4 获取角色 NPC 刷怪报表

```
POST /info/npc-kills
```

**请求体**：`NpcKillRequest`

---

### 7.5 获取所有角色汇总刷怪报表

```
POST /info/npc-kills/all
```

**请求体**：`NpcKillAllRequest`

---

## 8. 系统钱包（用户端）

> 基础路径：`/operation/wallet`，需要 JWT

### 8.1 查询我的钱包余额

```
POST /operation/wallet/my
```

---

### 8.2 查询我的交易记录

```
POST /operation/wallet/my/transactions
```

---

## 9. 商店（用户端）

> 基础路径：`/shop`，需要 JWT

### 9.1 获取商品列表

```
POST /shop/products
```

**请求体**：

```json
{
  "current": 1,
  "size": 20,
  "type": "normal"
}
```

| 字段      | 类型     | 必填 | 说明                          |
| --------- | -------- | ---- | ----------------------------- |
| `current` | `int`    | 否   | 页码                          |
| `size`    | `int`    | 否   | 每页大小                      |
| `type`    | `string` | 否   | 商品类型：`normal` / `redeem` |

---

### 9.2 获取商品详情

```
POST /shop/product/detail
```

**请求体**：

```json
{ "product_id": 1 }
```

---

### 9.3 购买商品

```
POST /shop/buy
```

**请求体**：`BuyRequest`（包含 `product_id`、`quantity` 等）

---

### 9.4 获取我的订单

```
POST /shop/orders
```

**请求体**：

```json
{
  "current": 1,
  "size": 20,
  "status": "pending"
}
```

---

### 9.5 获取我的兑换码

```
POST /shop/redeem/list
```

**请求体**：

```json
{ "current": 1, "size": 20 }
```

---

## 10. SRP 补损

> 基础路径：`/srp`，需要 JWT

### 10.1 舰船价格表（公开）

```
GET /srp/prices
```

| 参数      | 类型  | 必填 | 说明           |
| --------- | ----- | ---- | -------------- |
| `keyword` | query | 否   | 舰船名称关键词 |

---

### 10.2 添加/更新舰船价格

```
POST /srp/prices
```

> 需要权限：`srp:price:add`

**请求体**：`UpsertShipPriceRequest`

---

### 10.3 删除舰船价格

```
DELETE /srp/prices/:id
```

> 需要权限：`srp:price:delete`

---

### 10.4 提交补损申请

```
POST /srp/applications
```

**请求体**：`SubmitApplicationRequest`（包含 killmail 链接等）

---

### 10.5 查询我的补损申请列表

```
GET /srp/applications/me
```

| 参数      | 类型  | 必填 | 说明                |
| --------- | ----- | ---- | ------------------- |
| `current` | query | 否   | 页码（默认 1）      |
| `size`    | query | 否   | 每页大小（默认 20） |

---

### 10.6 查询我的 Killmail 列表

```
GET /srp/killmails/me
```

| 参数           | 类型  | 必填 | 说明        |
| -------------- | ----- | ---- | ----------- |
| `character_id` | query | 否   | 指定角色 ID |

---

### 10.7 查询舰队 Killmail 列表

```
GET /srp/killmails/fleet/:fleet_id
```

---

### 10.8 获取 Killmail 详情

```
POST /srp/killmails/detail
```

**请求体**：`KillmailDetailRequest`（包含 killmail_id 和 hash）

---

### 10.9 在 EVE 客户端打开角色信息窗口

```
POST /srp/open-info-window
```

**请求体**：`OpenInfoWindowRequest`

---

### 10.10 查询补损申请列表（管理审核）

```
GET /srp/applications
```

> 需要权限：`srp:review`

| 参数            | 类型  | 必填 | 说明         |
| --------------- | ----- | ---- | ------------ |
| `current`       | query | 否   | 页码         |
| `size`          | query | 否   | 每页大小     |
| `review_status` | query | 否   | 审核状态过滤 |
| `payout_status` | query | 否   | 发放状态过滤 |
| `fleet_id`      | query | 否   | 按舰队过滤   |
| `character_id`  | query | 否   | 按角色过滤   |

---

### 10.11 获取补损申请详情（管理）

```
GET /srp/applications/:id
```

> 需要权限：`srp:review`

---

### 10.12 审核补损申请

```
PUT /srp/applications/:id/review
```

> 需要权限：`srp:review`

**请求体**：`ReviewApplicationRequest`（包含审核结果和备注）

---

### 10.13 发放补损

```
PUT /srp/applications/:id/payout
```

> 需要权限：`srp:review`

**请求体**：`SrpPayoutRequest`

---

## 11. ESI 刷新队列

> 基础路径：`/esi/refresh`，需要 JWT + **Admin 角色**

### 11.1 获取任务列表

```
GET /esi/refresh/tasks
```

### 11.2 获取任务执行状态

```
GET /esi/refresh/statuses
```

### 11.3 运行任务

```
POST /esi/refresh/run
```

### 11.4 按名称运行指定任务

```
POST /esi/refresh/run-task
```

### 11.5 运行所有任务

```
POST /esi/refresh/run-all
```

---

## 12. 系统管理（Admin）

> 基础路径：`/system`，需要 JWT + **Admin 角色**

---

### 12.1 NPC 刷怪（公司级）

#### 获取全公司成员刷怪报表

```
POST /system/npc-kills
```

**请求体**：`NpcKillCorpRequest`

---

### 12.2 联盟 PAP 管理

#### 查询所有成员月度 PAP 汇总

```
GET /system/pap
```

| 参数      | 类型  | 必填 | 说明                |
| --------- | ----- | ---- | ------------------- |
| `year`    | query | 否   | 年份（默认当年）    |
| `month`   | query | 否   | 月份（默认当月）    |
| `current` | query | 否   | 页码（默认 1）      |
| `size`    | query | 否   | 每页大小（默认 20） |

#### 手动触发 PAP 数据拉取

```
POST /system/pap/fetch
```

| 参数    | 类型  | 必填 | 说明 |
| ------- | ----- | ---- | ---- |
| `year`  | body | 否   | 年份 |
| `month` | body | 否   | 月份 |

#### 从SEAT或表格导入PAP数据

```
POST /system/pap/import
```

| 参数    | 类型  | 必填 | 说明 |
| ------- | ----- | ---- | ---- |
| `year`  | body | 是   | 年份 |
| `month` | body | 是   | 月份 |
| `data` | object | 是   | 角色 PAP 信息 |

`data` 定义：

```json
{
  "primary_character_name": "角色名",
  "monthly_pap": 100.0,
  "calculated_at": "2024-01-01 00:00:00"
}
```

#### 查询 PAP 兑换配置

```
GET /system/pap/config
```

#### 更新 PAP 兑换配置

```
PUT /system/pap/config
```

**请求体**：`SetExchangeConfigRequest`

#### 月度归档结算

```
POST /system/pap/settle
```

**请求体**：

```json
{
  "year": 2024,
  "month": 12,
  "wallet_convert": true
}
```

---

### 12.3 菜单管理

| 方法     | 路径                | 说明           |
| -------- | ------------------- | -------------- |
| `GET`    | `/system/menu/tree` | 获取完整菜单树 |
| `POST`   | `/system/menu`      | 创建菜单       |
| `PUT`    | `/system/menu/:id`  | 更新菜单       |
| `DELETE` | `/system/menu/:id`  | 删除菜单       |

---

### 12.4 角色管理

| 方法     | 路径                     | 说明             |
| -------- | ------------------------ | ---------------- |
| `GET`    | `/system/role`           | 角色列表（分页） |
| `GET`    | `/system/role/all`       | 获取所有角色     |
| `GET`    | `/system/role/:id`       | 角色详情         |
| `POST`   | `/system/role`           | 创建角色         |
| `PUT`    | `/system/role/:id`       | 更新角色         |
| `DELETE` | `/system/role/:id`       | 删除角色         |
| `GET`    | `/system/role/:id/menus` | 获取角色菜单权限 |
| `PUT`    | `/system/role/:id/menus` | 设置角色菜单权限 |

---

### 12.5 用户管理

| 方法     | 路径                           | 说明                     |
| -------- | ------------------------------ | ------------------------ |
| `GET`    | `/system/user`                 | 用户列表（分页）         |
| `GET`    | `/system/user/:id`             | 用户详情                 |
| `PUT`    | `/system/user/:id`             | 更新用户                 |
| `DELETE` | `/system/user/:id`             | 删除用户                 |
| `GET`    | `/system/user/:id/roles`       | 获取用户角色列表         |
| `PUT`    | `/system/user/:id/roles`       | 设置用户角色             |
| `POST`   | `/system/user/:id/impersonate` | 模拟登录（仅超级管理员） |

---

### 12.6 系统钱包管理

| 方法   | 路径                          | 说明             |
| ------ | ----------------------------- | ---------------- |
| `POST` | `/system/wallet/list`         | 所有用户钱包列表 |
| `POST` | `/system/wallet/detail`       | 指定用户钱包详情 |
| `POST` | `/system/wallet/adjust`       | 手动调整钱包余额 |
| `POST` | `/system/wallet/transactions` | 交易记录查询     |
| `POST` | `/system/wallet/logs`         | 操作日志查询     |

---

### 12.7 商店管理（商品）

| 方法   | 路径                          | 说明     |
| ------ | ----------------------------- | -------- |
| `POST` | `/system/shop/product/list`   | 商品列表 |
| `POST` | `/system/shop/product/add`    | 创建商品 |
| `POST` | `/system/shop/product/edit`   | 更新商品 |
| `POST` | `/system/shop/product/delete` | 删除商品 |

**创建商品请求体**：

```json
{
  "name": "商品名称",
  "description": "描述",
  "image": "图片URL",
  "price": 100.0,
  "stock": -1,
  "max_per_user": 0,
  "type": "normal",
  "need_approval": false,
  "status": 1,
  "sort_order": 0
}
```

> `type`：`normal`（普通商品）/ `redeem`（兑换码商品）
> `stock`：`-1` 表示无限库存

---

### 12.8 商店管理（订单）

| 方法   | 路径                         | 说明         |
| ------ | ---------------------------- | ------------ |
| `POST` | `/system/shop/order/list`    | 订单列表     |
| `POST` | `/system/shop/order/approve` | 审核通过订单 |
| `POST` | `/system/shop/order/reject`  | 拒绝订单     |

---

### 12.9 商店管理（兑换码）

| 方法   | 路径                       | 说明       |
| ------ | -------------------------- | ---------- |
| `POST` | `/system/shop/redeem/list` | 兑换码列表 |

---

### 12.10 自动权限映射管理

| 方法     | 路径                                       | 说明                      |
| -------- | ------------------------------------------ | ------------------------- |
| `GET`    | `/system/auto-role/esi-roles`              | 获取 ESI 军团角色列表     |
| `GET`    | `/system/auto-role/esi-role-mappings`      | ESI 角色→系统角色映射列表 |
| `POST`   | `/system/auto-role/esi-role-mappings`      | 创建 ESI 角色映射         |
| `DELETE` | `/system/auto-role/esi-role-mappings/:id`  | 删除 ESI 角色映射         |
| `GET`    | `/system/auto-role/corp-titles`            | 获取军团头衔列表          |
| `GET`    | `/system/auto-role/esi-title-mappings`     | 头衔→系统角色映射列表     |
| `POST`   | `/system/auto-role/esi-title-mappings`     | 创建头衔映射              |
| `DELETE` | `/system/auto-role/esi-title-mappings/:id` | 删除头衔映射              |
| `POST`   | `/system/auto-role/sync`                   | 手动触发自动角色同步      |

---

## 错误码说明

| code  | 含义                |
| ----- | ------------------- |
| `0`   | 成功                |
| `400` | 参数错误            |
| `401` | 未登录 / Token 无效 |
| `403` | 权限不足            |
| `404` | 资源不存在          |
| `500` | 业务错误            |
