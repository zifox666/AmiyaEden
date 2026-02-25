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
