import request from '@/utils/http'

export function fetchCorpStructureList(data: Api.CorpStructure.ListRequest) {
  return request.post<Api.CorpStructure.StructureList>({
    url: '/api/v1/operation/corp-structures/list',
    data
  })
}

export function fetchCorpIDs() {
  return request.get<number[]>({
    url: '/api/v1/operation/corp-structures/corps'
  })
}

export function fetchCorpStructureFuelSetting(corpID?: number) {
  return request.get<Api.CorpStructure.FuelSetting>({
    url: '/api/v1/operation/corp-structures/fuel/settings',
    params: corpID ? { corp_id: corpID } : undefined
  })
}

export function updateCorpStructureFuelSetting(data: Api.CorpStructure.FuelSettingUpdateRequest) {
  return request.put<null>({
    url: '/api/v1/operation/corp-structures/fuel/settings',
    data
  })
}

export function claimCorpStructureFuelTask(structureID: number) {
  return request.post<null>({
    url: `/api/v1/operation/corp-structures/${structureID}/fuel/claim`
  })
}

export function cancelCorpStructureFuelTask(structureID: number) {
  return request.post<null>({
    url: `/api/v1/operation/corp-structures/${structureID}/fuel/cancel`
  })
}

export function settleCorpStructureFuelTask(structureID: number) {
  return request.post<Api.CorpStructure.FuelSettleResult>({
    url: `/api/v1/operation/corp-structures/${structureID}/fuel/settle`
  })
}

export function fetchCorpStructureFuelTaskList(data: Api.CorpStructure.FuelTaskListRequest) {
  return request.post<Api.CorpStructure.FuelTaskList>({
    url: '/api/v1/operation/corp-structures/fuel-tasks/list',
    data
  })
}

export function markCorpStructureFuelTaskIskPaid(taskID: number, note?: string) {
  return request.post<null>({
    url: `/api/v1/operation/corp-structures/fuel-tasks/${taskID}/isk/mark-paid`,
    data: { note }
  })
}
