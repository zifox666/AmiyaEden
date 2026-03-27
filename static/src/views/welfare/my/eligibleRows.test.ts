import assert from 'node:assert/strict'
import test from 'node:test'
import { sortEligibleRows } from './eligibleRows'

test('sortEligibleRows puts current options before future options', () => {
  const rows = sortEligibleRows([
    { canApplyNow: false, label: 'future-a' },
    { canApplyNow: true, label: 'current-a' },
    { canApplyNow: false, label: 'future-b' },
    { canApplyNow: true, label: 'current-b' }
  ])

  assert.deepEqual(
    rows.map((row) => row.label),
    ['current-a', 'current-b', 'future-a', 'future-b']
  )
})
