# ESI 数据刷新队列

## 概述

本模块 (`pkg/eve/esi/`) 负责管理和调度 EVE ESI 数据的定时刷新任务。

**核心特性：**

- 每种数据类型一个独立 `.go` 文件，方便扩展
- 任务自动注册机制（通过 `init()` + `Register()`）
- 支持任务优先级
- 不活跃角色自动降频刷新
- Redis 记录任务执行状态，支持可视化
- 并发控制，防止 ESI 限流
- **自动分页**：`GetPaginated` 自动处理 `x-pages` 多页响应，合并所有数据
- **限速感知**：`RateLimiter` 根据 `x-ratelimit-*` 响应头自动节流
- **420 重试**：遇到 ESI 限速 (HTTP 420) 自动指数退避重试

## 目录结构

```
pkg/eve/esi/
├── README.md              # 本文件
├── client.go              # ESI HTTP 客户端（基础方法）
├── request.go             # 增强请求客户端（分页、限速、元数据）
├── task.go                # 任务接口定义、优先级、注册表
├── queue.go               # 队列调度引擎
├── activity.go            # 角色活跃度检测
├── task_affiliation.go    # 角色归属（军团/联盟）
├── task_assets.go         # 角色资产
├── task_clones.go         # 克隆体/植入体/跳跃疲劳
├── task_contracts.go      # 角色合同
├── task_killmails.go      # 击杀邮件
├── task_notifications.go  # 角色通知
├── task_online.go         # 在线状态
├── task_titles.go         # 角色头衔
└── task_wallet.go         # 角色钱包
```

## 请求客户端

### 架构分层

```
client.go          底层 HTTP（Get / GetRaw / PostJSON / PutJSON）
    ↓
request.go         增强层（分页 + 限速 + 元数据 + 420 重试）
                   ├── GetPaginated()   自动分页合并
                   ├── GetWithMeta()    返回响应元数据
                   ├── RateLimiter      限速器
                   └── ResponseMeta     响应元数据
```

### ResponseMeta 响应元数据

每次请求后从 ESI 响应头中提取以下信息：

| 字段 | 响应头 | 说明 |
|------|--------|------|
| `CacheStatus` | `x-esi-cache-status` | 缓存状态（HIT / MISS） |
| `RequestID` | `x-esi-request-id` | 请求唯一标识 |
| `Pages` | `x-pages` | 总页数（分页端点） |
| `RateLimitGroup` | `x-ratelimit-group` | 限速组（如 char-wallet） |
| `RateLimitLimit` | `x-ratelimit-limit` | 限速窗口（如 150/15m） |
| `RateLimitRemain` | `x-ratelimit-remaining` | 窗口内剩余请求数 |
| `RateLimitUsed` | `x-ratelimit-used` | 窗口内已用请求数 |
| `ETag` | `ETag` | 缓存标签 |

### GetPaginated —— 自动分页

ESI 的分页端点会在响应头中返回 `x-pages` 表示总页数。`GetPaginated` 会：

1. 请求第 1 页，从响应头获取 `x-pages` 总页数
2. 并发拉取剩余页面（最多 10 并发，受限速器约束）
3. 将所有页面的 JSON 数组合并为完整切片
4. 返回合并后的结果 + 第 1 页的 `ResponseMeta`

```go
// 之前（不支持分页，只能拿到第 1 页数据）
var assets []AssetItem
if err := ctx.Client.Get(bgCtx, path, token, &assets); err != nil { ... }

// 现在（自动拉取全部页，对调用方透明）
var assets []AssetItem
if _, err := ctx.Client.GetPaginated(bgCtx, path, token, &assets); err != nil { ... }
```

**不需要分页的端点** 继续使用 `Get()` 即可，`GetPaginated` 在 `x-pages ≤ 1` 时行为等同于 `Get`。

### GetWithMeta —— 带元数据的 GET

如果不需要分页但想获取限速/缓存状态：

```go
meta, err := ctx.Client.GetWithMeta(bgCtx, path, token, &result)
if err != nil { ... }
fmt.Println(meta.CacheStatus)     // "HIT"
fmt.Println(meta.RateLimitRemain) // 139
```

### RateLimiter —— 自动限速

限速器根据 ESI 响应头中的 `x-ratelimit-*` 信息自动追踪各限速组的配额：

- **remaining > 10**：正常请求
- **5 < remaining ≤ 10**：节流（500ms 延迟）
- **remaining ≤ 5**：暂停，等待窗口重置
- **HTTP 420**：指数退避重试（2s → 4s → 8s，最多 3 次）

限速器对调用方完全透明，由 `Client` 内部自动管理。

### 420 限速重试

当 ESI 返回 HTTP 420（Rate Limited）时，`doRequest` 会自动执行指数退避重试：
- 第 1 次重试：等待 2 秒
- 第 2 次重试：等待 4 秒
- 第 3 次重试：等待 8 秒
- 超过 3 次：返回错误

## 如何添加新的刷新任务

### 1. 创建任务文件

在 `pkg/eve/esi/` 下新建 `task_xxx.go` 文件：

