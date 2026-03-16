import request from '@/utils/http'

/** 获取角色钱包流水 */
export function fetchInfoWallet(data: Api.EveInfo.WalletRequest) {
  return request.post<Api.EveInfo.WalletResponse>({ url: '/api/v1/info/wallet', data })
}

/** 获取角色技能列表与队列 */
export function fetchInfoSkills(data: Api.EveInfo.SkillRequest) {
  return request.post<Api.EveInfo.SkillResponse>({ url: '/api/v1/info/skills', data })
}

/** 获取角色可用舰船列表 */
export function fetchInfoShips(data: Api.EveInfo.ShipRequest) {
  return request.post<Api.EveInfo.ShipResponse>({ url: '/api/v1/info/ships', data })
}

/** 获取角色克隆体/植入体信息 */
export function fetchInfoImplants(data: Api.EveInfo.ImplantsRequest) {
  return request.post<Api.EveInfo.ImplantsResponse>({ url: '/api/v1/info/implants', data })
}

/** 获取用户所有角色的装配列表 */
export function fetchInfoFittings(data: Api.EveInfo.FittingsRequest) {
  return request.post<Api.EveInfo.FittingsListResponse>({ url: '/api/v1/info/fittings', data })
}

/** 保存装配（同步 ESI） */
export function saveInfoFitting(data: Api.EveInfo.SaveFittingRequest) {
  return request.post<Api.EveInfo.FittingResponse>({ url: '/api/v1/info/fittings/save', data })
}

/** 获取用户所有角色的资产列表 */
export function fetchInfoAssets(data: Api.EveInfo.AssetsRequest) {
  return request.post<Api.EveInfo.AssetsResponse>({ url: '/api/v1/info/assets', data })
}

/** 获取用户所有角色的合同列表（分页） */
export function fetchInfoContracts(data: Api.EveInfo.ContractsRequest) {
  return request.post<Api.Common.PaginatedResponse<Api.EveInfo.ContractItem>>({
    url: '/api/v1/info/contracts',
    data
  })
}

/** 获取指定合同的物品与竞标详情 */
export function fetchInfoContractDetail(data: Api.EveInfo.ContractDetailRequest) {
  return request.post<Api.EveInfo.ContractDetailResponse>({
    url: '/api/v1/info/contracts/detail',
    data
  })
}
