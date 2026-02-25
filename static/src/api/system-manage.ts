import request from '@/utils/http'
import { AppRouteRecord } from '@/types/router'

// 获取用户列表（Admin+）
export function fetchGetUserList(params?: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: '/api/v1/users',
    params
  })
}

// 删除用户（Admin+）
export function fetchDeleteUser(id: number) {
  return request.del({
    url: `/api/v1/users/${id}`
  })
}

// 修改用户角色（Admin+）
export function fetchUpdateUserRole(id: number, role: string) {
  return request.request({
    url: `/api/v1/users/${id}/role`,
    method: 'PATCH',
    data: { role }
  })
}

// 获取角色列表（前端静态，与后端角色常量一致）
export function fetchGetRoleList(_params?: Api.SystemManage.RoleSearchParams) {
  // 系统内置角色，与后端 model.Role* 保持一致
  const roles: Api.SystemManage.RoleListItem[] = [
    {
      roleId: 1,
      roleName: '超级管理员',
      roleCode: 'super_admin',
      description: '拥有系统的全部权限，包括用户管理、数据管理、系统设置等',
      enabled: true,
      createTime: '-'
    },
    {
      roleId: 2,
      roleName: '管理员',
      roleCode: 'admin',
      description: '所有非技术性权限，包括用户管理、数据管理等',
      enabled: true,
      createTime: '-'
    },
    {
      roleId: 3,
      roleName: '已认证用户',
      roleCode: 'user',
      description: '基本访问权限，包括查看数据和使用系统功能',
      enabled: true,
      createTime: '-'
    },
    {
      roleId: 4,
      roleName: '访客',
      roleCode: 'guest',
      description: '有限访问权限，主要用于查看公开信息',
      enabled: true,
      createTime: '-'
    }
  ]
  return Promise.resolve({
    records: roles,
    current: 1,
    size: 10,
    total: roles.length
  } as Api.SystemManage.RoleList)
}

// 获取菜单列表（后端动态路由，根据当前登录用户角色返回）
export function fetchGetMenuList() {
  return request.get<AppRouteRecord[]>({
    url: '/api/v1/menu'
  })
}
