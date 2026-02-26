import request from '@/utils/http'

/** 批量查询 ID → 名称映射 */
export function fetchNames(data: {
  language?: string
  ids?: Record<string, number[]>
  esi?: number[]
}) {
  return request.post<Record<number, string>>({
    url: '/api/v1/sde/names',
    data
  })
}
