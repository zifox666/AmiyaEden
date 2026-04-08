import assert from 'node:assert/strict'
import test from 'node:test'

import { formatWelfareHistoryReviewerName } from './reviewerName'

test('formatWelfareHistoryReviewerName returns reviewer nickname when present', () => {
  assert.equal(
    formatWelfareHistoryReviewerName({
      reviewerName: 'Amiya',
      reviewedBy: 77,
      status: 'delivered',
      reviewedAt: '2026-04-08T10:00:00Z',
      systemLabel: 'System'
    }),
    'Amiya'
  )
})

test('formatWelfareHistoryReviewerName returns system label for auto-delivered claims', () => {
  assert.equal(
    formatWelfareHistoryReviewerName({
      reviewerName: '',
      reviewedBy: 0,
      status: 'delivered',
      reviewedAt: '2026-04-08T10:00:00Z',
      systemLabel: 'System'
    }),
    'System'
  )
})

test('formatWelfareHistoryReviewerName keeps imported history without processed timestamp blank', () => {
  assert.equal(
    formatWelfareHistoryReviewerName({
      reviewerName: '',
      reviewedBy: 0,
      status: 'delivered',
      reviewedAt: null,
      systemLabel: 'System'
    }),
    '-'
  )
})
