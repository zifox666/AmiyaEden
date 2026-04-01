import type { AppRouteRecord } from '../../types/router'

const routeBadgeFieldMap = {
  NewbroMentorDashboard: 'mentor_pending_applications',
  WelfareMy: 'welfare_eligible',
  WelfareApproval: 'welfare_pending',
  SrpManage: 'srp_pending',
  ShopOrderManage: 'order_pending'
} as const satisfies Record<string, keyof Api.Badge.BadgeCounts>

function clearTextBadge(route: AppRouteRecord): void {
  delete route.meta.showTextBadge
}

function getRouteBadgeCount(route: AppRouteRecord, badgeCounts: Api.Badge.BadgeCounts): number {
  const routeName = typeof route.name === 'string' ? route.name : ''
  const field = routeBadgeFieldMap[routeName as keyof typeof routeBadgeFieldMap]
  if (!field) {
    return 0
  }

  return badgeCounts[field] ?? 0
}

function applyBadgeCountsToRoute(
  route: AppRouteRecord,
  badgeCounts: Api.Badge.BadgeCounts
): number {
  if (!route.children || route.children.length === 0) {
    const count = getRouteBadgeCount(route, badgeCounts)
    if (count > 0) {
      route.meta.showTextBadge = String(count)
      return count
    }

    clearTextBadge(route)
    return 0
  }

  let childTotal = 0
  for (const child of route.children) {
    childTotal += applyBadgeCountsToRoute(child, badgeCounts)
  }

  if (childTotal > 0) {
    route.meta.showTextBadge = String(childTotal)
    return childTotal
  }

  clearTextBadge(route)
  return 0
}

export function applyBadgeCountsToMenu(
  menuList: AppRouteRecord[],
  badgeCounts: Api.Badge.BadgeCounts
): void {
  for (const route of menuList) {
    applyBadgeCountsToRoute(route, badgeCounts)
  }
}
