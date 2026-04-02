import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const manageHookSource = readFileSync(
  new URL('../../../hooks/srp/useSrpManage.ts', import.meta.url),
  'utf8'
)
const workflowHookSource = readFileSync(
  new URL('../../../hooks/srp/useSrpWorkflow.ts', import.meta.url),
  'utf8'
)

test('srp manage uses the shared copy button for the character column and shared clipboard hook for copy flows', () => {
  assert.match(
    manageHookSource,
    /prop:\s*'character_name'[\s\S]*h\(ArtCopyButton,[\s\S]*text:\s*row\.character_name/
  )
  assert.match(workflowHookSource, /useClipboardCopy/)
  assert.doesNotMatch(workflowHookSource, /navigator\.clipboard\.writeText/)
})

test('srp batch payout copy text keeps exact ISK values instead of smart-abbreviated amounts', () => {
  assert.match(workflowHookSource, /formatBatchPayoutLine[\s\S]*formatIskPlain\(totalAmount\)/)
})
