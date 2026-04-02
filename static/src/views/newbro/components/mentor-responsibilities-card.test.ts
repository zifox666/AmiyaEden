import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(
  new URL('./mentor-responsibilities-card.vue', import.meta.url),
  'utf8'
)

test('mentor responsibilities card renders a localized title and four responsibility items', () => {
  assert.match(source, /newbro\.mentor\.responsibilitiesTitle/)
  assert.match(source, /newbro\.mentor\.responsibilitiesDescription/)

  for (const key of [
    'newbro.mentor.responsibilityItems.registration',
    'newbro.mentor.responsibilityItems.basics',
    'newbro.mentor.responsibilityItems.pvp',
    'newbro.mentor.responsibilityItems.scope'
  ]) {
    assert.match(source, new RegExp(key.replace(/\./g, '\\.')))
  }
})
