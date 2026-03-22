import { fetchNames } from '@/api/sde'
import { useUserStore } from '@/store/modules/user'
import {
  filterUnresolvedESIIDs,
  filterUnresolvedNameIDs,
  getResolvedName,
  mergeResolvedNames
} from './useNameResolver.helpers'

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
 * getName(587, '-', 'type') // => "Rifter"
 * ```
 */
export function useNameResolver() {
  const nameMap = ref<Record<number, string>>({})
  const namesByNamespace = ref<Record<string, Record<number, string>>>({})
  const loading = ref(false)

  /**
   * 批量解析 ID → 名称。
   * 结果会合并进 nameMap（不会清除之前已有的映射）。
   */
  async function resolve(params: Api.Sde.ResolveNamesRequest) {
    // 过滤掉空数组 & 已解析过的 ID，减少请求体积
    const ids = filterUnresolvedNameIDs(params.ids, namesByNamespace.value)
    const esi = filterUnresolvedESIIDs(params.esi, namesByNamespace.value)

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
      mergeResolvedNames(namesByNamespace.value, nameMap.value, result)
    } catch (e) {
      console.warn('[useNameResolver] 名称解析失败', e)
    } finally {
      loading.value = false
    }
  }

  /** 便捷方法：根据 id 从 nameMap 获取名称，未找到时返回 fallback */
  function getName(id: number | null | undefined, fallback?: string, namespace?: string): string {
    return getResolvedName(id, fallback, namespace, namesByNamespace.value, nameMap.value)
  }

  return { nameMap, namesByNamespace, loading, resolve, getName }
}
