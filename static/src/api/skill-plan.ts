import request from '@/utils/http'

/** 获取军团技能计划列表 */
export function fetchSkillPlanList(params?: Api.SkillPlan.SkillPlanSearchParams) {
  return request.get<Api.SkillPlan.SkillPlanList>({
    url: '/api/v1/skill-planning/skill-plans',
    params
  })
}

/** 获取军团技能计划详情 */
export function fetchSkillPlanDetail(id: number, lang?: string) {
  return request.get<Api.SkillPlan.SkillPlanDetail>({
    url: `/api/v1/skill-planning/skill-plans/${id}`,
    params: lang ? { lang } : undefined
  })
}

/** 创建军团技能计划 */
export function createSkillPlan(data: Api.SkillPlan.CreateSkillPlanParams, lang?: string) {
  return request.post<Api.SkillPlan.SkillPlanDetail>({
    url: '/api/v1/skill-planning/skill-plans',
    data,
    params: lang ? { lang } : undefined
  })
}

/** 更新军团技能计划 */
export function updateSkillPlan(
  id: number,
  data: Api.SkillPlan.UpdateSkillPlanParams,
  lang?: string
) {
  return request.put<Api.SkillPlan.SkillPlanDetail>({
    url: `/api/v1/skill-planning/skill-plans/${id}`,
    data,
    params: lang ? { lang } : undefined
  })
}

/** 删除军团技能计划 */
export function deleteSkillPlan(id: number) {
  return request.del({
    url: `/api/v1/skill-planning/skill-plans/${id}`
  })
}

/** 获取技能完成度检查角色选择 */
export function fetchSkillPlanCheckSelection() {
  return request.get<Api.SkillPlan.CheckSelection>({
    url: '/api/v1/skill-planning/skill-plans/check/selection'
  })
}

/** 保存技能完成度检查角色选择 */
export function saveSkillPlanCheckSelection(data: Api.SkillPlan.CheckSelection) {
  return request.put<Api.SkillPlan.CheckSelection>({
    url: '/api/v1/skill-planning/skill-plans/check/selection',
    data
  })
}

/** 执行技能完成度检查 */
export function runSkillPlanCompletionCheck(data?: Api.SkillPlan.CompletionCheckParams) {
  return request.post<Api.SkillPlan.CompletionCheckResult>({
    url: '/api/v1/skill-planning/skill-plans/check/run',
    data
  })
}
