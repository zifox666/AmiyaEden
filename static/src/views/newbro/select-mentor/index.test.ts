import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('mentor cards show mentor contact details', () => {
  assert.match(source, /mentor\.qq/)
  assert.match(source, /mentor\.discord_id/)
  assert.match(source, /newbro\.mentor\.qq/)
  assert.match(source, /newbro\.mentor\.discordId/)
})

test('current relationship card shows mentor contact details', () => {
  assert.match(source, /currentRelationship\.mentor_qq/)
  assert.match(source, /currentRelationship\.mentor_discord_id/)
})

test('available mentor section is hidden after the relationship becomes active', () => {
  assert.match(
    source,
    /const hasActiveRelationship = computed\(\(\) => currentRelationship\.value\?\.status === 'active'\)/
  )
  assert.match(source, /<ElCard shadow="never" v-if="!hasActiveRelationship">/)
})

test('mentor selection page allows mentor cards to grow with content', () => {
  assert.doesNotMatch(source, /<template>\s*<div class="[^"]*\bart-full-height\b[^"]*">/)
})

test('mentor selection page includes the mentor responsibilities card', () => {
  assert.match(source, /<MentorResponsibilitiesCard\b/)
})
