import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const viewSource = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')
const apiSource = readFileSync(new URL('../../../api/mentor.ts', import.meta.url), 'utf8')
const typeSource = readFileSync(new URL('../../../types/api/api.d.ts', import.meta.url), 'utf8')
const zhLocaleSource = readFileSync(new URL('../../../locales/langs/zh.json', import.meta.url), 'utf8')
const enLocaleSource = readFileSync(new URL('../../../locales/langs/en.json', import.meta.url), 'utf8')

test('mentor reward stages page wires mentor eligibility settings separately', () => {
  assert.match(apiSource, /export function fetchMentorSettings\(/)
  assert.match(apiSource, /export function updateMentorSettings\(/)

  assert.match(typeSource, /interface Settings\s*\{[\s\S]*max_character_sp: number/)
  assert.match(typeSource, /interface Settings\s*\{[\s\S]*max_account_age_days: number/)
  assert.match(typeSource, /interface UpdateSettingsParams\s*\{[\s\S]*max_character_sp: number/)
  assert.match(typeSource, /interface UpdateSettingsParams\s*\{[\s\S]*max_account_age_days: number/)

  assert.match(viewSource, /fetchMentorSettings/)
  assert.match(viewSource, /updateMentorSettings/)
  assert.match(viewSource, /system\.mentorRewardStages\.eligibilityTitle/)
  assert.match(viewSource, /system\.mentorRewardStages\.maxCharacterSP/)
  assert.match(viewSource, /system\.mentorRewardStages\.maxAccountAgeDays/)
})

test('mentor reward stages locales include eligibility settings copy', () => {
  assert.match(zhLocaleSource, /"eligibilityTitle"\s*:/)
  assert.match(zhLocaleSource, /"maxCharacterSP"\s*:/)
  assert.match(zhLocaleSource, /"maxAccountAgeDays"\s*:/)
  assert.match(zhLocaleSource, /"saveEligibilitySuccess"\s*:/)

  assert.match(enLocaleSource, /"eligibilityTitle"\s*:/)
  assert.match(enLocaleSource, /"maxCharacterSP"\s*:/)
  assert.match(enLocaleSource, /"maxAccountAgeDays"\s*:/)
  assert.match(enLocaleSource, /"saveEligibilitySuccess"\s*:/)
})

test('mentor reward stage and eligibility number inputs are configured for integers only', () => {
  const inputNumbers = viewSource.match(/<ElInputNumber[\s\S]*?\/>/g) ?? []

  assert.equal(inputNumbers.length, 5)

  for (const inputNumber of inputNumbers) {
    assert.match(inputNumber, /:controls="false"/)
    assert.doesNotMatch(inputNumber, /:precision="/)
    assert.doesNotMatch(inputNumber, /0\.01/)
  }

  assert.match(
    viewSource,
    /v-model="mentorSettings\.max_character_sp"[\s\S]*?:step="1000000"[\s\S]*?step-strictly/
  )
  assert.match(
    viewSource,
    /v-model="mentorSettings\.max_account_age_days"[\s\S]*?:step="1"[\s\S]*?step-strictly/
  )
})
