import request from '@/utils/http'

export function fetchMumbleProfile() {
  return request.get<Api.Mumble.Profile>({
    url: '/api/v1/voice/mumble'
  })
}

export function resetMumblePassword() {
  return request.post<Api.Mumble.Account>({
    url: '/api/v1/voice/mumble/reset-password'
  })
}

export function fetchMumbleConfig() {
  return request.get<Api.Mumble.Config>({
    url: '/api/v1/system/mumble-config'
  })
}

export function updateMumbleConfig(data: Api.Mumble.UpdateConfigParams) {
  return request.put({
    url: '/api/v1/system/mumble-config',
    data
  })
}

export function fetchMumbleRoleGroups() {
  return request.get<Api.Mumble.RoleGroupMapping[]>({
    url: '/api/v1/system/mumble-role-groups'
  })
}

export function updateMumbleRoleGroups(data: Api.Mumble.UpdateRoleGroupsParams) {
  return request.put({
    url: '/api/v1/system/mumble-role-groups',
    data
  })
}
