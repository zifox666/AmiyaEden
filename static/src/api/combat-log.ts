import request from '@/utils/http'

export function uploadCombatLog(file: File) {
  const data = new FormData()
  data.append('file', file)
  return request.post<Api.CombatLog.Detail>({
    url: '/api/v1/combat-logs',
    data,
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

export function getCombatLogReports(params: {
  page: number
  page_size: number
  scope?: 'mine' | 'all'
}) {
  return request.get<Api.CombatLog.ReportList>({ url: '/api/v1/combat-logs', params })
}

export function getCombatLogDetail(id: number) {
  return request.get<Api.CombatLog.Detail>({ url: `/api/v1/combat-logs/${id}` })
}
