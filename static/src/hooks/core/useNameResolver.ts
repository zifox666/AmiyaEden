import { fetchNames } from '@/api/sde'
import { useUserStore } from '@/store/modules/user'

/**
 * useNameResolver — 批量收集 ID，调用 SDE /names 接口解析为名称。
 *
 * 支持两种来源：
 *   1. SDE 翻译表（type / solar_system / group / category 等）
 *   2. ESI /universe/names（character / corporation / alliance）
 *
 * 用法：
 * ```ts
 * const { nameMap, resolve } = useNameResolver()
 * await resolve({
 *   ids: { type: [587], solar_system: [30002187] },
 *   esi: [95465499]
 * })
 * nameMap.value[587] // => "Rifter"
 * ```
 */
export function useNameResolver() {
  const nameMap = ref<Record<number, string>>({})
  const loading = ref(false)

  /**
   * 批量解析 ID → 名称。
   * 结果会合并进 nameMap（不会清除之前已有的映射）。
   */
  async function resolve(params: { ids?: Record<string, number[]>; esi?: number[] }) {
    // 过滤掉空数组 & 已解析过的 ID，减少请求体积
    const ids: Record<string, number[]> = {}
    if (params.ids) {
      for (const [key, arr] of Object.entries(params.ids)) {
        const unique = [...new Set(arr)].filter((id) => id != null && !(id in nameMap.value))
        if (unique.length) ids[key] = unique
      }
    }
    const esi = params.esi
      ? [...new Set(params.esi)].filter((id) => id != null && id > 0 && !(id in nameMap.value))
      : []

    if (!Object.keys(ids).length && !esi.length) return

    // 推断当前 i18n 语言
    const userStore = useUserStore()
    const language = (userStore.language as string) || 'en'

    loading.value = true
    try {
      const result = await fetchNames({
        language,
        ids: Object.keys(ids).length ? ids : undefined,
        esi: esi.length ? esi : undefined
      })
      if (result) {
        // result 可能是 Record<string, string>（key 为数字字符串）
        for (const [k, v] of Object.entries(result)) {
          nameMap.value[Number(k)] = v
        }
      }
    } catch (e) {
      console.warn('[useNameResolver] 名称解析失败', e)
    } finally {
      loading.value = false
    }
  }

  /** 便捷方法：根据 id 从 nameMap 获取名称，未找到时返回 fallback */
  function getName(id: number | null | undefined, fallback?: string): string {
    if (id == null) return fallback ?? '-'
    return nameMap.value[id] ?? fallback ?? String(id)
  }

  return { nameMap, loading, resolve, getName }
}
