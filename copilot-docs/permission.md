# 权限管理系统文档

## 一、概述

本系统采用 **RBAC（基于角色的访问控制）** 模型，通过 4 张核心表实现菜单级 + 按钮级权限控制。

- **多角色**：一个用户可拥有多个角色
- **菜单权限**：角色与菜单建立关联，控制可访问的页面
- **按钮权限**：菜单中 `type=button` 的节点作为操作权限标识
- **超级管理员**：`super_admin` 角色自动绕过所有权限检查

## 二、数据模型

### 2.1 数据库表

| 表名 | 说明 | 主要字段 |
|------|------|----------|
| `role` | 角色 | `code`, `name`, `description`, `is_system`, `sort`, `status` |
| `menu` | 菜单/按钮 | `parent_id`, `type`(dir/menu/button), `name`, `path`, `component`, `permission`, `title`, `icon`, `sort`, `status` |
| `role_menu` | 角色-菜单关联 | `role_id`, `menu_id` (联合主键) |
| `user_role` | 用户-角色关联 | `user_id`, `role_id` (联合主键) |

### 2.2 菜单类型

| 类型 | 说明 | 示例 |
|------|------|------|
| `dir` | 目录 | 系统管理、运营管理 |
| `menu` | 页面 | 用户管理、角色管理 |
| `button` | 操作权限 | `srp:price:edit`、`srp:review` |

### 2.3 系统预置角色

| Code | 名称 | Sort | 说明 |
|------|------|------|------|
| `super_admin` | 超级管理员 | 100 | 绕过所有权限检查，拥有全部菜单和按钮 |
| `admin` | 管理员 | 90 | 系统管理权限 |
| `srp` | SRP管理员 | 80 | SRP 审核与价格管理 |
| `fc` | FC | 70 | 舰队指挥官 |
| `user` | 已认证用户 | 10 | 基础功能访问 |
| `guest` | 访客 | 0 | 最小权限 |

## 三、后端架构

### 3.1 分层结构

```
server/internal/
├── model/        # 数据模型定义 + 种子数据
│   ├── role.go   # Role, RoleMenu, UserRole, 角色常量, 辅助函数
│   └── menu.go   # Menu, MenuItem(前端格式), MenuMeta, 种子数据
├── repository/   # 数据访问层
│   ├── role.go   # Role/RoleMenu/UserRole 增删改查
│   └── menu.go   # Menu 增删改查 + 树构建
├── service/      # 业务逻辑层
│   ├── role.go   # 角色CRUD, 权限缓存, 种子初始化
│   ├── menu.go   # 菜单CRUD, 用户菜单树
│   └── user.go   # 用户管理 (简化版)
├── handler/      # HTTP 处理层
│   ├── role.go   # 角色管理 + 角色-菜单 + 用户-角色
│   ├── menu.go   # 菜单管理 + 用户菜单
│   ├── user.go   # 用户列表/详情/编辑/删除
│   └── me.go     # 当前用户信息（含角色、权限）
├── middleware/
│   └── auth.go   # JWT认证, RequireRole, RequirePermission
└── router/
    └── router.go # 路由注册
```

### 3.2 中间件

#### JWTAuth
- 从 `Authorization: Bearer <token>` 或 `?token=` 提取 JWT
- 解析后加载用户角色（Redis 缓存 30min）和权限
- 写入 Gin Context（`userID`, `characterID`, `roles`, `permissions`）

#### RequireRole(codes ...string)
- 检查用户角色是否匹配任一指定角色
- `super_admin` 自动通过

#### RequirePermission(perms ...string)
- 检查用户是否拥有任一指定按钮权限
- `super_admin` 自动通过

### 3.3 缓存策略

| Key 格式 | TTL | 内容 |
|----------|-----|------|
| `user_roles:{userID}` | 30min | 用户所有角色 Code 列表 (JSON) |
| `user_perms:{userID}` | 30min | 用户所有按钮权限列表 (JSON) |

缓存失效时机：
- 修改用户角色 → 清除该用户缓存
- 修改角色菜单 → 清除该角色所有成员缓存

### 3.4 API 路由

#### 公开接口
| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/eve/sso/login` | SSO 登录 |
| GET | `/api/v1/eve/sso/callback` | SSO 回调 |

#### 需要登录
| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/me` | 获取当前用户信息（含角色+权限） |
| GET | `/api/v1/menu/list` | 获取用户菜单树（前端路由格式） |

