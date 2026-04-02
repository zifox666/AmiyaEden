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

test('welfare approval row description tooltip stays visible long enough after mouse leave', () => {
  assert.doesNotMatch(source, /utils\/ui\/tooltip/)
  assert.match(
    source,
    /const\s+ROW_DESCRIPTION_TOOLTIP_HIDE_DELAY_MS\s*=\s*800/,
    'expected welfare approval to own its local tooltip delay constant'
  )
  assert.match(
    source,
    /handleCellMouseLeave\(\)\s*\{[\s\S]*?ROW_DESCRIPTION_TOOLTIP_HIDE_DELAY_MS[\s\S]*?\)/,
    'expected handleCellMouseLeave to use the local tooltip hide delay constant'
  )
  assert.match(source, /effect="dark"/)
  assert.doesNotMatch(source, /:show-after="0"/)
})

test('welfare approval character rows use the shared copy button instead of page-local clipboard logic', () => {
  assert.match(
    source,
    /import ArtCopyButton from '@\/components\/core\/forms\/art-copy-button\/index.vue'/
  )
  assert.match(
    source,
    /prop:\s*'character_name'[\s\S]*h\(ArtCopyButton,[\s\S]*text:\s*row\.character_name/
  )
  assert.doesNotMatch(source, /const copyText = async \(text: string\)/)
})
