import request from '@/utils/http'

// ─── 舰队 CRUD ───

/** 创建舰队 */
export function createFleet(data: Api.Fleet.CreateFleetParams) {
  return request.post<Api.Fleet.FleetItem>({
    url: '/api/v1/operation/fleets',
    data
  })
}

/** 获取舰队列表 */
export function fetchFleetList(params?: Api.Fleet.FleetSearchParams) {
  return request.get<Api.Fleet.FleetList>({
    url: '/api/v1/operation/fleets',
    params
  })
}

/** 获取舰队详情 */
export function fetchFleetDetail(id: string) {
  return request.get<Api.Fleet.FleetItem>({
    url: `/api/v1/operation/fleets/${id}`
  })
}

/** 更新舰队 */
export function updateFleet(id: string, data: Api.Fleet.UpdateFleetParams) {
  return request.put<Api.Fleet.FleetItem>({
    url: `/api/v1/operation/fleets/${id}`,
    data
  })
}

/** 从 ESI 刷新舰队 ID */
export function refreshFleetESI(id: string) {
  return request.post<Api.Fleet.FleetItem>({
    url: `/api/v1/operation/fleets/${id}/refresh-esi`
  })
}

/** 删除舰队 */
export function deleteFleet(id: string) {
  return request.del({
    url: `/api/v1/operation/fleets/${id}`
  })
}

// ─── 舰队成员 ───

/** 获取舰队成员列表 */
export function fetchFleetMembers(fleetId: string) {
  return request.get<Api.Fleet.FleetMember[]>({
    url: `/api/v1/operation/fleets/${fleetId}/members`
  })
}

/** 从 ESI 同步舰队成员 */
export function syncESIFleetMembers(fleetId: string) {
  return request.post<Api.Fleet.ESIFleetMember[]>({
    url: `/api/v1/operation/fleets/${fleetId}/members/sync`
  })
}

// ─── PAP ───

/** 发放 PAP */
export function issuePap(fleetId: string) {
  return request.post({
    url: `/api/v1/operation/fleets/${fleetId}/pap`
  })
}

/** 获取舰队 PAP 记录 */
export function fetchFleetPapLogs(fleetId: string) {
  return request.get<Api.Fleet.PapLog[]>({
    url: `/api/v1/operation/fleets/${fleetId}/pap`
  })
}

/** 获取我的 PAP 记录 */
export function fetchMyPapLogs() {
  return request.get<Api.Fleet.PapLog[]>({
    url: '/api/v1/operation/fleets/pap/me'
  })
}

// ─── 邀请链接 ───

/** 创建舰队邀请链接 */
export function createFleetInvite(fleetId: string) {
  return request.post<Api.Fleet.FleetInvite>({
    url: `/api/v1/operation/fleets/${fleetId}/invites`
  })
}

/** 获取舰队邀请链接列表 */
export function fetchFleetInvites(fleetId: string) {
  return request.get<Api.Fleet.FleetInvite[]>({
    url: `/api/v1/operation/fleets/${fleetId}/invites`
  })
}

/** 禁用邀请链接 */
export function deactivateFleetInvite(inviteId: number) {
  return request.del({
    url: `/api/v1/operation/fleets/invites/${inviteId}`
  })
}

/** 通过邀请码加入舰队 */
export function joinFleet(data: Api.Fleet.JoinFleetParams) {
  return request.post({
    url: '/api/v1/operation/fleets/join',
    data
  })
}

// ─── ESI 舰队 ───

/** 获取角色当前 ESI 舰队信息 */
export function fetchCharacterFleetInfo(characterId: number) {
  return request.get<Api.Fleet.CharacterFleetInfo>({
    url: `/api/v1/operation/fleets/esi/${characterId}`
  })
}
