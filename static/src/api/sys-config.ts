import request from '@/utils/http'

export function fetchBasicConfig() {
  return request.get<Api.SysConfig.BasicConfig>({
    url: '/api/v1/system/basic-config'
  })
}

export function updateBasicConfig(data: Api.SysConfig.UpdateBasicConfigParams) {
  return request.put({
    url: '/api/v1/system/basic-config',
    data
  })
}

// ─── SeAT 配置 ───

export function fetchSeatConfig() {
  return request.get<Api.SysConfig.SeatConfig>({
    url: '/api/v1/system/seat-config'
  })
}

export function updateSeatConfig(data: Api.SysConfig.UpdateSeatConfigParams) {
  return request.put({
    url: '/api/v1/system/seat-config',
    data
  })
}
