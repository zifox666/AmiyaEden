# 前端路由迁移计划（移除后端菜单系统）

**创建日期**: 2026-03-26
**最后更新**: 2026-03-27
**状态**: In Progress
**预计工期**: 1 天
**优先级**: High

---

## 迁移进度

### 阶段零：后端路由同步到前端 ✅ 已完成

- [x] 提取后端菜单数据
- [x] 分析后端菜单结构和权限
- [x] 对比前端现有路由
- [x] 创建或更新前端路由模块
- [x] 验证路由迁移完整性

**更新内容**：

#### [shop.ts](../../static/src/router/modules/shop.ts)
- ✅ 为 `ShopManage` 路由添加按钮权限（4 个按钮）
  - 新增商品
  - 编辑商品
  - 删除商品
  - 审批订单

#### [info.ts](../../static/src/router/modules/info.ts)
- ✅ 添加缺失的路由（2 个）
  - `EveInfoAssets` (/info/assets)
  - `EveInfoContracts` (/info/contracts)

#### [system.ts](../../static/src/router/modules/system.ts)
- ✅ 添加缺失的路由（2 个）
  - `RoleManage` (/system/role)
  - `PAPExchange` (/system/pap-exchange)
- ✅ 添加按钮权限（7 个页面，16 个按钮）
  - `User`: 删除用户、分配角色
  - `RoleManage`: 新增角色、编辑角色、删除角色、权限设置
  - `ESIRefresh`: 执行任务
  - `SystemWallet`: 调整余额、查看日志
  - `AlliancePAP`: 手动拉取
  - `PAPExchange`: 编辑兑换率
  - `Menus`: 新增、编辑、删除（已存在）

#### [srp.ts](../../static/src/router/modules/srp.ts)
- ✅ 添加按钮权限（2 个页面，4 个按钮）
  - `SrpManage`: 审批
  - `SrpPrices`: 新增价格、删除价格

#### [dashboard.ts](../../static/src/router/modules/dashboard.ts)
- ✅ 配置正确，无需修改

#### [operation.ts](../../static/src/router/modules/operation.ts)
- ✅ 权限配置正确，无需修改

#### [skill-planning.ts](../../static/src/router/modules/skill-planning.ts)
- ✅ 权限配置正确，无需修改

#### [welfare.ts](../../static/src/router/modules/welfare.ts)
- ✅ 权限配置正确，无需修改

#### [index.ts](../../static/src/router/modules/index.ts)
- ✅ 修正路由模块顺序，与后端菜单 Sort 值保持一致
  - 将 `srpRoutes` 从最后移到 `systemRoutes` 之前
  - 正确顺序：dashboard → operation → skill-planning → info → welfare → shop → srp → system → exception

#### [exception.ts](../../static/src/router/modules/exception.ts)
- ✅ 修复异常页面路由显示在菜单中的问题
  - 为 `exceptionRoutes` 添加 `isHide: true` 属性
  - 异常页面（403、404、500）不应出现在前端菜单中

**总计**：
- 新增路由：4 个
- 新增按钮权限：20 个
- 修正路由顺序：1 处
- 修复路由显示问题：1 处
- 涉及文件：7 个

---

## 一、背景与目标

### 1.1 当前问题

- 项目同时支持前端静态路由和后端动态菜单两种模式，但实际只使用后端模式
- 后端菜单系统存在数据一致性风险（种子数据 vs 数据库）
- 每次启动需要同步 ~200 次数据库操作，影响启动性能
- 没有动态菜单需求，没有 SaaS 多租户需求，几乎不用后台菜单管理功能
- PAPExchange 等菜单在数据库中持久化后无法自动清理

### 1.2 迁移目标

- 移除后端菜单系统，简化架构
- 只使用前端静态路由模式
- 提升系统性能（启动时间、菜单加载时间）
- 降低维护成本
- 消除数据一致性风险

---

## 二、前后端路由对比分析

### 2.1 前端路由

