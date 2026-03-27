import type { AppRouteRecord } from '../../types/router'
import { RoutesAlias } from '../routesAlias'

export function hasNonGuestRole(roles: string[]): boolean {
  return roles.some((role) => role !== 'guest')
}

export function applyMenuAccessFilter(
  menu: AppRouteRecord[],
  roles: string[],
  isCurrentlyNewbro?: boolean
): AppRouteRecord[] {
  return menu.reduce((acc: AppRouteRecord[], item) => {
    const itemRoles = item.meta?.roles
    const requiresLogin = item.meta?.login === true
    const requiresNewbro = item.meta?.requiresNewbro === true
    const hasRolePermission = !itemRoles || itemRoles.some((role) => roles.includes(role))
    const hasLoginPermission = !requiresLogin || hasNonGuestRole(roles)
    const hasNewbroPermission = !requiresNewbro || isCurrentlyNewbro === true
    const hasPermission = hasRolePermission && hasLoginPermission && hasNewbroPermission

    if (!hasPermission) {
      return acc
    }

    const filteredItem = { ...item }
    if (filteredItem.children?.length) {
      filteredItem.children = applyMenuAccessFilter(filteredItem.children, roles, isCurrentlyNewbro)
    }

    acc.push(filteredItem)
    return acc
  }, [])
}

export function pruneEmptyMenus(menuList: AppRouteRecord[]): AppRouteRecord[] {
  return menuList
    .map((item) => {
      if (item.children && item.children.length > 0) {
        return {
          ...item,
          children: pruneEmptyMenus(item.children)
        }
      }

      return item
    })
    .filter((item) => {
      // Directory menus: keep only if they still have children after pruning
      if ('children' in item) {
        return item.children !== undefined && item.children.length > 0
      }

      // Leaf nodes: keep iframes, external links, or real components
      if (item.meta?.isIframe === true || item.meta?.link) {
        return true
      }

      return !!item.component && item.component !== '' && item.component !== RoutesAlias.Layout
    })
}
