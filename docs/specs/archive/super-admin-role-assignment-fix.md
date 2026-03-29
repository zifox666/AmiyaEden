---
status: archived
doc_type: spec
implemented: 2026-03-29
owner: engineering
last_reviewed: 2026-03-29
source_of_truth:
  - server/internal/service/role.go
  - server/internal/service/role_test.go
  - static/src/views/system/user/modules/user-role-dialog.vue
---

# Super Admin 角色分配修复

## 状态：已实现 ✅

2026-03-29 实现并验证通过。

## 问题描述

super_admin 通过角色管理界面给自己添加 `admin`、`captain` 等任意非 `super_admin` 角色时，后端返回错误 "超级管理员角色仅通过配置文件管理，不可手动分配"。

## 根因分析

### 调用链

1. 前端 `user-role-dialog.vue` 加载用户当前角色 `["super_admin"]`，`super_admin` checkbox 为 disabled（不可取消勾选）
2. super_admin 勾选 `admin` → `selectedRoleCodes = ["super_admin", "admin"]`
3. 提交到 `PUT /api/v1/system/user/:id/roles`，body: `{"role_codes": ["super_admin", "admin"]}`
4. 后端 `SetUserRoles` → `normalizeAssignedRoleCodes` → `requestedCodes = ["super_admin", "admin"]`
5. 第 109 行 `ContainsAnyRole(requestedCodes, RoleSuperAdmin)` → `true` → **无条件报错**

### ElCheckbox disabled 行为

`disabled=true` 的 checkbox，如果值已在 `v-model` 数组中，会保留在数组中但不可取消。因此 super_admin 编辑自己时，`super_admin` 必然出现在提交数据里。

### 额外风险

`repo.SetUserRoles` 是**全量替换**（DELETE + INSERT）。如果 super_admin 给自己提交 `["admin"]`，`super_admin` 角色会被覆盖删除，直到下次 SSO 登录 `SyncConfigSuperAdmins` 才会恢复，存在权限丢失窗口。

## 已实现的修复

### 修改文件

#### 1. `server/internal/service/role.go` — `SetUserRoles` 核心逻辑

**替换原有无条件拒绝逻辑**，改为分支处理：

```go
if model.IsSuperAdmin(operatorRoles) {
    requestedCodes = filterOutRole(requestedCodes, model.RoleSuperAdmin)
    if model.ContainsAnyRole(currentCodes, model.RoleSuperAdmin) {
        requestedCodes = append([]string{model.RoleSuperAdmin}, requestedCodes...)
    }
} else {
    if model.ContainsAnyRole(requestedCodes, model.RoleSuperAdmin) {
        return errors.New("超级管理员角色仅通过配置文件管理，不可手动分配")
    }
    if model.ContainsAnyRole(currentCodes, model.RoleSuperAdmin) {
        return errors.New("超级管理员角色仅通过配置文件管理，不可手动修改")
    }
}
```

#### 2. `server/internal/service/role.go` — 新增 `filterOutRole` 辅助函数

```go
func filterOutRole(codes []string, target string) []string {
    result := make([]string, 0, len(codes))
    for _, code := range codes {
        if code != target {
            result = append(result, code)
        }
    }
    return result
}
```

#### 3. `server/internal/service/role_test.go` — 新增 `TestFilterOutRole` 测试

4 个子用例：移除目标角色、目标不存在返回原样、空输入返回空、全部是目标角色返回空。

全部 18 个测试用例通过（含原有 11 个 `TestValidateSetUserRolesPermission` + 3 个 `TestNormalizeAssignedRoleCodes` + 4 个 `TestFilterOutRole`）。

## 保证的不变量

| 不变量 | 修复前 | 修复后 |
|---|---|---|
| `super_admin` 仅通过配置文件授予 | ✅ | ✅ |
| `super_admin` 可操作自己非 `super_admin` 的其他角色 | ❌ 报错 | ✅ 静默剥离后处理 |
| `super_admin` 可操作任何其他用户的角色 | ✅ | ✅ |
| `super_admin` 用户的角色不能被任何 admin 编辑 | ✅ | ✅ |
| `super_admin` 角色不会被全量替换意外删除 | ❌ 存在窗口期风险 | ✅ 自动保留 |

## 已完成的文档更新

- ✅ `docs/architecture/auth-and-permissions.md` — 角色分配规则描述与矩阵表格已更新
- ✅ `docs/features/current/administration.md` — 关键不变量描述已更新