| 特性 | 描述 |
|------|------|
| **数据源** | `static/src/router/modules/*.ts`（10 个模块） |
| **类型安全** | TypeScript 编译时检查 |
| **数据一致性** | 单一数据源，无一致性风险 |
| **权限控制** | `meta.roles` + `meta.login` |
| **按钮权限** | `meta.authList` + `v-auth` 指令 |
| **维护成本** | 修改文件 + 重新部署 |
| **启动开销** | 无数据库查询 |
| **运行时开销** | 纯内存操作 |
| **动态调整** | 需要重新部署 |
| **测试难度** | 纯前端测试，无需数据库 |

### 2.2 后端路由

| 特性 | 描述 |
|------|------|
| **数据源** | `menu` 表 + `GetSystemMenuSeeds()` |
| **类型安全** | Go 编译时检查 |
| **数据一致性** | 需要同步（种子数据 ↔ 数据库） |
| **权限控制** | 角色-菜单关联表 |
| **按钮权限** | `type=button` + `permission` |
| **维护成本** | 修改代码 + 同步到数据库 |
| **启动开销** | 每次启动需要同步菜单（~200 次数据库操作） |
| **运行时开销** | 每次登录查询数据库 |
| **动态调整** | 可以后台调整（但几乎不用） |
| **测试难度** | 需要数据库环境 |

### 2.3 对比结论

前端路由在后端模式不使用的情况下，优势明显：
- 简洁性：单一数据源，无同步开销
- 性能：无数据库查询，启动和运行更快
- 维护：修改路由配置更简单，无需考虑数据库同步
- 测试：纯前端测试，环境搭建更容易

---

## 三、迁移阶段详解

### 阶段零：后端路由同步到前端 ✅ 已完成

后端菜单数据已完全同步到前端路由配置，包括所有路由、权限和按钮权限。

---

### 阶段一：验证前端模式可行性（1-2 天） ✅ 已完成

- [x] 修改环境变量（VITE_ACCESS_MODE = frontend）
- [x] 重启前端服务
- [x] 测试登录流程
- [x] 验证不同角色用户看到的菜单正确
- [x] 测试路由跳转功能
- [x] 验证权限验证功能
- [x] 测试按钮权限
- [x] 测试异常页面
- [x] 验证刷新页面后状态保持
- [x] 检查并补充前端路由配置

**验证结果**：
- ✅ 前端模式功能完整
- ✅ 所有角色菜单显示正确
- ✅ 路由和权限验证正常
- ✅ 按钮权限生效
- ✅ 无需后端菜单系统支持

**完成日期**：2026-03-27

---

### 阶段二：删除后端菜单相关代码（2-3 天） 🔄 进行中

- [x] 后端文件删除
- [x] 后端代码修改（router、db、role）
- [ ] 前端文件删除
- [ ] 前端代码修改
- [ ] 前端依赖关系检查
- [ ] 前端编译测试

**完成内容**：

#### 后端文件删除（5个文件）

- ✅ [model/menu.go](../../server/internal/model/menu.go) - 菜单模型和种子数据（235 行）
- ✅ [repository/menu.go](../../server/internal/repository/menu.go) - 菜单数据访问层（217 行）
- ✅ [service/menu.go](../../server/internal/service/menu.go) - 菜单业务逻辑层（178 行）
- ✅ [service/menu_test.go](../../server/internal/service/menu_test.go) - 菜单测试文件
- ✅ [handler/menu.go](../../server/internal/handler/menu.go) - 菜单HTTP处理器（167 行）

#### 后端代码修改（5个文件）

- ✅ [router.go](../../server/internal/router/router.go)
  - 删除菜单 Handler 初始化
  - 删除 7 个菜单相关 API 路由：
    - `GET /auth/menu/list` - 获取当前用户菜单
    - `GET /admin/menu/tree` - 获取菜单树
    - `POST /admin/menu` - 创建菜单
    - `PUT /admin/menu/:id` - 更新菜单
    - `DELETE /admin/menu/:id` - 删除菜单
    - `GET /admin/role/:id/menus` - 获取角色菜单
    - `PUT /admin/role/:id/menus` - 设置角色菜单

- ✅ [db.go](../../server/bootstrap/db.go)
  - 删除 `&model.Menu{}` 表迁移
  - 删除 `&model.RoleMenu{}` 表迁移
  - 删除 `roleSvc.SeedSystemMenus()` 种子数据初始化
  - 更新注释：种子数据流程简化

