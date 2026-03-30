import request from '@/utils/http'

export function fetchBadgeCounts() {
  return request.get<Api.Badge.BadgeCounts>({
    url: '/api/v1/badge-counts'
  })
}
