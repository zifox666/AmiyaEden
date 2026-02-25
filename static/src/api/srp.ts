import request from '@/utils/http'

// ─── 舰船价格表 ───

/** 获取舰船价格列表 */
export function fetchShipPrices(keyword?: string) {
  return request.get<Api.Srp.ShipPrice[]>({
    url: '/api/v1/srp/prices',
    params: keyword ? { keyword } : undefined
  })
}

/** 新增或更新舰船价格 */
export function upsertShipPrice(data: Api.Srp.UpsertShipPriceParams) {
  return request.post<Api.Srp.ShipPrice>({
    url: '/api/v1/srp/prices',
    data
  })
}

/** 删除舰船价格 */
export function deleteShipPrice(id: number) {
  return request.del({
    url: `/api/v1/srp/prices/${id}`
  })
}

// ─── 补损申请（用户端）───

/** 提交补损申请 */
export function submitApplication(data: Api.Srp.SubmitApplicationParams) {
  return request.post<Api.Srp.Application>({
    url: '/api/v1/srp/applications',
    data
  })
}

/** 获取我的补损申请列表 */
export function fetchMyApplications(params?: Partial<Api.Common.CommonSearchParams>) {
  return request.get<Api.Srp.ApplicationList>({
    url: '/api/v1/srp/applications/my',
    params
  })
}

/** 获取舰队范围内可用的 KM 列表（快捷申请） */
export function fetchFleetKillmails(fleetId: string) {
  return request.get<Api.Srp.FleetKillmailItem[]>({
    url: '/api/v1/srp/fleet-killmails',
    params: { fleet_id: fleetId }
  })
}

/** 获取当前用户所有角色的全部 KM 列表（不限舰队；可按 characterId 筛选） */
export function fetchMyKillmails(characterId?: number) {
  return request.get<Api.Srp.FleetKillmailItem[]>({
    url: '/api/v1/srp/my-killmails',
    params: characterId ? { character_id: characterId } : undefined
  })
}

// ─── 补损申请（管理端）───

/** 获取全部申请列表（管理端） */
export function fetchApplicationList(params?: Api.Srp.ApplicationSearchParams) {
  return request.get<Api.Srp.ApplicationList>({
    url: '/api/v1/srp/manage/applications',
    params
  })
}

/** 获取单条申请详情 */
export function fetchApplicationDetail(id: number) {
  return request.get<Api.Srp.Application>({
    url: `/api/v1/srp/manage/applications/${id}`
  })
}

/** 审批补损申请 */
export function reviewApplication(id: number, data: Api.Srp.ReviewParams) {
  return request.request<Api.Srp.Application>({
    url: `/api/v1/srp/manage/applications/${id}/review`,
    method: 'PATCH',
    data
  })
}

/** 发放补损 */
export function payoutApplication(id: number, data?: Api.Srp.PayoutParams) {
  return request.request<Api.Srp.Application>({
    url: `/api/v1/srp/manage/applications/${id}/payout`,
    method: 'PATCH',
    data: data ?? {}
  })
}
