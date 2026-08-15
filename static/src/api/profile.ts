import request from '@/utils/http'

export function fetchProfile() {
  return request.get<Api.Profile.Profile>({ url: '/api/v1/profile' })
}

export function updateProfile(data: Api.Profile.UpdateRequest) {
  return request.put<Api.Profile.Profile>({ url: '/api/v1/profile', data })
}
