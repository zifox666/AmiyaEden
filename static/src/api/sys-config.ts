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

/** 获取允许军团列表 */
export function fetchAllowCorporations() {
  return request.get<Api.SysConfig.AllowCorporationsConfig>({
    url: '/api/v1/system/basic-config/allow-corporations'
  })
}

/** 更新允许军团列表 */
export function updateAllowCorporations(data: Api.SysConfig.UpdateAllowCorporationsParams) {
  return request.put({
    url: '/api/v1/system/basic-config/allow-corporations',
    data
  })
}

/** 获取 SDE 配置 */
export function fetchSDEConfig() {
  return request.get<Api.SysConfig.SDEConfig>({
    url: '/api/v1/system/sde-config'
  })
}

/** 更新 SDE 配置 */
export function updateSDEConfig(data: Api.SysConfig.UpdateSDEConfigParams) {
  return request.put({
    url: '/api/v1/system/sde-config',
    data
  })
}
