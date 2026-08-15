import request from '@/utils/http'

export const atApi = {
  listMatches: () => request.get<Api.AT.Match[]>({ url: '/api/v1/at/matches' }),
  getMatch: (id: number) => request.get<Api.AT.MatchDetail>({ url: `/api/v1/at/matches/${id}` }),
  getMatchByUUID: (uuid: string) =>
    request.get<Api.AT.MatchDetail>({ url: `/api/v1/at/matches/by-uuid/${uuid}` }),
  createMatch: (data: Api.AT.MatchCreate) =>
    request.post<Api.AT.Match>({ url: '/api/v1/at/matches', data }),
  registerMatch: (id: number) => request.post({ url: `/api/v1/at/matches/${id}/register` }),
  selectMatchPosition: (id: number, data: Omit<Api.AT.MatchPositionInput, 'user_id'>) =>
    request.post({ url: `/api/v1/at/matches/${id}/positions/self`, data }),
  assignMatchPosition: (id: number, data: Api.AT.MatchPositionInput) =>
    request.put({ url: `/api/v1/at/matches/${id}/positions`, data }),
  getMatchPositionEFT: (id: number, slot: number) =>
    request.get<{ slot_id: number; slot_order: number; eft: string }>({
      url: `/api/v1/at/matches/${id}/positions/${slot}/eft`
    }),
  searchMatchUsers: (keyword: string) =>
    request.get<Api.AT.ContributionUserOption[]>({
      url: '/api/v1/at/matches/users/search',
      params: { keyword }
    }),
  transitionMatch: (id: number, data: { status: string; reason: string }) =>
    request.put({ url: `/api/v1/at/matches/${id}/status`, data }),
  decideRegistration: (id: number, userId: number, status: string) =>
    request.put({
      url: `/api/v1/at/matches/${id}/registrations`,
      data: { user_id: userId, status }
    }),
  createRound: (id: number, roster: Api.AT.RoundRoster[]) =>
    request.post<Api.AT.MatchRound>({ url: `/api/v1/at/matches/${id}/rounds`, data: { roster } }),
  setRoundRoster: (id: number, roster: Api.AT.RoundRoster[]) =>
    request.put({ url: `/api/v1/at/matches/rounds/${id}/roster`, data: { roster } }),
  getRoundRoster: (id: number) =>
    request.get<Api.AT.RoundRoster[]>({ url: `/api/v1/at/matches/rounds/${id}/roster` }),
  finishRound: (id: number) => request.post({ url: `/api/v1/at/matches/rounds/${id}/finish` }),
  reopenRound: (id: number, reason: string) =>
    request.post<Api.AT.MatchRound>({
      url: `/api/v1/at/matches/rounds/${id}/reopen`,
      data: { reason }
    }),
  confirmRound: (id: number, result: 'a_win' | 'b_win' | 'draw') =>
    request.post({ url: `/api/v1/at/matches/rounds/${id}/confirm`, data: { result } }),
  resetMatch: (id: number, reason: string) =>
    request.post({ url: `/api/v1/at/matches/${id}/reset`, data: { reason } }),

  listSets: () => request.get<Api.AT.Set[]>({ url: '/api/v1/at/sets' }),
  listFits: () => request.get<Api.AT.Fit[]>({ url: '/api/v1/at/fits' }),
  getFit: (id: number) => request.get<Api.AT.FitDetail>({ url: `/api/v1/at/fits/${id}` }),
  saveFit: (id: number | null, data: Api.AT.FitInput) =>
    id
      ? request.put<Api.AT.FitDetail>({ url: `/api/v1/at/fits/${id}`, data })
      : request.post<Api.AT.FitDetail>({ url: '/api/v1/at/fits', data }),
  deleteFit: (id: number) => request.del({ url: `/api/v1/at/fits/${id}` }),
  getSet: (id: number) => request.get<Api.AT.SetDetail>({ url: `/api/v1/at/sets/${id}` }),
  saveSet: (id: number | null, data: Api.AT.SetInput) =>
    id
      ? request.put<Api.AT.SetDetail>({ url: `/api/v1/at/sets/${id}`, data })
      : request.post<Api.AT.SetDetail>({ url: '/api/v1/at/sets', data }),
  parseSetEFT: (eft: string) =>
    request.post<Api.AT.ParsedEFT>({ url: '/api/v1/at/sets/parse-eft', data: { eft } }),
  getSetEFT: (id: number) => request.get<Api.AT.SetEFT>({ url: `/api/v1/at/sets/${id}/eft` }),

  listScoreRules: () => request.get<Api.AT.ScoreRule[]>({ url: '/api/v1/at/score-rules' }),
  getScoreRule: (id: number) =>
    request.get<Api.AT.ScoreRuleDetail>({ url: `/api/v1/at/score-rules/${id}` }),
  importScoreRule: (file: File) => {
    const data = new FormData()
    data.append('file', file)
    return request.post<Api.AT.ScoreRuleDetail>({
      url: '/api/v1/at/score-rules/import',
      data,
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  publishScoreRule: (id: number) => request.post({ url: `/api/v1/at/score-rules/${id}/publish` }),

  submitKillmails: (text: string) =>
    request.post<Api.AT.KMSubmitResult>({ url: '/api/v1/at/killmails/submit', data: { text } }),
  listPendingKillmails: () =>
    request.get<Api.AT.Killmail[]>({ url: '/api/v1/at/killmails/pending' }),
  listKillmails: () => request.get<Api.AT.Killmail[]>({ url: '/api/v1/at/killmails' }),
  assignKillmail: (id: string, matchId: number, roundId: number) =>
    request.post<{ status: string }>({
      url: `/api/v1/at/killmails/${id}/assign`,
      data: { match_id: matchId, round_id: roundId }
    }),

  latestBR: (roundId: number) =>
    request.get<Api.AT.BRDetail>({ url: `/api/v1/at/br/rounds/${roundId}/latest` }),
  listBRs: () => request.get<Api.AT.BRListItem[]>({ url: '/api/v1/at/br' }),
  buildBR: (roundId: number, revisionReason = '') =>
    request.post<Api.AT.BRDetail>({
      url: `/api/v1/at/br/rounds/${roundId}/build`,
      data: { revision_reason: revisionReason }
    }),

  listSeasons: () => request.get<Api.AT.Season[]>({ url: '/api/v1/at/contributions/seasons' }),
  listMatchTypes: () =>
    request.get<Api.AT.MatchType[]>({ url: '/api/v1/at/contributions/match-types' }),
  saveSeason: (data: Partial<Api.AT.Season> & Pick<Api.AT.Season, 'name'>) =>
    data.id
      ? request.put<Api.AT.Season>({ url: `/api/v1/at/contributions/seasons/${data.id}`, data })
      : request.post<Api.AT.Season>({ url: '/api/v1/at/contributions/seasons', data }),
  setCleanupShips: (seasonId: number, typeIds: number[]) =>
    request.put({
      url: `/api/v1/at/contributions/seasons/${seasonId}/cleanup-ships`,
      data: { type_ids: typeIds }
    }),
  listCleanupShips: (seasonId: number, language: string) =>
    request.get<Api.AT.CleanupShip[]>({
      url: `/api/v1/at/contributions/seasons/${seasonId}/cleanup-ships`,
      params: { language }
    }),
  saveMatchType: (data: Partial<Api.AT.MatchType> & Pick<Api.AT.MatchType, 'name'>) =>
    data.id
      ? request.put<Api.AT.MatchType>({
          url: `/api/v1/at/contributions/match-types/${data.id}`,
          data
        })
      : request.post<Api.AT.MatchType>({ url: '/api/v1/at/contributions/match-types', data }),
  manualContribution: (data: {
    season_id: number
    user_id: number
    amount: number
    reason: string
  }) => request.post({ url: '/api/v1/at/contributions/manual', data }),
  searchContributionUsers: (keyword: string) =>
    request.get<Api.AT.ContributionUserOption[]>({
      url: '/api/v1/at/contributions/users/search',
      params: { keyword }
    }),
  leaderboard: (seasonId?: number) =>
    request.get<Api.AT.ContributionRow[]>({
      url: '/api/v1/at/contributions/leaderboard',
      params: seasonId ? { season_id: seasonId } : undefined
    }),
  myContributionLedger: (seasonId?: number) =>
    request.get<Api.AT.ContributionLedger[]>({
      url: '/api/v1/at/contributions/me',
      params: seasonId ? { season_id: seasonId } : undefined
    })
}

export const recruitmentApi = {
  uploadImage: (file: File) => {
    const data = new FormData()
    data.append('file', file)
    return request.post<{ url: string }>({
      url: '/api/v1/recruitment/images',
      data,
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  publicContext: (token: string) =>
    request.get<Api.AT.RecruitmentContext>({ url: `/api/v1/recruitment/public/${token}` }),
  submit: (token: string, answers: Record<string, unknown>) =>
    request.post<Api.AT.RecruitmentApplication>({
      url: `/api/v1/recruitment/public/${token}/applications`,
      data: { answers }
    }),
  withdraw: (id: number) => request.put({ url: `/api/v1/recruitment/applications/${id}/withdraw` }),
  listMyApplications: () =>
    request.get<Api.AT.RecruitmentApplication[]>({ url: '/api/v1/recruitment/my-applications' }),
  listQuestionnaires: () =>
    request.get<Api.AT.Questionnaire[]>({ url: '/api/v1/recruitment/questionnaires' }),
  getQuestionnaire: (id: number) =>
    request.get<Api.AT.QuestionnaireDetail>({ url: `/api/v1/recruitment/questionnaires/${id}` }),
  createQuestionnaire: (data: Api.AT.QuestionnaireInput) =>
    request.post<Api.AT.QuestionnaireDetail>({ url: '/api/v1/recruitment/questionnaires', data }),
  updateQuestionnaire: (id: number, data: Api.AT.QuestionnaireInput, publish: boolean) =>
    request.put<Api.AT.QuestionnaireDetail>({
      url: `/api/v1/recruitment/questionnaires/${id}`,
      params: { publish },
      data
    }),
  listLinks: () => request.get<Api.AT.RecruitmentLink[]>({ url: '/api/v1/recruitment/links' }),
  createLink: (data: {
    questionnaire_id: number
    expires_at: string | null
    max_submissions: number
  }) => request.post<Api.AT.CreatedRecruitmentLink>({ url: '/api/v1/recruitment/links', data }),
  setLinkActive: (id: number, active: boolean) =>
    request.put({ url: `/api/v1/recruitment/links/${id}/active`, data: { active } }),
  listApplications: () =>
    request.get<Api.AT.RecruitmentApplication[]>({ url: '/api/v1/recruitment/applications' }),
  getApplication: (id: number) =>
    request.get<Api.AT.RecruitmentApplicationDetail>({
      url: `/api/v1/recruitment/applications/${id}`
    }),
  reviewApplication: (id: number, status: string, reason: string) =>
    request.put({ url: `/api/v1/recruitment/applications/${id}/review`, data: { status, reason } }),
  privacyCleanup: (reason: string) =>
    request.post<{ applications_cleaned: number }>({
      url: '/api/v1/recruitment/privacy-cleanup',
      data: { reason }
    })
}