- ✅ [handler/role.go](../../server/internal/handler/role.go)
  - 删除 `GetRoleMenus` 方法
  - 删除 `SetRoleMenus` 方法

- ✅ [repository/role.go](../../server/internal/repository/role.go)
  - 删除 `Delete` 方法中的 `RoleMenu` 删除逻辑
  - 删除 `GetRoleMenuIDs` 方法
  - 删除 `SetRoleMenus` 方法
  - 删除 `GetMenuIDsByRoles` 方法

- ✅ [model/role.go](../../server/internal/model/role.go)
  - 删除 `Role` 结构体的 `MenuIDs` 字段
  - 删除 `RoleMenu` 结构体定义
  - 删除 `RoleMenu.TableName()` 方法

#### 后端编译测试

- ✅ 后端编译成功，无错误
- ✅ 后端代码中已无任何 menu/Menu 相关引用

**后端完成日期**：2026-03-27

#### 2.2 前端删除清单

#### 1.1 修改环境变量

```bash
# static/.env.development
VITE_ACCESS_MODE = frontend

# static/.env.production
VITE_ACCESS_MODE = frontend
```

#### 1.2 重启前端服务

```bash
pnpm dev
```

#### 1.3 测试清单

```
□ 登录流程正常
  - 打开登录页
  - EVE SSO 登录
  - 回调处理正确
  - 登录成功后跳转到首页

□ 不同角色用户看到的菜单正确
  - super_admin: 看到所有菜单
  - admin: 看到管理菜单（除用户模拟登录）
  - fc: 看到舰队管理、技能计划
  - user: 看到用户菜单（舰队、商店、福利等）
  - guest: 只看到基础菜单（仪表盘、角色探索）

□ 路由跳转正常
  - 点击侧边栏菜单
  - 手动输入 URL
  - 刷新页面后保持状态

□ 权限验证正常
  - 无权限用户访问受限页面跳转到 403
  - 超级管理员可以访问所有页面
  - meta.login 保护的路由需要登录

□ 按钮权限正常
  - v-auth 指令正确隐藏/显示按钮

□ 异常页面正常
  - 404 页面（访问不存在的路由）
  - 403 页面（无权限访问）
  - 500 页面（服务器错误）

□ 刷新页面后状态保持
  - 用户登录状态保持
  - 菜单保持展开/折叠状态
  - 工作标签页保持
```

#### 1.4 检查并补充前端路由配置

**需要检查的文件**：

- `static/src/router/modules/system.ts`
- `static/src/router/modules/operation.ts`
- `static/src/router/modules/skill-planning.ts`
- `static/src/router/modules/info.ts`
- `static/src/router/modules/welfare.ts`
- `static/src/router/modules/shop.ts`
- `static/src/router/modules/srp.ts`
- `static/src/router/modules/role.ts`
- `static/src/router/modules/exception.ts`
- `static/src/router/modules/dashboard.ts`

**检查要点**：
1. `meta.roles` 是否正确设置（哪些路由需要特定角色）
2. `meta.login` 是否正确设置（哪些路由需要登录）
3. 路由路径是否与后端 API 对应
4. 组件路径是否正确
5. 是否有遗漏的路由（对比后端 GetSystemMenuSeeds）

---

### 阶段二：删除后端菜单相关代码（2-3 天）

#### 2.1 后端文件删除清单

##### 完整删除的文件

```
server/internal/model/menu.go                    # 删除 (235 行)
server/internal/repository/menu.go                # 删除 (217 行)
server/internal/service/menu.go                   # 删除 (178 行)
server/internal/service/menu_test.go              # 删除 (测试文件)
server/internal/handler/menu.go                   # 删除 (167 行)
```

##### 需要修改的文件

**1. `server/internal/router/router.go`**

删除以下代码块：

```go
// 删除第 77-78 行：菜单 Handler 初始化和路由
menuH := handler.NewMenuHandler()
auth.GET("/menu/list", menuH.GetMenuList) // 当前用户可用菜单

// 删除第 286-289 行：菜单管理路由组
adminMenu := admin.Group("/menu")
{
    adminMenu.GET("/tree", menuH.GetMenuTree)
    adminMenu.POST("", menuH.CreateMenu)
    adminMenu.PUT("/:id", menuH.UpdateMenu)
    adminMenu.DELETE("/:id", menuH.DeleteMenu)
}
```