#### 需要管理员 (admin)
| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/system/role` | 角色列表（分页） |
| POST | `/api/v1/system/role` | 创建角色 |
| GET | `/api/v1/system/role/:id` | 角色详情 |
| PUT | `/api/v1/system/role/:id` | 更新角色 |
| DELETE | `/api/v1/system/role/:id` | 删除角色 |
| GET | `/api/v1/system/role/:id/menus` | 获取角色菜单ID列表 |
| PUT | `/api/v1/system/role/:id/menus` | 设置角色菜单 |
| GET | `/api/v1/system/user` | 用户列表（分页） |
| GET | `/api/v1/system/user/:id` | 用户详情 |
| PUT | `/api/v1/system/user/:id` | 更新用户 |
| DELETE | `/api/v1/system/user/:id` | 删除用户 |
| GET | `/api/v1/system/user/:id/roles` | 获取用户角色 |
| PUT | `/api/v1/system/user/:id/roles` | 设置用户角色 |
| GET | `/api/v1/system/menu/tree` | 菜单树（管理用，含全部） |
| POST | `/api/v1/system/menu` | 创建菜单 |
| PUT | `/api/v1/system/menu/:id` | 更新菜单 |
| DELETE | `/api/v1/system/menu/:id` | 删除菜单 |

#### 按钮权限控制示例
| 接口 | 需要权限 |
|------|---------|
| `PUT /api/v1/srp/price` | `srp:price:edit` |
| `POST /api/v1/srp/review` | `srp:review` |

## 四、前端架构

### 4.1 权限数据流

```
用户登录 → GET /api/v1/me → { roles, permissions }
                    ↓
         Pinia userStore.info = {
           roles: ['admin', 'fc'],
           buttons: ['srp:price:edit', ...]
         }
                    ↓
    ┌───────────────┼───────────────┐
    ↓               ↓               ↓
路由守卫          v-auth指令      useAuth Hook
(meta.roles)    (按钮权限)      (编程式检查)
```

### 4.2 前端权限模式

通过 `VITE_ACCESS_MODE` 环境变量切换：

| 模式 | 路由来源 | 权限检查方式 |
|------|----------|-------------|
| `frontend` | 前端静态路由 | `meta.roles` 匹配用户角色 |
| `backend` | 后端 `/api/v1/menu/list` 返回 | 路由 `meta.authList` |

### 4.3 角色标识使用

路由 meta 中直接使用后端角色编码：

```typescript
// router/modules/system.ts
meta: {
  roles: ['super_admin', 'admin']
}
```

### 4.4 按钮权限指令

```vue
<ElButton v-auth="'srp:price:edit'">编辑价格</ElButton>
```

编程式检查：

```typescript
const { hasAuth } = useAuth()
if (hasAuth('srp:review')) { /* ... */ }
```

### 4.5 关键文件

| 文件 | 职责 |
|------|------|
| `api/auth.ts` | 获取用户信息并映射角色/权限 |
| `api/system-manage.ts` | 角色/菜单/用户管理 API |
| `types/api/api.d.ts` | 类型定义 |
| `store/modules/user.ts` | 用户状态存储 |
| `hooks/core/useAuth.ts` | 权限检查 Hook |
| `views/system/role/` | 角色管理页面（CRUD + 菜单权限分配） |
| `views/system/user/` | 用户管理页面（含多角色分配对话框） |

## 五、初始化流程

服务启动时 `bootstrap/db.go` 自动执行：

1. **AutoMigrate** — 创建/更新表结构（`Role`, `Menu`, `RoleMenu`, `UserRole`）
2. **SeedSystemRoles** — Upsert 6 个系统角色
3. **SeedSystemMenus** — Upsert 菜单树 + 绑定默认角色-菜单关系
4. **MigrateExistingUsers** — 将 `user.role` 字段迁移到 `user_role` 表

## 六、扩展指南

### 添加新角色

1. 后台管理界面 → 角色管理 → 新增角色
2. 或在 `model/role.go` 的 `SystemRoleSeeds` 中添加系统角色

### 添加新菜单

1. 后台管理界面 → 菜单管理 → 新增菜单
2. 或在 `model/menu.go` 的 `GetSystemMenuSeeds()` 中添加种子数据

### 添加新按钮权限

1. 在菜单管理中对应页面下新增 `type=button` 子节点，设置 `permission` 字段
2. 后端使用 `middleware.RequirePermission("permission_code")` 保护路由
3. 前端使用 `v-auth="'permission_code'"` 控制按钮显示
