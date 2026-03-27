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

### 阶段一：验证前端模式可行性（1-2 天）

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

#### 2.1 删除文件

```
server/internal/model/menu.go                    # 删除 (~250 行)
server/internal/repository/menu.go                # 删除 (~150 行)
server/internal/service/menu.go                   # 删除 (~200 行)
server/internal/service/menu_test.go              # 删除 (~100 行)
server/internal/handler/menu.go                   # 删除部分代码 (~170 行)
```

**注意**：`handler/menu.go` 需要检查是否有被其他地方引用。

#### 2.2 删除数据库相关代码

```go
// server/bootstrap/db.go
func InitializeDatabase() {
    // 删除这行
    roleSvc.SeedSystemMenus()  // ← 删除
}
```

#### 2.3 删除 API 路由

```go
// server/internal/router/router.go
// 删除以下路由
auth.GET("/menu/list", menuH.GetMenuList)  // ← 删除

adminMenu := admin.Group("/menu")           // ← 删除整个组
{
    adminMenu.GET("/tree", menuH.GetMenuTree)
    adminMenu.POST("", menuH.CreateMenu)
    adminMenu.PUT("/:id", menuH.UpdateMenu)
    adminMenu.DELETE("/:id", menuH.DeleteMenu)
}
```

#### 2.4 删除数据库表

```sql
-- 删除菜单表
DROP TABLE IF EXISTS menu;
DROP TABLE IF EXISTS role_menu;
```

#### 2.5 删除前端菜单管理相关页面

```
static/src/views/system/menu/                # 删除整个目录
static/src/api/menu.ts                       # 删除菜单 API
static/src/views/system/menu.vue             # 删除菜单管理页面
```

#### 2.6 更新前端路由

删除前端路由中的菜单管理路由：

```typescript
// static/src/router/modules/system.ts
// 删除这个路由
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

#### 2.7 更新权限验证逻辑

**前端路由守卫调整**：

```typescript
// static/src/router/guard.ts
// 确认不再从后端获取菜单，直接使用前端路由
```

**后端 JWT 中间件调整**：

```go
// server/internal/middleware/jwt.go
// 确认不再验证菜单权限，只验证角色权限
```

---

### 阶段三：调整前端权限验证逻辑（1 天）

#### 3.1 前端权限指令

```typescript
// static/src/directives/auth.ts
// 确认 v-auth 指令正确使用路由 meta.authList
```

#### 3.2 前端路由守卫

```typescript
// static/src/router/guard.ts
// 确认路由守卫使用 meta.roles 和 meta.login
```

#### 3.3 后端权限验证

```go
// server/internal/middleware/role.go
// 确认后端验证用户角色，不再验证菜单权限
```

---

### 阶段四：数据库清理（1 天）

#### 4.1 备份数据库

```bash
# 备份整个数据库
pg_dump -U username -d database_name > backup.sql
```

#### 4.2 清理菜单相关数据

```sql
-- 删除菜单表和关联表
DROP TABLE IF EXISTS role_menu;
DROP TABLE IF EXISTS menu;

-- 确认删除成功
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN ('menu', 'role_menu');
```

#### 4.3 验证其他表完整性

```sql
-- 确认用户、角色等核心表正常
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM roles;
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

- [ ] 验证通过后：阶段一 - 验证前端模式可行性（1-2 天）
- [ ] 验证通过后：阶段二 - 删除后端菜单相关代码（2-3 天）
- [ ] 同步进行：阶段三 - 调整前端权限验证逻辑（1 天）
- [ ] 可选执行：阶段四 - 数据库清理（1 天）
- [ ] 全面测试：阶段五 - 全面测试（2-3 天）
- [ ] 文档更新：阶段六 - 文档更新（1 天）
- [ ] 部署上线：阶段七 - 部署上线（1 天）

---

## 六、参考文档

- [后端菜单模型](../../server/internal/model/menu.go)
- [前端路由配置](../../static/src/router/modules/)
- [权限验证中间件](../../server/internal/middleware/role.go)
- [前端路由守卫](../../static/src/router/guard.ts)
