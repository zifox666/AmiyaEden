import request from '@/utils/http'

export interface PAPTypeRate {
  pap_type: string
  display_name: string
  rate: number
  updated_at: string
}

/** 管理员：获取 PAP 类型兑换汇率列表 */
export function fetchPAPTypeRates() {
  return request.get<PAPTypeRate[]>({ url: '/api/v1/system/pap-exchange/rates' })
}

/** 管理员：批量更新 PAP 类型兑换汇率 */
export function updatePAPTypeRates(
  data: Pick<PAPTypeRate, 'pap_type' | 'display_name' | 'rate'>[]
) {
  return request.put<PAPTypeRate[]>({ url: '/api/v1/system/pap-exchange/rates', data })
}