**2. `server/bootstrap/db.go`**

删除以下内容：

```go
// 在 autoMigrate() 函数中删除
&model.Menu{},
&model.RoleMenu{},

// 在 autoMigrate() 函数末尾删除
roleSvc.SeedSystemMenus()
```

**3. `server/internal/service/role.go`**

删除以下内容：

```go
// 删除结构体字段（第 20 行）
menuRepo *repository.MenuRepository

// 删除 NewRoleService 中的初始化（第 32 行）
menuRepo: repository.NewMenuRepository(),

// 删除 SeedSystemMenus 方法（第 357-398 行）
// 删除 removeObsoleteSystemMenus 方法（第 404-422 行）
// 删除 seedDefaultRoleMenus 方法（第 424-483 行）
// 删除 seedAdminMenus 方法（第 544-581 行）
// 删除 reconcileGuestMenuRestrictions 方法（第 486-541 行）

// 删除 GetUserPermissions 方法中的菜单相关逻辑（第 89-95 行）
// 删除所有 menuRepo 的调用
```

#### 2.3 前端代码修改

##### 完全删除的文件和目录

```
static/src/views/system/menu/                # 删除整个目录（包括子模块）
  ├── index.vue                              # 菜单管理主页面
  └── modules/
      └── menu-dialog.vue                    # 菜单编辑对话框
```

##### 需要修改的文件

**1. `static/src/api/system-manage.ts`**

删除以下函数：

```typescript
// 删除菜单管理相关 API
export function fetchGetMenuTree() { ... }
export function fetchCreateMenu(data) { ... }
export function fetchUpdateMenu(id, data) { ... }
export function fetchDeleteMenu(id) { ... }

// 删除用户菜单 API
export function fetchGetMenuList() { ... }

// 删除角色菜单相关 API
export function fetchGetRoleMenus(roleId) { ... }
export function fetchSetRoleMenus(roleId, menuIds) { ... }
```

**2. `static/src/router/modules/system.ts`**

删除菜单管理路由（如果还存在）：

```typescript
// 删除这个路由配置
{
  path: 'menu',
  name: 'Menus',
  component: '/system/menu',
  meta: {
    title: 'menus.system.menu',
    keepAlive: true,
    roles: ['super_admin', 'admin']
  }
}
```

**3. `static/src/store/modules/menu.ts`**

如果 store 中有从后端获取菜单的逻辑，需要删除：

```typescript
// 删除或注释掉从后端获取菜单的代码
// 如果该 store 仅用于前端模式，则保持不变
```

**4. 其他可能引用菜单 API 的文件**

需要检查并清理以下文件中的菜单相关引用：

```
static/src/router/core/RouteValidator.ts
static/src/router/core/RoutePermissionValidator.ts
static/src/router/core/MenuProcessor.ts
static/src/router/guards/beforeEach.ts
static/src/types/api/api.d.ts
static/src/locales/langs/zh.json
static/src/locales/langs/en.json
```

#### 2.4 前端依赖关系检查

在删除之前，需要确认以下依赖关系：

##### 后端依赖检查

```bash
# 检查是否有其他服务引用 MenuService
cd server
grep -r "MenuService" --include="*.go" .

# 检查是否有其他服务引用 MenuRepository
grep -r "MenuRepository" --include="*.go" .

# 检查是否有其他地方引用 Menu 模型
grep -r "model\.Menu" --include="*.go" .
```

##### 前端依赖检查

```bash
# 检查是否有其他地方引用菜单 API
cd static
grep -r "fetchGetMenuList\|fetchGetMenuTree" --include="*.ts" --include="*.vue" src/

# 检查是否有其他地方引用菜单管理页面
grep -r "/system/menu" --include="*.ts" --include="*.vue" src/
```

#### 2.5 前端删除步骤建议

1. **第一步：删除前端文件**
   - 删除菜单管理页面目录
   - 删除 API 函数
   - 删除路由配置
   - 检查并清理其他引用

