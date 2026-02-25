# AmiyaEden 开发者维护指南

> 本文档面向后续接手维护/开发的人员，涵盖项目结构、核心机制（权限控制、动态路由、ESI Token 管理）以及 **添加新模块的完整流程**。

---

## 目录

1. [项目概览](#1-项目概览)
2. [技术栈](#2-技术栈)
3. [项目结构](#3-项目结构)
4. [核心机制](#4-核心机制)
   - 4.1 [权限控制](#41-权限控制)
   - 4.2 [动态路由](#42-动态路由)
   - 4.3 [ESI Token 管理](#43-esi-token-管理)
   - 4.4 [ESI 刷新任务框架](#44-esi-刷新任务框架)
5. [添加新模块完整流程](#5-添加新模块完整流程)
6. [附录](#6-附录)

---

## 1. 项目概览

AmiyaEden 是一个为 EVE Online 联盟/军团打造的管理平台，包含：

- **EVE SSO 登录**：通过 EVE Online OAuth 2.0 进行身份认证，支持多角色绑定
- **舰队行动管理**（Operation/Fleet）：创建舰队、ESI 成员同步、PAP 发放
- **SRP 补损系统**：击杀邮件关联补损申请、审批、发放
- **ESI 数据刷新**：自动定时从 ESI API 拉取角色数据（资产、击杀、合同等）
- **SDE 静态数据**：自动从 GitHub Release 下载并导入最新 EVE 静态数据

---

## 2. 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go + Gin + GORM (MySQL) + Redis + robfig/cron |
| 前端 | Vue 3 + TypeScript + Vite + Pinia + Vue Router |
| 认证 | EVE SSO OAuth 2.0 + 自制 HMAC-SHA256 JWT |
| 日志 | zap + lumberjack（切割） |
| 缓存 | Redis（SSO state、ESI 活跃度、任务状态） |

---

## 3. 项目结构

### 3.1 后端 (`server/`)

```
server/
├── main.go                     # 入口：初始化各组件并启动 HTTP 服务
├── Makefile                    # 构建脚本
├── config/
│   ├── config.go               # 配置结构体定义
│   ├── config.yaml             # 配置文件（数据库、Redis、JWT、EVE SSO 等）
│   └── scopes.json             # ESI scope 列表（参考用）
├── bootstrap/                  # 启动引导
│   ├── config.go               # 读取配置 → global.Config
│   ├── logger.go               # zap 日志 → global.Logger
│   ├── db.go                   # GORM MySQL → global.DB + AutoMigrate
│   ├── redis.go                # Redis → global.Redis
│   ├── cron.go                 # 定时任务 → global.Cron
│   ├── router.go               # Gin 路由引擎（注册全局中间件 + 业务路由）
│   └── scopes.go               # ESI Task 的 scope 注册到 SSO 服务
├── global/
│   └── global.go               # 全局变量（Config / Logger / DB / Redis / Cron）
├── internal/
│   ├── handler/                # HTTP 处理器（Controller 层）
│   │   ├── eve_sso.go          # SSO 登录/回调/绑定角色
│   │   ├── user.go             # 用户管理 + /me
│   │   ├── menu.go             # 动态菜单
│   │   ├── fleet.go            # 舰队行动
│   │   ├── srp.go              # SRP 补损
│   │   ├── sde.go              # SDE 数据查询
│   │   └── esi_refresh.go      # ESI 刷新队列管理
│   ├── service/                # 业务逻辑层
│   │   ├── eve_sso.go          # SSO 业务 + Scope 注册机制
│   │   ├── user.go             # 用户 CRUD
│   │   ├── fleet.go            # 舰队业务（含 ESI 交互）
│   │   ├── srp.go              # 补损业务
│   │   └── sde.go              # SDE 更新与查询
│   ├── repository/             # 数据访问层（DAO）
│   │   ├── user.go
│   │   ├── eve_character.go
│   │   ├── fleet.go
│   │   ├── srp.go
│   │   └── sde.go
│   ├── model/                  # 数据模型（GORM Model）
│   │   ├── base.go             # BaseModel（ID / CreatedAt / UpdatedAt / DeletedAt）
│   │   ├── user.go             # 用户
│   │   ├── role.go             # 角色常量 + 权限继承逻辑
│   │   ├── menu.go             # 动态菜单定义 + 权限过滤
│   │   ├── eve_character.go    # EVE 角色（含 Token）
│   │   ├── esi_data.go         # ESI 数据表（资产/通知/头衔/克隆/KM/合同）
│   │   ├── fleet.go            # 舰队/成员/PAP/邀请/钱包
│   │   ├── srp.go              # 补损价格表/申请
│   │   ├── sde.go              # SDE 静态数据表
│   │   └── operation_log.go    # API 操作日志
│   ├── middleware/             # 中间件
│   │   ├── auth.go             # JWT 鉴权 + RequireRole
│   │   ├── apikey.go           # API Key 鉴权
│   │   ├── cors.go             # 跨域
│   │   ├── response.go         # 统一响应包装
│   │   ├── operation_log.go    # 操作日志
│   │   ├── logger.go           # 请求日志
│   │   ├── recovery.go         # Panic 恢复
│   │   └── requestid.go        # 请求 ID
│   └── router/
│       └── router.go           # 路由注册（所有 API 路由定义在此）
├── jobs/                       # 定时任务
│   ├── jobs.go                 # 任务统一注册
│   ├── esi_refresh.go          # ESI 刷新队列定时调度
│   └── sde.go                  # SDE 自动更新
├── pkg/                        # 公共工具包
│   ├── cache/cache.go          # Redis 缓存封装
│   ├── eve/
│   │   ├── sso.go              # EVE SSO OAuth 客户端
│   │   └── esi/                # ESI 刷新任务框架
│   │       ├── client.go       # ESI HTTP 客户端
│   │       ├── task.go         # 任务接口 + 注册表
│   │       ├── queue.go        # 队列调度引擎
│   │       ├── activity.go     # 角色活跃度检测
│   │       ├── task_*.go       # 各具体刷新任务
│   │       └── README.md       # 框架文档
│   ├── jwt/jwt.go              # HMAC-SHA256 JWT
│   ├── response/response.go    # 统一响应格式
│   └── utils/password.go       # 密码工具
└── tmp/sde/                    # SDE 临时下载目录
```

### 3.2 前端 (`static/`)

```
static/src/
├── main.ts                     # 入口
├── App.vue                     # 根组件
├── api/                        # API 调用层
│   ├── auth.ts                 # SSO 登录 / 用户信息
│   ├── fleet.ts                # 舰队操作
│   ├── srp.ts                  # SRP 补损
│   ├── esi-refresh.ts          # ESI 刷新管理
│   └── system-manage.ts        # 系统管理（用户/角色/菜单）
├── router/
│   ├── index.ts                # 路由实例
│   ├── routesAlias.ts          # 路由公共别名
│   ├── core/                   # 路由核心类（注册/加载/权限校验）
│   │   ├── RouteRegistry.ts    # 动态路由注册/卸载
│   │   ├── ComponentLoader.ts  # 视图组件动态加载
│   │   ├── RouteTransformer.ts # 后端→前端路由格式转换
│   │   ├── RouteValidator.ts   # 路由配置校验
│   │   ├── MenuProcessor.ts    # 菜单处理（前端/后端模式）
│   │   ├── RoutePermissionValidator.ts  # 路由权限校验
│   │   └── IframeRouteManager.ts        # iframe 路由管理
│   ├── guards/
│   │   ├── beforeEach.ts       # 前置守卫（登录态/动态路由注册/权限校验）
│   │   └── afterEach.ts        # 后置守卫
│   ├── modules/                # 前端路由模块定义（仅前端模式使用）
│   │   ├── index.ts            # 汇总导出
│   │   ├── dashboard.ts        # 仪表盘
│   │   ├── operation.ts        # 舰队行动
│   │   ├── system.ts           # 系统管理
│   │   ├── result.ts           # 结果页
│   │   ├── exception.ts        # 异常页
│   │   └── srp.ts              # SRP（待完善）
│   └── routes/
│       ├── staticRoutes.ts     # 静态路由（登录/回调/异常页）
│       └── asyncRoutes.ts      # 异步路由入口
├── store/modules/
│   ├── user.ts                 # 用户状态（登录态/角色/Token）
│   ├── menu.ts                 # 菜单状态（menuList/homePath）
│   ├── setting.ts              # 应用设置
│   ├── worktab.ts              # 标签页
│   └── table.ts                # 表格设置
├── types/
│   ├── api/api.d.ts            # 所有 API 类型定义
│   └── router/index.ts         # 路由类型
├── views/                      # 页面视图组件
├── components/                 # 公共组件
├── locales/                    # 国际化
└── utils/                      # 工具函数
```

---

## 4. 核心机制

### 4.1 权限控制

#### 4.1.1 角色体系

系统采用 **角色优先级继承** 模型，定义在 `server/internal/model/role.go`：

| 角色 | 常量 | 优先级 | 说明 |
|------|------|--------|------|
| `super_admin` | `RoleSuperAdmin` | 100 | 超级管理员，全部权限 |
| `admin` | `RoleAdmin` | 50 | 管理员，非技术性管理 |
| `srp` | `RoleSRP` | 40 | 补损管理员 |
| `fc` | `RoleFC` | 30 | 舰队指挥 |
| `user` | `RoleUser` | 10 | 已认证用户 |
| `guest` | `RoleGuest` | 0 | 访客 |

**继承规则**：`HasRole(userRole, requiredRole)` — 用户角色优先级 ≥ 所需角色优先级即视为有权限。

例：`admin` (50) 可以访问需要 `fc` (30) 权限的接口。

#### 4.1.2 后端权限控制

权限在路由层通过中间件实现，定义在 `server/internal/middleware/auth.go`：

```go
// JWT 鉴权（所有需要登录的接口）
middleware.JWTAuth()

// 角色权限检查（基于继承）
middleware.RequireRole(model.RoleAdmin)   // 需要 Admin 或以上
middleware.RequireRole(model.RoleFC)      // 需要 FC 或以上

// 精确角色匹配（不走继承）
middleware.RequireAnyRole(model.RoleSRP, model.RoleFC)
```

JWT Token 中包含的信息：`uid`（用户 ID）、`cid`（角色 ID）、`role`（角色名）、`exp`（过期时间）。

#### 4.1.3 前端权限控制

前端使用 **角色映射** 将后端角色转让为前端角色代码（`auth.ts` 中的 `ROLE_MAP`）：

| 后端角色 | 前端角色 |
|---------|---------|
| `super_admin` | `R_SUPER` |
| `admin` | `R_ADMIN` |
| `srp` | `R_SRP` |
| `fc` | `R_FC` |
| `user` | `R_USER` |
| `guest` | `R_GUEST` |

前端路由通过 `meta.roles` 做权限过滤：
```typescript
meta: { roles: ['R_SUPER', 'R_ADMIN'] }  // 只有这些角色可见
```

---

### 4.2 动态路由

系统支持 **前端模式** 和 **后端模式** 两种动态路由方案：

#### 4.2.1 后端模式（默认，推荐）

**流程**：

```
用户登录
  → 获取 JWT Token
  → 路由守卫检测到未注册动态路由
  → 调用 GET /api/v1/me 获取用户信息
  → 调用 GET /api/v1/menu 获取菜单列表（后端根据角色过滤）
  → RouteRegistry.register() 注册动态路由
  → 存储到 MenuStore
```

**后端菜单定义**在 `server/internal/model/menu.go` 的 `allMenus` 数组中：

```go
var allMenus = []*menuItemWithRoles{
    {
        MenuItem: MenuItem{
            Path:      "/dashboard",
            Name:      "Dashboard",
            Component: "/index/index",   // 对应 views/index/index.vue（布局容器）
            Meta:      MenuMeta{Title: "menus.dashboard.title", Icon: "ri:pie-chart-line"},
        },
        requiredRole: "",      // 空 = 所有已登录用户
        children: []*menuItemWithRoles{
            {
                MenuItem: MenuItem{
                    Path:      "console",         // 子路由用相对路径
                    Name:      "Console",
                    Component: "/dashboard/console", // 对应 views/dashboard/console.vue
                    Meta:      MenuMeta{Title: "menus.dashboard.console"},
                },
                requiredRole: "",
            },
        },
    },
}
```

前端 `RouteTransformer` 收到菜单数据后：
1. 一级路由的 `component` 为 `/index/index` → 加载布局容器
2. 子路由的 `component` → 通过 `import.meta.glob('../../views/**/*.vue')` 动态加载对应视图

#### 4.2.2 前端模式

前端路由定义在 `static/src/router/modules/` 下各模块文件中，通过 `meta.roles` 做前端过滤。

#### 4.2.3 关键组件

| 组件 | 作用 |
|------|------|
| `MenuProcessor` | 根据模式获取菜单列表（前端过滤 / 后端请求） |
| `RouteRegistry` | 将菜单列表注册为 Vue Router 路由 |
| `ComponentLoader` | 通过 `import.meta.glob` 加载视图组件 |
| `RouteTransformer` | 将后端路由格式转换为 Vue Router 格式 |
| `RoutePermissionValidator` | 校验当前路径是否在菜单权限内 |

---

### 4.3 ESI Token 管理

#### 4.3.1 Token 存储

每个 EVE 角色（`eve_character` 表）保存以下 Token 信息：

| 字段 | 说明 |
|------|------|
| `access_token` | ESI API 的访问令牌（~20 分钟有效） |
| `refresh_token` | 刷新令牌（长期有效，用于获取新 access_token） |
| `token_expiry` | access_token 过期时间 |
| `scopes` | 授权的 ESI scope 列表（空格分隔） |

#### 4.3.2 Token 刷新机制

`EveSSOService.GetValidToken(ctx, characterID)` 是获取有效 Token 的统一入口：

```
调用 GetValidToken(characterID)
  → 从 DB 读取 eve_character
  → 检查 token_expiry 是否在 5 分钟内过期
  → 若即将过期：调用 EVE SSO RefreshAccessToken
  → 更新 DB 中的 access_token / refresh_token / token_expiry / scopes
  → 返回有效的 access_token
```

#### 4.3.3 Scope 注册机制

各模块在启动时通过 `service.RegisterScope()` 声明需要的 ESI scope：

```go
// ESI 刷新任务的 scope 由框架自动收集
bootstrap.InitScopes()  // 遍历所有 ESI Task 的 RequiredScopes()

// 其他模块手动注册
service.RegisterScope("fleet", "esi-fleets.read_fleet.v1", "读取舰队信息", true)
```

所有已注册的 scope 会在 SSO 登录时自动合并请求，确保用户授权后获得完整权限。

前端可通过 `GET /api/v1/sso/eve/scopes` 查看所有已注册 scope。

---

### 4.4 ESI 刷新任务框架

#### 4.4.1 架构

```
Cron (每 5 分钟) ──> Queue.Run()
                     ├─ 1. 获取所有有 Token 的角色
                     ├─ 2. 检测各角色活跃度（ESI online + Redis 缓存）
                     ├─ 3. 按优先级排序任务
                     ├─ 4. 过滤：scope 检查 + 间隔检查（活跃/不活跃不同间隔）
                     └─ 5. 并发执行（信号量控制，默认 5 并发）
```

#### 4.4.2 任务接口

每个刷新任务实现 `esi.RefreshTask` 接口：

```go
type RefreshTask interface {
    Name() string
    Description() string
    Priority() TaskPriority        // Critical / High / Normal / Low
    Interval() TaskInterval        // Active 间隔 / Inactive 间隔
    RequiredScopes() []TaskScope   // 需要的 ESI scope
    Execute(ctx context.Context, client *Client, char *model.EveCharacter) error
}
```

任务通过 `init()` + `Register()` 自动注册到全局注册表。

#### 4.4.3 活跃度机制

- **活跃角色**：7 天内有 ESI 在线记录 → 使用 `Interval().Active` 间隔刷新
- **不活跃角色**：超过 7 天未登录 → 使用 `Interval().Inactive` 间隔（降频）
- 活跃状态缓存在 Redis，key：`eve:activity:{characterID}`，TTL 1 小时

---

## 5. 添加新模块完整流程

以添加一个假想的 **「采矿管理」（Mining）** 模块为例，完整演示从后端到前端的全流程。

### 步骤 1：定义数据模型

**文件**：`server/internal/model/mining.go`

```go
package model

import "time"

// MiningSession 采矿会话记录
type MiningSession struct {
    BaseModel
    UserID        uint      `gorm:"not null;index"          json:"user_id"`
    CharacterID   int64     `gorm:"not null;index"          json:"character_id"`
    CharacterName string    `gorm:"size:128"                json:"character_name"`
    SolarSystemID int64     `gorm:"not null"                json:"solar_system_id"`
    StartAt       time.Time `gorm:"not null"                json:"start_at"`
    EndAt         *time.Time `gorm:""                       json:"end_at,omitempty"`
    OreTypeID     int64     `gorm:"not null"                json:"ore_type_id"`
    Quantity      int64     `gorm:"default:0"               json:"quantity"`
    Status        string    `gorm:"size:32;default:'active'" json:"status"` // active / completed
}

func (MiningSession) TableName() string { return "mining_session" }
```

### 步骤 2：注册数据库自动迁移

**文件**：`server/bootstrap/db.go` — 在 `autoMigrate` 函数中添加：

```go
func autoMigrate(db *gorm.DB) {
    if err := db.AutoMigrate(
        // ... 已有模型 ...
        // Mining
        &model.MiningSession{},
    ); err != nil {
        global.Logger.Fatal("数据库迁移失败", zap.Error(err))
    }
}
```

### 步骤 3：创建 Repository（数据访问层）

**文件**：`server/internal/repository/mining.go`

```go
package repository

import (
    "amiya-eden/global"
    "amiya-eden/internal/model"
)

type MiningRepository struct{}

func NewMiningRepository() *MiningRepository {
    return &MiningRepository{}
}

func (r *MiningRepository) Create(session *model.MiningSession) error {
    return global.DB.Create(session).Error
}

func (r *MiningRepository) GetByID(id uint) (*model.MiningSession, error) {
    var session model.MiningSession
    err := global.DB.First(&session, id).Error
    return &session, err
}

type MiningFilter struct {
    UserID      *uint
    CharacterID *int64
    Status      string
}

func (r *MiningRepository) List(page, pageSize int, filter MiningFilter) ([]model.MiningSession, int64, error) {
    var list []model.MiningSession
    var total int64

    db := global.DB.Model(&model.MiningSession{})
    if filter.UserID != nil {
        db = db.Where("user_id = ?", *filter.UserID)
    }
    if filter.CharacterID != nil {
        db = db.Where("character_id = ?", *filter.CharacterID)
    }
    if filter.Status != "" {
        db = db.Where("status = ?", filter.Status)
    }

    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    offset := (page - 1) * pageSize
    err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
    return list, total, err
}
```

### 步骤 4：创建 Service（业务逻辑层）

**文件**：`server/internal/service/mining.go`

```go
package service

import (
    "amiya-eden/internal/model"
    "amiya-eden/internal/repository"
    "errors"
    "time"
)

type MiningService struct {
    repo     *repository.MiningRepository
    charRepo *repository.EveCharacterRepository
}

func NewMiningService() *MiningService {
    return &MiningService{
        repo:     repository.NewMiningRepository(),
        charRepo: repository.NewEveCharacterRepository(),
    }
}

type CreateMiningSessionRequest struct {
    CharacterID   int64  `json:"character_id" binding:"required"`
    SolarSystemID int64  `json:"solar_system_id" binding:"required"`
    OreTypeID     int64  `json:"ore_type_id" binding:"required"`
    StartAt       string `json:"start_at" binding:"required"`
}

func (s *MiningService) CreateSession(userID uint, req *CreateMiningSessionRequest) (*model.MiningSession, error) {
    char, err := s.charRepo.GetByCharacterID(req.CharacterID)
    if err != nil || char.UserID != userID {
        return nil, errors.New("角色不属于当前用户")
    }

    startAt, err := time.Parse(time.RFC3339, req.StartAt)
    if err != nil {
        return nil, errors.New("时间格式错误")
    }

    session := &model.MiningSession{
        UserID:        userID,
        CharacterID:   req.CharacterID,
        CharacterName: char.CharacterName,
        SolarSystemID: req.SolarSystemID,
        OreTypeID:     req.OreTypeID,
        StartAt:       startAt,
        Status:        "active",
    }
    if err := s.repo.Create(session); err != nil {
        return nil, err
    }
    return session, nil
}

func (s *MiningService) ListSessions(page, pageSize int, filter repository.MiningFilter) ([]model.MiningSession, int64, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 20 }
    return s.repo.List(page, pageSize, filter)
}
```

### 步骤 5：创建 Handler（HTTP 处理器）

**文件**：`server/internal/handler/mining.go`

```go
package handler

import (
    "amiya-eden/internal/middleware"
    "amiya-eden/internal/repository"
    "amiya-eden/internal/service"
    "amiya-eden/pkg/response"
    "strconv"

    "github.com/gin-gonic/gin"
)

type MiningHandler struct {
    svc *service.MiningService
}

func NewMiningHandler() *MiningHandler {
    return &MiningHandler{svc: service.NewMiningService()}
}

// CreateSession POST /api/v1/mining/sessions
func (h *MiningHandler) CreateSession(c *gin.Context) {
    var req service.CreateMiningSessionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, response.CodeParamError, "参数错误: "+err.Error())
        return
    }
    userID := middleware.GetUserID(c)
    session, err := h.svc.CreateSession(userID, &req)
    if err != nil {
        response.Fail(c, response.CodeBizError, err.Error())
        return
    }
    response.OK(c, session)
}

// ListSessions GET /api/v1/mining/sessions
func (h *MiningHandler) ListSessions(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

    filter := repository.MiningFilter{
        Status: c.Query("status"),
    }

    records, total, err := h.svc.ListSessions(page, size, filter)
    if err != nil {
        response.Fail(c, response.CodeBizError, err.Error())
        return
    }
    response.OK(c, gin.H{
        "records": records,
        "current": page,
        "size":    size,
        "total":   total,
    })
}
```

### 步骤 6：注册路由

**文件**：`server/internal/router/router.go`

在 `RegisterRoutes` 中添加注册调用：

```go
func RegisterRoutes(r *gin.Engine) {
    v1 := r.Group("/api/v1")
    {
        // ... 已有路由 ...
        registerMiningRoutes(v1)  // ← 新增
    }
}
```

新增路由注册函数：

```go
// registerMiningRoutes 采矿管理路由
func registerMiningRoutes(rg *gin.RouterGroup) {
    h := handler.NewMiningHandler()
    mining := rg.Group("/mining", middleware.JWTAuth())
    {
        mining.POST("/sessions", h.CreateSession)
        mining.GET("/sessions", h.ListSessions)
        // 根据需求添加管理端路由
        // admin := mining.Group("", middleware.RequireRole(model.RoleAdmin))
        // admin.DELETE("/sessions/:id", h.DeleteSession)
    }
}
```

### 步骤 7：添加后端动态菜单

**文件**：`server/internal/model/menu.go`

在 `allMenus` 中添加采矿模块的菜单定义：

```go
{
    MenuItem: MenuItem{
        Path:      "/mining",
        Name:      "Mining",
        Component: "/index/index",     // 一级菜单固定用布局容器
        Meta: MenuMeta{
            Title: "menus.mining.title",
            Icon:  "ri:hammer-line",
        },
    },
    requiredRole: "",  // 所有已登录用户可见，或设为 RoleUser
    children: []*menuItemWithRoles{
        {
            MenuItem: MenuItem{
                Path:      "sessions",
                Name:      "MiningSessions",
                Component: "/mining/sessions",   // 对应 views/mining/sessions.vue
                Meta: MenuMeta{
                    Title:     "menus.mining.sessions",
                    KeepAlive: true,
                },
            },
            requiredRole: "",
        },
    },
},
```

### 步骤 8：创建前端视图

**文件**：`static/src/views/mining/sessions.vue`（或 `sessions/index.vue`）

```vue
<template>
  <div class="mining-sessions">
    <h2>采矿会话</h2>
    <!-- 你的页面内容 -->
  </div>
</template>

<script setup lang="ts">
// 页面逻辑
</script>
```

> **注意**：前端 `ComponentLoader` 通过 `import.meta.glob('../../views/**/*.vue')` 自动发现视图文件，组件路径与后端菜单中的 `Component` 字段对应。例如后端定义 `Component: "/mining/sessions"` → 前端查找 `views/mining/sessions.vue` 或 `views/mining/sessions/index.vue`。

### 步骤 9：前端 API 层

**文件**：`static/src/api/mining.ts`

```typescript
import request from '@/utils/http'

export function createMiningSession(data: Api.Mining.CreateSessionParams) {
  return request.post<Api.Mining.Session>({
    url: '/api/v1/mining/sessions',
    data
  })
}

export function fetchMiningSessionList(params?: Partial<Api.Common.CommonSearchParams>) {
  return request.get<Api.Mining.SessionList>({
    url: '/api/v1/mining/sessions',
    params
  })
}
```

### 步骤 10：前端类型定义

**文件**：`static/src/types/api/api.d.ts`

在 `Api` 命名空间中添加：

```typescript
namespace Mining {
  interface Session {
    id: number
    user_id: number
    character_id: number
    character_name: string
    solar_system_id: number
    ore_type_id: number
    quantity: number
    status: string
    start_at: string
    end_at?: string
    created_at: string
  }

  interface CreateSessionParams {
    character_id: number
    solar_system_id: number
    ore_type_id: number
    start_at: string
  }

  type SessionList = Api.Common.PaginatingQueryRecord<Session>
}
```

### 步骤 11：前端路由模块（仅前端模式需要）

如果需要同时支持前端模式路由，在 `static/src/router/modules/mining.ts` 添加：

```typescript
import { AppRouteRecord } from '@/types/router'

export const miningRoutes: AppRouteRecord = {
  path: '/mining',
  name: 'Mining',
  component: '/index/index',
  meta: { title: 'menus.mining.title', icon: 'ri:hammer-line' },
  children: [
    {
      path: 'sessions',
      name: 'MiningSessions',
      component: '/mining/sessions',
      meta: { title: 'menus.mining.sessions', keepAlive: true }
    }
  ]
}
```

并在 `static/src/router/modules/index.ts` 中导入并添加到 `routeModules` 数组。

### 步骤 12：国际化（可选）

在 `static/src/locales/langs/` 的对应语言文件中添加菜单翻译 key：

```json
{
  "menus": {
    "mining": {
      "title": "采矿管理",
      "sessions": "采矿会话"
    }
  }
}
```

### 流程总结清单

| # | 位置 | 文件 | 操作 |
|---|------|------|------|
| 1 | 后端 Model | `model/mining.go` | 定义数据模型 |
| 2 | 后端 Bootstrap | `bootstrap/db.go` | 添加 AutoMigrate |
| 3 | 后端 Repository | `repository/mining.go` | 实现数据访问 |
| 4 | 后端 Service | `service/mining.go` | 实现业务逻辑 |
| 5 | 后端 Handler | `handler/mining.go` | 实现 HTTP 处理器 |
| 6 | 后端 Router | `router/router.go` | 注册 API 路由 + 中间件 |
| 7 | 后端 Menu | `model/menu.go` | 添加菜单定义（指定 requiredRole） |
| 8 | 前端 View | `views/mining/sessions.vue` | 创建页面 |
| 9 | 前端 API | `api/mining.ts` | 创建 API 调用 |
| 10 | 前端 Types | `types/api/api.d.ts` | 定义接口类型 |
| 11 | 前端 Router | `router/modules/mining.ts` | 前端模式路由（可选） |
| 12 | 前端 i18n | `locales/langs/*.json` | 菜单翻译（可选） |

---

## 6. 附录

### 6.1 新增角色

如需添加新角色，修改 `server/internal/model/role.go`：

```go
const RoleNewRole = "new_role"

var rolePriority = map[string]int{
    // ... 在合适的优先级位置插入
    RoleNewRole: 35,
}
```

同时更新前端 `auth.ts` 中的 `ROLE_MAP`：
```typescript
const ROLE_MAP: Record<string, string> = {
    // ...
    new_role: 'R_NEW_ROLE',
}
```

### 6.2 新增 ESI 刷新任务

在 `server/pkg/eve/esi/` 下创建 `task_xxx.go`：

```go
package esi

import (
    "amiya-eden/internal/model"
    "context"
)

type xxxTask struct{}

func init() {
    Register(&xxxTask{})  // init() 自动注册
}

func (t *xxxTask) Name() string        { return "xxx" }
func (t *xxxTask) Description() string { return "描述" }
func (t *xxxTask) Priority() TaskPriority { return PriorityNormal }
func (t *xxxTask) Interval() TaskInterval {
    return TaskInterval{Active: 1 * time.Hour, Inactive: 24 * time.Hour}
}
func (t *xxxTask) RequiredScopes() []TaskScope {
    return []TaskScope{{Scope: "esi-xxx.read_xxx.v1", Description: "描述"}}
}
func (t *xxxTask) Execute(ctx context.Context, client *Client, char *model.EveCharacter) error {
    // 实现刷新逻辑
    return nil
}
```

新增的 ESI 数据表需要在 `model/esi_data.go` 中定义并在 `bootstrap/db.go` 中注册 AutoMigrate。

task 文件放在 `pkg/eve/esi/` 目录下即可，`jobs/jobs.go` 中的 `_ "amiya-eden/pkg/eve/esi"` 匿名导入会确保所有 `init()` 被调用。

### 6.3 统一响应格式

所有 API 响应遵循统一格式：

```json
{
    "code": 200,
    "msg": "success",
    "data": { ... }
}
```

| code | 含义 |
|------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 业务错误 |

Handler 中使用：
```go
response.OK(c, data)
response.Fail(c, response.CodeBizError, "错误信息")
```

### 6.4 中间件执行顺序

```
请求 →  RequestID → OperationLog → ResponseWrapper → ZapLogger → ZapRecovery → Cors → Handler
响应 ←  Cors → ZapRecovery → ZapLogger → ResponseWrapper(写biz_code) → OperationLog(读biz_code存DB)
```

### 6.5 配置文件结构

`server/config/config.yaml` 主要配置项：

```yaml
server:
  port: "8080"
  mode: "debug"          # debug / release / test

database:
  host: "127.0.0.1"
  port: 3306
  dbname: "amiya_eden"

redis:
  addr: "127.0.0.1:6379"

jwt:
  secret: "your-secret"
  expire_day: 7

eve_sso:
  client_id: "..."
  client_secret: "..."
  callback_url: "http://localhost:8080/api/v1/sso/eve/callback"

sde:
  api_key: "..."
  proxy: ""              # 可选，SDE 下载代理
```
