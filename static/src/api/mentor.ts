import request from '@/utils/http'

export function fetchMentorCandidates() {
  return request.get<Api.Mentor.MentorCandidate[]>({
    url: '/api/v1/mentor/mentors'
  })
}

export function fetchMyMentorStatus() {
  return request.get<Api.Mentor.MyStatusResponse>({
    url: '/api/v1/mentor/me'
  })
}

export function applyForMentor(data: Api.Mentor.ApplyParams) {
  return request.post<Api.Mentor.ApplyResponse>({
    url: '/api/v1/mentor/apply',
    data
  })
}

export function fetchMentorApplications() {
  return request.get<Api.Mentor.MenteeListItem[]>({
    url: '/api/v1/mentor/dashboard/applications'
  })
}

export function fetchMentorMentees(params?: Api.Mentor.MentorMenteesParams) {
  return request.get<Api.Mentor.MenteeListResponse>({
    url: '/api/v1/mentor/dashboard/mentees',
    params
  })
}

export function fetchMentorDashboardRewardStages() {
  return request.get<Api.Mentor.RewardStage[]>({
    url: '/api/v1/mentor/dashboard/reward-stages'
  })
}

export function acceptMentorApplication(data: Api.Mentor.RelationshipActionParams) {
  return request.post<Api.Mentor.EmptyResponse>({
    url: '/api/v1/mentor/dashboard/accept',
    data
  })
}

export function rejectMentorApplication(data: Api.Mentor.RelationshipActionParams) {
  return request.post<Api.Mentor.EmptyResponse>({
    url: '/api/v1/mentor/dashboard/reject',
    data
  })
}

export function fetchAdminMentorRelationships(params?: Api.Mentor.AdminRelationshipsParams) {
  return request.get<Api.Mentor.AdminRelationshipsResponse>({
    url: '/api/v1/system/mentor/relationships',
    params
  })
}

export function revokeMentorRelationship(data: Api.Mentor.RelationshipActionParams) {
  return request.post<Api.Mentor.EmptyResponse>({
    url: '/api/v1/system/mentor/revoke',
    data
  })
}

export function fetchMentorRewardStages() {
  return request.get<Api.Mentor.RewardStage[]>({
    url: '/api/v1/system/mentor/reward-stages'
  })
}

export function updateMentorRewardStages(data: Api.Mentor.UpdateRewardStagesParams) {
  return request.put<Api.Mentor.RewardStage[]>({
    url: '/api/v1/system/mentor/reward-stages',
    data
  })
}

export function runMentorRewardProcessing() {
  return request.post<Api.Mentor.RewardProcessResult>({
    url: '/api/v1/system/mentor/reward/process'
  })
}
