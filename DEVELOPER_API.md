# AmiyaEden API 接口文档

> **Base URL**：`/api/v1`
>
> **认证方式**：需要登录的接口在 `Authorization` 请求头中携带 JWT Token：`Bearer <token>`
>
> **统一响应格式**：
>
> ```json
> { "code": 200, "msg": "success", "data": {} }
> ```
>
> 分页响应：
>
> ```json
> {
>   "code": 200,
>   "msg": "success",
>   "data": { "list": [], "total": 0, "page": 1, "pageSize": 20 }
> }
> ```

> **AT 模式说明**：第 14 节是当前运行时允许名单。前文旧舰队、商店、SRP、钱包、语音等章节仅保留为历史契约，AT 模式不注册对应路由；源码和历史数据没有删除。

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
- [13. 语音中心 / Mumble](#13-语音中心--mumble)

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

若有 `redirect` 则 302 回跳并附加 `token`。Hash Router 地址会把参数加入 `#` 后的查询串，例如 `/#/auth/callback?redirect=%2F&token=<jwt>`，不会把令牌拼入 `redirect` 参数值。

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

## 13. 语音中心 / Mumble

### 13.1 获取当前用户 Mumble 账号

```
GET /voice/mumble
```

> 需要 JWT。若当前用户还没有 Mumble 账号，后端会自动创建。Mumble 登录账号固定为 AmiyaEden 用户 ID，密码为 16 位随机字符串。

**响应**：

```json
{
  "config": {
    "enabled": true,
    "url": "mumble.example.com",
    "port": 64738,
    "server_name": "AmiyaEden Mumble",
    "auth_secret_set": true
  },
  "account": {
    "user_id": 1,
    "username": "1",
    "password": "Random16Chars",
    "display_name": "Character Name",
    "groups": ["amiya_user"],
    "quick_url": "mumble://1:Random16Chars@mumble.example.com:64738/?version=1.2.0&title=AmiyaEden%20Mumble"
  }
}
```

### 13.2 重置当前用户 Mumble 密码

```
POST /voice/mumble/reset-password
```

> 需要 JWT。用户不能自定义密码，只能重置为新的 16 位随机密码。

**响应**：`account` 对象，字段同 13.1。

### 13.3 Mumble 配置（Admin）

| 方法  | 路径                    | 说明                |
| ----- | ----------------------- | ------------------- |
| `GET` | `/system/mumble-config` | 获取 Mumble 配置    |
| `PUT` | `/system/mumble-config` | 更新 Mumble 配置    |

**更新请求体**：

```json
{
  "enabled": true,
  "url": "mumble.example.com",
  "port": 64738,
  "server_name": "AmiyaEden Mumble",
  "auth_secret": "shared-secret-for-ice-authenticator"
}
```

### 13.4 Mumble 角色 Group 映射（Admin）

> AmiyaEden 只负责把平台角色映射为 Mumble groups，并在 ICE 认证时返回 groups。频道 ACL 仍由管理员在 Mumble 管理页面维护。

| 方法  | 路径                         | 说明                    |
| ----- | ---------------------------- | ----------------------- |
| `GET` | `/system/mumble-role-groups` | 获取角色与 group 映射   |
| `PUT` | `/system/mumble-role-groups` | 批量更新角色 group 映射 |

**响应**：

```json
[
  {
    "role_id": 1,
    "role_code": "user",
    "role_name": "用户",
    "group_name": "amiya_user",
    "enabled": true
  }
]
```

**更新请求体**：

```json
{
  "mappings": [
    {
      "role_code": "user",
      "group_name": "amiya_user",
      "enabled": true
    }
  ]
}
```

### 13.5 Mumble ICE Authenticator 内部认证桥

```
POST /voice/mumble/ice-auth
```

> 不需要 JWT，但必须携带 `X-Mumble-Auth-Secret`。该接口仅供 `server/cmd/mumble-ice-auth` 调用，不应暴露到公网。

**请求体**：

```json
{
  "username": "1",
  "password": "Random16Chars"
}
```

**响应**：

```json
{
  "allowed": true,
  "user_id": 1,
  "display_name": "Character Name",
  "groups": ["amiya_user"]
}
```

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

---

## 14. AT 赛用系统 API

除公开报名上下文、EVE/SeAT OAuth 和 `/health` 外，以下接口均要求 JWT；响应继续使用本文统一 `{code,msg,data}` 封装。细粒度所有权由 Service 校验。

### 14.1 健康与个人资料

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/health` | 检查 DATA、SDE、Redis；失败返回 HTTP 503 |
| GET | `/api/v1/profile` | QQ、昵称、比赛服身份、双语言和主角色来源 |
| PUT | `/api/v1/profile` | 更新资料；比赛服 ID 必须为唯一正整数 |

资料字段为 `qq`、`nickname`、`tournament_character_name`、`tournament_character_id`、`ui_language`、`game_language`。主角色来源只读返回 `main_character_source=seat|eve|none`。个人信息页面使用根级路由 `/profile`；访客也可访问自己的技能、配置、克隆和植入体信息。

### 14.2 报名

| 方法 | 路径 | 权限/说明 |
| --- | --- | --- |
| GET | `/api/v1/recruitment/public/:token` | 公开读取有效链接的问卷快照 |
| POST | `/api/v1/recruitment/public/:token/applications` | 登录后提交答案并固化技能/Public 快照 |
| GET | `/api/v1/recruitment/my-applications` | 当前用户查看自己的申请 |
| PUT | `/api/v1/recruitment/applications/:id/withdraw` | 当前用户撤回自己的申请 |
| GET/POST | `/api/v1/recruitment/questionnaires` | 招募官；列表/创建 |
| GET/PUT | `/api/v1/recruitment/questionnaires/:id` | 签发者或管理员；`publish=true` 发布 |
| POST | `/api/v1/recruitment/images` | 登录用户上传问卷图片；5 MiB，校验实际 JPEG/PNG/GIF/WebP 内容 |
| GET/POST | `/api/v1/recruitment/links` | 签发者或管理员；列表/创建 |
| PUT | `/api/v1/recruitment/links/:id/active` | 启停链接 |
| GET | `/api/v1/recruitment/applications` | 仅自己的链接；管理员全部 |
| GET | `/api/v1/recruitment/applications/:id` | 答案、技能/Public 和事件，不含配置/克隆/植入体 |
| PUT | `/api/v1/recruitment/applications/:id/review` | 状态审核和原因；通过时原子移除 `guest` 并授予 `user` 普通成员角色 |
| POST | `/api/v1/recruitment/privacy-cleanup` | 仅管理员；原因必填 |

可视化设计器提供 `text|textarea|radio|checkbox|select` 五种常用题型；后端继续兼容历史 `number|date|image` 快照。`field_key` 由系统生成且不向招募官暴露，字段另含 `label`、`description`、`required`、`sort`、`options`。OAuth/SSO 账号、QQ/昵称、比赛服角色、正式服主角色、Public 信息与技能快照是每份申请固定收集的系统项。应用状态为 `draft|submitted|reviewing|approved|rejected|withdrawn`。

### 14.3 分数规则与 Set

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v1/at/score-rules` | 规则版本列表 |
| POST | `/api/v1/at/score-rules/import` | 管理员上传 XLSX (`multipart/form-data` 的 `file`) |
| GET | `/api/v1/at/score-rules/:id` | 规则、舰船、禁用船和惩罚明细 |
| POST | `/api/v1/at/score-rules/:id/publish` | 管理员发布不可变版本 |
| GET/POST | `/api/v1/at/fits` | 装配库列表/FC 创建可复用舰船装配 |
| GET/PUT/DELETE | `/api/v1/at/fits/:id` | 可视化装配明细/更新/删除未被引用装配 |
| GET/POST | `/api/v1/at/sets` | 列表/创建 Set |
| POST | `/api/v1/at/sets/parse-eft` | FC/管理员解析官方英文 EFT，返回舰船和可替换装备/弹药 |
| GET/PUT | `/api/v1/at/sets/:id` | 读取/更新 Set、槽位和替换操作 |
| GET | `/api/v1/at/sets/:id/eft` | 整队官方英文 EFT |

`at_fit` 保存基础 EFT、EFT 已具备的基础标签和可选标签替换规则；名称由系统按“舰船 + 电子战/加成标签”生成。Set 使用 `format=7v7|10v10`、`rule_version_id` 和有序 `slots`，新槽位通过 `fit_id` 引用装配并保存该位置最终标签。历史槽位的内嵌 `base_eft` 继续只读兼容。保存及输出都会重新校验 200 分、禁用船、EFT 船型和冲突。

分数导入当前严格支持 `ATXXII Comp Creator V1.4`：读取 `New Static Values (Do Not Edit)` F/J 区域，并按每艘船的 `Inflation Value` 保存重复惩罚；特殊舰船不会被 Hull Type 默认值覆盖。

### 14.4 比赛与局

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET/POST | `/api/v1/at/matches` | 列表/FC 创建 |
| GET | `/api/v1/at/matches/by-uuid/:uuid` | UUID 比赛详情 |
| GET | `/api/v1/at/matches/:id` | 比赛、报名和所有局 |
| POST | `/api/v1/at/matches/:id/positions/self` | 允许自选时，成员报名具体阵容位置和标签 |
| PUT | `/api/v1/at/matches/:id/positions` | FC 指定位置成员并微调标签 |
| GET | `/api/v1/at/matches/:id/positions/:slot/eft` | 成员复制本人位置最终英文 EFT |
| GET | `/api/v1/at/matches/users/search?keyword=` | FC 搜索可指定用户 |
| PUT | `/api/v1/at/matches/:id/registrations` | FC 确认/拒绝/指定；请求体包含 `user_id`、`status` |
| PUT | `/api/v1/at/matches/:id/status` | 合法状态转换/提前结束原因 |
| POST | `/api/v1/at/matches/:id/rounds` | 兼容接口；新流程由状态机自动创建下一局 |
| PUT | `/api/v1/at/matches/rounds/:id/roster` | 锁定双方完整阵容 |
| POST | `/api/v1/at/matches/rounds/:id/finish` | 进入 BR 审核 |
| POST | `/api/v1/at/matches/rounds/:id/reopen` | 作废并生成递增 attempt |
| POST | `/api/v1/at/matches/rounds/:id/confirm` | 确认胜/负/平；要求 BR 已固化 |
| POST | `/api/v1/at/matches/:id/reset` | 仅管理员；整场重置和贡献冲正 |

比赛状态为 `draft → recruiting → ready → in_progress → completed|early_ended|cancelled`；固定 `format`、`series=BO3|BO5`、Set/规则/贡献值/清场船快照。新比赛使用 UUID 详情页；进入 `in_progress` 自动创建第一局，确认一局且未达到系列赛胜场时自动创建下一局。KM 延迟和对手阵容暂缺只在比赛过程中提示，最终固化/确认 BR 仍执行完整校验。

### 14.5 比赛服 KM 与 BR

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/v1/at/killmails/submit` | 拥有 `at:km:submit` 权限的成员批量提交中文剪贴板 KM |
| GET | `/api/v1/at/killmails` | 全部已提交 KM（普通成员可读） |
| GET | `/api/v1/at/killmails/pending` | 仅 FC/管理员；全局待审核库 |
| GET | `/api/v1/at/killmails/:id/suggestions` | 仅 FC/管理员；比赛/单局归属建议 |
| POST | `/api/v1/at/killmails/:id/assign` | 仅 FC/管理员；确认归属 |
| POST | `/api/v1/at/killmails/rounds/:id/resolve-conflict` | 仅 FC/管理员；选择重复死亡有效记录 |
| GET | `/api/v1/at/br/rounds/:id/latest` | 最新站内 BR 版本 |
| GET | `/api/v1/at/br` | 每局最新 BR 总览（普通成员可读） |
| POST | `/api/v1/at/br/rounds/:id/build` | FC 固化；后续修订仅管理员且原因必填 |

KMID 为 UUIDv7，`normalized_hash` 为规范文本 SHA-256。角色/军团仅保存比赛服文本；舰船和武器批量关联 SDE TypeID。未解析、15 分钟超窗或未解决死亡冲突会阻止 BR 固化。

### 14.6 赛季与贡献

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET/POST/PUT | `/api/v1/at/contributions/seasons[/:id]` | 查询；管理员新增/更新赛季 |
| GET | `/api/v1/at/contributions/seasons/:id/cleanup-ships` | 查询赛季清场舰船及本地化名称 |
| PUT | `/api/v1/at/contributions/seasons/:id/cleanup-ships` | 管理员独立维护清场舰船；仅接受 SDE 舰船 TypeID |
| GET/POST/PUT | `/api/v1/at/contributions/match-types[/:id]` | 查询；管理员维护比赛类型/贡献值 |
| GET | `/api/v1/at/contributions/users/search?keyword=` | 管理员按昵称、比赛服角色或用户编号搜索贡献调整对象 |
| POST | `/api/v1/at/contributions/manual` | 管理员正负手调，非零且原因必填 |
| GET | `/api/v1/at/contributions/leaderboard?season_id=` | 当前/历史赛季或生涯榜单 |
| GET | `/api/v1/at/contributions/me?season_id=` | 当前用户的比赛、手调和冲正账目明细 |

`at_contribution_ledger` 不允许业务更新或删除。比赛、手工和冲正均以唯一 `source_key` 保证幂等；一名实际参加有效局的队员每场只结算一次。

### 14.7 技能规划

技能规划沿用 `/api/v1/operation/skill-plans` 契约并在 AT 模式恢复。普通用户可读取规划及检查自己的正式服主角色；FC/管理员可创建、更新、删除规划并检查所有用户的主角色。历史副角色技能不会进入 AT 技能规划结果。

### 14.8 EVE 战斗日志分析

所有接口需要 JWT 和正式成员角色（`user|fc|recruiter|admin|super_admin`）。原始日志只在上传请求内解析，不写入文件系统或数据库；数据库仅保存派生报表、按秒时间序列、飞行员/舰船/武器拆分以及输出/承受伤害的命中质量分布。

| 方法 | 路径 | 权限/说明 |
| --- | --- | --- |
| POST | `/api/v1/combat-logs` | `multipart/form-data` 的 `file`；UTF-8 `.txt`/`.log`，最大 20 MiB |
| GET | `/api/v1/combat-logs?page=1&page_size=20&scope=mine` | 当前成员自己的报表；`scope=all` 仅管理员可用 |
| GET | `/api/v1/combat-logs/:id` | 报表详情；普通成员仅可读取自己的记录，管理员可读取全部 |

分析类别固定为 `damage_out|damage_in|logistics_out|logistics_in|cap_transfer_out|cap_transfer_in|cap_damage_out|cap_damage_in|mining`。伤害命中质量固定为 `barely_scratches|hits|smashes|penetrates|wrecks|glances_off`，服务端按输出/承受方向聚合事件次数、伤害量并计算次数占比。采矿行会通过 SDE 批量解析矿石体积并返回 `mining_units`、`mining_m3` 和拆分行 `volume_m3`；未匹配到 SDE 物品时体积为 0。详情响应为：

```json
{
  "report": {
    "id": 1,
    "user_id": 2,
    "user_nickname": "Pilot",
    "character_name": "Capsuleer",
    "language": "english",
    "started_at": "2026-08-15T10:00:01Z",
    "ended_at": "2026-08-15T10:05:01Z",
    "duration_seconds": 300,
    "event_count": 120,
    "damage_out": 100000,
    "damage_in": 50000
  },
  "breakdowns": [],
  "timeline": [],
	"hit_qualities": [
	  {
		"direction": "damage_out",
		"quality": "hits",
		"hit_count": 32,
		"damage_amount": 24000,
		"percentage": 64
	  }
	],
  "average_dps_out": 333.33,
  "average_dps_in": 166.67
}
```

重复上传同一用户的相同文件（SHA-256 一致）会返回已有分析，并用本次请求中的日志补算/刷新命中质量分布，不重复创建报表。