```go
package esi

import (
    "amiya-eden/global"
    "context"
    "fmt"
    "time"

    "go.uber.org/zap"
)

// ─────────────────────────────────────────────
//  任务名称和说明
//  对应 ESI 接口：GET /characters/{character_id}/xxx
//  默认刷新间隔: X Hours / 不活跃: Y Days
// ─────────────────────────────────────────────

func init() {
    Register(&XxxTask{})
}

// XxxTask 你的任务描述
type XxxTask struct{}

func (t *XxxTask) Name() string        { return "character_xxx" }
func (t *XxxTask) Description() string { return "任务可读描述" }
func (t *XxxTask) Priority() Priority  { return PriorityNormal }

func (t *XxxTask) Interval() RefreshInterval {
    return RefreshInterval{
        Active:   6 * time.Hour,       // 活跃角色刷新间隔
        Inactive: 7 * 24 * time.Hour,  // 不活跃角色刷新间隔
    }
}

func (t *XxxTask) RequiredScopes() []TaskScope {
    return []TaskScope{
        {Scope: "esi-xxx.read_xxx.v1", Description: "scope 说明"},
    }
}

func (t *XxxTask) Execute(ctx *TaskContext) error {
    bgCtx := context.Background()
    path := fmt.Sprintf("/characters/%d/xxx/", ctx.CharacterID)

    // 方式 1：需要分页的端点（如 assets、contracts、killmails）
    var result []YourStruct
    if _, err := ctx.Client.GetPaginated(bgCtx, path, ctx.AccessToken, &result); err != nil {
        return fmt.Errorf("fetch xxx: %w", err)
    }

    // 方式 2：不需要分页的端点
    // if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &result); err != nil { ... }

    // 方式 3：需要响应元数据（缓存状态、限速信息等）
    // meta, err := ctx.Client.GetWithMeta(bgCtx, path, ctx.AccessToken, &result)
    // if err != nil { ... }
    // fmt.Println(meta.CacheStatus, meta.RateLimitRemain)

    global.Logger.Debug("[ESI] xxx 刷新完成",
        zap.Int64("character_id", ctx.CharacterID),
        zap.Int("count", len(result)),
    )

    // TODO: 将 result 入库

    return nil
}
```

### 2. 关键说明

| 项目 | 说明 |
|------|------|
| **`init()` 中调用 `Register()`** | 任务会自动注册到全局注册表，无需其他配置 |
| **`Name()`** | 全局唯一的任务标识，建议格式 `character_xxx` |
| **`Priority()`** | `PriorityCritical(1)` > `PriorityHigh(10)` > `PriorityNormal(50)` > `PriorityLow(90)` |
| **`Interval()`** | 分别设置活跃/不活跃角色的刷新间隔 |
| **`RequiredScopes()`** | 角色必须拥有这些 scope 才会执行该任务 |
| **`Execute()`** | 核心执行逻辑，`ctx` 中携带 characterID、accessToken、client |

### 3. 请求方法选择

| 方法 | 使用场景 | 示例端点 |
|------|---------|---------|
| `Client.Get()` | 单页端点、不关心元数据 | `/characters/{id}/online/` |
| `Client.GetWithMeta()` | 单页但需要缓存/限速信息 | 带 ETag 条件请求 |
| `Client.GetPaginated()` | 分页端点（自动合并所有页） | `/characters/{id}/assets/` |
| `Client.PostJSON()` | POST 请求 | `/characters/affiliation/` |

### 4. Scope 注册（可选）

如果你希望新任务的 scope 出现在 SSO 登录授权页，在任务文件中额外添加：

```go
func init() {
    Register(&XxxTask{})

    // 注册到 SSO scope 列表
    service.RegisterScope(
        "xxx",                       // 模块名
        "esi-xxx.read_xxx.v1",       // scope
        "读取角色 xxx 数据",           // 描述
        false,                       // 是否必选
    )
}
```

### 5. 批量任务（可选）

如果需要批量处理多个角色（如 affiliation），实现 `BatchTask` 接口：

```go
func (t *XxxTask) ExecuteBatch(client *Client, characterIDs []int64) error {
    // 批量处理逻辑
    return nil
}
```

## 优先级参考

| 优先级 | 值 | 适用场景 |
|--------|----|---------| 
| `PriorityCritical` | 1 | 高频关键数据（killmail） |
| `PriorityHigh` | 10 | 重要但不需要太频繁 |
| `PriorityNormal` | 50 | 标准数据 |
| `PriorityLow` | 90 | 非关键低频数据 |

## 刷新间隔参考

| 任务 | 活跃角色 | 不活跃角色 | 分页 |
|------|---------|-----------|------|
| killmails | 20m | 3d | ✓ |
| online | 30m | 2h | ✗ |
| affiliation | 2h | 2h | ✗ |
| titles / clones | 6h | 7d | ✗ |
| wallet | 12h | 7d | ✓ |
| assets / notifications / contracts | 1d | 7d | ✓/✗/✓ |

## 活跃度判定

- 通过 ESI `GET /characters/{character_id}/online` 获取 `last_login`
- 7 天未登录视为不活跃
- 活跃状态缓存 1 小时（Redis）
- 查询失败时默认视为活跃

## 架构图

```
Cron (每5分钟) ──> Queue.Run()
                     │
                     ├─ 1. 获取所有有 Token 的角色
                     ├─ 2. 检测各角色活跃度
                     ├─ 3. 按优先级排序任务
                     ├─ 4. 过滤（scope 检查 + 间隔检查）
                     └─ 5. 并发执行（信号量控制）
                           │
                           ├─ GetValidToken（自动刷新）
                           ├─ task.Execute()
                           │     │
                           │     ├─ GetPaginated()  ─── 分页端点
                           │     │     ├─ doRequest(page=1) → 解析 x-pages
                           │     │     ├─ 并发 doRequest(page=2..N)
                           │     │     │     └─ RateLimiter.Wait() 限速控制
                           │     │     └─ mergeJSONArrays() 合并结果
                           │     │
                           │     ├─ Get()           ─── 单页端点
                           │     └─ GetWithMeta()   ─── 需要元数据
                           │
                           └─ 更新 Redis 状态
```
