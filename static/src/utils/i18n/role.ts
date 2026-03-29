import i18n, { $t } from '@/locales'

interface RoleDefinition {
  code: string
  name: string
  i18nKey?: string
  description?: string
  sort?: number
}

const ROLE_I18N_KEY_PREFIX = 'userAdmin.roles.'

export function getRoleName(role: RoleDefinition | string, fallbackKey?: string): string {
  let code: string
  let fallbackName: string | undefined
  let i18nKey: string | undefined

  if (typeof role === 'string') {
    code = role
    i18nKey = fallbackKey || `${ROLE_I18N_KEY_PREFIX}${code}`
  } else {
    code = role.code
    fallbackName = role.name
    i18nKey = role.i18nKey || fallbackKey || `${ROLE_I18N_KEY_PREFIX}${code}`
  }

  if (i18nKey && i18n.global.te(i18nKey)) {
    return $t(i18nKey)
  }

  return fallbackName || code
}

export function getRoleNames(roles: (RoleDefinition | string)[]): string[] {
  return roles.map((role) => getRoleName(role))
}

export function getRoleLabel(role: RoleDefinition | string): string {
  const name = getRoleName(role)
  if (typeof role === 'string') {
    return `${name} (${role})`
  }
  return role.name === name ? name : `${name} (${role.code})`
}