2. **第二步：前端依赖关系检查**
   - 检查是否有其他地方引用菜单 API
   - 检查是否有其他地方引用菜单管理页面

3. **第三步：前端编译测试**
   ```bash
   cd static
   pnpm build
   ```

---

### 阶段三：数据库清理（1 天）

#### 3.1 数据库备份

##### 完整备份

```bash
# 备份整个数据库
pg_dump -U username -h localhost -d amiya_eden > backup_$(date +%Y%m%d_%H%M%S).sql

# 或者只备份表结构
pg_dump -U username -h localhost -d amiya_eden --schema-only > schema_$(date +%Y%m%d_%H%M%S).sql
```

##### 备份关键表数据

```sql
-- 备份角色表（确保角色数据安全）
COPY roles TO '/tmp/roles_backup.csv' CSV HEADER;

-- 备份用户-角色关联表
COPY user_role TO '/tmp/user_role_backup.csv' CSV HEADER;
```

#### 3.2 清理菜单相关数据

##### 删除菜单表和关联表

```sql
-- 开启事务（确保操作可回滚）
BEGIN;

-- 查看当前表数据（确认删除前状态）
SELECT COUNT(*) as menu_count FROM menu;
SELECT COUNT(*) as role_menu_count FROM role_menu;

-- 删除角色-菜单关联表
DROP TABLE IF EXISTS role_menu;

-- 删除菜单表
DROP TABLE IF EXISTS menu;

-- 提交事务
COMMIT;
```

##### 验证删除成功

```sql
-- 确认删除成功
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('menu', 'role_menu');
-- 应该返回空结果

-- 检查是否还有外键引用
SELECT
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND (ccu.table_name = 'menu' OR ccu.table_name = 'role_menu');
-- 应该返回空结果
```

#### 3.3 验证其他表完整性

##### 验证核心表

```sql
-- 确认用户表正常
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM users WHERE status = 1;

-- 确认角色表正常
SELECT COUNT(*) FROM role;
SELECT code, name FROM role WHERE is_system = true;

-- 确认用户-角色关联表正常
SELECT COUNT(*) FROM user_role;
```

#### 3.4 回滚方案

如果需要回滚数据库更改：

```sql
-- 恢复菜单表结构（从备份）
psql -U username -h localhost -d amiya_eden -f backup_YYYYMMDD_HHMMSS.sql

-- 或者恢复关键表数据
COPY roles FROM '/tmp/roles_backup.csv' CSV HEADER;
COPY user_role FROM '/tmp/user_role_backup.csv' CSV HEADER;

-- 如果菜单表被删除，需要重新创建并同步种子数据
-- 重启后端服务，会自动执行 SeedSystemMenus()
```

---

### 阶段四：调整前端权限验证逻辑（1 天）

#### 3.1 前端权限指令验证

**目标**：确认 `v-auth` 指令正确使用前端路由的 `meta.authList`

```typescript
// static/src/directives/auth.ts
// 确认指令实现逻辑
import { useUserStore } from '@/store/modules/user'

const vAuth = {
  mounted(el, binding) {
    const { value } = binding
    const userStore = useUserStore()
    const userRoles = userStore.roles || []
    
    // 检查用户是否具有所需权限
    // 权限应该来自前端路由的 meta.authList
    if (!hasPermission(userRoles, value)) {
      el.parentNode?.removeChild(el)
    }
  }
}

// 示例用法：<button v-auth="'user:create'">创建用户</button>
```

#### 3.2 前端路由守卫调整

**目标**：确认路由守卫使用前端路由的 `meta.roles` 和 `meta.login`

```typescript
// static/src/router/guards/beforeEach.ts
import { useUserStore } from '@/store/modules/user'

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  
  // 检查是否需要登录
  if (to.meta.login && !userStore.isLogin) {
    next('/login')
    return
  }
  
  // 检查角色权限
  if (to.meta.roles && to.meta.roles.length > 0) {
    const userRoles = userStore.roles || []
    const hasRole = to.meta.roles.some(role => 
      userRoles.includes(role)
    )
    
    if (!hasRole) {
      next('/403')
      return
    }
  }
  
  next()
})
```

