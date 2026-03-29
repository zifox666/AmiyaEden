import request from '@/utils/http'

// ─── 用户管理 ───

export function fetchGetUserList(params?: Api.SystemManage.UserSearchParams) {
  return request.get<Api.SystemManage.UserList>({
    url: '/api/v1/system/user',
    params
  })
}

export function fetchGetUser(id: number) {
  return request.get<Api.SystemManage.UserDetail>({
    url: `/api/v1/system/user/${id}`
  })
}

export function fetchUpdateUser(
  id: number,
  data: { nickname?: string; qq?: string; discord_id?: string; status?: number }
) {
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
  return request.post<{ token: string; user: Api.SystemManage.UserDetail }>({
    url: `/api/v1/system/user/${id}/impersonate`
  })
}

// ─── 用户职权分配 ───

export function fetchGetUserRoles(userId: number) {
  return request.get<Api.SystemManage.RoleDefinition[]>({
    url: `/api/v1/system/user/${userId}/roles`
  })
}

export function fetchSetUserRoles(userId: number, roleCodes: string[]) {
  return request.put({
    url: `/api/v1/system/user/${userId}/roles`,
    data: { role_codes: roleCodes }
  })
}

// ─── 职权定义 ───

export function fetchGetRoleDefinitions() {
  return request.get<Api.SystemManage.RoleDefinition[]>({
    url: '/api/v1/system/role/definitions'
  })
}

// ─── 自动权限映射 ───

/** 获取所有 ESI 军团职权名列表 */
export function fetchGetAllEsiRoles() {
  return request.get<string[]>({
    url: '/api/v1/system/auto-role/esi-roles'
  })
}

/** 获取所有 ESI 军团职权映射 */
export function fetchGetEsiRoleMappings() {
  return request.get<Api.SystemManage.EsiRoleMapping[]>({
    url: '/api/v1/system/auto-role/esi-role-mappings'
  })
}

/** 创建 ESI 军团职权映射 */
export function fetchCreateEsiRoleMapping(data: { esi_role: string; role_code: string }) {
  return request.post<Api.SystemManage.EsiRoleMapping>({
    url: '/api/v1/system/auto-role/esi-role-mappings',
    data
  })
}

/** 删除 ESI 军团职权映射 */
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
export function fetchCreateEsiTitleMapping(data: {
  corporation_id: number
  title_id: number
  title_name?: string
  role_code: string
}) {
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
