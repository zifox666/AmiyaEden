import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('info wallet renders translated ref type labels instead of raw ref_type values', () => {
  assert.match(source, /formatJournalTypeLabel\(row\.ref_type\)/)
  assert.doesNotMatch(source, /=> row\.ref_type/)
  assert.match(source, /walletAdmin\.refTypes\.\$\{value\}/)
})
