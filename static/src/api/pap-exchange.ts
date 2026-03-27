import request from '@/utils/http'

export type PAPExchangeRate = Api.PapExchange.RateItem

export type PAPExchangeConfig = Api.PapExchange.ConfigResponse

export type UpdatePAPExchangeConfigParams = Api.PapExchange.UpdateConfigParams

/** 管理员：获取 PAP 兑换配置 */
export function fetchPAPExchangeConfig() {
  return request.get<PAPExchangeConfig>({ url: '/api/v1/system/pap-exchange/rates' })
}

/** 管理员：更新 PAP 兑换配置 */
export function updatePAPExchangeConfig(data: UpdatePAPExchangeConfigParams) {
  return request.put<PAPExchangeConfig>({ url: '/api/v1/system/pap-exchange/rates', data })
}
