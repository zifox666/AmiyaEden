import request from '@/utils/http'

// ─── 技能规划 CRUD ───

/** 创建技能规划 */
export function createSkillPlan(data: Api.SkillPlan.CreateSkillPlanRequest) {
  return request.post<Api.SkillPlan.SkillPlanDTO>({
    url: '/api/v1/operation/skill-plans',
    data
  })
}

/** 获取技能规划列表（分页，管理用） */
export function fetchSkillPlanList(params?: { current?: number; size?: number }) {
  return request.get<Api.Common.PaginatedResponse<Api.SkillPlan.SkillPlanDTO>>({
    url: '/api/v1/operation/skill-plans',
    params
  })
}

/** 获取全部技能规划（下拉选项用） */
export function fetchAllSkillPlans() {
  return request.get<Api.SkillPlan.SkillPlanDTO[]>({
    url: '/api/v1/operation/skill-plans/all'
  })
}

/** 获取技能规划详情 */
export function fetchSkillPlanDetail(id: number, lang?: string) {
  return request.get<Api.SkillPlan.SkillPlanDTO>({
    url: `/api/v1/operation/skill-plans/${id}`,
    params: lang ? { lang } : undefined
  })
}

/** 更新技能规划 */
export function updateSkillPlan(id: number, data: Api.SkillPlan.UpdateSkillPlanRequest) {
  return request.put<Api.SkillPlan.SkillPlanDTO>({
    url: `/api/v1/operation/skill-plans/${id}`,
    data
  })
}

/** 删除技能规划 */
export function deleteSkillPlan(id: number) {
  return request.del({
    url: `/api/v1/operation/skill-plans/${id}`
  })
}

// ─── 技能检查 ───

/** 检查所有角色（管理员/FC） */
export function checkAllCharacters(id: number, lang?: string) {
  return request.get<Api.SkillPlan.SkillCheckSummary>({
    url: `/api/v1/operation/skill-plans/${id}/check`,
    params: lang ? { lang } : undefined
  })
}

/** 检查当前用户角色 */
export function checkMyCharacters(id: number, lang?: string) {
  return request.get<Api.SkillPlan.SkillCheckSummary>({
    url: `/api/v1/operation/skill-plans/${id}/check/me`,
    params: lang ? { lang } : undefined
  })
}
