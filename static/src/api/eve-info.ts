import request from '@/utils/http'

/** 获取角色钱包流水 */
export function fetchInfoWallet(data: Api.EveInfo.WalletRequest) {
  return request.post<Api.EveInfo.WalletResponse>({ url: '/api/v1/info/wallet', data })
}

/** 获取角色技能列表与队列 */
export function fetchInfoSkills(data: Api.EveInfo.SkillRequest) {
  return request.post<Api.EveInfo.SkillResponse>({ url: '/api/v1/info/skills', data })
}
