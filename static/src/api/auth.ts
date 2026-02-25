import request from '@/utils/http'

/**
 * 获取 EVE SSO 授权 URL（通过后端接口获取，前端直接跳转）
 * @param scopes 额外 ESI scopes（可选）
 */
export async function getEveSSOLoginURL(scopes?: string[]): Promise<string> {
  // hash 模式下 callback 路径在 # 后面
  const callbackURL = window.location.origin + '/#/auth/callback'
  const params: Record<string, string> = { redirect: callbackURL }
  if (scopes && scopes.length > 0) {
    params.scopes = scopes.join(',')
  }
  const data = await request.get<{ url: string }>({
    url: '/api/v1/sso/eve/login',
    params
  })
  return data.url
}

/**
 * 获取已注册的 ESI Scope 列表
 */
export function fetchEveSSOScopes() {
  return request.get<Api.Auth.RegisteredScope[]>({
    url: '/api/v1/sso/eve/scopes'
  })
}

/**
 * 获取当前登录用户的 EVE 角色列表
 */
export function fetchMyCharacters() {
  return request.get<Api.Auth.EveCharacter[]>({
    url: '/api/v1/sso/eve/characters'
  })
}

/**
 * 获取「绑定新角色」的 EVE SSO 授权 URL
 * @param scopes 额外 ESI scopes（可选）
 */
export async function getEveBindURL(scopes?: string[]): Promise<string> {
  const callbackURL = window.location.origin + '/#/auth/callback'
  const params: Record<string, string> = { redirect: callbackURL }
  if (scopes && scopes.length > 0) {
    params.scopes = scopes.join(',')
  }
  const data = await request.get<{ url: string }>({
    url: '/api/v1/sso/eve/bind',
    params
  })
  return data.url
}

/**
 * 设置主角色
 * @param characterId EVE 角色 ID
 */
export function setPrimaryCharacter(characterId: number) {
  return request.put({
    url: `/api/v1/sso/eve/primary/${characterId}`
  })
}

/**
 * 解绑角色
 * @param characterId EVE 角色 ID
 */
export function unbindCharacter(characterId: number) {
  return request.del({
    url: `/api/v1/sso/eve/characters/${characterId}`
  })
}

/**
 * 获取当前登录用户信息（从 /me 接口获取并封装成统一格式）
 * @returns 用户信息
 */
export async function fetchGetUserInfo(): Promise<Api.Auth.UserInfo> {
  const data = await request.get<Api.Auth.MeResponse>({
    url: '/api/v1/me'
  })

  const { user, characters, roles: backendRoles, permissions } = data

  // 主角色：根据 primary_character_id 查找，找不到则用第一个，再 fallback 到用户信息
  const primaryChar =
    characters?.find((c) => c.character_id === user.primary_character_id) ?? characters?.[0]

  // 直接使用后端角色编码，回退到 user.role
  const roles = backendRoles && backendRoles.length > 0 ? backendRoles : [user.role ?? 'user']

  return {
    userId: user.id,
    userName: primaryChar?.character_name ?? user.nickname ?? `Capsuleer#${user.id}`,
    avatar: primaryChar?.portrait_url ?? user.avatar ?? '',
    roles,
    buttons: permissions ?? [],
    characters: characters ?? [],
    primaryCharacterId: user.primary_character_id ?? 0
  }
}