#### 3.3 前端权限验证核心模块检查

**需要检查的文件**：

1. **`static/src/router/core/RoutePermissionValidator.ts`**
   ```typescript
   // 确认使用前端路由的 meta.roles 进行验证
   // 删除或注释掉从后端获取权限的逻辑
   ```

2. **`static/src/router/core/MenuProcessor.ts`**
   ```typescript
   // 确认处理的是前端路由配置
   // 删除从后端 API 获取菜单的逻辑
   ```

3. **`static/src/store/modules/user.ts`**
   ```typescript
   // 确认用户信息中包含 roles 字段
   // roles 应该来自登录时后端返回的用户信息
   // 不应该从菜单 API 获取权限
   ```

#### 3.4 后端权限验证调整

**目标**：确认后端只验证用户角色，不再验证菜单权限

**需要检查的文件**：

1. **`server/internal/middleware/jwt.go`**
   ```go
   // 确认 JWT 中间件只验证用户身份和角色
   // 不再从 menu 表获取权限
   ```

2. **`server/internal/middleware/role.go`**
   ```go
   // 确认角色验证中间件只检查 user_role 表
   // 不再检查 role_menu 表
   ```

3. **`server/internal/service/role.go`**
   ```go
   // GetUserPermissions 方法应该返回角色权限
   // 不应该从菜单表获取按钮权限
   
   // 可能的实现方案：
   // 1. 直接返回角色编码列表
   // 2. 或者维护一个角色到权限的映射表
   ```

#### 3.5 权限验证逻辑调整建议

##### 前端权限验证

| 场景 | 验证方式 | 数据来源 |
|------|---------|---------|
| 路由访问权限 | 路由守卫检查 `meta.roles` | 前端路由配置 + 用户角色 |
| 按钮显示权限 | `v-auth` 指令检查 `meta.authList` | 前端路由配置 + 用户角色 |
| 页面元素权限 | `v-if` + 用户角色判断 | 用户角色 |

##### 后端权限验证

| 场景 | 验证方式 | 数据来源 |
|------|---------|---------|
| API 访问权限 | 中间件检查用户角色 | `user_role` 表 |
| 敏感操作权限 | 中间件检查特定角色 | `user_role` 表 |
| 资源访问权限 | 业务逻辑检查用户角色 | `user_role` 表 |

#### 3.6 需要修改的代码清单

##### 后端修改

```go
// server/internal/service/role.go
// 修改 GetUserPermissions 方法
func (s *RoleService) GetUserPermissions(ctx context.Context, userID uint) ([]string, error) {
    cacheKey := fmt.Sprintf("%s%d", userPermsCachePrefix, userID)
    val, err := global.Redis.Get(ctx, cacheKey).Result()
    if err == nil && val != "" {
        var perms []string
        if json.Unmarshal([]byte(val), &perms) == nil {
            return perms, nil
        }
    }

    // 获取用户角色
    roleCodes, err := s.repo.GetUserRoleCodes(userID)
    if err != nil {
        return nil, err
    }

    // TODO: 定义角色到权限的映射
    // 方案1: 直接返回角色编码
    // 方案2: 维护一个权限映射表
    
    // 临时方案：直接返回角色编码
    if data, err := json.Marshal(roleCodes); err == nil {
        global.Redis.Set(ctx, cacheKey, string(data), cacheTTL)
    }
    return roleCodes, nil
}
```

##### 前端修改

```typescript
// static/src/router/core/RoutePermissionValidator.ts
// 确认只使用前端路由配置
export class RoutePermissionValidator {
  validate(route: AppRouteRecord, userRoles: string[]): boolean {
    if (!route.meta?.roles || route.meta.roles.length === 0) {
      return true
    }
    
    return route.meta.roles.some(role => userRoles.includes(role))
  }
}
```

---

### 阶段八：验证与监控（持续）

#### 8.1 功能验证

```
□ 用户登录/登出流程
□ 所有菜单访问和权限控制
□ 不同角色的菜单显示正确
□ 路由跳转和页面状态保持
□ 按钮权限控制（v-auth 指令）
□ 异常页面（404、403、500）
```

#### 8.2 性能监控

