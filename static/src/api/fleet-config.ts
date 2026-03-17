import request from '@/utils/http'

// ─── 舰队配置 CRUD ───

/** 创建舰队配置 */
export function createFleetConfig(data: Api.FleetConfig.CreateFleetConfigParams) {
  return request.post<Api.FleetConfig.FleetConfigItem>({
    url: '/api/v1/operation/fleet-configs',
    data
  })
}

/** 获取舰队配置列表 */
export function fetchFleetConfigList(params?: { current?: number; size?: number }) {
  return request.get<Api.FleetConfig.FleetConfigList>({
    url: '/api/v1/operation/fleet-configs',
    params
  })
}

/** 获取舰队配置详情 */
export function fetchFleetConfigDetail(id: number) {
  return request.get<Api.FleetConfig.FleetConfigItem>({
    url: `/api/v1/operation/fleet-configs/${id}`
  })
}

/** 更新舰队配置 */
export function updateFleetConfig(id: number, data: Api.FleetConfig.UpdateFleetConfigParams) {
  return request.put<Api.FleetConfig.FleetConfigItem>({
    url: `/api/v1/operation/fleet-configs/${id}`,
    data
  })
}

/** 删除舰队配置 */
export function deleteFleetConfig(id: number) {
  return request.del({
    url: `/api/v1/operation/fleet-configs/${id}`
  })
}

// ─── 装配导入 / 导出 ───

/** 从用户 ESI 装配导入为英文 EFT（供编辑表单预填充） */
export function importFittingFromUser(data: Api.FleetConfig.ImportFittingParams) {
  return request.post<Api.FleetConfig.ImportFittingResponse>({
    url: '/api/v1/operation/fleet-configs/import-fitting',
    data
  })
}

/** 将配置中的装配导出到 ESI */
export function exportFittingToESI(data: Api.FleetConfig.ExportToESIParams) {
  return request.post({
    url: '/api/v1/operation/fleet-configs/export-esi',
    data
  })
}

/** 获取舰队配置所有装配的本地化 EFT 文本 */
export function fetchFleetConfigEFT(id: number, lang?: string) {
  return request.get<Api.FleetConfig.EFTResponse>({
    url: `/api/v1/operation/fleet-configs/${id}/eft`,
    params: lang ? { lang } : undefined
  })
}

// ─── 装备详情 & 设置 ───

/** 获取装配物品详情（含重要性、惩罚、替代品） */
export function fetchFittingItems(configId: number, fittingId: number, lang?: string) {
  return request.get<Api.FleetConfig.FittingItemsResponse>({
    url: `/api/v1/operation/fleet-configs/${configId}/fittings/${fittingId}/items`,
    params: lang ? { lang } : undefined
  })
}

/** 批量更新装配物品设置（重要性、惩罚、替代品） */
export function updateFittingItemsSettings(configId: number, fittingId: number, data: Api.FleetConfig.UpdateItemsSettingsParams) {
  return request.put({
    url: `/api/v1/operation/fleet-configs/${configId}/fittings/${fittingId}/items/settings`,
    data
  })
}
