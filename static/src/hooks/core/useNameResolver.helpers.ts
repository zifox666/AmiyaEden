export function filterUnresolvedNameIDs(
  ids: Record<string, number[]> | undefined,
  namesByNamespace: Record<string, Record<number, string>>
): Record<string, number[]> {
  const filtered: Record<string, number[]> = {}

  if (!ids) return filtered

  for (const [key, arr] of Object.entries(ids)) {
    const namespaceMap = namesByNamespace[key] ?? {}
    const unique = [...new Set(arr)].filter((id) => id != null && !(id in namespaceMap))
    if (unique.length) filtered[key] = unique
  }

  return filtered
}

export function filterUnresolvedESIIDs(
  ids: number[] | undefined,
  namesByNamespace: Record<string, Record<number, string>>
): number[] {
  const esiMap = namesByNamespace.esi ?? {}
  return ids ? [...new Set(ids)].filter((id) => id != null && id > 0 && !(id in esiMap)) : []
}

export function mergeResolvedNames(
  targetNamesByNamespace: Record<string, Record<number, string>>,
  targetFlatMap: Record<number, string>,
  result: Api.Sde.ResolveNamesResponse | null | undefined
) {
  if (result?.names) {
    for (const [namespace, entries] of Object.entries(result.names)) {
      if (!targetNamesByNamespace[namespace]) {
        targetNamesByNamespace[namespace] = {}
      }
      Object.assign(targetNamesByNamespace[namespace], entries)
    }
  }

  if (result?.flat) {
    for (const [key, value] of Object.entries(result.flat)) {
      targetFlatMap[Number(key)] = value
    }
  }
}

export function getResolvedName(
  id: number | null | undefined,
  fallback: string | undefined,
  namespace: string | undefined,
  namesByNamespace: Record<string, Record<number, string>>,
  flatMap: Record<number, string>
): string {
  if (id == null) return fallback ?? '-'

  if (namespace) {
    const namespaceName = namesByNamespace[namespace]?.[id]
    if (namespaceName != null) return namespaceName
  }

  return flatMap[id] ?? fallback ?? String(id)
}