```
□ 前端启动时间
□ 菜单加载时间
□ 页面切换速度
□ API 响应时间
□ 数据库查询次数
```

#### 8.3 日志监控

```
□ 后端日志异常
□ 前端控制台错误
□ 权限验证失败日志
□ 性能慢查询
```

#### 8.4 用户反馈

```
□ 收集用户反馈
□ 监控问题报告
□ 分析用户行为
□ 持续优化体验
```

#### 4.1 数据库备份

##### 完整备份

```bash
# 备份整个数据库
pg_dump -U username -h localhost -d amiya_eden > backup_$(date +%Y%m%d_%H%M%S).sql

# 或者只备份表结构
pg_dump -U username -h localhost -d amiya_eden --schema-only > schema_$(date +%Y%m%d_%H%M%S).sql
```

##### 备份关键表数据

```sql
-- 备份角色表（确保角色数据安全）
COPY roles TO '/tmp/roles_backup.csv' CSV HEADER;

-- 备份用户-角色关联表
COPY user_role TO '/tmp/user_role_backup.csv' CSV HEADER;
```

#### 4.2 清理菜单相关数据

##### 删除菜单表和关联表

```sql
-- 开启事务（确保操作可回滚）
BEGIN;

-- 查看当前表数据（确认删除前状态）
SELECT COUNT(*) as menu_count FROM menu;
SELECT COUNT(*) as role_menu_count FROM role_menu;

-- 删除角色-菜单关联表
DROP TABLE IF EXISTS role_menu;

-- 删除菜单表
DROP TABLE IF EXISTS menu;

-- 提交事务
COMMIT;
```

##### 验证删除成功

```sql
-- 确认删除成功
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('menu', 'role_menu');
-- 应该返回空结果

-- 检查是否还有外键引用
SELECT
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
  AND (ccu.table_name = 'menu' OR ccu.table_name = 'role_menu');
-- 应该返回空结果
```

#### 4.3 验证其他表完整性

##### 验证核心表

```sql
-- 确认用户表正常
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM users WHERE status = 1;

-- 确认角色表正常
SELECT COUNT(*) FROM roles;
SELECT code, name FROM roles ORDER BY sort DESC;

-- 确认用户-角色关联表正常
SELECT COUNT(*) FROM user_role;
SELECT user_id, role_id FROM user_role LIMIT 10;
```

##### 验证业务表

```sql
-- 验证舰队表
SELECT COUNT(*) FROM fleets;

-- 验证商店表
SELECT COUNT(*) FROM shop_products;

-- 验证 SRP 表
SELECT COUNT(*) FROM srp_applications;
```

#### 4.4 清理 Redis 缓存

```bash
# 连接到 Redis
redis-cli

# 清理用户角色缓存（包含菜单权限缓存）
KEYS user_roles:*
KEYS user_perms:*

# 批量删除缓存
EVAL "return redis.call('del', unpack(redis.call('keys', 'user_roles:*')))" 0
EVAL "return redis.call('del', unpack(redis.call('keys', 'user_perms:*')))" 0

# 或者清空所有缓存（谨慎操作）
FLUSHDB
```

#### 4.5 回滚方案

如果删除后发现异常，可以按以下步骤回滚：

##### 从备份恢复

```bash
# 恢复完整备份
psql -U username -h localhost -d amiya_eden < backup_20260327_120000.sql
```

##### 从 CSV 恢复

```sql
-- 恢复角色表
COPY roles FROM '/tmp/roles_backup.csv' CSV HEADER;

-- 恢复用户-角色关联表
COPY user_role FROM '/tmp/user_role_backup.csv' CSV HEADER;

-- 如果菜单表被删除，需要重新创建并同步种子数据
-- 重启后端服务，会自动执行 SeedSystemMenus()
```

---

### 阶段五：全面测试（2-3 天）

#### 5.1 功能测试

```
□ 用户登录/登出
□ 所有菜单访问
□ 权限验证（不同角色）
□ 按钮权限控制
□ 页面缓存（keepAlive）
□ 路由跳转
□ 刷新页面状态保持
```

#### 5.2 性能测试

```
□ 启动时间（后端）
□ 菜单加载时间（前端）
□ 页面切换速度
□ 数据库查询次数
```

