import { defineStore } from 'pinia'
import { ref } from 'vue'
import { fetchBadgeCounts as fetchBadgeCountsApi } from '@/api/badge'
import { applyBadgeCountsToMenu } from './badge.helpers'
import { useMenuStore } from './menu'

export const useBadgeStore = defineStore('badgeStore', () => {
  const badgeCounts = ref<Api.Badge.BadgeCounts>({})
  const menuStore = useMenuStore()

  const fetchBadgeCounts = async () => {
    const counts = await fetchBadgeCountsApi()
    badgeCounts.value = counts
    applyBadgeCountsToMenu(menuStore.menuList, counts)
    return counts
  }

  const clearBadgeCounts = () => {
    badgeCounts.value = {}
    applyBadgeCountsToMenu(menuStore.menuList, {})
  }

  return {
    badgeCounts,
    fetchBadgeCounts,
    clearBadgeCounts
  }
})
