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
    url: '/api/v1/srp/applications/me',
    params
  })
}

/** 获取舰队范围内可用的 KM 列表（快捷申请） */
export function fetchFleetKillmails(fleetId: string) {
  return request.get<Api.Srp.FleetKillmailItem[]>({
    url: `/api/v1/srp/killmails/fleet/${fleetId}`
  })
}

/** 获取当前用户所有角色的全部 KM 列表（不限舰队；可按 characterId 筛选） */
export function fetchMyKillmails(characterId?: number) {
  return request.get<Api.Srp.FleetKillmailItem[]>({
    url: '/api/v1/srp/killmails/me',
    params: characterId ? { character_id: characterId } : undefined
  })
}

/** 获取 KM 装配详情 */
export function fetchKillmailDetail(data: Api.Srp.KillmailDetailRequest) {
  return request.post<Api.Srp.KillmailDetailResponse>({
    url: '/api/v1/srp/killmails/detail',
    data
  })
}

// ─── 补损申请（管理端）───

/** 获取全部申请列表（管理端） */
export function fetchApplicationList(params?: Api.Srp.ApplicationSearchParams) {
  return request.get<Api.Srp.ApplicationList>({
    url: '/api/v1/srp/applications',
    params
  })
}

/** 获取单条申请详情 */
export function fetchApplicationDetail(id: number) {
  return request.get<Api.Srp.Application>({
    url: `/api/v1/srp/applications/${id}`
  })
}

/** 审批补损申请 */
export function reviewApplication(id: number, data: Api.Srp.ReviewParams) {
  return request.put<Api.Srp.Application>({
    url: `/api/v1/srp/applications/${id}/review`,
    data
  })
}

/** 发放补损 */
export function payoutApplication(id: number, data?: Api.Srp.PayoutParams) {
  return request.put<Api.Srp.Application>({
    url: `/api/v1/srp/applications/${id}/payout`,
    data: data ?? {}
  })
}

/** 获取批量发放汇总 */
export function fetchBatchPayoutSummary() {
  return request.get<Api.Srp.BatchPayoutSummary[]>({
    url: '/api/v1/srp/applications/batch-payout-summary'
  })
}

/** 按用户批量发放 SRP */
export function batchPayoutByUser(userId: number) {
  return request.put<Api.Srp.BatchPayoutSummary>({
    url: `/api/v1/srp/applications/users/${userId}/payout`,
    data: {}
  })
}

/** 通过 ESI 在客户端打开角色信息窗口 */
export function openInfoWindow(data: { character_id: number; target_id: number }) {
  return request.post({
    url: '/api/v1/srp/open-info-window',
    data
  })
}
