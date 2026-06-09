import request from '@/utils/http'

/** 获取军团建筑列表 */
export function fetchCorpStructureList(data: Api.CorpStructure.ListRequest) {
  return request.post<Api.CorpStructure.StructureList>({
    url: '/api/v1/operation/corp-structures/list',
    data
  })
}

/** 获取用户关联的军团 ID 列表 */
export function fetchCorpIDs() {
  return request.get<number[]>({
    url: '/api/v1/operation/corp-structures/corps'
  })
}

/** 获取建筑承接与贡献设置 */
export function fetchCorpStructureFuelSetting(corpID?: number) {
  return request.get<Api.CorpStructure.FuelSetting>({
    url: '/api/v1/operation/corp-structures/fuel/settings',
    params: corpID ? { corp_id: corpID } : undefined
  })
}

/** 更新建筑承接与贡献设置 */
export function updateCorpStructureFuelSetting(data: Api.CorpStructure.FuelSettingUpdateRequest) {
  return request.put<null>({
    url: '/api/v1/operation/corp-structures/fuel/settings',
    data
  })
}

/** 承接建筑加油任务 */
export function claimCorpStructureFuelTask(structureID: number) {
  return request.post<null>({
    url: `/api/v1/operation/corp-structures/${structureID}/fuel/claim`
  })
}

/** 结算建筑加油贡献 */
export function settleCorpStructureFuelTask(structureID: number) {
  return request.post<Api.CorpStructure.FuelSettleResult>({
    url: `/api/v1/operation/corp-structures/${structureID}/fuel/settle`
  })
}

/** 标记ISK手动发放完成 */
export function markCorpStructureFuelTaskIskPaid(taskID: number, note?: string) {
  return request.post<null>({
    url: `/api/v1/operation/corp-structures/fuel-tasks/${taskID}/isk/mark-paid`,
    data: { note }
  })
}
