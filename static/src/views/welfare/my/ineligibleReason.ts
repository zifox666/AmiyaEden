export interface WelfareReasonMessages {
  pap: string
  skill: string
  papSkill: string
  skillPlan: (plans: string) => string
  papSkillPlan: (plans: string) => string
  planSeparator: string
}

export function formatWelfareIneligibleReason(
  reason: Api.Welfare.EligibleWelfare['ineligible_reason'] | undefined,
  skillPlanNames: string[] | undefined,
  messages: WelfareReasonMessages
) {
  const planNames = (skillPlanNames ?? []).map((name) => name.trim()).filter(Boolean)
  const joinedPlanNames = planNames.join(messages.planSeparator)

  if (reason === 'skill' && joinedPlanNames) {
    return messages.skillPlan(joinedPlanNames)
  }
  if (reason === 'pap_skill' && joinedPlanNames) {
    return messages.papSkillPlan(joinedPlanNames)
  }
  if (reason === 'pap_skill') {
    return messages.papSkill
  }
  if (reason === 'pap') {
    return messages.pap
  }
  return messages.skill
}
