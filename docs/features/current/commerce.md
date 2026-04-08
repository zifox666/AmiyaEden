---
status: active
doc_type: feature
owner: engineering
last_reviewed: 2026-04-03
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

- 用户侧伏羲币与流水；流水表展示操作人
- 商品浏览、购买、订单、兑换码
- 管理员商品管理、订单发放/拒绝、订单历史、兑换码列表；福利官可访问订单发放/拒绝与订单历史
- 订单发放成功后，系统会以执行发放官员的主人物为发件人尽力发送一封双语游戏内邮件；若发件人未绑定可用主人物、未授权 `esi-mail.send_mail.v1` 或 ESI 发送失败，不影响发放结果
- 若订单已成功发放但邮件发送失败，订单管理界面会继续显示发放成功，并额外弹出一条包含后端错误内容的警告提示
- 管理员伏羲币调整、流水、日志

## 货币展示边界

- 本模块展示的是伏羲币，不适用仓库级 ISK 格式化标准。
- 若同一页面未来同时出现 ISK 与伏羲币，ISK 部分必须遵循 `docs/standards/isk-formatting.md`，伏羲币继续使用本模块既有约定。

## 入口

### 前端页面

- `static/src/views/shop/browse`
- `static/src/views/shop/manage`
- `static/src/views/shop/order-manage`
- `static/src/views/shop/wallet`
- `static/src/views/system/wallet`

### 后端路由

- `/api/v1/shop/*`
- `/api/v1/system/wallet/*`
- `/api/v1/system/shop/*`

## 权限边界

- 用户侧能力要求 `Login`
- `/system/wallet/*` 默认要求 `admin`
- `/system/shop/product/*` 与 `/system/shop/redeem/*` 默认要求 `admin`
- `/system/shop/order/*` 允许 `admin` 与 `welfare`

## 订单状态

订单只有三种状态：

| 状态 | 含义 |
| --- | --- |
| `requested` | 已下单，钱包已扣款，等待管理员发放 |
| `delivered` | 管理员已发放 |
| `rejected` | 管理员拒绝，钱包已退款 |

## 购买流程

```text
用户点击购买
  → 校验商品上架、库存、限购、余额
  → 事务：扣减库存 + 创建订单（status=requested）
  → 立即扣款（DebitUser）
  → 返回订单
```

管理员处理：

```text
订单管理（待发放）
  → 发放：status=delivered；若为兑换码类商品则生成兑换码
  → 拒绝：退款（CreditUser）+ 恢复库存 + status=rejected
```

## 订单号格式

8 位随机大写字母+数字（去掉易混淆字符），例如：`A3KM9ZQ2`。

## 商品

商品价格为整数，默认值为 1。

商品类型：`normal`（普通）、`redeem`（兑换码/服务）。

## 订单快照字段

下单时会从用户档案中快照以下信息，存入订单记录，后续不随用户资料变更而改变：

- `main_character_name`：主人物名
- `nickname`：昵称
- `qq`
- `discord_id`

## 管理后台订单视图

- **订单管理**（`shop/order-manage` → 订单管理 Tab）：仅展示 `requested` 状态订单，支持按商品名/主人物名/昵称关键字搜索，可执行发放或拒绝操作。
- **订单历史**（`shop/order-manage` → 订单历史 Tab）：展示 `delivered` + `rejected` 订单，支持相同关键字搜索，只读，并展示操作人与发放备注。
- **我的订单**（`shop/browse` → 我的订单 Tab）：展示订单状态，以及在已发放/已拒绝时展示操作人。
- 两个 Tab 都为 `订单号` 与 `主人物` 提供共享内联复制按钮，便于发放时复制合同描述或转账备注所需文本。
- 若商品名包含 `ISK`，订单管理与订单历史会在数量前展示 `ISK总和` 列，按 `total_price * 1,000,000` 计算，使用 compact 风格显示，并提供复制原始 ISK 数值的内联按钮。

`/shop/manage` 仍保留为管理员专用的商品管理入口。

管理后台订单视图的表格列包含：订单号、主人物、昵称、联系方式（QQ/Discord）、商品、`ISK总和`、数量、总价、操作人；历史视图额外显示发放备注。

## 关键不变量

- 伏羲币是多个模块共享的资金载体，不能按单一页面理解
- 钱包流水与调整日志属于不同查询面
- 用户侧钱包页面与其后端接口当前都归属 `Shop` 模块
- 用户侧 `/shop/wallet` 交易流水，以及管理端钱包列表、钱包流水、钱包操作日志，都按 ledger 视图处理
- 管理端钱包列表支持按当前用户昵称或任意已绑定人物名搜索
- 管理端钱包流水的用户筛选按 `/system/user` 一致语义执行，支持昵称或任意已绑定人物名搜索
- 商店、兑换码虽然都在 `Shop` 目录下，但用户态与管理态接口是分开的
- 商店商品图片上传当前通过 `/upload/image` 返回 base64 data URL，不写入项目文件夹；大小上限 2MB，仅支持 jpeg/png/webp
- 钱包在下单时立即扣款；拒绝订单时通过 `CreditUser` 退款，流水类型为 `shop_refund`
- `requested -> delivered` 成功后，服务会尽力向下单用户主人物发送一封双语发放通知邮件，发件人为执行发放的官员主人物；邮件失败只记录告警、不回滚订单发放，并在成功响应里附带 `mail_error` 供前端提示
- 若 ESI 接受了发信请求，成功响应还可能附带邮件调试信息；具体字段以代码契约为准

## 主要代码文件

- `server/internal/model/shop.go`
- `server/internal/service/shop.go`
- `server/internal/service/sys_wallet.go`
- `server/internal/repository/shop.go`
- `server/internal/router/router.go`
- `static/src/api/shop.ts`
- `static/src/api/sys-wallet.ts`
- `static/src/views/shop`
- `static/src/views/system/wallet`
