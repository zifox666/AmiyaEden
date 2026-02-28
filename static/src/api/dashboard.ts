import request from '@/utils/http'

/** 获取工作台数据 */
export function fetchDashboard() {
  return request.post<Api.Dashboard.DashboardResult>({
    url: '/api/v1/dashboard'
  })
}
