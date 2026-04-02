import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('welfare approval history tab keeps ledger defaults and uses full-height tab layout', () => {
  assert.doesNotMatch(source, /:pagination-options="historyPaginationOptions"/)
  assert.match(source, /apiParams: \{ current: 1, size: 200, status: 'delivered,rejected' \}/)
  assert.match(source, /visual-variant="ledger"/)
  assert.match(source, /:deep\(\.el-card__body\)\s*\{[\s\S]*display:\s*flex/)
  assert.match(source, /:deep\(\.el-tabs__content\)\s*\{[\s\S]*overflow:\s*hidden/)
  assert.match(source, /:deep\(\.el-tab-pane\)\s*\{[\s\S]*height:\s*100%/)
})
