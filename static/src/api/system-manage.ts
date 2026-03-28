import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'

// ─── 用户管理 ───

export function fetchGetUserList(params?: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: '/api/v1/system/user',
    params
  })
}

export function fetchGetUser(id: number) {
  return request.get<Api.SystemManage.UserListItem>({
    url: `/api/v1/system/user/${id}`
  })
}

export function fetchUpdateUser(id: number, data: { nickname?: string; status?: number }) {
  return request.put({
    url: `/api/v1/system/user/${id}`,
    data
  })
}

export function fetchDeleteUser(id: number) {
  return request.del({
    url: `/api/v1/system/user/${id}`
  })
}

export function fetchImpersonateUser(id: number) {
  return request.post<{ token: string; user: Api.SystemManage.UserListItem }>({
    url: `/api/v1/system/user/${id}/impersonate`
  })
}

// ─── 用户角色分配 ───

export function fetchGetUserRoles(userId: number) {
  return request.get<Api.SystemManage.RoleItem[]>({
    url: `/api/v1/system/user/${userId}/roles`
  })
}

export function fetchSetUserRoles(userId: number, roleIds: number[]) {
  return request.put({
    url: `/api/v1/system/user/${userId}/roles`,
    data: { role_ids: roleIds }
  })
}

// ─── 角色管理 ───

export function fetchGetRoleList(params?: Api.SystemManage.RoleSearchParams) {
  return request.get<Api.SystemManage.RoleList>({
    url: '/api/v1/system/role',
    params
  })
}

export function fetchGetAllRoles() {
  return request.get<Api.SystemManage.RoleItem[]>({
    url: '/api/v1/system/role/all'
  })
}

export function fetchGetRole(id: number) {
  return request.get<Api.SystemManage.RoleItem>({
    url: `/api/v1/system/role/${id}`
  })
}

export function fetchCreateRole(data: Api.SystemManage.CreateRoleParams) {
  return request.post<Api.SystemManage.RoleItem>({
    url: '/api/v1/system/role',
    data
  })
}

export function fetchUpdateRole(id: number, data: Api.SystemManage.UpdateRoleParams) {
  return request.put({
    url: `/api/v1/system/role/${id}`,
    data
  })
}

export function fetchDeleteRole(id: number) {
  return request.del({
    url: `/api/v1/system/role/${id}`
  })
}

// ─── 角色权限（菜单）管理 ───

export function fetchGetRoleMenus(roleId: number) {
  return request.get<number[]>({
    url: `/api/v1/system/role/${roleId}/menus`
  })
}

export function fetchSetRoleMenus(roleId: number, menuIds: number[]) {
  return request.put({
    url: `/api/v1/system/role/${roleId}/menus`,
    data: { menu_ids: menuIds }
  })
}

// ─── 菜单管理 ───

export function fetchGetMenuTree() {
  return request.get<Api.SystemManage.MenuItem[]>({
    url: '/api/v1/system/menu/tree'
  })
}

export function fetchCreateMenu(data: Api.SystemManage.CreateMenuParams) {
  return request.post<Api.SystemManage.MenuItem>({
    url: '/api/v1/system/menu',
    data
  })
}

export function fetchUpdateMenu(id: number, data: Api.SystemManage.UpdateMenuParams) {
  return request.put({
    url: `/api/v1/system/menu/${id}`,
    data
  })
}

export function fetchDeleteMenu(id: number) {
  return request.del({
    url: `/api/v1/system/menu/${id}`
  })
}

// ─── 用户菜单（前端路由） ───

export function fetchGetMenuList() {
  return request.get<AppRouteRecord[]>({
    url: '/api/v1/menu/list'
  })
}

// ─── 自动权限映射 ───

/** 获取所有 ESI 军团角色名列表 */
export function fetchGetAllEsiRoles() {
  return request.get<string[]>({
    url: '/api/v1/system/auto-role/esi-roles'
  })
}

/** 获取所有 ESI 军团角色映射 */
export function fetchGetEsiRoleMappings() {
  return request.get<Api.SystemManage.EsiRoleMapping[]>({
    url: '/api/v1/system/auto-role/esi-role-mappings'
  })
}

/** 创建 ESI 军团角色映射 */
export function fetchCreateEsiRoleMapping(data: Api.SystemManage.CreateEsiRoleMappingParams) {
  return request.post<Api.SystemManage.EsiRoleMapping>({
    url: '/api/v1/system/auto-role/esi-role-mappings',
    data
  })
}

/** 删除 ESI 军团角色映射 */
export function fetchDeleteEsiRoleMapping(id: number) {
  return request.del({
    url: `/api/v1/system/auto-role/esi-role-mappings/${id}`
  })
}

/** 获取所有 ESI 头衔映射 */
export function fetchGetEsiTitleMappings() {
  return request.get<Api.SystemManage.EsiTitleMapping[]>({
    url: '/api/v1/system/auto-role/esi-title-mappings'
  })
}

/** 创建 ESI 头衔映射 */
export function fetchCreateEsiTitleMapping(data: Api.SystemManage.CreateEsiTitleMappingParams) {
  return request.post<Api.SystemManage.EsiTitleMapping>({
    url: '/api/v1/system/auto-role/esi-title-mappings',
    data
  })
}

/** 删除 ESI 头衔映射 */
export function fetchDeleteEsiTitleMapping(id: number) {
  return request.del({
    url: `/api/v1/system/auto-role/esi-title-mappings/${id}`
  })
}

/** 获取数据库中所有军团头衔（用于前端下拉选择） */
export function fetchGetCorpTitles() {
  return request.get<Api.SystemManage.CorpTitleInfo[]>({
    url: '/api/v1/system/auto-role/corp-titles'
  })
}

/** 手动触发自动权限同步 */
export function fetchTriggerAutoRoleSync() {
  return request.post({
    url: '/api/v1/system/auto-role/sync'
  })
}

/** 分页查询自动权限操作日志 */
export function fetchGetAutoRoleLogs(params: { current: number; size: number }) {
  return request.get<Api.SystemManage.AutoRoleLogList>({
    url: '/api/v1/system/auto-role/logs',
    params
  })
}
