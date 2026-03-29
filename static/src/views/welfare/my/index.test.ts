import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('welfare my tables use ArtTable emptyText instead of standalone ElEmpty blocks', () => {
  assert.match(source, /<ArtTable[\s\S]*?:empty-text="t\('welfareMy\.noEligibleWelfares'\)"/)
  assert.match(source, /<ArtTable[\s\S]*?:empty-text="t\('welfareMy\.noApplications'\)"/)
  assert.doesNotMatch(source, /<ElEmpty/)
})
