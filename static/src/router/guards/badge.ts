export interface BadgeLoader {
  fetchBadgeCounts: () => Promise<unknown>
  clearBadgeCounts: () => void
}

export interface BadgeLogger {
  error: (...args: unknown[]) => void
}

export async function loadBadgeCounts(
  badgeLoader: BadgeLoader,
  logger: BadgeLogger = console
): Promise<void> {
  try {
    await badgeLoader.fetchBadgeCounts()
  } catch (error) {
    badgeLoader.clearBadgeCounts()
    logger.error('[RouteGuard] 导航徽章加载失败:', error)
  }
}