#### 5.3 兼容性测试

```
□ 不同浏览器
□ 不同角色
□ 不同权限组合
```

---

### 阶段六：文档更新（1 天）

#### 6.1 更新开发文档

- API 文档：删除菜单相关 API
- 路由文档：更新为前端路由
- 权限文档：更新为基于角色的权限控制

#### 6.2 更新部署文档

- 环境变量：VITE_ACCESS_MODE
- 数据库迁移：删除菜单表创建脚本

#### 6.3 更新运维文档

- 故障排查：菜单相关部分删除
- 日志监控：删除菜单相关日志监控

---

### 阶段七：部署上线（1 天）

#### 7.1 测试环境部署

```bash
# 后端
cd server
git pull
go build -o amiya-eden
./amiya-eden

# 前端
cd static
git pull
pnpm install
pnpm build
```

#### 7.2 生产环境部署

```bash
# 备份数据库
pg_dump -U username -d database_name > backup_$(date +%Y%m%d).sql

# 部署后端
cd server
git pull
go build -o amiya-eden
./amiya-eden

# 部署前端
cd static
git pull
pnpm install
pnpm build
```

#### 7.3 验证上线

```
□ 检查服务状态
□ 检查日志
□ 检查功能正常
□ 检查性能指标
```

---

## 四、风险与回滚

### 4.1 潜在风险

1. **权限配置错误**：前端路由配置可能遗漏某些权限
2. **角色映射错误**：后端角色到前端路由的映射可能不完整
3. **隐藏路由丢失**：某些隐藏路由可能被误删
4. **缓存问题**：前端路由缓存可能导致旧配置生效

### 4.2 回滚方案

1. **保留后端菜单代码**：在分支中保留完整后端菜单代码
2. **数据库备份**：删除菜单表前完整备份
3. **快速切换**：通过环境变量快速切换回后端模式
4. **版本回退**：使用 Git 回退到之前版本

---

## 五、总结

### 5.1 迁移收益

- **性能提升**：启动时间减少 ~2-3 秒（不需要同步菜单）
- **维护简化**：无需同步种子数据和数据库
- **风险降低**：消除数据一致性风险
- **架构清晰**：单一数据源，逻辑更清晰

### 5.2 迁移完成状态

- [x] 阶段零：后端路由同步到前端（1-2 天）✅ 已完成
  - ✅ 提取后端菜单数据（数据库或代码）
  - ✅ 分析后端菜单结构和权限
  - ✅ 对比前端现有路由
  - ✅ 创建或更新前端路由模块
  - ✅ 验证路由迁移完整性
  - ✅ 输出路由对比文档

- [x] 紧急修复：修复发现的问题（1 天）✅ 已完成
  - ✅ 修复 Info 模块路由顺序（NpcKills 位置错误）
  - ✅ 修复 System 模块路由顺序（多个路由位置错误）
  - ✅ 删除 system/menu 路由（后端菜单系统将被移除）
  - ✅ 为 SystemWallet 添加按钮权限（调整余额、查看日志）
  - ✅ 为 SrpManage 添加按钮权限（审批）
  - ✅ 为 SrpPrices 添加按钮权限（新增价格、删除价格）
  - ✅ 确认 PAPExchange 路由是否应该保留

- [x] 阶段零：后端路由同步到前端（1-2 天）✅ 已完成
- [x] 阶段一：验证前端模式可行性（1-2 天）✅ 已完成
- [ ] 阶段二：删除后端菜单相关代码（2-3 天）🔄 进行中
- [ ] 阶段三：数据库清理（1 天）
- [ ] 阶段四：调整前端权限验证逻辑（1 天）
- [ ] 阶段五：全面测试（2-3 天）
- [ ] 阶段六：文档更新（1 天）
- [ ] 阶段七：部署上线（1 天）
- [ ] 阶段八：验证与监控（持续）

---

## 六、参考文档

- [后端菜单模型](../../server/internal/model/menu.go)
- [前端路由配置](../../static/src/router/modules/)
- [权限验证中间件](../../server/internal/middleware/role.go)
- [前端路由守卫](../../static/src/router/guard.ts)
