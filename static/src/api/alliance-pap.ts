import request from '@/utils/http'

export interface AlliancePAPFleet {
  id: number
  main_character: string
  character_id: string
  character_name: string
  fleet_id: string
  year: number
  month: number
  start_at: string
  end_at?: string
  title: string
  level: string
  pap: number
  ship_group_id: string
  ship_group_name: string
  ship_type_id: string
  ship_type_name: string
  is_archived: boolean
}

export interface AlliancePAPSummary {
  id: number
  main_character: string
  year: number
  month: number
  corporation_id: string
  total_pap: number
  yearly_total_pap: number
  monthly_rank: number
  yearly_rank: number
  global_monthly_rank: number
  global_yearly_rank: number
  total_in_corp: number
  total_global: number
  calculated_at: string
  is_archived: boolean
}

export interface AlliancePAPResult {
  summary: AlliancePAPSummary | null
  fleets: AlliancePAPFleet[]
}

export interface AlliancePAPAllResult {
  year: number
  month: number
  list: AlliancePAPSummary[]
}

/** 获取我的联盟 PAP（当前用户主角色，默认当月） */
export function fetchMyAlliancePAP(params?: { year?: number; month?: number }) {
  return request.get<AlliancePAPResult>({
    url: '/api/v1/operation/fleets/pap/alliance',
    params
  })
}

/** 管理员：分页获取所有成员某月联盟 PAP 汇总 */
export function fetchAllAlliancePAP(
  params?: Api.Common.CommonSearchParams & { year?: number; month?: number }
) {
  return request.get<Api.Common.PaginatedResponse<AlliancePAPSummary>>({
    url: '/api/v1/system/pap',
    params
  })
}

/** 管理员：手动触发拉取 */
export function triggerAlliancePAPFetch(params?: { year?: number; month?: number }) {
  return request.post({
    url: '/api/v1/system/pap/fetch',
    params
  })
}

/** 管理员：通过表格导入 PAP 数据 */
export interface PAPImportInfo {
  primary_character_name: string
  monthly_pap: number
  calculated_at: string
}

export function importAlliancePAP(params?: { year?: number, month?: number, data: PAPImportInfo }) {
  return request.post<PAPImportInfo>({ url: '/api/v1/system/pap/import', params })
}

export interface PAPExchangeConfig {
  wallet_per_pap: number
  enabled: boolean
}

export interface SettleMonthResult {
  year: number
  month: number
  total_users: number
  skipped_users: number
  total_wallet: number
}

/** 管理员：获取 PAP 兑换配置 */
export function fetchPAPExchangeConfig() {
  return request.get<PAPExchangeConfig>({ url: '/api/v1/system/pap/config' })
}

/** 管理员：更新 PAP 兑换配置 */
export function updatePAPExchangeConfig(data: PAPExchangeConfig) {
  return request.put<PAPExchangeConfig>({ url: '/api/v1/system/pap/config', data })
}

/** 管理员：月度归档 + 可选兑换系统钱包 */
export function settleAlliancePAPMonth(data: {
  year: number
  month: number
  wallet_convert: boolean
}) {
  return request.post<SettleMonthResult>({ url: '/api/v1/system/pap/settle', data })
}
