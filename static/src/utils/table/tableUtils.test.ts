import assert from 'node:assert/strict'
import test from 'node:test'
import { updatePaginationFromResponse } from './tableUtils'

test('updatePaginationFromResponse syncs server-provided page size before clamping', () => {
  const pagination: Api.Common.PaginationParams = {
    current: 2,
    size: 260,
    total: 0
  }

  updatePaginationFromResponse(pagination, {
    list: [],
    total: 20,
    current: 2,
    size: 20
  })

  assert.equal(pagination.size, 20)
  assert.equal(pagination.current, 1)
  assert.equal(pagination.total, 20)
})
