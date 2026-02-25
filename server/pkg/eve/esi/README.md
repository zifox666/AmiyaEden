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

## 目录结构

```
pkg/eve/esi/
├── README.md              # 本文件
├── client.go              # ESI HTTP 客户端
├── task.go                # 任务接口定义、优先级、注册表
├── queue.go               # 队列调度引擎
├── activity.go            # 角色活跃度检测
├── task_affiliation.go    # 角色归属（军团/联盟）
├── task_assets.go         # 角色资产
├── task_notifications.go  # 角色通知
├── task_titles.go         # 角色头衔
├── task_clones.go         # 克隆体/植入体/跳跃疲劳
├── task_contracts.go      # 角色合同
└── task_killmails.go      # 击杀邮件
```

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

    var result []YourStruct
    if err := ctx.Client.Get(bgCtx, path, ctx.AccessToken, &result); err != nil {
        return fmt.Errorf("fetch xxx: %w", err)
    }

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

### 3. Scope 注册（可选）

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

### 4. 批量任务（可选）

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

| 任务 | 活跃角色 | 不活跃角色 |
|------|---------|-----------|
| killmails | 20m | 3d |
| affiliation | 2h | 2h |
| titles / clones | 6h | 7d |
| assets / notifications / contracts | 1d | 7d |

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
                           └─ 更新 Redis 状态
```
