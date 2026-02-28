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

/** 管理员：获取所有成员某月联盟 PAP 汇总 */
export function fetchAllAlliancePAP(params?: { year?: number; month?: number }) {
  return request.get<AlliancePAPAllResult>({
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
