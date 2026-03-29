import request from '@/utils/http'

/** 获取个人刷怪报表 */
export function fetchNpcKills(data: Api.NpcKill.NpcKillRequest) {
  return request.post<Api.NpcKill.NpcKillResponse>({ url: '/api/v1/info/npc-kills', data })
}

/** 获取名下所有人物的汇总刷怪报表 */
export function fetchNpcKillsAll(data: Api.NpcKill.NpcKillAllRequest) {
  return request.post<Api.NpcKill.NpcKillResponse>({ url: '/api/v1/info/npc-kills/all', data })
}

/** 获取公司刷怪报表（管理员） */
export function fetchCorpNpcKills(data: Api.NpcKill.NpcKillCorpRequest) {
  return request.post<Api.NpcKill.NpcKillCorpResponse>({ url: '/api/v1/system/npc-kills', data })
}
