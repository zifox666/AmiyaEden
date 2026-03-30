import assert from 'node:assert/strict'
import test from 'node:test'

import { formatWelfareIneligibleReason, type WelfareReasonMessages } from './ineligibleReason'

const messages: WelfareReasonMessages = {
  pap: '军团PAP数不足',
  skill: '技能未达标',
  papSkill: '技能未达标，军团PAP数不足',
  skillPlan: (plans) => `技能规划${plans}未达成`,
  papSkillPlan: (plans) => `技能规划${plans}未达成，军团PAP数不足`,
  planSeparator: '或'
}

test('formatWelfareIneligibleReason falls back to the existing skill message when no plan names exist', () => {
  assert.equal(formatWelfareIneligibleReason('skill', [], messages), '技能未达标')
})

test('formatWelfareIneligibleReason joins multiple skill plans with the localized separator', () => {
  assert.equal(
    formatWelfareIneligibleReason('skill', ['护盾方案', '装甲方案'], messages),
    '技能规划护盾方案或装甲方案未达成'
  )
})

test('formatWelfareIneligibleReason preserves the PAP warning when both checks fail', () => {
  assert.equal(
    formatWelfareIneligibleReason('pap_skill', ['护盾方案', '装甲方案'], messages),
    '技能规划护盾方案或装甲方案未达成，军团PAP数不足'
  )
})
