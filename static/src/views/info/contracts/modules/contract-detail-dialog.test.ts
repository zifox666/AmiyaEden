import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const dialogSource = readFileSync(new URL('./contract-detail-dialog.vue', import.meta.url), 'utf8')

test('contract detail dialog skips loading when character or contract ids are missing', () => {
  assert.match(dialogSource, /if \(!props\.characterId \|\| !props\.contractId\) return/)
})
